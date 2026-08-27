package repair

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

func createS7ReviewPlanSeed(root string, round, baselineGeneration int, impact ChangeImpact, contractRef ContractRef, taskIDs []string, impactPath string) (ArtifactRef, error) {
	contractBytes, err := readArtifact(root, ArtifactRef{Path: contractRef.Path, SHA256: contractRef.SHA256}, "repair-contract.schema.json")
	if err != nil {
		return ArtifactRef{}, err
	}
	var contract map[string]any
	if err := json.Unmarshal(contractBytes, &contract); err != nil {
		return ArtifactRef{}, fmt.Errorf("decode RepairContract for S7 seed: %w", err)
	}
	planID := fmt.Sprintf("review-plan-s9-round-%d", round)
	frozen := make([]any, 0, len(impact.ChangedArtifacts))
	coverage := make([]any, 0, len(impact.ChangedArtifacts))
	for _, artifact := range impact.ChangedArtifacts {
		frozen = append(frozen, map[string]any{"path": artifact.Path, "sha256": artifact.SHA256, "kind": "post_repair_changed_artifact"})
		coverage = append(coverage, map[string]any{"id": "coverage-" + safeSeedID(artifact.Path), "kind": "post_repair_change", "source_ref": artifact.Path, "target": artifact.Path, "lens": "qa"})
	}
	changedPaths := make([]string, 0, len(impact.ChangedArtifacts))
	for _, artifact := range impact.ChangedArtifacts {
		changedPaths = append(changedPaths, artifact.Path)
	}
	qaSources := append([]string{impactPath}, changedPaths...)
	claims := []any{
		map[string]any{"claim_id": "claim-s9-delivery", "lens": "delivery", "target": stringField(contract["repair_contract_id"]), "assertion": "the approved RepairContract is fully represented by the post-repair review", "oracle": "every Contract unit and changed artifact has a traceable review source", "method": "traceability", "applicability": "required", "source_refs": append([]string{contractRef.Path}, taskIDs...)},
		map[string]any{"claim_id": "claim-s9-qa", "lens": "qa", "target": "post-repair changed surface", "assertion": "the repaired invariant and scope boundary hold", "oracle": "the affected code passes the Contract assertions without scope drift", "method": "static review and targeted checks", "applicability": "required", "source_refs": qaSources},
		map[string]any{"claim_id": "claim-s9-e2e-na", "lens": "e2e", "target": "post-repair user surface", "assertion": "E2E applicability is decided by the complete S7 plan", "oracle": "the S7 Planner either creates behavior Claims or records an evidence-backed N/A decision", "method": "impact analysis", "applicability": "not_applicable", "na_rationale": "S9 creates a seed only; S7 must decide E2E applicability from the affected surface", "source_refs": []string{impactPath}},
	}
	assignments := []any{
		map[string]any{"assignment_id": "assignment-s9-delivery", "lens": "delivery", "claim_ids": []string{"claim-s9-delivery"}, "non_overlap_boundary": "owns Contract traceability and changed-surface completeness", "execution_wave": "static"},
		map[string]any{"assignment_id": "assignment-s9-qa", "lens": "qa", "claim_ids": []string{"claim-s9-qa"}, "non_overlap_boundary": "owns invariant, maintainability and scope verification", "execution_wave": "static"},
	}
	seed := map[string]any{
		"schema_version": "1.0.0", "review_plan_id": planID, "review_round": round, "baseline_generation": baselineGeneration,
		"frozen_subjects": frozen, "coverage_inventory": coverage, "change_impact": map[string]any{"summary": "S9 post-repair seed; Planner must refine the exact Claim set", "source_refs": []string{impactPath}},
		"claims": claims, "assignments": assignments, "e2e_coverage_state": "not_applicable", "verification_artifact_workspace": nil,
		"dispatch_capacity_policy": "coverage_complete", "created_by": "s9-handoff", "created_at": nowOr(time.Time{}),
	}
	seedPath := artifactRoot + "/s7-seeds/" + planID + ".json"
	return writeImmutable(root, seedPath, "review-plan.schema.json", seed)
}

func seedBaselineDigest(artifacts []ArtifactRef) string {
	lines := make([]string, 0, len(artifacts))
	for _, artifact := range artifacts {
		lines = append(lines, normalizePath(artifact.Path)+":"+artifact.SHA256)
	}
	sort.Strings(lines)
	return sha256Bytes([]byte(strings.Join(lines, "\n")))
}

func safeSeedID(path string) string {
	return strings.NewReplacer("/", "-", "\\", "-", ".", "-").Replace(normalizePath(path))
}
