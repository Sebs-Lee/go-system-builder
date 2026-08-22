package review

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	loopruntime "github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/schema"
	"github.com/entroforge/go-system-builder/internal/semantic"
)

// SubmitRequest drives `runtime review-result submit`.
type SubmitRequest struct {
	ExpectedRevision int
	AssignmentID     string
	ResultPath       string
	// CaptureDir optionally points at a captures directory; findings with an
	// empty encounter timeline absorb the buffered steps (L3-S7 §3.6).
	CaptureDir       string
	OccurredAt       time.Time
}

// SubmitResult is the single entry point for a Canonical ReviewResult
// (L3-S7 §9.1). One CAS transaction:
//
//  1. validates the result against review-plan coordinates, the Assignment's
//     exact Claim set, producer identity and Builder/Reviewer independence;
//  2. persists the result envelope plus one immutable Finding per fail Claim;
//  3. advances the reviewer Agent (working -> reported) and marks the
//     Assignment consumed;
//  4. updates the claim disposition projection; a finding flips the round to
//     cannot_clean (critical P0 seals immediately; ordinary findings drain);
//  5. when the final required Claim disposition lands, the round consumer
//     runs in the same transaction: findings -> sealed ObservationBatch,
//     no findings -> machine CleanRound;
//  6. pause verdicts (req_change_required / release_blocked) create the one
//     authoritative pause checkpoint here; TR-010/TR-011 only move the cursor.
func SubmitResult(
	root, statePath, journalPath string,
	request SubmitRequest,
) (loopruntime.Snapshot, error) {
	if request.AssignmentID == "" || request.ResultPath == "" {
		return loopruntime.Snapshot{}, fmt.Errorf("--assignment-id and --result are required")
	}
	data, err := os.ReadFile(request.ResultPath)
	if err != nil {
		return loopruntime.Snapshot{}, fmt.Errorf("read ReviewResult: %w", err)
	}
	if err := schema.NewValidator(root).ValidateBytes("review-result.schema.json", data); err != nil {
		return loopruntime.Snapshot{}, fmt.Errorf("ReviewResult schema: %w", err)
	}
	var result Result
	if err := json.Unmarshal(data, &result); err != nil {
		return loopruntime.Snapshot{}, fmt.Errorf("decode ReviewResult: %w", err)
	}
	if result.AssignmentID != request.AssignmentID {
		return loopruntime.Snapshot{}, fmt.Errorf("ReviewResult assignment_id %s does not match --assignment-id %s", result.AssignmentID, request.AssignmentID)
	}
	// Capture-buffer merge: findings whose encounter timeline is empty absorb
	// the buffered steps; reviewer-written timelines are never rewritten.
	if request.CaptureDir != "" {
		mergeCapturedTimeline(result.Findings, LoadCaptureSteps(request.CaptureDir))
	}

	stateData, err := os.ReadFile(statePath)
	if err != nil {
		return loopruntime.Snapshot{}, fmt.Errorf("read runtime: %w", err)
	}
	var current map[string]any
	if err := json.Unmarshal(stateData, &current); err != nil {
		return loopruntime.Snapshot{}, fmt.Errorf("decode runtime: %w", err)
	}
	lifecycle, _ := current["lifecycle"].(map[string]any)
	if state, _ := lifecycle["state"].(string); state != "verification" {
		return loopruntime.Snapshot{}, fmt.Errorf("ReviewResults can only be submitted in the verification stage (current state: %s)", lifecycle["state"])
	}
	plan, ptr, err := LoadPlan(root, current)
	if err != nil {
		return loopruntime.Snapshot{}, err
	}
	round := currentReviewRound(current)
	generation := baselineGeneration(current)
	switch ptr.Status {
	case "running", "cannot_clean", "discovery_draining":
	default:
		return loopruntime.Snapshot{}, fmt.Errorf("ReviewPlan %s is %s; results are only accepted while the round is running or draining", ptr.PlanID, ptr.Status)
	}
	if result.ReviewPlanID != plan.ReviewPlanID {
		return loopruntime.Snapshot{}, fmt.Errorf("ReviewResult binds plan %s but the registered plan is %s", result.ReviewPlanID, plan.ReviewPlanID)
	}
	if result.ReviewRound != round {
		return loopruntime.Snapshot{}, fmt.Errorf("ReviewResult declares review_round %d but the runtime is at round %d", result.ReviewRound, round)
	}
	if result.BaselineGeneration != generation {
		return loopruntime.Snapshot{}, fmt.Errorf("ReviewResult declares baseline_generation %d but the runtime is at generation %d", result.BaselineGeneration, generation)
	}
	if digest := SubjectDigest(plan); result.SubjectDigest != digest {
		return loopruntime.Snapshot{}, fmt.Errorf("subject_digest mismatch: the result binds %s but the frozen baseline digests to %s; a drifted baseline makes the round stale, not submittable", result.SubjectDigest, digest)
	}

	assignment := findPlanAssignment(plan, result.AssignmentID)
	if assignment == nil {
		return loopruntime.Snapshot{}, fmt.Errorf("assignment %s is not part of ReviewPlan %s (known: %s)", result.AssignmentID, plan.ReviewPlanID, strings.Join(planAssignmentIDs(plan), ", "))
	}
	if err := validateClaimResultSet(assignment, &result); err != nil {
		return loopruntime.Snapshot{}, err
	}
	if err := validateVerdictConsistency(&result); err != nil {
		return loopruntime.Snapshot{}, err
	}
	if err := validateFindings(plan, assignment, &result); err != nil {
		return loopruntime.Snapshot{}, err
	}
	if err := validateProducerIndependence(current, &result); err != nil {
		return loopruntime.Snapshot{}, err
	}
	if err := verifyResultArtifactDigest(root, plan, ptr, &result, assignment.Lens); err != nil {
		return loopruntime.Snapshot{}, err
	}

	runtimeID, _ := current["runtime_id"].(string)
	occurredAt := request.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	lens := assignment.Lens
	responsibility := LensToResponsibility(lens)

	// Persist artifacts before the CAS (same pattern as the S6 Builder
	// Result): bytes on disk are what the evidence index fingerprints.
	resultRel := filepath.ToSlash(filepath.Join(
		".claude", "evidence", runtimeID, fmt.Sprintf("g%d", generation),
		"reviews", result.ProducerAgentID, result.ResultID+".json"))
	resultEnvelope := map[string]any{
		"schema_version":          "1.0.0",
		"evidence_id":             result.ResultID,
		"kind":                    "review_result",
		"runtime_id":              runtimeID,
		"baseline_generation":     generation,
		"review_round":            round,
		"producer_agent_id":       result.ProducerAgentID,
		"producer_responsibility": responsibility,
		"subject_refs":            []any{},
		"conclusion":              result.Verdict,
		"review_plan_id":          plan.ReviewPlanID,
		"assignment_id":           result.AssignmentID,
		"subject_digest":          result.SubjectDigest,
		"claim_results":           result.ClaimResults,
		"checks":                  result.Checks,
		"deviations":              result.Deviations,
		"verdict":                 result.Verdict,
		"created_at":              occurredAt.UTC().Format(time.RFC3339Nano),
	}
	resultBytes, err := marshalArtifact(resultEnvelope)
	if err != nil {
		return loopruntime.Snapshot{}, fmt.Errorf("encode review result envelope: %w", err)
	}
	if err := writeArtifact(root, resultRel, resultBytes); err != nil {
		return loopruntime.Snapshot{}, err
	}
	resultSHA := sha256Of(resultBytes)

	findingArtifacts := make([]findingArtifact, 0, len(result.Findings))
	for _, finding := range result.Findings {
		rel := filepath.ToSlash(filepath.Join(
			".claude", "evidence", runtimeID, fmt.Sprintf("g%d", generation),
			"findings", finding.FindingID+".json"))
		bytes, err := marshalArtifact(finding)
		if err != nil {
			return loopruntime.Snapshot{}, fmt.Errorf("encode finding %s: %w", finding.FindingID, err)
		}
		if err := writeArtifact(root, rel, bytes); err != nil {
			return loopruntime.Snapshot{}, err
		}
		findingArtifacts = append(findingArtifacts, findingArtifact{finding: finding, rel: rel, sha: sha256Of(bytes)})
	}

	// The round consumer runs in the same transaction: project the final
	// dispositions with this result applied, then decide seal / clean.
	projected := projectDispositions(current, assignment, &result)
	complete := roundCompleteWith(projected)
	var batchRel, batchSHA string
	var batchID string
	var cleanRel, cleanSHA string
	sealNow := complete && len(RoundFindings(current))+len(result.Findings) > 0
	if !sealNow {
		for _, f := range result.Findings {
			if f.Severity == "P0" {
				// Critical findings stop the line: seal immediately with the
				// unobserved Claims made explicit (L3-S7 §3.7, §5.2).
				sealNow = true
				break
			}
		}
	}
	cleanNow := complete && !sealNow && len(RoundFindings(current)) == 0 && len(result.Findings) == 0

	if (sealNow || cleanNow) && ptr.VerificationArtifactWorkspace != "" {
		if err := verifySealedArtifactDigests(root, ptr, projectedAssignments(current, assignment, &result)); err != nil {
			return loopruntime.Snapshot{}, err
		}
	}
	if sealNow {
		batchID = fmt.Sprintf("observation-batch-r%d", round)
		batch, err := buildObservationBatch(state_view{current}, plan, ptr, projected, &result, complete, occurredAt)
		if err != nil {
			return loopruntime.Snapshot{}, err
		}
		batchRel = filepath.ToSlash(filepath.Join(
			".claude", "evidence", runtimeID, fmt.Sprintf("g%d", generation),
			"review", batchID+".json"))
		batchBytes, err := marshalArtifact(batch)
		if err != nil {
			return loopruntime.Snapshot{}, fmt.Errorf("encode ObservationBatch: %w", err)
		}
		if err := schema.NewValidator(root).ValidateBytes("observation-batch.schema.json", batchBytes); err != nil {
			return loopruntime.Snapshot{}, fmt.Errorf("ObservationBatch schema: %w", err)
		}
		if err := writeArtifact(root, batchRel, batchBytes); err != nil {
			return loopruntime.Snapshot{}, err
		}
		batchSHA = sha256Of(batchBytes)
	}
	if cleanNow {
		cleanRel = filepath.ToSlash(filepath.Join(
			".claude", "evidence", runtimeID, fmt.Sprintf("g%d", generation),
			"review", fmt.Sprintf("clean-round-r%d.json", round)))
		snapshot := buildCleanRoundSnapshot(current, plan, resultEnvelope, resultSHA, resultRel, occurredAt)
		cleanBytes, err := marshalArtifact(snapshot)
		if err != nil {
			return loopruntime.Snapshot{}, fmt.Errorf("encode CleanRound: %w", err)
		}
		if err := writeArtifact(root, cleanRel, cleanBytes); err != nil {
			return loopruntime.Snapshot{}, err
		}
		cleanSHA = sha256Of(cleanBytes)
	}

	cursor := map[string]any{"state": lifecycle["state"], "phase": lifecycle["phase"]}
	resultRepoPath := repositoryPath(root, request.ResultPath)

	store := loopruntime.NewWriter(statePath, journalPath, root, semantic.RuntimeCandidateValidator{})
	return store.Update(request.ExpectedRevision, loopruntime.Mutation{
		EventID:        fmt.Sprintf("evt-review-result-%s-r%d", result.ResultID, request.ExpectedRevision+1),
		TransitionID:   "REVIEW-RESULT",
		Event:          "review_result_submitted",
		Actor:          "orchestrator",
		IdempotencyKey: fmt.Sprintf("runtime:review-result:%s:%d", result.ResultID, request.ExpectedRevision),
		RuntimeID:      runtimeID,
		From:           cursor,
		To:             cursor,
		EvidenceIDs:    append([]string{result.ResultID}, findingIDs(result.Findings)...),
		Message: fmt.Sprintf("Consumed Canonical ReviewResult %s for assignment %s (verdict %s)",
			result.ResultID, result.AssignmentID, result.Verdict),
		OccurredAt: occurredAt,
		Apply: func(state map[string]any) error {
			if err := applyResultConsumption(state, plan, assignment, &result, resultRel, resultSHA, responsibility, generation, round, occurredAt); err != nil {
				return err
			}
			if err := applyFindings(state, &result, findingArtifacts, round, occurredAt); err != nil {
				return err
			}
			if err := advanceReviewerAgent(state, &result, resultRepoPath, occurredAt); err != nil {
				return err
			}
			reviewMap := state["review"].(map[string]any)
			lifecycleMap := state["lifecycle"].(map[string]any)
			switch result.Verdict {
			case "req_change_required", "release_blocked":
				// One authoritative checkpoint in the verdict transaction;
				// TR-010/TR-011 then carry only the result evidence
				// (L3-S7 §9.2, §13.1 dual-carrier removal).
				if err := capturePauseCheckpoint(state, "S7 review verdict: "+result.Verdict, occurredAt); err != nil {
					return err
				}
				setPlanStatus(reviewMap, lifecycleMap, "paused")
				return nil
			}
			if len(result.Findings) > 0 {
				if status, _ := reviewMap["plan"].(map[string]any)["status"].(string); status == "running" {
					setPlanStatus(reviewMap, lifecycleMap, "cannot_clean")
				} else if status == "cannot_clean" && !sealNow {
					setPlanStatus(reviewMap, lifecycleMap, "discovery_draining")
				}
			}
			if sealNow {
				if err := applyObservationBatch(state, batchID, batchRel, batchSHA, result.Findings, round, occurredAt); err != nil {
					return err
				}
				setPlanStatus(reviewMap, lifecycleMap, "observation_sealed")
				return nil
			}
			if cleanNow {
				if err := applyCleanRound(state, cleanRel, cleanSHA, round, occurredAt); err != nil {
					return err
				}
				setPlanStatus(reviewMap, lifecycleMap, "clean")
				reviewMap["clean_round"] = round
				return nil
			}
			return nil
		},
	})
}

// ---------------------------------------------------------------------------
// validation helpers
// ---------------------------------------------------------------------------

func findPlanAssignment(plan *Plan, assignmentID string) *PlanAssignment {
	for i := range plan.Assignments {
		if plan.Assignments[i].AssignmentID == assignmentID {
			return &plan.Assignments[i]
		}
	}
	return nil
}

func planAssignmentIDs(plan *Plan) []string {
	ids := make([]string, 0, len(plan.Assignments))
	for _, assignment := range plan.Assignments {
		ids = append(ids, assignment.AssignmentID)
	}
	sort.Strings(ids)
	return ids
}

// validateClaimResultSet proves claim_results == the Assignment's exact
// Claim set: no missing, no extras, no duplicates (L3-S7 §3.5).
func validateClaimResultSet(assignment *PlanAssignment, result *Result) error {
	want := map[string]bool{}
	for _, claimID := range assignment.ClaimIDs {
		want[claimID] = true
	}
	seen := map[string]bool{}
	for _, claimResult := range result.ClaimResults {
		if !want[claimResult.ClaimID] {
			return fmt.Errorf("claim_results contains %s which is not part of assignment %s; a Reviewer never adds Claims (L3-S7 §3.5)", claimResult.ClaimID, assignment.AssignmentID)
		}
		if seen[claimResult.ClaimID] {
			return fmt.Errorf("claim_results contains %s twice", claimResult.ClaimID)
		}
		seen[claimResult.ClaimID] = true
	}
	for claimID := range want {
		if !seen[claimID] {
			return fmt.Errorf("claim_results is missing %s; the Assignment's Claim set must be answered exactly (L3-S7 §3.5)", claimID)
		}
	}
	return nil
}

// validateVerdictConsistency keeps the verdict from contradicting the local
// claim results (L3-S7 §3.5).
func validateVerdictConsistency(result *Result) error {
	failures := 0
	for _, claimResult := range result.ClaimResults {
		if claimResult.Conclusion == "fail" {
			failures++
		}
	}
	failedChecks := 0
	for _, check := range result.Checks {
		if check.Result == "fail" {
			failedChecks++
		}
	}
	switch result.Verdict {
	case "pass":
		if failures > 0 || len(result.Findings) > 0 || failedChecks > 0 {
			return fmt.Errorf("verdict=pass contradicts %d fail claim(s), %d finding(s), %d failed check(s)", failures, len(result.Findings), failedChecks)
		}
		if len(result.Deviations) > 0 {
			return fmt.Errorf("verdict=pass contradicts %d recorded deviation(s)", len(result.Deviations))
		}
	case "finding":
		if failures == 0 || len(result.Findings) == 0 {
			return fmt.Errorf("verdict=finding requires at least one fail claim and one Finding with a real encounter")
		}
	case "req_change_required", "release_blocked":
		// Pause verdicts route to the human gateway; claim results still
		// cover the exact set so the resumed round can rely on them.
	default:
		return fmt.Errorf("unknown verdict %q", result.Verdict)
	}
	return nil
}

// validateFindings binds every fail Claim to exactly one Finding, validates
// each Finding against the schema, and enforces investigation readiness for
// ordinary findings (L3-S7 §3.6/§3.7).
func validateFindings(plan *Plan, assignment *PlanAssignment, result *Result) error {
	validator := schema.NewEmbeddedValidator()
	failByClaim := map[string]bool{}
	for _, claimResult := range result.ClaimResults {
		if claimResult.Conclusion == "fail" {
			failByClaim[claimResult.ClaimID] = true
		}
	}
	covered := map[string]int{}
	claimLens := map[string]string{}
	for _, claim := range plan.Claims {
		claimLens[claim.ClaimID] = claim.Lens
	}
	for _, finding := range result.Findings {
		data, err := json.Marshal(finding)
		if err != nil {
			return fmt.Errorf("encode finding %s: %w", finding.FindingID, err)
		}
		if err := validator.ValidateBytes("finding.schema.json", data); err != nil {
			return fmt.Errorf("finding %s schema: %w", finding.FindingID, err)
		}
		if !failByClaim[finding.ClaimID] {
			return fmt.Errorf("finding %s references claim %s which has no fail conclusion in this result", finding.FindingID, finding.ClaimID)
		}
		covered[finding.ClaimID]++
		if finding.Lens != assignment.Lens || claimLens[finding.ClaimID] != assignment.Lens {
			return fmt.Errorf("finding %s lens %q contradicts the assignment/claim lens %q", finding.FindingID, finding.Lens, assignment.Lens)
		}
		if err := validateInvestigationReadiness(finding); err != nil {
			return err
		}
	}
	for claimID := range failByClaim {
		if covered[claimID] == 0 {
			return fmt.Errorf("fail claim %s has no Finding; every fail Claim must reference one (L3-S7 §3.5)", claimID)
		}
		if covered[claimID] > 1 {
			return fmt.Errorf("fail claim %s has %d Findings; keep exactly one immutable Finding per fail Claim", claimID, covered[claimID])
		}
	}
	return nil
}

// validateInvestigationReadiness: ordinary findings need a complete failure
// boundary and step-bound evidence; P0 findings may stop dangerous capture
// but must say so explicitly (L3-S7 §3.7, §14.1).
func validateInvestigationReadiness(finding Finding) error {
	if finding.Severity == "P0" {
		if finding.Encounter.LastGoodCheckpoint == "" && len(finding.Encounter.CaptureGaps) == 0 {
			return fmt.Errorf("finding %s is P0 with no last_good_checkpoint; safety-stop findings must record explicit capture_gaps", finding.FindingID)
		}
		return nil
	}
	if finding.Encounter.LastGoodCheckpoint == "" {
		return fmt.Errorf("finding %s misses last_good_checkpoint; S8 investigates from last-good -> wall -> first-bad, not from a bare symptom (L3-S7 §3.6)", finding.FindingID)
	}
	if finding.Encounter.TerminalState == "" {
		return fmt.Errorf("finding %s misses terminal_state", finding.FindingID)
	}
	for _, step := range finding.Encounter.Timeline {
		if len(step.EvidenceRefs) == 0 {
			return fmt.Errorf("finding %s timeline step %d has no evidence_refs; ordinary findings are not investigation-ready without step-bound evidence", finding.FindingID, step.Sequence)
		}
	}
	return nil
}

// validateProducerIndependence enforces the Builder/Reviewer separation
// edge (L3-S7 §1.4): the producer must not be an Agent that delivered S6
// product results this generation.
func validateProducerIndependence(state map[string]any, result *Result) error {
	if result.ProducerAgentID == "" {
		return fmt.Errorf("producer_agent_id is required")
	}
	evidence, _ := state["evidence"].([]any)
	generation := baselineGeneration(state)
	for _, raw := range evidence {
		item, _ := raw.(map[string]any)
		if item == nil || item["kind"] != "completion_report" {
			continue
		}
		if intField(item["baseline_generation"]) != generation {
			continue
		}
		producers, _ := item["produced_by"].([]any)
		for _, producer := range producers {
			if id, _ := producer.(string); id == result.ProducerAgentID {
				return fmt.Errorf("producer %s delivered an S6 Builder Result this generation and cannot review it (L3-S7 §1.4 role independence)", result.ProducerAgentID)
			}
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// CAS Apply helpers
// ---------------------------------------------------------------------------

// applyResultConsumption registers the result evidence and updates the
// claim/assignment projections.
func applyResultConsumption(
	state map[string]any,
	plan *Plan,
	assignment *PlanAssignment,
	result *Result,
	resultRel, resultSHA, responsibility string,
	generation, round int,
	occurredAt time.Time,
) error {
	reviewMap, ok := state["review"].(map[string]any)
	if !ok {
		return fmt.Errorf("runtime review section must be an object")
	}
	assignments, _ := reviewMap["assignments"].(map[string]any)
	row, _ := assignments[assignment.AssignmentID].(map[string]any)
	if row == nil {
		return fmt.Errorf("assignment %s is not registered in the runtime projection; dispatch it via `runtime register-workgroup` first", assignment.AssignmentID)
	}
	status, _ := row["status"].(string)
	if status == "result_submitted" || status == "consumed" {
		return fmt.Errorf("assignment %s already has a consumed ReviewResult; one Assignment submits exactly one Result (L3-S7 §3.5)", assignment.AssignmentID)
	}
	agentID, _ := row["agent_id"].(string)
	if agentID == "" {
		return fmt.Errorf("assignment %s has no dispatched Agent; dispatch it via `runtime register-workgroup` before submitting", assignment.AssignmentID)
	}
	if agentID != result.ProducerAgentID {
		return fmt.Errorf("ReviewResult producer %s does not match the dispatched Agent %s for assignment %s", result.ProducerAgentID, agentID, assignment.AssignmentID)
	}

	claims, _ := reviewMap["claims"].(map[string]any)
	failFindings := map[string][]string{}
	for _, finding := range result.Findings {
		failFindings[finding.ClaimID] = append(failFindings[finding.ClaimID], finding.FindingID)
	}
	for _, claimResult := range result.ClaimResults {
		claimRow, _ := claims[claimResult.ClaimID].(map[string]any)
		if claimRow == nil {
			return fmt.Errorf("claim %s is missing from the runtime projection", claimResult.ClaimID)
		}
		if claimRow["applicability"] == "not_applicable" {
			return fmt.Errorf("claim %s is not_applicable in the plan; N/A is a plan disposition, never a Reviewer conclusion (L3-S7 §9.3)", claimResult.ClaimID)
		}
		if claimResult.Conclusion == "pass" {
			claimRow["disposition"] = "pass"
		} else {
			claimRow["disposition"] = "finding"
		}
		claimRow["result_id"] = result.ResultID
		ids := make([]any, 0, len(failFindings[claimResult.ClaimID]))
		for _, id := range failFindings[claimResult.ClaimID] {
			ids = append(ids, id)
		}
		claimRow["finding_ids"] = ids
	}
	row["status"] = "consumed"
	row["result_ref"] = resultRel
	if result.VerificationArtifactDigest != nil && *result.VerificationArtifactDigest != "" {
		row["artifact_digest"] = *result.VerificationArtifactDigest
	}

	return appendEvidence(state, map[string]any{
		"id":                  result.ResultID,
		"kind":                "review_result",
		"path":                resultRel,
		"sha256":              resultSHA,
		"status":              "valid",
		"baseline_generation": generation,
		"review_round":        round,
		"produced_by":         []any{result.ProducerAgentID},
		"invalidated_by":      nil,
		"invalidation_rule":   nil,
		"invalidation_reason": nil,
		"responsibility_id":   responsibility,
		"scope_refs":          []any{},
	})
}

// findingArtifact pairs one Finding with its persisted evidence coordinates.
type findingArtifact struct {
	finding Finding
	rel     string
	sha     string
}

// applyFindings appends immutable finding rows and their evidence index
// entries. Finding identity is global: an id seen in any round collides.
func applyFindings(
	state map[string]any,
	result *Result,
	artifacts []findingArtifact,
	round int,
	occurredAt time.Time,
) error {
	entities, ok := state["entities"].(map[string]any)
	if !ok {
		return fmt.Errorf("runtime entities must be an object")
	}
	if len(artifacts) == 0 {
		return nil
	}
	findings, _ := entities["findings"].([]any)
	for _, artifact := range artifacts {
		for _, raw := range findings {
			row, _ := raw.(map[string]any)
			if row != nil && row["finding_id"] == artifact.finding.FindingID {
				return fmt.Errorf("finding %s already exists; Findings are immutable — add a FindingSupplement instead of reusing the id", artifact.finding.FindingID)
			}
		}
		findings = append(findings, map[string]any{
			"finding_id":       artifact.finding.FindingID,
			"path":             artifact.rel,
			"sha256":           artifact.sha,
			"claim_id":         artifact.finding.ClaimID,
			"assignment_id":    result.AssignmentID,
			"lens":             artifact.finding.Lens,
			"severity":         artifact.finding.Severity,
			"observation_mode": artifact.finding.ObservationMode,
			"original_finder":  result.ProducerAgentID,
			"review_round":     round,
			"created_at":       occurredAt.UTC().Format(time.RFC3339Nano),
		})
		if err := appendEvidence(state, map[string]any{
			"id":                  artifact.finding.FindingID,
			"kind":                "finding",
			"path":                artifact.rel,
			"sha256":              artifact.sha,
			"status":              "valid",
			"baseline_generation": baselineGeneration(state),
			"review_round":        round,
			"produced_by":         []any{result.ProducerAgentID},
			"invalidated_by":      nil,
			"invalidation_rule":   nil,
			"invalidation_reason": nil,
			"responsibility_id":   LensToResponsibility(artifact.finding.Lens),
			"scope_refs":          []any{},
		}); err != nil {
			return err
		}
	}
	entities["findings"] = findings
	return nil
}

// advanceReviewerAgent moves the reviewer working -> reported, mirroring the
// completion_reported lifecycle event (entity_lifecycles.agent). Kept local
// so review stays a leaf package (transition -> verification -> review must
// not cycle).
func advanceReviewerAgent(state map[string]any, result *Result, resultRepoPath string, occurredAt time.Time) error {
	entities, _ := state["entities"].(map[string]any)
	agents, _ := entities["agents"].([]any)
	for _, raw := range agents {
		agent, _ := raw.(map[string]any)
		if agent["id"] != result.ProducerAgentID {
			continue
		}
		currentState, _ := agent["state"].(string)
		if currentState == "reported" || currentState == "done" {
			return nil
		}
		if currentState != "working" {
			return fmt.Errorf("reviewer Agent %s is %s; a ReviewResult requires working state (canonical Agent states: spawned, reading, understanding_submitted, understanding_approved, activated, working, reported, done, blocked, stopped)", result.ProducerAgentID, currentState)
		}
		agent["state"] = "reported"
		if resultRepoPath != "" {
			agent["completion_reported_ref"] = resultRepoPath
		}
		agent["updated_at"] = occurredAt.UTC().Format(time.RFC3339Nano)
		return nil
	}
	return fmt.Errorf("Agent %s is not registered", result.ProducerAgentID)
}

// applyObservationBatch registers the sealed batch pointer and evidence.
func applyObservationBatch(
	state map[string]any,
	batchID, batchRel, batchSHA string,
	newFindings []Finding,
	round int,
	occurredAt time.Time,
) error {
	reviewMap := state["review"].(map[string]any)
	findingIDs := []any{}
	for _, row := range RoundFindings(state) {
		findingIDs = append(findingIDs, row["finding_id"])
	}
	reviewMap["observation_batch"] = map[string]any{
		"batch_id":     batchID,
		"path":         batchRel,
		"sha256":       batchSHA,
		"finding_ids":  findingIDs,
		"drain_policy": drainPolicyOf(newFindings),
		"sealed_at":    occurredAt.UTC().Format(time.RFC3339Nano),
	}
	return appendEvidence(state, map[string]any{
		"id":                  batchID,
		"kind":                "observation_batch",
		"path":                batchRel,
		"sha256":              batchSHA,
		"status":              "valid",
		"baseline_generation": baselineGeneration(state),
		"review_round":        round,
		"produced_by":         []any{"round-consumer"},
		"invalidated_by":      nil,
		"invalidation_rule":   nil,
		"invalidation_reason": nil,
		"responsibility_id":   "Orchestrator",
		"scope_refs":          []any{},
	})
}

// applyCleanRound registers the machine CleanRound evidence (L3-S7 §10.2).
func applyCleanRound(state map[string]any, cleanRel, cleanSHA string, round int, occurredAt time.Time) error {
	return appendEvidence(state, map[string]any{
		"id":                  fmt.Sprintf("clean-round-r%d", round),
		"kind":                "clean_round",
		"path":                cleanRel,
		"sha256":              cleanSHA,
		"status":              "valid",
		"baseline_generation": baselineGeneration(state),
		"review_round":        round,
		"produced_by":         []any{"round-consumer"},
		"invalidated_by":      nil,
		"invalidation_rule":   nil,
		"invalidation_reason": nil,
		"responsibility_id":   "Clean Round Evaluator",
		"scope_refs":          []any{},
	})
}

// appendEvidence registers one evidence index row, rejecting id collisions.
func appendEvidence(state map[string]any, entry map[string]any) error {
	items, ok := state["evidence"].([]any)
	if !ok {
		return fmt.Errorf("runtime evidence must be an array")
	}
	for _, raw := range items {
		item, _ := raw.(map[string]any)
		if item != nil && item["id"] == entry["id"] {
			return fmt.Errorf("evidence %s is already registered", entry["id"])
		}
	}
	state["evidence"] = append(items, entry)
	state["updated_at"] = time.Now().UTC().Format(time.RFC3339Nano)
	return nil
}

// setPlanStatus moves the ReviewPlan status and — when the status is also a
// verification phase — the lifecycle phase projection (L3-S7 §11.1). paused
// and stale are plan-only statuses: the cursor moves to the paused STATE via
// TR-010/TR-011, never to a verification phase.
func setPlanStatus(reviewMap, lifecycleMap map[string]any, status string) {
	if plan, ok := reviewMap["plan"].(map[string]any); ok {
		plan["status"] = status
	}
	switch status {
	case "running", "cannot_clean", "discovery_draining", "observation_sealed", "clean":
		if lifecycleMap != nil {
			lifecycleMap["phase"] = status
			lifecycleMap["phase_revision"] = intField(lifecycleMap["phase_revision"]) + 1
		}
	}
}

// ---------------------------------------------------------------------------
// round consumer projections
// ---------------------------------------------------------------------------

// projectDispositions returns the claim disposition map as it will look
// after this result is consumed.
func projectDispositions(state map[string]any, assignment *PlanAssignment, result *Result) map[string]ClaimDisposition {
	dispositions := Dispositions(state)
	failByClaim := map[string]bool{}
	for _, finding := range result.Findings {
		failByClaim[finding.ClaimID] = true
	}
	for _, claimResult := range result.ClaimResults {
		disp := dispositions[claimResult.ClaimID]
		disp.ResultID = result.ResultID
		if failByClaim[claimResult.ClaimID] {
			disp.Disposition = "finding"
		} else {
			disp.Disposition = "pass"
		}
		dispositions[claimResult.ClaimID] = disp
	}
	return dispositions
}

func roundCompleteWith(dispositions map[string]ClaimDisposition) bool {
	for _, disp := range dispositions {
		if disp.Applicability == "not_applicable" {
			continue
		}
		switch disp.Disposition {
		case "pass", "finding", "blocked":
		default:
			return false
		}
	}
	return true
}

func drainPolicyOf(newFindings []Finding) string {
	for _, finding := range newFindings {
		if finding.Severity == "P0" {
			return "immediate_stop"
		}
	}
	return "complete_required_claims"
}

func findingIDs(findings []Finding) []string {
	ids := make([]string, 0, len(findings))
	for _, finding := range findings {
		ids = append(ids, finding.FindingID)
	}
	return ids
}

// state_view exposes the pre-transaction state to the batch builder.
type state_view struct {
	state map[string]any
}

// buildObservationBatch assembles the sealed handoff document (L3-S7 §3.7).
func buildObservationBatch(
	view state_view,
	plan *Plan, ptr *PlanPointer,
	projected map[string]ClaimDisposition,
	result *Result,
	complete bool,
	occurredAt time.Time,
) (map[string]any, error) {
	state := view.state
	round := currentReviewRound(state)
	batchFindingIDs := []string{}
	readiness := []any{}
	routes := []any{}
	for _, row := range RoundFindings(state) {
		batchFindingIDs = append(batchFindingIDs, stringField(row["finding_id"]))
		readiness = append(readiness, map[string]any{
			"finding_id":   row["finding_id"],
			"status":       "ready",
			"capture_gaps": []any{},
		})
		routes = append(routes, map[string]any{
			"finding_id":    row["finding_id"],
			"agent_id":      row["original_finder"],
			"assignment_id": row["assignment_id"],
		})
	}
	for _, finding := range result.Findings {
		batchFindingIDs = append(batchFindingIDs, finding.FindingID)
		status := "ready"
		gaps := finding.Encounter.CaptureGaps
		if finding.Severity == "P0" && len(gaps) > 0 {
			status = "ready_with_safety_gaps"
		}
		gapValues := make([]any, 0, len(gaps))
		for _, gap := range gaps {
			gapValues = append(gapValues, gap)
		}
		readiness = append(readiness, map[string]any{
			"finding_id":   finding.FindingID,
			"status":       status,
			"capture_gaps": gapValues,
		})
		routes = append(routes, map[string]any{
			"finding_id":    finding.FindingID,
			"agent_id":      result.ProducerAgentID,
			"assignment_id": result.AssignmentID,
		})
	}
	sort.Strings(batchFindingIDs)

	summary := map[string]any{"pass": 0, "finding": 0, "not_applicable": 0, "blocked": 0}
	total := 0
	for _, disp := range projected {
		if disp.Applicability == "not_applicable" {
			summary["not_applicable"] = summary["not_applicable"].(int) + 1
			continue
		}
		total++
		switch disp.Disposition {
		case "pass":
			summary["pass"] = summary["pass"].(int) + 1
		case "finding":
			summary["finding"] = summary["finding"].(int) + 1
		case "blocked":
			summary["blocked"] = summary["blocked"].(int) + 1
		}
	}
	summary["total_required"] = total
	summary["plan_revision"] = ptr.Revision

	drainPolicy := drainPolicyOf(result.Findings)
	unobserved := []string{}
	stopReason := ""
	if !complete {
		// immediate_stop only: ordinary batches seal with an empty unobserved
		// set (L3-S7 §3.7 seal condition).
		for _, claimID := range UndispositionedRequired(state) {
			if disp, ok := projected[claimID]; ok {
				switch disp.Disposition {
				case "pass", "finding", "blocked":
					continue
				}
			}
			unobserved = append(unobserved, claimID)
		}
		sort.Strings(unobserved)
		stopReason = "P0/security/data-destructive finding: stop-the-line with explicit safety gaps"
	}

	runtimeID, _ := state["runtime_id"].(string)
	return map[string]any{
		"schema_version":        "1.0.0",
		"observation_batch_id":  fmt.Sprintf("observation-batch-r%d", round),
		"conclusion":            "sealed",
		"evidence_id":           fmt.Sprintf("observation-batch-r%d", round),
		"kind":                  "observation_batch",
		"runtime_id":            runtimeID,
		"producer_agent_id":     "round-consumer",
		"producer_responsibility": "Orchestrator",
		"review_plan_id":        plan.ReviewPlanID,
		"review_round":          round,
		"baseline_generation":   baselineGeneration(state),
		"subject_digest":        SubjectDigest(plan),
		"finding_ids":           batchFindingIDs,
		"drained_assignment_ids": drainedAssignments(state),
		"drain_policy":          drainPolicy,
		"claim_coverage_summary": summary,
		"cancelled_or_non_gating_assignment_ids": []any{},
		"unobserved_claim_ids": unobserved,
		"original_finder_routes": routes,
		"investigation_readiness": readiness,
		"severity_summary":      severitySummary(state, result.Findings),
		"stop_reason":           stopReason,
		"sealed_at":             occurredAt.UTC().Format(time.RFC3339Nano),
		"sealed_by":             "round-consumer",
		"revision":              1,
	}, nil
}

func drainedAssignments(state map[string]any) []any {
	out := []any{}
	reviewMap, _ := state["review"].(map[string]any)
	assignments, _ := reviewMap["assignments"].(map[string]any)
	ids := make([]string, 0, len(assignments))
	for id := range assignments {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		row, _ := assignments[id].(map[string]any)
		if row != nil && row["status"] == "consumed" {
			out = append(out, id)
		}
	}
	return out
}

func severitySummary(state map[string]any, newFindings []Finding) string {
	counts := map[string]int{}
	for _, row := range RoundFindings(state) {
		counts[stringField(row["severity"])]++
	}
	for _, finding := range newFindings {
		counts[finding.Severity]++
	}
	parts := []string{}
	for _, severity := range []string{"P0", "P1", "P2", "P3"} {
		if counts[severity] > 0 {
			parts = append(parts, fmt.Sprintf("%s=%d", severity, counts[severity]))
		}
	}
	return strings.Join(parts, ", ")
}

// buildCleanRoundSnapshot assembles the immutable CleanRound document
// (L3-S7 §10.2): minimal references and digests for recomputation, not a
// copy of the raw evidence.
func buildCleanRoundSnapshot(
	state map[string]any,
	plan *Plan,
	resultEnvelope map[string]any,
	resultSHA, resultRel string,
	occurredAt time.Time,
) map[string]any {
	round := currentReviewRound(state)
	resultRefs := []any{}
	seen := map[string]bool{}
	evidence, _ := state["evidence"].([]any)
	byID := map[string]map[string]any{}
	for _, raw := range evidence {
		item, _ := raw.(map[string]any)
		if item != nil {
			byID[stringField(item["id"])] = item
		}
	}
	for _, disp := range Dispositions(state) {
		if disp.ResultID == "" || seen[disp.ResultID] {
			continue
		}
		seen[disp.ResultID] = true
		if item, ok := byID[disp.ResultID]; ok {
			resultRefs = append(resultRefs, map[string]any{
				"result_id": disp.ResultID,
				"path":      item["path"],
				"sha256":    item["sha256"],
			})
		}
	}
	resultRefs = append(resultRefs, map[string]any{
		"result_id": resultEnvelope["evidence_id"],
		"path":      resultRel,
		"sha256":    resultSHA,
	})
	runtimeID, _ := state["runtime_id"].(string)
	return map[string]any{
		"schema_version":      "1.0.0",
		"clean_round_id":      fmt.Sprintf("clean-round-r%d", round),
		// envelope identity fields: the gate verifies the persisted document
		// against the evidence index row (same contract as the batch).
		"evidence_id":           fmt.Sprintf("clean-round-r%d", round),
		"kind":                  "clean_round",
		"runtime_id":            runtimeID,
		"producer_agent_id":     "round-consumer",
		"producer_responsibility": "Clean Round Evaluator",
		"review_plan_id":      plan.ReviewPlanID,
		"review_round":        round,
		"baseline_generation": baselineGeneration(state),
		"subject_digest":      SubjectDigest(plan),
		"result_refs":         resultRefs,
		"conclusion":          "pass",
		"evaluated_at":        occurredAt.UTC().Format(time.RFC3339Nano),
		"evaluated_by":        "round-consumer",
		"evaluator_version":   "review.SubmitResult/1",
	}
}

// capturePauseCheckpoint mirrors transition.CapturePauseCheckpoint's field
// shape (engine.go documentFingerprints + pauseRequiredAction) so the S7
// verdict transaction creates the one authoritative checkpoint without
// importing transition (transition -> verification -> review must stay
// acyclic).
func capturePauseCheckpoint(state map[string]any, reason string, occurredAt time.Time) error {
	if existing, ok := state["pause"]; ok && existing != nil {
		return fmt.Errorf("pause checkpoint already exists; would overwrite")
	}
	lifecycle, ok := state["lifecycle"].(map[string]any)
	if !ok {
		return fmt.Errorf("lifecycle missing")
	}
	documents := []any{}
	for _, raw := range stateDocuments(state) {
		doc, _ := raw.(map[string]any)
		if doc == nil || doc["path"] == nil {
			continue
		}
		documents = append(documents, map[string]any{
			"path":    doc["path"],
			"version": doc["version"],
			"sha256":  doc["sha256"],
		})
	}
	state["pause"] = map[string]any{
		"from_state":            lifecycle["state"],
		"from_phase":            lifecycle["phase"],
		"phase_revision":        intField(lifecycle["phase_revision"]),
		"baseline_generation":   baselineGeneration(state),
		"review_round":          currentReviewRound(state),
		"reason":                reason,
		"required_human_action": "review the blocking finding or REQ change, then resume or re-lock the REQ",
		"document_fingerprints": documents,
		"paused_at":             occurredAt.UTC().Format(time.RFC3339Nano),
	}
	return nil
}

func stateDocuments(state map[string]any) []any {
	documents, _ := state["documents"].([]any)
	return documents
}

// ---------------------------------------------------------------------------
// artifact IO
// ---------------------------------------------------------------------------

func marshalArtifact(value any) ([]byte, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func writeArtifact(root, rel string, data []byte) error {
	abs := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return fmt.Errorf("create evidence dir: %w", err)
	}
	if err := os.WriteFile(abs, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", rel, err)
	}
	return nil
}

func repositoryPath(root, path string) string {
	if relative, err := filepath.Rel(root, path); err == nil && relative != ".." && !filepath.IsAbs(relative) {
		return filepath.ToSlash(relative)
	}
	return filepath.ToSlash(path)
}
