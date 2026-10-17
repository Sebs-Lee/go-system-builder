package runtime

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/entroforge/go-system-builder/internal/fileview"
	"github.com/entroforge/go-system-builder/internal/workpacket"
)

func ValidateIssuedTaskPlanningPacket(root string, state map[string]any, p workpacket.TaskPlanningPacket, files fileview.Reader) error {
	return ValidateTaskPlanningPacket(root, state, p, files)
}

func prepareTaskPlanningSubmission(root string, state map[string]any, req EvidenceRequest, data []byte) (SubmissionRecord, error) {
	s := req.Submission
	p, found, err := FindTaskPlanningPacket(state, s.PacketID)
	if err != nil || !found {
		return SubmissionRecord{}, fmt.Errorf("S4 submission has no original issuance")
	}
	if req.Kind != "planning_task" || s.Owner != "Main" || len(req.ProducedBy) != 1 || req.ProducedBy[0] != "Main" || req.ResponsibilityID != "Orchestrator" || req.ReviewRound != nil {
		return SubmissionRecord{}, fmt.Errorf("S4 producer/kind differs from actual Main responsibility")
	}
	answer, issues := workpacket.ValidateTaskPlanningAnswer(p, s.AnswerJSON)
	if len(issues) > 0 {
		return SubmissionRecord{}, fmt.Errorf("contract answer %s: %s", issues[0].Pointer, issues[0].Expected)
	}
	digest, err := workpacket.TaskPlanningAnswerDigest(answer)
	if err != nil {
		return SubmissionRecord{}, err
	}
	if s.ContentHash != "" && s.ContentHash != digest {
		return SubmissionRecord{}, fmt.Errorf("S4 content digest differs")
	}
	prior, accepted, err := FindSubmission(state, p.ID)
	if err != nil {
		return SubmissionRecord{}, err
	}
	if accepted {
		if prior.ContentHash != digest {
			return prior, ErrSubmissionConflict
		}
		if prior.Owner != "Main" || prior.EvidenceID != req.ID {
			return prior, fmt.Errorf("S4 replay original domain differs")
		}
		if err := ValidateTaskPlanningEnvelope(p, req.ID, answer, data); err != nil {
			return prior, err
		}
		rows, _ := state["evidence"].([]any)
		for _, raw := range rows {
			row, _ := raw.(map[string]any)
			if row["id"] == req.ID && row["kind"] == req.Kind && row["path"] == req.Path && row["sha256"] == workpacket.Digest(data) {
				return prior, errSubmissionAccepted
			}
		}
		return prior, fmt.Errorf("S4 replay original evidence differs")
	}
	if err := ValidateIssuedTaskPlanningPacket(root, state, p, s.Files); err != nil {
		return SubmissionRecord{}, err
	}
	if s.Authorize == nil {
		return SubmissionRecord{}, fmt.Errorf("S4 requires trusted Main application authority")
	}
	if err := s.Authorize(root, state); err != nil {
		return SubmissionRecord{}, err
	}
	artifacts, err := TaskPlanningArtifacts(root, state, p, answer, s.Files)
	if err != nil {
		return SubmissionRecord{}, err
	}
	if err := ValidateTaskPlanningEnvelope(p, req.ID, answer, data); err != nil {
		return SubmissionRecord{}, err
	}
	var value map[string]any
	json.Unmarshal(data, &value)
	at, err := time.Parse(time.RFC3339Nano, stringValue(value["created_at"]))
	if err != nil {
		return SubmissionRecord{}, err
	}
	expected, err := workpacket.TaskPlanningEvidence(p, req.ID, artifacts, answer, at)
	if err != nil {
		return SubmissionRecord{}, err
	}
	expected["created_at"] = value["created_at"]
	if progressCanonicalJSON(value) != progressCanonicalJSON(expected) {
		return SubmissionRecord{}, fmt.Errorf("S4 planning envelope actual subjects differ")
	}
	key := workpacket.Digest([]byte("submission:" + p.ID))
	return SubmissionRecord{Key: key, PacketID: p.ID, Owner: "Main", ContentHash: digest, EvidenceID: req.ID, FollowupID: workpacket.Digest([]byte(key + ":after-domain")), ControlStatus: "pending", OriginState: "planning", OriginPhase: "tasks", OriginEntry: p.Binding.StageEntry}, nil
}

// ValidateTaskPlanningEnvelope verifies historical professional facts without
// reinterpreting them against a later project cursor or changing authoring file.
func ValidateTaskPlanningEnvelope(p workpacket.TaskPlanningPacket, id string, answer workpacket.TaskPlanningAnswer, data []byte) error {
	if err := workpacket.ValidateUniqueJSON(data); err != nil {
		return err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for _, key := range []string{"schema_version", "evidence_id", "kind", "runtime_id", "baseline_generation", "producer_agent_id", "producer_responsibility", "subject_refs", "conclusion", "created_at", "task_planning_answer"} {
		if _, ok := raw[key]; !ok {
			return fmt.Errorf("S4 planning envelope missing %s", key)
		}
	}
	if len(raw) != 11 {
		return fmt.Errorf("S4 planning envelope has unsupported fields")
	}
	var historical workpacket.TaskPlanningAnswer
	if err := json.Unmarshal(raw["task_planning_answer"], &historical); err != nil {
		return err
	}
	draft := p.Draft
	draft.Answer = historical
	encoded, _ := json.Marshal(draft)
	historical, issues := workpacket.ValidateTaskPlanningAnswer(p, encoded)
	if len(issues) > 0 {
		return fmt.Errorf("S4 historical answer invalid")
	}
	a, err := workpacket.TaskPlanningAnswerDigest(answer)
	if err != nil {
		return err
	}
	h, err := workpacket.TaskPlanningAnswerDigest(historical)
	if err != nil || a != h {
		return fmt.Errorf("S4 evidence differs from professional answer")
	}
	var subjects []workpacket.Subject
	if err := json.Unmarshal(raw["subject_refs"], &subjects); err != nil {
		return err
	}
	byPath := map[string]workpacket.Subject{}
	for _, ref := range subjects {
		if _, duplicate := byPath[ref.Path]; duplicate {
			return fmt.Errorf("S4 evidence duplicate subjects")
		}
		byPath[ref.Path] = ref
	}
	paths := append([]string{*answer.IndexPath}, answer.TaskPaths...)
	artifacts := []workpacket.Reference{}
	for _, path := range paths {
		ref, ok := byPath[path]
		if !ok {
			return fmt.Errorf("S4 evidence missing delivered contract")
		}
		artifacts = append(artifacts, workpacket.Reference{Path: path, Version: ref.Version, SHA256: ref.SHA256})
	}
	var value map[string]any
	json.Unmarshal(data, &value)
	at, err := time.Parse(time.RFC3339Nano, stringValue(value["created_at"]))
	if err != nil {
		return err
	}
	expected, err := workpacket.TaskPlanningEvidence(p, id, artifacts, answer, at)
	if err != nil {
		return err
	}
	expected["created_at"] = value["created_at"]
	if progressCanonicalJSON(value) != progressCanonicalJSON(expected) {
		return fmt.Errorf("S4 envelope differs from frozen original subjects and professional answer")
	}
	return nil
}

func FindTaskPlanningPacket(state map[string]any, id string) (workpacket.TaskPlanningPacket, bool, error) {
	entities, _ := state["entities"].(map[string]any)
	rows, _ := entities["work_packets"].([]any)
	for _, raw := range rows {
		row, _ := raw.(map[string]any)
		if row["packet_id"] != id || row["template_id"] != "s4-task-planning-v1" {
			continue
		}
		data, err := json.Marshal(row)
		if err != nil {
			return workpacket.TaskPlanningPacket{}, false, err
		}
		var p workpacket.TaskPlanningPacket
		err = json.Unmarshal(data, &p)
		return p, true, err
	}
	return workpacket.TaskPlanningPacket{}, false, nil
}

func RecordTaskPlanningPacket(root, statePath, journalPath string, p workpacket.TaskPlanningPacket, files fileview.Reader, validator CandidateValidator) (Snapshot, error) {
	writer := NewWriter(statePath, journalPath, root, validator)
	snap, err := writer.Snapshot()
	if err != nil {
		return Snapshot{}, err
	}
	if err := ValidateIssuedTaskPlanningPacket(root, snap.State, p, files); err != nil {
		return Snapshot{}, err
	}
	if old, found, err := FindTaskPlanningPacket(snap.State, p.ID); err != nil {
		return Snapshot{}, err
	} else if found {
		if !reflect.DeepEqual(old, p) {
			return Snapshot{}, fmt.Errorf("contract packet identity conflict")
		}
		return snap, nil
	}
	data, _ := json.Marshal(p)
	var row map[string]any
	json.Unmarshal(data, &row)
	cursor := map[string]any{"state": "planning", "phase": "tasks"}
	next, err := writer.UpdateCurrent(Mutation{EventID: "evt-task-planning-packet-" + p.ID, TransitionID: "PACKET-ISSUE", Event: "work_packet_issued", Actor: "orchestrator", IdempotencyKey: "packet:" + p.ID, RuntimeID: p.Binding.RuntimeID, From: cursor, To: cursor, RetainLastTransition: true, OccurredAt: time.Now().UTC(), Apply: func(state map[string]any) error {
		if err := ValidateIssuedTaskPlanningPacket(root, state, p, files); err != nil {
			return err
		}
		if _, found, err := FindTaskPlanningPacket(state, p.ID); err != nil {
			return err
		} else if found {
			return errSubmissionAccepted
		}
		entities := state["entities"].(map[string]any)
		rows, _ := entities["work_packets"].([]any)
		for _, raw := range rows {
			old, _ := raw.(map[string]any)
			if old["template_id"] != "s4-task-planning-v1" {
				continue
			}
			binding, _ := old["binding"].(map[string]any)
			if stringValue(binding["runtime_id"]) == p.Binding.RuntimeID && numberBindingGeneration(binding) == p.Binding.Generation && stringValue(binding["owner"]) == "Main" {
				entry, err := integerField(binding, "stage_entry")
				if err != nil {
					return err
				}
				if entry == p.Binding.StageEntry {
					return fmt.Errorf("contract stage already has a different frozen issuance")
				}
			}
		}
		entities["work_packets"] = append(rows, row)
		return nil
	}})
	if err != nil {
		latest, readErr := writer.Snapshot()
		if readErr == nil {
			prior, found, readErr := FindTaskPlanningPacket(latest.State, p.ID)
			if readErr == nil && found && reflect.DeepEqual(prior, p) {
				return latest, nil
			}
		}
	}
	return next, err
}
