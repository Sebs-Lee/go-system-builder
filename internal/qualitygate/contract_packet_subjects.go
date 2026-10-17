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

type verifiedContractSubmission struct {
	packet   workpacket.ContractPacket
	answer   workpacket.ContractAnswer
	subjects map[string]workpacket.Subject
	files    fileview.Reader
}

// verifiedContractPacketFacts projects facts from the original accepted S3
// packet. It rechecks the frozen source closure and asks the original runtime
// artifact resolver for the current complete delivery set, including any
// shared-model dependencies checked by the original S3 consumer.
func verifiedContractPacketFacts(input Input) ([]documentFact, bool, error) {
	verified, accepted, err := readAcceptedContractSubmission(input)
	if err != nil || !accepted {
		return nil, accepted, err
	}

	artifacts, err := runtime.ContractArtifacts(
		input.Root,
		input.Snapshot.State,
		verified.packet,
		verified.answer,
		verified.files,
	)
	if err != nil {
		return nil, true, fmt.Errorf("revalidate original S3 delivered artifacts: %w", err)
	}
	if len(artifacts) == 0 {
		return nil, true, fmt.Errorf("original S3 delivery has no index or contract artifacts")
	}

	expected := map[string]workpacket.Subject{}
	addExpected := func(subject workpacket.Subject) error {
		if subject.Path == "" {
			return fmt.Errorf("original S3 subject has an empty path")
		}
		if _, duplicate := expected[subject.Path]; duplicate {
			return fmt.Errorf("original S3 subject set repeats %s", subject.Path)
		}
		expected[subject.Path] = subject
		return nil
	}
	if err := addExpected(workpacket.Subject{
		Path:    verified.packet.Binding.Requirement.Path,
		Version: verified.packet.Binding.Requirement.Version,
		SHA256:  verified.packet.Binding.Requirement.SHA256,
	}); err != nil {
		return nil, true, err
	}
	for _, design := range verified.packet.Inputs.Designs {
		if err := addExpected(workpacket.Subject{Path: design.Path, Version: design.Version, SHA256: design.SHA256}); err != nil {
			return nil, true, err
		}
	}
	for _, source := range verified.packet.Inputs.Sources {
		if err := addExpected(source); err != nil {
			return nil, true, err
		}
	}
	for _, artifact := range artifacts {
		if artifact.Path == "" || artifact.SHA256 == "" {
			return nil, true, fmt.Errorf("original S3 artifact has an empty path or digest")
		}
		if err := addExpected(workpacket.Subject{Path: artifact.Path, Version: artifact.Version, SHA256: artifact.SHA256}); err != nil {
			return nil, true, err
		}
	}
	if len(expected) != len(verified.subjects) {
		return nil, true, fmt.Errorf("original S3 envelope subject set has %d paths; the validated source and delivery set has %d", len(verified.subjects), len(expected))
	}
	for path, want := range expected {
		got, ok := verified.subjects[path]
		if !ok {
			return nil, true, fmt.Errorf("original S3 envelope is missing current subject %s", path)
		}
		if got != want {
			return nil, true, fmt.Errorf("original S3 envelope subject differs from current source or delivered bytes: %s", path)
		}
	}

	facts, err := contractPacketSourceFacts(verified.packet, verified.subjects)
	if err != nil {
		return nil, true, err
	}
	indexPath := ""
	if verified.answer.IndexPath != nil {
		indexPath = *verified.answer.IndexPath
	}
	contractPaths := make(map[string]bool, len(verified.answer.ContractPaths))
	for _, path := range verified.answer.ContractPaths {
		contractPaths[path] = true
	}
	for _, artifact := range artifacts {
		kind, status, id := "source", "frozen", "shared-model:"+artifact.Path
		switch {
		case artifact.Path == indexPath:
			kind, status, id = "contract_index", "locked", "contract-index:"+artifact.ID
		case contractPaths[artifact.Path]:
			kind, status, id = "contract", "locked", "contract:"+artifact.ID
		case strings.HasPrefix(filepath.ToSlash(artifact.Path), "docs/dev/contracts/"):
			return nil, true, fmt.Errorf("original S3 checker returned an undeclared contract output: %s", artifact.Path)
		}
		facts = append(facts, documentFact{
			ID:         id,
			Kind:       kind,
			Path:       artifact.Path,
			Version:    artifact.Version,
			SHA256:     artifact.SHA256,
			Status:     status,
			Generation: verified.packet.Binding.Generation,
		})
	}
	if verifiedFileView, ok := input.Files.(interface{ Verify() error }); ok {
		if err := verifiedFileView.Verify(); err != nil {
			return nil, true, fmt.Errorf("original S3 file view moved during artifact projection: %w", err)
		}
	}
	return facts, true, nil
}

func readAcceptedContractSubmission(input Input) (*verifiedContractSubmission, bool, error) {
	state := input.Snapshot.State
	if !runtime.PacketEnabled(state) {
		return nil, false, nil
	}
	issuanceRows := currentContractIssuanceRows(state)
	if len(issuanceRows) == 0 {
		return nil, false, nil
	}
	acceptedIssuance := false
	for _, row := range issuanceRows {
		packetID := stringValue(row["packet_id"])
		if packetID != "" && countContractSubmissionRows(state, packetID) > 0 {
			acceptedIssuance = true
		}
	}
	if !acceptedIssuance {
		// A prepared packet without a received answer is the ordinary S3
		// questionnaire state. Let the existing disk fallback report missing
		// contract evidence until Main saves it.
		return nil, false, nil
	}
	if input.Root == "" || state == nil || input.Files == nil {
		return nil, true, fmt.Errorf("accepted S3 source projection requires the original root, runtime snapshot, and file view")
	}
	files, ok := input.Files.(fileview.Reader)
	if !ok {
		return nil, true, fmt.Errorf("accepted S3 source projection requires the original file view with directory reads")
	}

	binding, _, err := runtime.CurrentContractBinding(state)
	if err != nil {
		return nil, true, fmt.Errorf("resolve current S3 source binding: %w", err)
	}
	var packet workpacket.ContractPacket
	packetCount := 0
	for _, row := range issuanceRows {
		id := stringValue(row["packet_id"])
		if id == "" {
			return nil, true, fmt.Errorf("current S3 issuance row has no packet identity")
		}
		candidate, found, findErr := runtime.FindContractPacket(state, id)
		if findErr != nil {
			return nil, true, fmt.Errorf("decode original S3 issuance %s: %w", id, findErr)
		}
		if !found {
			return nil, true, fmt.Errorf("original S3 issuance %s is unavailable", id)
		}
		if candidate.Binding.RuntimeID == binding.RuntimeID &&
			candidate.Binding.Generation == binding.Generation &&
			candidate.Binding.StageEntry == binding.StageEntry &&
			candidate.Binding.Stage == "S3" && candidate.Binding.Owner == "Main" {
			packet = candidate
			packetCount++
		}
	}
	if packetCount != 1 {
		return nil, true, fmt.Errorf("expected exactly one original S3 issuance for the current binding; found %d", packetCount)
	}

	if countContractSubmissionRows(state, packet.ID) != 1 {
		return nil, true, fmt.Errorf("expected exactly one accepted S3 submission row")
	}
	submission, accepted, err := runtime.FindSubmission(state, packet.ID)
	if err != nil {
		return nil, true, fmt.Errorf("decode current S3 submission: %w", err)
	}
	if !accepted {
		return nil, false, nil
	}
	expectedKey := workpacket.Digest([]byte("submission:" + packet.ID))
	if submission.Key != expectedKey || submission.PacketID != packet.ID || submission.Owner != "Main" ||
		submission.EvidenceID == "" || submission.ContentHash == "" ||
		submission.FollowupID != workpacket.Digest([]byte(expectedKey+":after-domain")) ||
		submission.OriginState != "planning" || submission.OriginPhase != "contracts" ||
		submission.OriginEntry != packet.Binding.StageEntry || submission.RevisionPacketID != "" || submission.SupersededBy != "" {
		return nil, true, fmt.Errorf("current S3 submission is not the original Main planning.contracts reception")
	}

	evidence, found, err := contractEvidenceRow(state, submission.EvidenceID)
	if err != nil {
		return nil, true, fmt.Errorf("resolve original S3 evidence registration: %w", err)
	}
	if !found {
		return nil, true, fmt.Errorf("original S3 submission has no evidence registration")
	}
	if evidence["kind"] != "planning_contract" || evidence["status"] != "valid" || evidence["invalidated_by"] != nil ||
		stringValue(evidence["responsibility_id"]) != "Orchestrator" ||
		intValue(evidence["baseline_generation"]) != packet.Binding.Generation ||
		!singleProducer(evidence["produced_by"], "Main") {
		return nil, true, fmt.Errorf("original S3 evidence registration differs from the accepted Main planning_contract")
	}
	evidencePath := stringValue(evidence["path"])
	evidenceSHA := stringValue(evidence["sha256"])
	if !safeContractEvidencePath(evidencePath) || evidenceSHA == "" {
		return nil, true, fmt.Errorf("original S3 evidence registration has an invalid path or digest")
	}
	envelopeBytes, err := input.Files.ReadFile(evidencePath)
	if err != nil {
		return nil, true, fmt.Errorf("read original S3 planning_contract envelope %s: %w", evidencePath, err)
	}
	if sha256Hex(envelopeBytes) != evidenceSHA {
		return nil, true, fmt.Errorf("original S3 planning_contract envelope differs from its registered hash: %s", evidencePath)
	}
	var envelope struct {
		EvidenceID             string                    `json:"evidence_id"`
		Kind                   string                    `json:"kind"`
		RuntimeID              string                    `json:"runtime_id"`
		BaselineGeneration     int                       `json:"baseline_generation"`
		ProducerAgentID        string                    `json:"producer_agent_id"`
		ProducerResponsibility string                    `json:"producer_responsibility"`
		SubjectRefs            []workpacket.Subject      `json:"subject_refs"`
		Conclusion             string                    `json:"conclusion"`
		ContractAnswer         workpacket.ContractAnswer `json:"contract_answer"`
	}
	if err := json.Unmarshal(envelopeBytes, &envelope); err != nil {
		return nil, true, fmt.Errorf("decode original S3 planning_contract envelope %s: %w", evidencePath, err)
	}
	if envelope.EvidenceID != submission.EvidenceID || envelope.Kind != "planning_contract" ||
		envelope.RuntimeID != packet.Binding.RuntimeID || envelope.BaselineGeneration != packet.Binding.Generation ||
		envelope.ProducerAgentID != "Main" || envelope.ProducerResponsibility != "Orchestrator" || envelope.Conclusion != "pass" {
		return nil, true, fmt.Errorf("original S3 envelope header differs from the accepted Main binding: %s", evidencePath)
	}
	answerDigest, err := workpacket.ContractAnswerDigest(envelope.ContractAnswer)
	if err != nil || answerDigest != submission.ContentHash {
		return nil, true, fmt.Errorf("original S3 professional answer differs from the accepted submission hash: %s", evidencePath)
	}
	if err := runtime.ValidateContractEnvelope(packet, submission.EvidenceID, envelope.ContractAnswer, envelopeBytes); err != nil {
		return nil, true, fmt.Errorf("validate original S3 planning_contract envelope %s: %w", evidencePath, err)
	}
	if err := runtime.ValidateIssuedContractPacket(input.Root, state, packet, files); err != nil {
		return nil, true, fmt.Errorf("revalidate complete original S3 source closure: %w", err)
	}

	subjects := make(map[string]workpacket.Subject, len(envelope.SubjectRefs))
	for _, subject := range envelope.SubjectRefs {
		if subject.Path == "" {
			return nil, true, fmt.Errorf("original S3 envelope contains an empty subject path")
		}
		if _, duplicate := subjects[subject.Path]; duplicate {
			return nil, true, fmt.Errorf("original S3 envelope repeats subject %s", subject.Path)
		}
		subjects[subject.Path] = subject
	}
	if verifiedFileView, ok := input.Files.(interface{ Verify() error }); ok {
		if err := verifiedFileView.Verify(); err != nil {
			return nil, true, fmt.Errorf("original S3 file view moved during source projection: %w", err)
		}
	}
	return &verifiedContractSubmission{
		packet: packet, answer: envelope.ContractAnswer, subjects: subjects, files: files,
	}, true, nil
}

func contractPacketSourceFacts(packet workpacket.ContractPacket, subjects map[string]workpacket.Subject) ([]documentFact, error) {
	// ValidateContractEnvelope proves the frozen requirement, design, and
	// source refs. ContractInputs.Sources intentionally carry path and SHA only;
	// preserve the exact envelope version instead of deriving one from mutable
	// file text.
	facts := make([]documentFact, 0, len(packet.Inputs.Sources))
	for _, source := range packet.Inputs.Sources {
		subject, ok := subjects[source.Path]
		if !ok || subject.SHA256 != source.SHA256 {
			return nil, fmt.Errorf("original S3 envelope does not bind frozen source %s", source.Path)
		}
		facts = append(facts, documentFact{
			ID:         "contract-source:" + source.Path,
			Kind:       "source",
			Path:       source.Path,
			Version:    subject.Version,
			SHA256:     subject.SHA256,
			Status:     "frozen",
			Generation: packet.Binding.Generation,
		})
	}
	return facts, nil
}

func currentContractIssuanceRows(state map[string]any) []map[string]any {
	if state == nil {
		return nil
	}
	lifecycle, _ := state["lifecycle"].(map[string]any)
	runtimeID := stringValue(state["runtime_id"])
	generation := nestedInt(state, "baseline", "generation")
	stageEntry := intValue(lifecycle["phase_revision"])
	entities, _ := state["entities"].(map[string]any)
	rows, _ := entities["work_packets"].([]any)
	current := make([]map[string]any, 0, 1)
	for _, raw := range rows {
		row, _ := raw.(map[string]any)
		if row == nil || row["template_id"] != "s3-contract-v1" {
			continue
		}
		binding, _ := row["binding"].(map[string]any)
		if stringValue(binding["runtime_id"]) == runtimeID && intValue(binding["generation"]) == generation &&
			intValue(binding["stage_entry"]) == stageEntry && stringValue(binding["stage"]) == "S3" && stringValue(binding["owner"]) == "Main" {
			current = append(current, row)
		}
	}
	return current
}

func contractEvidenceRow(state map[string]any, id string) (map[string]any, bool, error) {
	raw, _ := state["evidence"].([]any)
	var result map[string]any
	for _, item := range raw {
		row, _ := item.(map[string]any)
		if row == nil || row["id"] != id {
			continue
		}
		if result != nil {
			return nil, false, fmt.Errorf("duplicate evidence identity %s", id)
		}
		result = row
	}
	return result, result != nil, nil
}

func countContractSubmissionRows(state map[string]any, packetID string) int {
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

func singleProducer(value any, want string) bool {
	producers, ok := value.([]any)
	return ok && len(producers) == 1 && producers[0] == want
}

func safeContractEvidencePath(value string) bool {
	clean := filepath.Clean(value)
	if value == "" || filepath.IsAbs(value) || filepath.ToSlash(clean) != value || clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return false
	}
	return len(value) < 4096 && !strings.ContainsAny(value, "\\\x00")
}
