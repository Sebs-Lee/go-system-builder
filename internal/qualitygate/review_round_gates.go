package qualitygate

import (
	"fmt"
	"sort"

	"github.com/entroforge/go-system-builder/internal/review"
	"github.com/entroforge/go-system-builder/internal/verification"
)

// applyCleanRoundGate recomputes the machine CleanRound over the current
// ReviewPlan's exact Claim set (L3-S7 §10). It runs unconditionally — like
// the S6 builder-batch evaluation — so the missing matrix names each unproven
// Claim instead of only the aggregate evidence token.
func applyCleanRoundGate(input Input, result *Evaluation) {
	evaluation := verification.EvaluateCleanRound(input.Snapshot.State)
	if evaluation.Passed {
		return
	}
	for _, check := range evaluation.Checks {
		if check.Passed {
			continue
		}
		result.Missing = append(result.Missing, "cleanround:"+check.Name)
	}
	result.Status = StatusNotReady
}

// applyObservationBatchGate proves the sealed ObservationBatch carries the
// exact current-round Finding set and that an ordinary batch sealed only
// after the full required Claim set was dispositioned (L3-S7 §3.7).
func applyObservationBatchGate(input Input, result *Evaluation) {
	state := input.Snapshot.State
	reviewMap, _ := state["review"].(map[string]any)
	plan, _ := reviewMap["plan"].(map[string]any)
	if plan == nil {
		result.Missing = append(result.Missing, "batch:review_plan_missing")
		result.Status = StatusNotReady
		return
	}
	if status, _ := plan["status"].(string); status != "observation_sealed" {
		result.Missing = append(result.Missing, "batch:plan_status="+status)
	}
	batch, _ := reviewMap["observation_batch"].(map[string]any)
	if batch == nil {
		result.Missing = append(result.Missing, "batch:observation_batch_not_sealed")
		result.Status = StatusNotReady
		return
	}
	batchIDs := map[string]bool{}
	if raw, ok := batch["finding_ids"].([]any); ok {
		for _, value := range raw {
			if id, _ := value.(string); id != "" {
				batchIDs[id] = true
			}
		}
	}
	actual := map[string]bool{}
	for _, row := range review.RoundFindings(state) {
		if id, ok := row["finding_id"].(string); ok {
			actual[id] = true
		}
	}
	var mismatch []string
	for id := range batchIDs {
		if !actual[id] {
			mismatch = append(mismatch, id+":not_a_current_round_finding")
		}
	}
	for id := range actual {
		if !batchIDs[id] {
			mismatch = append(mismatch, id+":missing_from_batch")
		}
	}
	if len(mismatch) > 0 {
		sort.Strings(mismatch)
		for _, entry := range mismatch {
			result.Missing = append(result.Missing, "batch:finding_set:"+entry)
		}
	}
	if policy, _ := batch["drain_policy"].(string); policy != "immediate_stop" {
		for _, claimID := range review.UndispositionedRequired(state) {
			result.Missing = append(result.Missing, "batch:unobserved_claim:"+claimID)
		}
	}
	if len(result.Missing) > 0 {
		result.Status = StatusNotReady
	}
}

// reviewRoundGateIDs are the two S7 exit gates whose semantics are computed
// over the ReviewPlan projection rather than evidence presence alone.
func isReviewRoundGate(gateID string) bool {
	return gateID == "GATE-VERIFY-CLEAN-ROUND-PASSED" || gateID == "GATE-VERIFY-BLOCKING-FINDING"
}

// describeReviewRoundGate renders the gate's missing-token legend rows.
func describeReviewRoundGate(gateID string) []string {
	if gateID == "GATE-VERIFY-CLEAN-ROUND-PASSED" {
		return []string{fmt.Sprintf("cleanround:<check> — the machine CleanRound check that still fails (run `loop-harness s7 status` for the claim-level board)")}
	}
	return []string{"batch:<token> — the ObservationBatch exact-set fact that is still missing"}
}
