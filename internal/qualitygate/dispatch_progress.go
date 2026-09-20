package qualitygate

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"time"
)

// BuilderProgress is shared by planned S6 gates, the board and dispatch checks.
// Only the TASK's current report and its owner's assignment can prove completion.
type BuilderProgress struct {
	State  string
	Reason string
}

func HasDispatchPlan(state map[string]any) bool {
	for _, d := range currentDocuments(state, nestedInt(state, "baseline", "generation")) {
		if d.Kind == "dispatch_plan" {
			return true
		}
	}
	return false
}
func PlannedBuilderProgress(input Input) map[string]BuilderProgress {
	result := map[string]BuilderProgress{}
	entities, _ := input.Snapshot.State["entities"].(map[string]any)
	tasks, _ := entities["tasks"].([]any)
	agents, _ := entities["agents"].([]any)
	generation := nestedInt(input.Snapshot.State, "baseline", "generation")
	runtimeID := stringValue(input.Snapshot.State["runtime_id"])
	evidence, _ := input.Snapshot.State["evidence"].([]any)
	for _, raw := range tasks {
		task, _ := raw.(map[string]any)
		id := stringValue(task["id"])
		state := stringValue(task["state"])
		if state == "candidate" || state == "reviewed" || state == "locked" {
			continue
		}
		progress := BuilderProgress{State: "running", Reason: "registered owner"}
		if state == "blocked" || state == "cancelled" {
			result[id] = BuilderProgress{State: "blocked", Reason: "TASK " + state}
			continue
		}
		report := stringValue(task["completion_report_ref"])
		if report == "" {
			result[id] = progress
			continue
		}
		progress = BuilderProgress{State: "reported", Reason: "current result/integration not verified"}
		var env evidenceEnvelope
		reportSHA := ""
		valid := false
		for _, entry := range evidence {
			e, _ := entry.(map[string]any)
			if e["path"] != report || e["kind"] != "completion_report" || e["status"] != "valid" || e["invalidated_by"] != nil || intValue(e["baseline_generation"]) != generation {
				continue
			}
			b, err := input.Files.ReadFile(report)
			if err != nil || sha256Hex(b) != e["sha256"] || json.Unmarshal(b, &env) != nil {
				continue
			}
			reportSHA = sha256Hex(b)
			valid = env.TaskID == id && env.RuntimeID == runtimeID && env.BaselineGeneration == generation && env.EvidenceID == e["id"] && env.InvalidatedBy == "" && env.Conclusion == "completed" && len(env.Checks) > 0 && len(failingEnvelopeChecks(env)) == 0 && len(env.ScopeDeviations) == 0
		}
		for _, doc := range currentDocuments(input.Snapshot.State, generation) {
			if doc.Kind == "task" && doc.ID == id {
				valid = valid && exactSubjects(env.SubjectRefs, []documentFact{doc})
			}
		}
		if !valid {
			result[id] = progress
			continue
		}
		owner := false
		owners, _ := task["owner_agent_ids"].([]any)
		for _, o := range owners {
			if o == env.ProducerAgentID {
				owner = true
			}
		}
		if !owner {
			result[id] = progress
			continue
		}
		assignment := ""
		for _, rawAgent := range agents {
			a, _ := rawAgent.(map[string]any)
			if a["id"] == env.ProducerAgentID {
				if a["state"] != "reported" && a["state"] != "done" && a["state"] != "stopped" {
					break
				}
				_, assignment, _ = strings.Cut(stringValue(a["prompt_ref"]), "#")
			}
		}
		if assignment == "" || path.Base(assignment) != assignment {
			result[id] = progress
			continue
		}
		b, err := input.Files.ReadFile(path.Join(".claude/evidence", runtimeID, fmt.Sprintf("g%d", generation), "worktree", assignment, "checkpoint.json"))
		var cp struct {
			TaskID       string `json:"task_id"`
			AssignmentID string `json:"assignment_id"`
			Generation   int    `json:"baseline_generation"`
			State        string `json:"state"`
			Updated      string `json:"verified_at"`
			ReportPath   string `json:"completion_report_path"`
			ReportSHA    string `json:"completion_report_sha256"`
		}
		if err == nil && json.Unmarshal(b, &cp) == nil && cp.TaskID == id && cp.AssignmentID == assignment && cp.Generation == generation && cp.ReportPath == report && cp.ReportSHA != "" && cp.ReportSHA == reportSHA {
			// A report submitted after this checkpoint needs fresh integration checks.
			var body struct {
				Created string `json:"created_at"`
			}
			rb, _ := input.Files.ReadFile(report)
			_ = json.Unmarshal(rb, &body)
			created, ce := time.Parse(time.RFC3339Nano, body.Created)
			updated, ue := time.Parse(time.RFC3339Nano, cp.Updated)
			if ce == nil && ue == nil && !updated.Before(created) {
				switch cp.State {
				case "verified", "acknowledged", "cleanup_pending", "complete":
					progress = BuilderProgress{State: "integrated", Reason: "current result and assignment integration verified"}
				}
			}
		}
		result[id] = progress
	}
	return result
}
