package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/entroforge/go-system-builder/internal/policy"
	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/workspace"
)

// Read the registered execution projection only; reminders never re-run the
// scheduler, workspace decoder/authority checks, Git, or physical check proofs.
func repairExecutionForChoice(state map[string]any, row repairDispatchChoice) (string, map[string]any) {
	owner := repairDispatchOwner(state, row)
	id := s9ListenerManifestAssignment(state, owner)
	executions := mapFieldCLI(mapFieldCLI(state, "workspace"), "executions")
	if execution := mapFieldCLI(executions, id); len(execution) > 0 {
		return id, execution
	}
	if execution := mapFieldCLI(executions, row.AssignmentID); len(execution) > 0 {
		return row.AssignmentID, execution
	}
	return id, nil
}

func repairExecutionCheckLocationConflict(state map[string]any, row repairDispatchChoice) error {
	_, execution := repairExecutionForChoice(state, row)
	frozen := stringValue(execution["check_location"])
	if row.CheckLocation != "" && frozen != "" && frozen != row.CheckLocation {
		return fmt.Errorf("Assignment %s already froze check_location=%s in its original execution; the changed choice %s was not applied", row.AssignmentID, frozen, row.CheckLocation)
	}
	return nil
}

func repairExecutionProjectionContext(root string, state map[string]any, row repairDispatchChoice) string {
	id, execution := repairExecutionForChoice(state, row)
	if len(execution) == 0 {
		return ""
	}
	if stringValue(execution["status"]) == "complete" {
		return fmt.Sprintf("Assignment %s is complete; Main integration/checks and cleanup are recorded. ", row.AssignmentID)
	}
	if stringValue(execution["delivery_ref"]) != "" {
		launcher := s7CaptureShellQuote(filepath.Join(root, ".claude/bin/loop-harness"))
		return fmt.Sprintf("Assignment %s: candidate received; Main runs %s runtime workspace integrate --root %s --assignment %s --agent %s and selects each actual necessary Main check with repeatable --check-command. ", row.AssignmentID, launcher, s7CaptureShellQuote(root), s7CaptureShellQuote(id), s7CaptureShellQuote(stringValue(execution["agent_id"])))
	}
	owner := stringValue(execution["agent_id"])
	path := stringValue(execution["worktree_path"])
	var b strings.Builder
	fmt.Fprintf(&b, "Original execution %s owner=%s status=%s check_location=%s; actual Worker root: %s; target branch=%s. ", id, owner, stringValue(execution["status"]), stringValue(execution["check_location"]), path, stringValue(execution["target_branch"]))
	if stringValue(execution["status"]) != "ready" {
		b.WriteString("This execution is not ready; retain it and follow its original preparation/recovery responsibility. ")
	}
	contextPath := filepath.Join(path, filepath.FromSlash(workspace.DispatchContextPath))
	if _, err := os.Stat(contextPath); err == nil {
		fmt.Fprintf(&b, "Assignment context: %s. ", contextPath)
	}
	packets, err := s9ListenerQuestionnaires(state, owner, true)
	if err != nil {
		fmt.Fprintf(&b, "Current Result path projection is unavailable: %v. ", err)
	} else {
		found := false
		for _, packet := range packets {
			if packet.binding.RepairAssignmentID != row.AssignmentID {
				continue
			}
			found = true
			fmt.Fprintf(&b, "Original RepairResult view: %s; exact professional answer: %s. ", filepath.Join(path, filepath.FromSlash(packet.view)), filepath.Join(path, filepath.FromSlash(packet.answer)))
		}
		if !found {
			b.WriteString("The original RepairResult questionnaire is not prepared yet; Main/system must prepare it during the normal work startup and supply its exact path. No Result answer path was guessed. ")
		}
	}
	return b.String()
}

// This context reads the existing division choice and reports normal startup
// commands. Neither observation nor saving prepares a workspace.
func stageRepairExecutionPrepareContext(root string, request policy.Input, snapshot runtime.Snapshot) string {
	if !mainQuestionnaireViewer(request) || stringValue(mapFieldCLI(snapshot.State, "lifecycle")["state"]) != "bug_resolution" || stringValue(repairPointerCLI(snapshot.State)["status"]) != "repairing" {
		return ""
	}
	paths := repairDispatchPaths(root, snapshot.State)
	if len(paths) != 1 {
		return "S9 execution preparation cannot resolve the original Builder division answer from the current Case/Session/Plan binding."
	}
	plan, err := repairDispatchObservedPlan(root, snapshot.State)
	if err != nil {
		return fmt.Sprintf("S9 execution preparation at %s cannot read the original Plan: %v", paths[0], err)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "S9 execution preparation uses the same Builder division answer: %s. ", paths[0])
	if questionnaireMissing(paths[0]) {
		b.WriteString("The same division answer is not prepared yet; Main/system must supply its rows from the original Plan and registered dispatch/execution facts during normal responsibility preparation. This event does not invent another report. ")
	}
	pointer := repairPointerCLI(snapshot.State)
	choices := map[string]repairDispatchChoice{}
	if data, err := os.ReadFile(paths[0]); err == nil {
		var rows []repairDispatchChoice
		if json.Unmarshal(data, &rows) == nil {
			for _, row := range rows {
				choices[row.AssignmentID] = row
			}
		}
	}
	priority := func(id string) int {
		_, execution := repairExecutionForChoice(snapshot.State, repairDispatchChoice{AssignmentID: id})
		if stringValue(execution["status"]) == "complete" {
			return 2
		}
		if stringValue(execution["delivery_ref"]) != "" {
			return 0
		}
		return 1
	}
	sort.SliceStable(plan.Assignments, func(i, j int) bool {
		return priority(plan.Assignments[i].AssignmentID) < priority(plan.Assignments[j].AssignmentID)
	})
	for _, assignment := range plan.Assignments {
		row := choices[assignment.AssignmentID]
		row.AssignmentID = assignment.AssignmentID
		if len(assignment.DependsOn) > 0 {
			fmt.Fprintf(&b, "Assignment %s depends_on=%v; Main must arrange the named predecessor's actual Result before starting dependent work. ", assignment.AssignmentID, assignment.DependsOn)
		}
		if len(assignment.ResourceLocks) > 0 {
			fmt.Fprintf(&b, "Assignment %s resource_locks=%v; Main must arrange work around these shared resources. ", assignment.AssignmentID, assignment.ResourceLocks)
		}
		owner := stringMapCLI(pointer["assignment_owners"])[assignment.AssignmentID]
		if owner == "" {
			fmt.Fprintf(&b, "Assignment %s awaits Builder selection at %s. ", assignment.AssignmentID, paths[0])
			continue
		}
		b.WriteString(prepareRepairExecutionFromChoice(root, snapshot.State, row, owner))
	}
	b.WriteString("Main owns Claude Code native spawn/resume and later integration; workspace commands apply only to an existing frozen legacy responsibility; this reminder only reads current facts. ")
	return b.String()
}

// Keep the existing save adapter as pure feedback; Main runs normal startup.
func prepareRepairExecutionFromChoice(root string, state map[string]any, row repairDispatchChoice, owner string) string {
	_, current := repairExecutionForChoice(state, row)
	for _, packet := range nativeS9PlanPackets(root, state) {
		if packet.Binding.OwnerAgentID != owner || packet.Binding.RepairAssignmentID != row.AssignmentID {
			continue
		}
		if stringValue(current["status"]) == "complete" {
			return fmt.Sprintf("S9 Assignment %s is already integrated and cleaned through the original Main consumer; continue remaining professional responsibilities. ", row.AssignmentID)
		}
		if stringValue(current["delivery_ref"]) != "" {
			return repairExecutionProjectionContext(root, state, row) + fmt.Sprintf("Native delivery received. Main selects actual necessary checks and runs original runtime workspace integrate --root %q --assignment %q --agent %q, supplying each actual check with --check-command. Do not launch another Worker. ", root, stringValue(current["assignment_id"]), owner)
		}
		brief, _, _ := nativeS9Paths(packet)
		return fmt.Sprintf("Main uses Claude Code native Agent execution/resume with this exact S9 brief: %s. Builder continues its own plan, repair, self-checks and Result in its native worktree. No workspace prepare/launch is required. ", filepath.Join(root, brief))
	}
	if stringValue(current["status"]) == "complete" || stringValue(current["delivery_ref"]) != "" {
		return repairExecutionProjectionContext(root, state, row)
	}
	if runtime.PacketEnabled(state) && (len(current) == 0 || current["purpose"] == "native_delivery") {
		return fmt.Sprintf("S9 Assignment %s owner=%s has no current prepared native PlanReport/brief. Main/system must restore the original division preparation at %s; no workspace startup or check_location selection was inferred from the missing native material. ", row.AssignmentID, owner, strings.Join(repairDispatchPaths(root, state), ", "))
	}
	if err := repairExecutionCheckLocationConflict(state, row); err != nil {
		return "S9 execution choice conflict: " + err.Error() + ". "
	}
	id, execution := repairExecutionForChoice(state, row)
	projection := repairExecutionProjectionContext(root, state, row)
	status := stringValue(execution["status"])
	if len(execution) > 0 && status != "preparing" && status != "ready" {
		return projection + "Retain this execution and follow its original recovery responsibility. "
	}
	if status == "ready" {
		packets, err := s9ListenerQuestionnaires(state, owner, true)
		if err == nil {
			for _, packet := range packets {
				if packet.binding.RepairAssignmentID != row.AssignmentID {
					continue
				}
				path := stringValue(execution["worktree_path"])
				_, viewErr := os.Stat(filepath.Join(path, filepath.FromSlash(packet.view)))
				_, answerErr := os.Stat(filepath.Join(path, filepath.FromSlash(packet.answer)))
				if viewErr == nil && answerErr == nil {
					return projection
				}
			}
		}
	}
	location := row.CheckLocation
	if location == "" {
		location = stringValue(execution["check_location"])
	}
	if location != "main" && location != "worker" {
		paths := repairDispatchPaths(root, state)
		return projection + fmt.Sprintf("Assignment %s awaits check_location main or worker in the existing division answer %s; save the selection, then Main runs the supplied startup command. ", row.AssignmentID, strings.Join(paths, ", "))
	}
	if id == "" {
		return projection + "Assignment " + row.AssignmentID + " has no registered manifest Assignment for " + owner + "; Main must finish its original dispatch. "
	}
	launcher := s7CaptureShellQuote(filepath.Join(root, ".claude/bin/loop-harness"))
	var b strings.Builder
	b.WriteString(projection)
	if state["workspace"] == nil {
		fmt.Fprintf(&b, "Main first runs: %s runtime workspace bind --root %s. ", launcher, s7CaptureShellQuote(root))
	}
	if status == "ready" {
		b.WriteString("The Worker is ready but its Result questionnaire is missing; finish normal startup before sending the author to work. ")
	}
	fmt.Fprintf(&b, "Main runs through a normal tool: %s runtime workspace prepare --root %s --assignment %s --agent %s --check-location %s. Use its actual Worker root and exact RepairResult paths. Saving the division answer did not run bind/prepare. ", launcher, s7CaptureShellQuote(root), s7CaptureShellQuote(id), s7CaptureShellQuote(owner), s7CaptureShellQuote(location))
	return b.String()
}
