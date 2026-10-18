package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/entroforge/go-system-builder/internal/policy"
	"github.com/entroforge/go-system-builder/internal/runtime"
)

// Entry questions reuse the original REQ document and a human-readable binding
// request. Preparing either draft never locks a REQ or records human approval.
func stageEntryQuestionnaireContext(root string, request policy.Input, snapshot runtime.Snapshot) string {
	if !mainQuestionnaireViewer(request) {
		return ""
	}
	lifecycle, _ := snapshot.State["lifecycle"].(map[string]any)
	if lifecycle["state"] != "inactive" {
		return ""
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return "S0/S1 questionnaire root is unavailable; Main must restore the project path: " + err.Error()
	}
	readonly := request.Event == "Stop" || request.Event == "SubagentStop" || request.Event == "TeammateIdle"
	candidates := classifyRequirements(root)
	var locked []reqSummary
	var drafts []reqSummary
	used := map[string]bool{}
	for _, candidate := range candidates {
		used[candidate.ID] = true
		if candidate.Bindable {
			locked = append(locked, candidate)
		} else if candidate.Status != "archived" && candidate.Status != "locked" {
			drafts = append(drafts, candidate)
		}
	}
	if len(locked) > 0 {
		path := filepath.Join(root, ".claude/submissions/requirement-binding.md")
		var body strings.Builder
		body.WriteString("# Requirement binding request\n\nRead the selected locked requirement; obtain the actual human binding decision.\n\n## Candidate requirement paths\n\n")
		for _, candidate := range locked {
			fmt.Fprintf(&body, "- %s\n", filepath.Join(root, filepath.FromSlash(candidate.Path)))
		}
		body.WriteString("\n## Questions\n\n- Which requirement is selected?\n- Who actually authorized this binding? Record that human decision; do not infer it from Git identity.\n- Which development branch was selected?\n- What is the explicit release destination?\n\nRecord the real human answers here. Main then executes the original req bind using those exact choices. Saving this request does not approve or bind a requirement.\n")
		if readonly {
			if _, err := os.Stat(path); os.IsNotExist(err) {
				return "S1 binding questionnaire is not prepared at " + path + "; the next tool listener must prepare it. Main must obtain the actual human binding approval and explicit branch destinations."
			}
		} else {
			if err := prepareQuestionnaireText(path, body.String()); err != nil {
				return "S1 binding questionnaire preparation failed at " + path + ": " + err.Error()
			}
		}
		return "S1 binding questionnaire: " + path + ". Missing: selected locked REQ, actual human binding approval, explicit development branch and release destination. Collect only missing choices from the human, then Main uses the original req bind with those exact values. A saved request is not human approval or a Runtime transition."
	}
	if len(drafts) > 0 {
		var body strings.Builder
		for _, draft := range drafts {
			fmt.Fprintf(&body, "S0 requirement questionnaire: %s. Complete the existing requirement interview and unresolved professional questions; obtain the human's explicit lock decision before S1. Saving a draft does not lock or bind it.\n", filepath.Join(root, filepath.FromSlash(draft.Path)))
		}
		if len(drafts) > 1 {
			body.WriteString("Main must confirm which existing draft is the current responsibility before editing; no draft was selected automatically.")
		}
		return strings.TrimSpace(body.String())
	}
	id := ""
	for n := 1; ; n++ {
		candidate := fmt.Sprintf("REQ-%03d", n)
		if !used[candidate] {
			id = candidate
			break
		}
	}
	path := filepath.Join(root, "docs/requirements", id+".md")
	templatePath := filepath.Join(root, "docs/requirements/REQ-template.md")
	if readonly {
		return "S0 questionnaire is not prepared yet. The next tool listener must prepare " + path + " from " + templatePath + "; Main owns preparation. No human confirmation or lock is recorded."
	}
	template, err := os.ReadFile(templatePath)
	if err != nil {
		return "S0 questionnaire preparation failed: template " + templatePath + " is unavailable; Main must restore the installed requirement template: " + err.Error()
	}
	draft := strings.ReplaceAll(string(template), "REQ-{id}", id)
	draft = strings.Replace(draft, "> 状态：draft / locked / changed / archived", "> 状态：draft", 1)
	draft = strings.Replace(draft, "> UI impact：none / changed / unknown", "> UI impact：unknown", 1)
	draft = strings.Replace(draft, "YYYY-MM-DD", time.Now().UTC().Format("2006-01-02"), 1)
	if err := prepareQuestionnaireText(path, draft); err != nil {
		return "S0 questionnaire preparation failed at " + path + ": " + err.Error()
	}
	return "S0 requirement questionnaire: " + path + "; source template: " + templatePath + ". Fill the existing requirement interview with the real user's goals, constraints, acceptance criteria and unresolved decisions. Save the draft; obtain explicit human confirmation and lock before S1. No answers or human approval were prefilled."
}

// Choose the responsibility's questions; this makes no tool permission decision.
func mainQuestionnaireViewer(request policy.Input) bool {
	owner := request.EffectiveAgentID()
	return (owner == "" || owner == "Main") && !(request.Event == "SubagentStart" && owner == "")
}

func prepareQuestionnaireText(path, text string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if os.IsExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	_, writeErr := file.WriteString(text)
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}

func currentQuestionnaireLifecycleContext(root string, request policy.Input) string {
	if request.Event == "SubagentStart" && request.EffectiveAgentID() == "" {
		return ""
	}
	snapshot, err := runtime.NewStore(filepath.Join(root, ".claude/loop-state.json"), filepath.Join(root, ".claude/loop-events.jsonl")).Snapshot()
	if err == nil {
		lifecycle, _ := snapshot.State["lifecycle"].(map[string]any)
		if runtime.PacketEnabled(snapshot.State) && lifecycle["state"] == "bug_resolution" && s9ProfessionalPhase(stringValue(lifecycle["phase"])) && request.EffectiveAgentID() == "" && (request.AgentType != "" || request.AgentTranscriptPath != "" || request.Event == "SubagentStop" || request.Event == "TeammateIdle") {
			return "S9 native Worker responsibility is not identified by this event. Main must provide its exact registered questionnaire path in the actual task prompt; saving that path uses its original registered owner. No Main responsibility was inferred."
		}
		if lifecycle["state"] == "building" {
			return stageBuilderQuestionLifecycleContext(root, request, snapshot)
		}
		if context := stageEntryQuestionnaireContext(root, request, snapshot); context != "" {
			return context
		}
		if context := stageRepairLeadQuestionnaireContext(root, request, snapshot); context != "" {
			if runtime.PacketEnabled(snapshot.State) && lifecycle["state"] == "bug_resolution" && s9ProfessionalPhase(stringValue(lifecycle["phase"])) {
				return strings.TrimSpace(strings.Join([]string{stageRepairVerifierDispatchContext(root, request, snapshot), context}, "\n"))
			}
			return context
		}
		if runtime.PacketEnabled(snapshot.State) && lifecycle["state"] == "bug_resolution" && s9ProfessionalPhase(stringValue(lifecycle["phase"])) {
			dispatch := ""
			if request.EffectiveAgentID() != "" || request.AgentType == "" && request.AgentTranscriptPath == "" {
				dispatch = stageRepairDispatchContext(root, request, snapshot)
			}
			if context, handled := s9PacketQuestionnaireLifecycleContext(root, request, snapshot); handled {
				return strings.TrimSpace(strings.Join([]string{dispatch, stageRepairVerifierDispatchContext(root, request, snapshot), context}, "\n"))
			}
			return strings.TrimSpace(strings.Join([]string{dispatch, stageRepairVerifierDispatchContext(root, request, snapshot)}, "\n"))
		}
		if context := stageReviewPlanQuestionnaireContext(root, request, snapshot); context != "" {
			return context
		}
		stopFeedback := request.Event == "Stop" || request.Event == "SubagentStop" || request.Event == "TeammateIdle"
		if lifecycle["state"] == "bug_resolution" && lifecycle["phase"] == "investigation" && runtime.PacketEnabled(snapshot.State) {
			ctx, cancel := context.WithTimeout(contextForHook(request), 4*time.Second)
			defer cancel()
			lead := stageInvestigationLeadQuestionnaireContext(root, request, snapshot)
			questions := stageHypothesisResultLifecycleContext(ctx, root, request, nil, snapshot, request.EffectiveAgentID())
			return strings.TrimSpace(strings.Join([]string{lead, questions}, "\n"))
		}
		if lifecycle["state"] == "verification" && runtime.PacketEnabled(snapshot.State) && (request.Event == "PreToolUse" || request.Event == "SessionStart" || request.Event == "SubagentStart" || request.Event == "PreCompact" || stopFeedback) {
			ctx, cancel := context.WithTimeout(contextForHook(request), 4*time.Second)
			defer cancel()
			// No service is needed to read current S7 facts. The original issuer
			// initializes it only if a relevant Pre actually needs preparation.
			return stageReviewResultLifecycleContext(ctx, root, request, nil, snapshot, request.EffectiveAgentID())
		}
		if context := stageAuditQuestionnaireContext(root, request, snapshot); context != "" {
			return context
		}
		if context := stageHumanReleaseQuestionnaireContext(root, request, snapshot); context != "" {
			return context
		}
	}
	return stageAnswerLifecycleContext(root, request)
}
