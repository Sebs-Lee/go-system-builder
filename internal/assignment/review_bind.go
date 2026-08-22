package assignment

import (
	"fmt"
	"sort"
	"strings"

	"github.com/entroforge/go-system-builder/internal/review"
)

// bindReviewPlanAssignments binds a reviewer workgroup's manifest
// assignments to the registered ReviewPlan (L3-S7 §3.4/§8 Assignment
// generator): every manifest assignment must name a planned plan assignment
// of the same lens, carry its exact Claim set, and respect the static ->
// behavior wave gate. On success the runtime projection records the
// dispatched Agent and flips the covered Claims to running.
//
// Runs inside the register-workgroup CAS so a plan revision and a dispatch
// never interleave.
func bindReviewPlanAssignments(root string, state map[string]any, value manifest) error {
	kind := value.WorkgroupKind
	if kind != "delivery_verifier" && kind != "qa" && kind != "e2e_browser" {
		return nil
	}
	reviewMap, _ := state["review"].(map[string]any)
	if reviewMap == nil {
		return fmt.Errorf("runtime review section missing; register the ReviewPlan via `runtime review-plan` first")
	}
	plan, ptr, err := review.LoadPlan(root, state)
	if err != nil {
		return err
	}
	if ptr.Status != "running" && ptr.Status != "cannot_clean" && ptr.Status != "discovery_draining" {
		return fmt.Errorf("ReviewPlan %s is %s; reviewer dispatch requires a running or draining round", ptr.PlanID, ptr.Status)
	}
	assignmentsProjection, _ := reviewMap["assignments"].(map[string]any)
	claimsProjection, _ := reviewMap["claims"].(map[string]any)
	if assignmentsProjection == nil || claimsProjection == nil {
		return fmt.Errorf("runtime review projection missing; re-register the ReviewPlan")
	}

	staticSettled := review.StaticClaimsSettled(state, plan)
	for _, item := range value.Assignments {
		row, _ := assignmentsProjection[item.AssignmentID].(map[string]any)
		if row == nil {
			return fmt.Errorf("assignment %s is not part of ReviewPlan %s (known: %s)", item.AssignmentID, ptr.PlanID, strings.Join(knownAssignmentIDs(assignmentsProjection), ", "))
		}
		lens, _ := row["lens"].(string)
		if review.LensToWorkgroupKind(lens) != kind {
			return fmt.Errorf("assignment %s belongs to lens %s but the workgroup kind is %s; a workgroup never mixes lenses (L3-S7 §3.4)", item.AssignmentID, lens, kind)
		}
		status, _ := row["status"].(string)
		if status != "planned" {
			return fmt.Errorf("assignment %s is already %s; one plan Assignment dispatches once", item.AssignmentID, status)
		}
		if len(item.ClaimIDs) == 0 {
			return fmt.Errorf("assignment %s carries no claim_ids; the manifest must bind the exact Claim set from the ReviewPlan", item.AssignmentID)
		}
		if err := exactClaimSet(row["claim_ids"], item.ClaimIDs); err != nil {
			return fmt.Errorf("assignment %s claim set mismatch: %w", item.AssignmentID, err)
		}
		planAssignment := findPlanAssignmentByID(plan, item.AssignmentID)
		if planAssignment != nil && planAssignment.ExecutionWave == "behavior" && !staticSettled {
			return fmt.Errorf("assignment %s is a behavior-wave (E2E/specialty) Assignment but required static Claims are not all dispositioned yet; behavior E2E unlocks only after the static set settles (L3-S7 §5.2-5.3)", item.AssignmentID)
		}
		row["status"] = "dispatched"
		row["agent_id"] = item.AgentID
		for _, claimID := range item.ClaimIDs {
			if claimRow, _ := claimsProjection[claimID].(map[string]any); claimRow != nil {
				claimRow["disposition"] = "running"
			}
		}
	}
	return nil
}

func knownAssignmentIDs(projection map[string]any) []string {
	ids := make([]string, 0, len(projection))
	for id := range projection {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func exactClaimSet(planValue any, manifestIDs []string) error {
	planIDs := map[string]bool{}
	if raw, ok := planValue.([]any); ok {
		for _, value := range raw {
			if id, _ := value.(string); id != "" {
				planIDs[id] = true
			}
		}
	}
	manifestSet := map[string]bool{}
	for _, id := range manifestIDs {
		manifestSet[id] = true
	}
	for id := range manifestSet {
		if !planIDs[id] {
			return fmt.Errorf("claim %s is not part of the plan assignment", id)
		}
	}
	for id := range planIDs {
		if !manifestSet[id] {
			return fmt.Errorf("plan claim %s is missing from the manifest assignment", id)
		}
	}
	return nil
}

func findPlanAssignmentByID(plan *review.Plan, assignmentID string) *review.PlanAssignment {
	for i := range plan.Assignments {
		if plan.Assignments[i].AssignmentID == assignmentID {
			return &plan.Assignments[i]
		}
	}
	return nil
}
