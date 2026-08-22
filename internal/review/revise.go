package review

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	loopruntime "github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/schema"
	"github.com/entroforge/go-system-builder/internal/semantic"
)

// ReviseRequest drives `runtime review-plan revise` — the one controlled
// ReviewPlan revision per round (L3-S7 §3.2, §5.3).
type ReviseRequest struct {
	ExpectedRevision int
	PlanPath         string
	SourceRef        string
	AffectedSurface  string
	OccurredAt       time.Time
}

// RevisePlan applies the one permitted controlled revision. Rules:
//
//   - the round must be running / cannot_clean / discovery_draining and the
//     plan must still be at revision 1 (one revision per round, ever);
//   - v2 must declare the same plan id, round and baseline generation;
//   - every Claim whose content changed (added / modified / removed) must
//     trace to the triggering evidence: source_refs must contain SourceRef
//     and the claim target must sit under AffectedSurface;
//   - unchanged Claims keep their dispositions; changed Claims return to
//     planned; consumed Results covering changed Claims are invalidated
//     (invalidation_reason=plan_revision) so they cannot satisfy v2.
func RevisePlan(
	root, statePath, journalPath string,
	request ReviseRequest,
) (loopruntime.Snapshot, error) {
	if request.PlanPath == "" || request.SourceRef == "" || request.AffectedSurface == "" {
		return loopruntime.Snapshot{}, fmt.Errorf("revise requires --file, --source-ref and --affected-surface; a revision without a concrete source is unbounded scope creep (L3-S7 §5.3)")
	}
	data, err := os.ReadFile(request.PlanPath)
	if err != nil {
		return loopruntime.Snapshot{}, fmt.Errorf("read revised ReviewPlan: %w", err)
	}
	if err := schema.NewValidator(root).ValidateBytes("review-plan.schema.json", data); err != nil {
		return loopruntime.Snapshot{}, fmt.Errorf("revised ReviewPlan schema: %w", err)
	}
	var next Plan
	if err := json.Unmarshal(data, &next); err != nil {
		return loopruntime.Snapshot{}, fmt.Errorf("decode revised ReviewPlan: %w", err)
	}
	if err := ValidatePlan(&next); err != nil {
		return loopruntime.Snapshot{}, fmt.Errorf("revised ReviewPlan coverage: %w", err)
	}

	stateData, err := os.ReadFile(statePath)
	if err != nil {
		return loopruntime.Snapshot{}, fmt.Errorf("read runtime: %w", err)
	}
	var current map[string]any
	if err := json.Unmarshal(stateData, &current); err != nil {
		return loopruntime.Snapshot{}, fmt.Errorf("decode runtime: %w", err)
	}
	currentPlan, ptr, err := LoadPlan(root, current)
	if err != nil {
		return loopruntime.Snapshot{}, err
	}
	switch ptr.Status {
	case "running", "cannot_clean", "discovery_draining":
	default:
		return loopruntime.Snapshot{}, fmt.Errorf("ReviewPlan %s is %s; revisions are only legal while the round is still discovering", ptr.PlanID, ptr.Status)
	}
	if ptr.Revision != 1 {
		return loopruntime.Snapshot{}, fmt.Errorf("ReviewPlan %s is already at revision %d; at most one controlled revision per round (L3-S7 §5.3)", ptr.PlanID, ptr.Revision)
	}
	if next.ReviewPlanID != currentPlan.ReviewPlanID {
		return loopruntime.Snapshot{}, fmt.Errorf("a revision keeps the plan id %s (got %s)", currentPlan.ReviewPlanID, next.ReviewPlanID)
	}
	if next.ReviewRound != currentPlan.ReviewRound || next.BaselineGeneration != currentPlan.BaselineGeneration {
		return loopruntime.Snapshot{}, fmt.Errorf("a revision keeps the review_round and baseline_generation; a changed baseline is a stale round, not a revision")
	}

	changed, err := diffClaims(currentPlan, &next, request.SourceRef, request.AffectedSurface)
	if err != nil {
		return loopruntime.Snapshot{}, err
	}
	if len(changed) == 0 {
		return loopruntime.Snapshot{}, fmt.Errorf("the revision changes no Claim; nothing to revise")
	}

	// Persist v2 over the pinned path; the pointer sha updates in the CAS.
	planBytes := append(canonicalJSON(data), '\n')
	planAbs := filepath.Join(root, filepath.FromSlash(ptr.Path))
	if err := os.WriteFile(planAbs, planBytes, 0o644); err != nil {
		return loopruntime.Snapshot{}, fmt.Errorf("write revised ReviewPlan: %w", err)
	}
	planSHA := sha256Of(planBytes)

	runtimeID, _ := current["runtime_id"].(string)
	occurredAt := request.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	lifecycle, _ := current["lifecycle"].(map[string]any)
	cursor := map[string]any{"state": lifecycle["state"], "phase": lifecycle["phase"]}
	changedSet := map[string]bool{}
	for _, claimID := range changed {
		changedSet[claimID] = true
	}

	store := loopruntime.NewWriter(statePath, journalPath, root, semantic.RuntimeCandidateValidator{})
	return store.Update(request.ExpectedRevision, loopruntime.Mutation{
		EventID:        fmt.Sprintf("evt-review-revise-%s-r%d", next.ReviewPlanID, request.ExpectedRevision+1),
		TransitionID:   "REVIEW-PLAN-REVISE",
		Event:          "review_plan_revised",
		Actor:          "orchestrator",
		IdempotencyKey: fmt.Sprintf("runtime:review-revise:%s:%d", next.ReviewPlanID, request.ExpectedRevision),
		RuntimeID:      runtimeID,
		From:           cursor,
		To:             cursor,
		Message: fmt.Sprintf("Revised ReviewPlan %s to revision 2 (%d claims changed via %s)",
			next.ReviewPlanID, len(changed), request.SourceRef),
		OccurredAt: occurredAt,
		Apply: func(state map[string]any) error {
			reviewMap, ok := state["review"].(map[string]any)
			if !ok {
				return fmt.Errorf("runtime review section must be an object")
			}
			planMap, _ := reviewMap["plan"].(map[string]any)
			if planMap == nil {
				return fmt.Errorf("review plan pointer missing")
			}
			planMap["revision"] = 2
			planMap["sha256"] = planSHA

			claimsProjection, _ := reviewMap["claims"].(map[string]any)
			assignmentsProjection, _ := reviewMap["assignments"].(map[string]any)

			// Rebuild the projection from v2: unchanged claims keep their
			// disposition; changed ones return to planned.
			v2Claims := map[string]Claim{}
			for _, claim := range next.Claims {
				v2Claims[claim.ClaimID] = claim
			}
			ownerByClaim := map[string]string{}
			newAssignments := map[string]any{}
			for _, assignment := range next.Assignments {
				claimIDs := make([]any, 0, len(assignment.ClaimIDs))
				for _, claimID := range assignment.ClaimIDs {
					claimIDs = append(claimIDs, claimID)
					ownerByClaim[claimID] = assignment.AssignmentID
				}
				existing, _ := assignmentsProjection[assignment.AssignmentID].(map[string]any)
				status := "planned"
				agentID := any(nil)
				resultRef := any(nil)
				if existing != nil {
					status, _ = existing["status"].(string)
					agentID = existing["agent_id"]
					resultRef = existing["result_ref"]
				}
				// An assignment containing a changed claim must produce a
				// new result: reset consumption so review-result submit is
				// reachable again.
				touched := false
				for _, claimID := range assignment.ClaimIDs {
					if changedSet[claimID] {
						touched = true
						break
					}
				}
				if touched {
					status = "dispatched"
					if agentID == nil {
						status = "planned"
					}
					resultRef = nil
				}
				newAssignments[assignment.AssignmentID] = map[string]any{
					"lens": assignment.Lens, "claim_ids": claimIDs,
					"status": status, "agent_id": agentID, "result_ref": resultRef,
				}
			}
			newClaims := map[string]any{}
			for _, claim := range next.Claims {
				if changedSet[claim.ClaimID] {
					disposition := "planned"
					if claim.Applicability == "not_applicable" {
						disposition = "not_applicable"
					}
					newClaims[claim.ClaimID] = map[string]any{
						"lens": claim.Lens, "applicability": claim.Applicability,
						"disposition": disposition, "assignment_id": ownerByClaim[claim.ClaimID],
						"result_id": nil, "finding_ids": []any{},
					}
					continue
				}
				if existing, ok := claimsProjection[claim.ClaimID].(map[string]any); ok {
					existing["assignment_id"] = ownerByClaim[claim.ClaimID]
					newClaims[claim.ClaimID] = existing
					continue
				}
				return fmt.Errorf("claim %s is unchanged but missing from the projection — the projection is out of sync; run `runtime reconcile`", claim.ClaimID)
			}
			reviewMap["claims"] = newClaims
			reviewMap["assignments"] = newAssignments

			// Invalidate consumed results that covered changed claims: the
			// evidence stays in the journal but no longer satisfies v2.
			evidence, _ := state["evidence"].([]any)
			invalidate := map[string]bool{}
			for claimID := range changedSet {
				if old, ok := claimsProjection[claimID].(map[string]any); ok {
					if resultID, _ := old["result_id"].(string); resultID != "" {
						invalidate[resultID] = true
					}
				}
			}
			for _, raw := range evidence {
				entry, _ := raw.(map[string]any)
				if entry == nil {
					continue
				}
				if id, _ := entry["id"].(string); invalidate[id] && entry["status"] == "valid" {
					entry["status"] = "invalid"
					entry["invalidated_by"] = next.ReviewPlanID + "@r2"
					entry["invalidation_reason"] = "plan_revision"
					entry["invalidation_rule"] = "review_plan_revision"
				}
			}
			state["updated_at"] = occurredAt.UTC().Format(time.RFC3339Nano)
			return nil
		},
	})
}

// diffClaims computes which claims changed between v1 and v2 and enforces
// the source/surface binding on every change.
func diffClaims(v1, v2 *Plan, sourceRef, surface string) ([]string, error) {
	byID1 := map[string]Claim{}
	for _, claim := range v1.Claims {
		byID1[claim.ClaimID] = claim
	}
	byID2 := map[string]Claim{}
	for _, claim := range v2.Claims {
		byID2[claim.ClaimID] = claim
	}
	var changed []string
	check := func(claimID string, claim *Claim) error {
		if claim == nil {
			// Removed claim: bind the removal to the surface via the v1 row.
			original := byID1[claimID]
			if !strings.HasPrefix(original.Target, surface) {
				return fmt.Errorf("claim %s (target %s) is removed but its target sits outside the affected surface %s; removals need the same source binding", claimID, original.Target, surface)
			}
			return nil
		}
		hasSource := false
		for _, ref := range claim.SourceRefs {
			if ref == sourceRef {
				hasSource = true
				break
			}
		}
		if !hasSource {
			return fmt.Errorf("changed claim %s lacks source_ref %s; revisions must trace to the triggering Result/Finding evidence (L3-S7 §5.3)", claimID, sourceRef)
		}
		if !strings.HasPrefix(claim.Target, surface) && !strings.Contains(claim.Target, surface) {
			return fmt.Errorf("changed claim %s targets %q, outside the affected surface %q", claimID, claim.Target, surface)
		}
		return nil
	}
	for claimID, claim2 := range byID2 {
		claim1, ok := byID1[claimID]
		if !ok {
			if err := check(claimID, &claim2); err != nil {
				return nil, err
			}
			changed = append(changed, claimID)
			continue
		}
		left, _ := json.Marshal(claim1)
		right, _ := json.Marshal(claim2)
		if string(left) != string(right) {
			if err := check(claimID, &claim2); err != nil {
				return nil, err
			}
			changed = append(changed, claimID)
		}
	}
	for claimID := range byID1 {
		if _, ok := byID2[claimID]; !ok {
			if err := check(claimID, nil); err != nil {
				return nil, err
			}
			changed = append(changed, claimID)
		}
	}
	return changed, nil
}
