package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/entroforge/go-system-builder/internal/policy"
	"github.com/entroforge/go-system-builder/internal/repair"
	"github.com/entroforge/go-system-builder/internal/review"
	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/submission"
	"github.com/entroforge/go-system-builder/internal/workpacket"
	"github.com/entroforge/go-system-builder/internal/workspace"
	"github.com/entroforge/go-system-builder/internal/workstate"
)

func receiveSavedQuestionnaire(root string, request policy.Input, stdout, stderr io.Writer) (bool, int) {
	for _, receive := range []func(string, policy.Input, io.Writer, io.Writer) (bool, int){
		runSavedRepairLeadQuestionnaire,
		runSavedReviewPlanQuestionnaire, runSavedAuditQuestionnaire,
		runSavedBuilderQuestionnaire, runSavedStageAnswerHook,
	} {
		if handled, code := receive(root, request, stdout, stderr); handled {
			return true, code
		}
	}
	return false, 0
}

type observedQuestionnaire struct {
	path     string
	draft    any
	s6Result bool
	s7Plan   bool
}

// A successful script or MCP tool can save a questionnaire without exposing a
// file_path. Read the graph's existing exact files, never parse command text or
// invent a tool policy. Unchanged issued drafts do not invoke a domain consumer.
func runObservedQuestionnaireSaves(root string, request policy.Input, stdout, stderr io.Writer) (bool, int) {
	switch request.ToolName {
	case "Read", "Glob", "Grep":
		return false, 0
	}
	if request.Event != "PostToolUse" || request.Error != "" {
		return false, 0
	}
	if failed, _ := request.ToolResponse["is_error"].(bool); failed {
		return false, 0
	}
	if _, explicit := savedQuestionnairePath(request); explicit {
		return false, 0
	}
	snapshot, err := runtime.NewStore(filepath.Join(root, ".claude/loop-state.json"), filepath.Join(root, ".claude/loop-events.jsonl")).Snapshot()
	if err != nil {
		return false, 0
	}
	var contexts []string
	seen := map[string]bool{}
	for _, q := range observableQuestionnaires(root, snapshot) {
		if q.path == "" || seen[q.path] {
			continue
		}
		seen[q.path] = true
		info, err := os.Stat(q.path)
		if err != nil || !info.Mode().IsRegular() || info.Size() > 20<<20 {
			continue
		}
		data, err := os.ReadFile(q.path)
		if err != nil {
			continue
		}
		var answer any
		if json.Unmarshal(data, &answer) == nil {
			if q.s6Result && !s6ObservedResultHasProfessionalContent(data) {
				continue
			}
			if q.s7Plan && !s7ObservedPlanHasProfessionalContent(data) {
				continue
			}
			if q.draft != nil && reflect.DeepEqual(answer, q.draft) {
				continue
			}
			if q.draft == nil {
				fields, _ := answer.(map[string]any)
				if _, audit := fields["manifest_type"]; audit && stringValue(fields["outcome"]) == "" {
					continue
				}
				if _, targeted := fields["performing_assignment_id"]; targeted && stringValue(fields["result"]) == "" {
					continue
				}
				changeTypes, _ := fields["change_types"].([]any)
				decisions, _ := fields["decisions"].([]any)
				if _, impact := fields["changed_artifacts"]; impact && len(changeTypes) == 0 && len(decisions) == 0 {
					continue
				}
			}
		}
		// Normalize an observed file save only for the existing adapters. The
		// actual host event and audit log keep their original tool identity.
		observed := request
		observed.ToolName = "Write"
		observed.ToolInput = map[string]any{"file_path": q.path}
		var out bytes.Buffer
		if handled, _ := receiveSavedQuestionnaire(root, observed, &out, stderr); !handled {
			continue
		}
		var envelope struct {
			Output struct {
				Context string `json:"additionalContext"`
			} `json:"hookSpecificOutput"`
		}
		if json.Unmarshal(out.Bytes(), &envelope) == nil && envelope.Output.Context != "" {
			contexts = append(contexts, envelope.Output.Context)
		}
	}
	if len(contexts) == 0 {
		return false, 0
	}
	_ = json.NewEncoder(stdout).Encode(map[string]any{"hookSpecificOutput": map[string]any{"hookEventName": "PostToolUse", "additionalContext": strings.Join(contexts, "\n")}})
	return true, 0
}

// DraftPlanForRoot always leaves the QA/delivery oracles for Main to answer.
// Inspect those existing professional fields instead of rebuilding the product
// baseline just to distinguish a machine draft from an observed script save.
// The original consumer still validates incomplete or invalid authored plans.
func s7ObservedPlanHasProfessionalContent(data []byte) bool {
	var plan review.Plan
	if json.Unmarshal(data, &plan) != nil {
		return true
	}
	placeholder := false
	for _, claim := range plan.Claims {
		oracle := strings.TrimSpace(claim.Oracle)
		if claim.Applicability == "required" && (claim.Lens == "qa" || claim.Lens == "delivery") && oracle != "" && !strings.Contains(oracle, "TODO(planner)") {
			return true
		}
		for _, field := range []string{claim.Target, claim.Assertion, claim.Oracle, claim.Method} {
			placeholder = placeholder || strings.Contains(field, "TODO(planner)")
		}
	}
	// A custom complete plan can use other lenses; leave its validity to the
	// original consumer rather than imposing another planning contract here.
	return !placeholder
}

func observableQuestionnaires(root string, snapshot runtime.Snapshot) []observedQuestionnaire {
	state := snapshot.State
	lifecycle, _ := state["lifecycle"].(map[string]any)
	stage, _, _ := projectNext(stringValue(lifecycle["state"]), stringValue(lifecycle["phase"]), root)
	if stage == "S0" || stage == "S1" || stage == "S11" {
		return nil
	}
	var out []observedQuestionnaire
	if stage == "S7" {
		facts, _ := state["review"].(map[string]any)
		if ptr := review.PlanPointerFromState(state); ptr == nil || ptr.ReviewRound != tolerantInt(facts["round"]) {
			out = append(out, observedQuestionnaire{path: reviewPlanQuestionnairePath(root, state), s7Plan: true})
		}
	}
	if stage == "S9" {
		for _, path := range repairLeadQuestionnairePaths(root, state) {
			out = append(out, observedQuestionnaire{path: path})
		}
	}
	if kind := currentAuditKind(state); kind != "" {
		if status := inspectS10Artifact(root, state, kind); status.ManifestPath != "" {
			out = append(out, observedQuestionnaire{path: filepath.Join(root, filepath.FromSlash(status.ManifestPath))})
		} else if path, err := auditQuestionnairePath(root, state, kind); err == nil {
			out = append(out, observedQuestionnaire{path: filepath.Join(root, filepath.FromSlash(path))})
		}
	}
	entities, _ := state["entities"].(map[string]any)
	rows, _ := entities["work_packets"].([]any)
	registry, _ := workspace.Decode(state)
	for _, raw := range rows {
		row, _ := raw.(map[string]any)
		template := stringValue(row["template_id"])
		if !strings.HasPrefix(template, strings.ToLower(stage)+"-") {
			continue
		}
		if stage == "S9" {
			switch stringValue(repairPointerCLI(state)["status"]) {
			case "impact_reconciliation", "targeted_reverification", "ready_for_full_review":
				// SubmitResult can leave lifecycle.fixing while the original
				// repair pointer has already moved to Main or verifier work.
				continue
			}
			phase := stringValue(lifecycle["phase"])
			if phase == "fixing" && template != "s9-repair-result-v1" {
				continue
			}
			if (phase == "planning" || phase == "reproducing") && template != "s9-plan-report-v1" {
				continue
			}
			if phase != "fixing" && phase != "planning" && phase != "reproducing" {
				continue
			}
		}
		path, source := stringValue(row["answer_path"]), root
		binding, _ := row["binding"].(map[string]any)
		if stage == "S2" || stage == "S3" || stage == "S4" {
			current, issued, verified, err := currentIssuedPlanningAnswerPath(state, stringValue(lifecycle["phase"]))
			if err != nil || !issued || !verified || path != current {
				continue
			}
		}
		if round := tolerantInt(binding["review_round"]); round > 0 {
			facts, _ := state["review"].(map[string]any)
			if round != tolerantInt(facts["round"]) {
				continue
			}
		}
		owner, assignment := stringValue(binding["owner_agent_id"]), stringValue(binding["assignment_id"])
		switch stringValue(row["template_id"]) {
		case "s9-repair-result-v1":
			data, _ := json.Marshal(row)
			var packet workpacket.RepairResultPacket
			if json.Unmarshal(data, &packet) != nil {
				continue
			}
			if path, data, err := repair.RepairResultResolutionDraft(root, state, packet.ID, packet.Binding.Sources.OwnerAgentID); err == nil {
				var draft any
				if json.Unmarshal(data, &draft) == nil {
					out = append(out, observedQuestionnaire{path: filepath.Join(root, filepath.FromSlash(path)), draft: draft})
				}
			}
			var err error
			source, _, err = submission.RepairResultRegisteredSource(root, state, packet)
			if err != nil {
				continue
			}
		case "s7-review-result-v1", "s8-hypothesis-result-v1":
			if template == "s7-review-result-v1" {
				facts, _ := state["review"].(map[string]any)
				assignments, _ := facts["assignments"].(map[string]any)
				assignmentRow, _ := assignments[assignment].(map[string]any)
				if activity := workstate.ReviewAssignmentActivity(assignmentRow); activity == "consumed" || activity == "result_submitted" {
					continue
				}
			}
			var err error
			source, err = questionnaireSourceRoot(root, state, assignment, owner, stringValue(binding["runtime_id"]), tolerantInt(binding["baseline_generation"]))
			if err != nil {
				continue
			}
		case "s5-document-review-v1":
			data, _ := json.Marshal(row)
			var packet workpacket.Packet
			if json.Unmarshal(data, &packet) != nil {
				continue
			}
			if execution, exists, err := workspace.ReviewExecution(state, packet); err != nil {
				continue
			} else if exists {
				if registry == nil || execution.Status != "ready" {
					continue
				}
				paths, err := registry.WorkerSubmissionPaths(execution)
				if err != nil {
					continue
				}
				source, path = execution.Path, paths.Answer
			}
		}
		if path != "" {
			out = append(out, observedQuestionnaire{path: filepath.Join(source, filepath.FromSlash(path)), draft: row["draft"]})
		}
	}
	if stage == "S6" {
		questions, _ := projectS6BuilderQuestionnaires(root, "", snapshot)
		for _, q := range questions {
			activity := workstate.AgentActivity(contextAgent(state, q.AgentID))
			if activity == "reported" || activity == "stopped" {
				continue
			}
			data, err := os.ReadFile(q.ContextPath)
			if err != nil {
				continue
			}
			var context map[string]any
			if json.Unmarshal(data, &context) != nil {
				continue
			}
			if activity != "working" {
				out = append(out, observedQuestionnaire{path: q.PlanPath, draft: context["plan_draft"]})
			}
			out = append(out, observedQuestionnaire{path: q.ResultPath, draft: context["result_draft"], s6Result: q.CanonicalRecord})
		}
	}
	return out
}

// PLAN acceptance can change only machine-owned fields in a canonical S6
// Result. A generic successful-tool observer must not submit that projection
// as if the Builder had authored an answer. Explicit path saves still use the
// same original Result consumer and its normal validation.
func s6ObservedResultHasProfessionalContent(data []byte) bool {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return true // Keep malformed changed files on the original validation path.
	}
	if s6NonNull(fields["summary"]) || s6NonNull(fields["risks_blockers"]) || s6NonNull(fields["scope_deviations"]) {
		return true
	}
	if raw := fields["assertion_evidence"]; len(raw) > 0 && string(bytes.TrimSpace(raw)) != "null" {
		var rows []struct {
			EvidenceRefs []string `json:"evidence_refs"`
		}
		if err := json.Unmarshal(raw, &rows); err != nil {
			return true
		}
		for _, row := range rows {
			if len(row.EvidenceRefs) > 0 {
				return true
			}
		}
	}
	if raw := fields["closing_assessment"]; len(raw) > 0 && string(bytes.TrimSpace(raw)) != "null" {
		var assessment struct {
			Disposition json.RawMessage `json:"disposition"`
			Conditions  []struct {
				Judgment  json.RawMessage `json:"judgment"`
				Rationale json.RawMessage `json:"rationale"`
			} `json:"conditions"`
		}
		if err := json.Unmarshal(raw, &assessment); err != nil {
			return true
		}
		if s6NonNull(assessment.Disposition) {
			return true
		}
		for _, condition := range assessment.Conditions {
			if s6NonNull(condition.Judgment) || s6NonNull(condition.Rationale) {
				return true
			}
		}
	}
	return false
}

func s6NonNull(raw json.RawMessage) bool {
	return len(raw) > 0 && string(bytes.TrimSpace(raw)) != "null"
}
