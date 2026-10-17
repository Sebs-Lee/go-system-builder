package qualitygate

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/entroforge/go-system-builder/internal/fileview"
	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/workpacket"
)

// verifiedTaskPlanningPacketFacts exposes only the source and output facts
// bound by the originally accepted S4 planning_task envelope. The accepted
// bit distinguishes an unfilled questionnaire from a broken accepted receipt.
// This is a read-only bridge for the existing quality-gate subject matcher.
func verifiedTaskPlanningPacketFacts(input Input) ([]documentFact, bool, error) {
	state := input.Snapshot.State
	if input.Root == "" || state == nil || input.Files == nil {
		return nil, false, fmt.Errorf("S4 source projection requires the original root, runtime snapshot, and file view")
	}
	files, ok := input.Files.(fileview.Reader)
	if !ok {
		return nil, false, fmt.Errorf("S4 source projection requires the original file view with directory reads")
	}

	binding, err := runtime.CurrentTaskPlanningBinding(state)
	if err != nil {
		return nil, false, fmt.Errorf("resolve current S4 source binding: %w", err)
	}
	entities, _ := state["entities"].(map[string]any)
	rows, _ := entities["work_packets"].([]any)
	var packet workpacket.TaskPlanningPacket
	packetCount := 0
	for _, raw := range rows {
		row, _ := raw.(map[string]any)
		if row == nil || row["template_id"] != "s4-task-planning-v1" {
			continue
		}
		id, _ := row["packet_id"].(string)
		if id == "" {
			return nil, false, fmt.Errorf("S4 issuance row has no packet identity")
		}
		candidate, found, findErr := runtime.FindTaskPlanningPacket(state, id)
		if findErr != nil {
			return nil, false, fmt.Errorf("decode original S4 issuance %s: %w", id, findErr)
		}
		if !found {
			return nil, false, fmt.Errorf("original S4 issuance %s is unavailable", id)
		}
		if candidate.Binding.RuntimeID == binding.RuntimeID &&
			candidate.Binding.Generation == binding.Generation &&
			candidate.Binding.StageEntry == binding.StageEntry &&
			candidate.Binding.Stage == "S4" && candidate.Binding.Owner == "Main" {
			packet = candidate
			packetCount++
		}
	}
	if packetCount == 0 {
		submissions, _ := entities["submissions"].([]any)
		for _, raw := range submissions {
			row, _ := raw.(map[string]any)
			if row != nil && row["origin_state"] == "planning" && row["origin_phase"] == "tasks" && intValue(row["origin_entry"]) == binding.StageEntry {
				return nil, false, fmt.Errorf("current S4 submission references a missing original planning_task issuance")
			}
		}
		return nil, false, nil
	}
	if packetCount != 1 {
		return nil, false, fmt.Errorf("expected exactly one original S4 issuance for the current binding; found %d", packetCount)
	}

	submissionCount := countTaskPlanningSubmissionRows(state, packet.ID)
	if submissionCount == 0 {
		return nil, false, nil
	}
	if submissionCount != 1 {
		return nil, false, fmt.Errorf("expected exactly one accepted S4 submission row; found %d", submissionCount)
	}
	submission, foundSubmission, err := runtime.FindSubmission(state, packet.ID)
	if err != nil {
		return nil, false, fmt.Errorf("decode current S4 submission: %w", err)
	}
	if !foundSubmission {
		return nil, false, nil
	}
	fail := func(err error) ([]documentFact, bool, error) { return nil, true, err }
	expectedKey := workpacket.Digest([]byte("submission:" + packet.ID))
	if submission.Key != expectedKey || submission.PacketID != packet.ID || submission.Owner != "Main" ||
		submission.EvidenceID == "" || submission.ContentHash == "" ||
		submission.FollowupID != workpacket.Digest([]byte(expectedKey+":after-domain")) ||
		(submission.ControlStatus != "pending" && submission.ControlStatus != "no_transition") || submission.OriginState != "planning" || submission.OriginPhase != "tasks" ||
		submission.OriginEntry != packet.Binding.StageEntry || submission.TransitionEventID != "" ||
		submission.RevisionPacketID != "" || submission.SupersededBy != "" {
		return fail(fmt.Errorf("current S4 submission is not the original Main planning.tasks reception"))
	}

	evidence, found, err := contractEvidenceRow(state, submission.EvidenceID)
	if err != nil {
		return fail(fmt.Errorf("resolve original S4 evidence registration: %w", err))
	}
	if !found {
		return fail(fmt.Errorf("original S4 submission has no evidence registration"))
	}
	if evidence["kind"] != "planning_task" || evidence["status"] != "valid" || evidence["invalidated_by"] != nil ||
		stringValue(evidence["responsibility_id"]) != "Orchestrator" ||
		intValue(evidence["baseline_generation"]) != packet.Binding.Generation ||
		!singleProducer(evidence["produced_by"], "Main") {
		return fail(fmt.Errorf("original S4 evidence registration differs from the accepted Main planning_task"))
	}
	evidencePath := stringValue(evidence["path"])
	evidenceSHA := stringValue(evidence["sha256"])
	if !safeTaskPlanningEvidencePath(evidencePath) || evidenceSHA == "" {
		return fail(fmt.Errorf("original S4 evidence registration has an invalid path or digest"))
	}
	envelopeBytes, err := input.Files.ReadFile(evidencePath)
	if err != nil {
		return fail(fmt.Errorf("read original S4 planning_task envelope: %w", err))
	}
	if sha256Hex(envelopeBytes) != evidenceSHA {
		return fail(fmt.Errorf("original S4 planning_task envelope differs from its registered hash"))
	}
	var envelope struct {
		SchemaVersion          string                        `json:"schema_version"`
		EvidenceID             string                        `json:"evidence_id"`
		Kind                   string                        `json:"kind"`
		RuntimeID              string                        `json:"runtime_id"`
		BaselineGeneration     int                           `json:"baseline_generation"`
		ProducerAgentID        string                        `json:"producer_agent_id"`
		ProducerResponsibility string                        `json:"producer_responsibility"`
		SubjectRefs            []workpacket.Subject          `json:"subject_refs"`
		Conclusion             string                        `json:"conclusion"`
		TaskPlanningAnswer     workpacket.TaskPlanningAnswer `json:"task_planning_answer"`
	}
	if err := json.Unmarshal(envelopeBytes, &envelope); err != nil {
		return fail(fmt.Errorf("decode original S4 planning_task envelope: %w", err))
	}
	if envelope.SchemaVersion != "1.0.0" || envelope.EvidenceID != submission.EvidenceID || envelope.Kind != "planning_task" ||
		envelope.RuntimeID != packet.Binding.RuntimeID || envelope.BaselineGeneration != packet.Binding.Generation ||
		envelope.ProducerAgentID != "Main" || envelope.ProducerResponsibility != "Orchestrator" || envelope.Conclusion != "pass" {
		return fail(fmt.Errorf("original S4 envelope header differs from the accepted Main binding"))
	}
	answerDigest, err := workpacket.TaskPlanningAnswerDigest(envelope.TaskPlanningAnswer)
	if err != nil || answerDigest != submission.ContentHash {
		return fail(fmt.Errorf("original S4 professional answer differs from the accepted submission hash"))
	}
	if err := runtime.ValidateTaskPlanningEnvelope(packet, submission.EvidenceID, envelope.TaskPlanningAnswer, envelopeBytes); err != nil {
		return fail(fmt.Errorf("validate original S4 planning_task envelope: %w", err))
	}
	if err := runtime.ValidateIssuedTaskPlanningPacket(input.Root, state, packet, files); err != nil {
		return fail(fmt.Errorf("revalidate complete original S4 source closure: %w", err))
	}
	artifacts, err := runtime.TaskPlanningArtifacts(input.Root, state, packet, envelope.TaskPlanningAnswer, files)
	if err != nil {
		return fail(fmt.Errorf("revalidate original S4 TASK batch and dispatch index: %w", err))
	}

	// Envelope validation proves the closed set of source and output paths.
	// Also compare output refs against the actual refs returned by the runtime
	// resolver so a rewritten envelope/evidence hash cannot bless changed TASKs.
	subjects := make(map[string]workpacket.Subject, len(envelope.SubjectRefs))
	for _, subject := range envelope.SubjectRefs {
		if subject.Path == "" {
			return fail(fmt.Errorf("original S4 envelope contains an empty subject path"))
		}
		if _, duplicate := subjects[subject.Path]; duplicate {
			return fail(fmt.Errorf("original S4 envelope repeats subject %s", subject.Path))
		}
		subjects[subject.Path] = subject
	}
	registeredPaths := map[string]bool{packet.Binding.Requirement.Path: true}
	for _, contract := range packet.Inputs.Contracts {
		registeredPaths[contract.Path] = true
	}
	facts := make([]documentFact, 0, len(packet.Inputs.Sources)+len(artifacts))
	for _, source := range packet.Inputs.Sources {
		subject, ok := subjects[source.Path]
		if !ok || subject.SHA256 != source.SHA256 {
			return fail(fmt.Errorf("original S4 envelope does not bind frozen source %s", source.Path))
		}
		if registeredPaths[source.Path] {
			continue
		}
		facts = append(facts, documentFact{
			ID:         "task-planning-source:" + source.Path,
			Kind:       "source",
			Path:       source.Path,
			Version:    subject.Version,
			SHA256:     subject.SHA256,
			Status:     "frozen",
			Generation: packet.Binding.Generation,
		})
	}

	artifactPaths := map[string]bool{*envelope.TaskPlanningAnswer.IndexPath: true}
	for _, path := range envelope.TaskPlanningAnswer.TaskPaths {
		artifactPaths[path] = true
	}
	if len(artifacts) != len(artifactPaths) {
		return fail(fmt.Errorf("original S4 envelope artifact set differs from the validated TASK batch"))
	}
	for _, artifact := range artifacts {
		subject, ok := subjects[artifact.Path]
		if !ok || subject.Version != artifact.Version || subject.SHA256 != artifact.SHA256 || !artifactPaths[artifact.Path] {
			return fail(fmt.Errorf("original S4 envelope artifact differs from the actual validated output %s", artifact.Path))
		}
		if artifact.Path == *envelope.TaskPlanningAnswer.IndexPath {
			facts = append(facts, documentFact{
				ID: "task-planning-index:" + artifact.ID, Kind: "task_index", Path: artifact.Path,
				Version: artifact.Version, SHA256: artifact.SHA256, Status: "complete", Generation: packet.Binding.Generation,
			})
			continue
		}
		data, err := files.ReadFile(artifact.Path)
		if err != nil || sha256Hex(data) != artifact.SHA256 {
			return fail(fmt.Errorf("read original S4 TASK output through source view: %s", artifact.Path))
		}
		status := parseMarkdownStatusField(string(data))
		if status != "complete" && status != "cancelled" {
			return fail(fmt.Errorf("original S4 TASK output has unsupported status %q: %s", status, artifact.Path))
		}
		facts = append(facts, documentFact{
			ID: "task-planning-output:" + artifact.ID, Kind: "task", Path: artifact.Path,
			Version: artifact.Version, SHA256: artifact.SHA256, Status: status, Generation: packet.Binding.Generation,
		})
	}
	if verified, ok := input.Files.(interface{ Verify() error }); ok {
		if err := verified.Verify(); err != nil {
			return fail(fmt.Errorf("original S4 file view moved during source projection: %w", err))
		}
	}
	return facts, true, nil
}

func countTaskPlanningSubmissionRows(state map[string]any, packetID string) int {
	entities, _ := state["entities"].(map[string]any)
	rows, _ := entities["submissions"].([]any)
	count := 0
	for _, item := range rows {
		row, _ := item.(map[string]any)
		if row != nil && row["packet_id"] == packetID {
			count++
		}
	}
	return count
}

func safeTaskPlanningEvidencePath(value string) bool {
	clean := filepath.Clean(value)
	return value != "" && !filepath.IsAbs(value) && filepath.ToSlash(clean) == value && clean != "." && clean != ".." &&
		!strings.HasPrefix(clean, ".."+string(filepath.Separator)) && len(value) < 4096 && !strings.ContainsAny(value, "\\\x00")
}
