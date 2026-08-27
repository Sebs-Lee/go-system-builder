package repair

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

func CreateChangeImpact(root string, request ChangeImpactRequest) (ChangeImpact, ArtifactRef, error) {
	if !strings.HasPrefix(request.ImpactID, "impact-") {
		return ChangeImpact{}, ArtifactRef{}, fmt.Errorf("impact_id must carry the impact- prefix so Runtime can bind it (got %q)", request.ImpactID)
	}
	if strings.TrimSpace(request.RuntimeID) == "" || strings.TrimSpace(request.ReqID) == "" || strings.TrimSpace(request.AnalyzedBy) == "" {
		return ChangeImpact{}, ArtifactRef{}, errors.New("runtime_id, req_id, and analyzed_by are required")
	}
	if request.BaselineGeneration < 1 || len(request.SourceBugIDs) == 0 && len(request.SourceCaseIDs) == 0 || len(request.ChangeTypes) == 0 || len(request.Decisions) == 0 || len(request.ChangedArtifacts) == 0 {
		return ChangeImpact{}, ArtifactRef{}, errors.New("ChangeImpact requires positive baseline, a source Bug or Case id, change types, changed artifacts, and decisions")
	}
	changed := append([]ArtifactRef(nil), request.ChangedArtifacts...)
	for i := range changed {
		if changed[i].ID == "" {
			changed[i].ID = "changed-" + strings.NewReplacer("/", "-", "\\", "-").Replace(changed[i].Path)
		}
		if changed[i].Path == "" || changed[i].SHA256 == "" {
			return ChangeImpact{}, ArtifactRef{}, fmt.Errorf("changed artifact %d requires path and sha256", i)
		}
	}
	for _, artifact := range changed {
		covered := false
		for _, decision := range request.Decisions {
			for _, scope := range decision.Scope {
				if pathMatches(artifact.Path, scope) || pathMatches(scope, artifact.Path) {
					covered = true
					break
				}
			}
			if covered {
				break
			}
		}
		if !covered {
			return ChangeImpact{}, ArtifactRef{}, fmt.Errorf("ChangeImpact decision coverage missing for changed artifact %q; add a decision scope and rationale", artifact.Path)
		}
	}
	impact := ChangeImpact{
		SchemaVersion: "1.0.0", RecordType: "change_impact", ImpactID: request.ImpactID,
		RuntimeID: request.RuntimeID, ReqID: request.ReqID, BaselineGeneration: request.BaselineGeneration,
		SourceBugIDs: sortedStrings(request.SourceBugIDs), SourceCaseIDs: sortedStrings(request.SourceCaseIDs), ChangeTypes: sortedStrings(request.ChangeTypes), ChangedArtifacts: changed,
		Decisions: append([]ImpactDecision(nil), request.Decisions...), EscalationLevel: request.EscalationLevel,
		InvalidatedEvidenceIDs: sortedStrings(request.InvalidatedEvidenceIDs), SupersededEvidenceIDs: sortedStrings(request.SupersededEvidenceIDs),
		RetainedEvidenceIDs: sortedStrings(request.RetainedEvidenceIDs), RequiredReverificationIDs: sortedStrings(request.RequiredReverificationIDs),
		AnalyzedBy: request.AnalyzedBy, AnalyzedAt: nowOr(request.AnalyzedAt),
	}
	if impact.EscalationLevel == "" {
		impact.EscalationLevel = "assignment"
	}
	ref, err := writeImmutable(root, filepath.Join(artifactRoot, "change-impact", request.ImpactID+".json"), "review-evidence.schema.json", impact)
	return impact, ref, err
}

func ValidateChangeImpact(root string, ref ArtifactRef) (ChangeImpact, error) {
	var impact ChangeImpact
	if err := decodeArtifact(root, ref, "review-evidence.schema.json", &impact); err != nil {
		return ChangeImpact{}, err
	}
	if impact.RecordType != "change_impact" {
		return ChangeImpact{}, fmt.Errorf("artifact %s is %q, not change_impact", ref.Path, impact.RecordType)
	}
	return impact, nil
}

func CreateTargetedReverification(root string, request TargetedReverificationRequest) (TargetedReverification, ArtifactRef, error) {
	if !strings.HasPrefix(request.ReverificationID, "reverify-") {
		return TargetedReverification{}, ArtifactRef{}, fmt.Errorf("request.ReverificationID must carry the reverify- prefix so Runtime can bind it (got %q)", request.ReverificationID)
	}
	if strings.TrimSpace(request.RuntimeID) == "" || strings.TrimSpace(request.BugID) == "" && strings.TrimSpace(request.CaseID) == "" || strings.TrimSpace(request.OriginalAssignmentID) == "" || strings.TrimSpace(request.PerformingAssignmentID) == "" || strings.TrimSpace(request.ImpactID) == "" {
		return TargetedReverification{}, ArtifactRef{}, errors.New("runtime_id, a bug_id or case_id, both assignment ids, and the impact id are required")
	}
	if request.OriginalAssignmentID == request.PerformingAssignmentID {
		return TargetedReverification{}, ArtifactRef{}, errors.New("targeted reverification requires an independent verifier: performing_assignment_id must differ from original_assignment_id")
	}
	if len(request.AssertionResults) == 0 {
		return TargetedReverification{}, ArtifactRef{}, errors.New("targeted reverification requires assertion results")
	}
	if request.ContinuityReason == "" {
		request.ContinuityReason = "independent verification after repair"
	}
	reverification := TargetedReverification{
		SchemaVersion: "1.0.0", RecordType: "targeted_reverification", ReverificationID: request.ReverificationID,
		RuntimeID: request.RuntimeID, BugID: request.BugID, CaseID: request.CaseID, BaselineGeneration: request.BaselineGeneration,
		OriginalAssignmentID: request.OriginalAssignmentID, PerformingAssignmentID: request.PerformingAssignmentID,
		ContinuityReason: request.ContinuityReason, ImpactID: request.ImpactID, AssertionResults: append([]AssertionResult(nil), request.AssertionResults...),
		ScopeCompliance: request.ScopeCompliance, Result: request.Result, FailureClass: request.FailureClass, PerformedAt: nowOr(request.PerformedAt),
	}
	if reverification.ScopeCompliance == "" {
		reverification.ScopeCompliance = "fail"
	}
	if reverification.Result == "" {
		reverification.Result = "blocked"
	}
	if reverification.Result != "pass" && reverification.FailureClass == "" {
		switch reverification.Result {
		case "blocked":
			reverification.FailureClass = "blocked"
		case "scope_changed":
			reverification.FailureClass = "scope_changed"
		default:
			reverification.FailureClass = "fail_same_cause"
		}
	}
	ref, err := writeImmutable(root, filepath.Join(artifactRoot, "reverification", request.ReverificationID+".json"), "review-evidence.schema.json", reverification)
	return reverification, ref, err
}

func ValidateTargetedReverification(root string, ref ArtifactRef) (TargetedReverification, error) {
	var value TargetedReverification
	if err := decodeArtifact(root, ref, "review-evidence.schema.json", &value); err != nil {
		return TargetedReverification{}, err
	}
	if value.RecordType != "targeted_reverification" {
		return TargetedReverification{}, fmt.Errorf("artifact %s is %q, not targeted_reverification", ref.Path, value.RecordType)
	}
	if value.OriginalAssignmentID == value.PerformingAssignmentID {
		return TargetedReverification{}, errors.New("targeted reverification is not independent")
	}
	seen := map[string]bool{}
	for _, assertion := range value.AssertionResults {
		if strings.TrimSpace(assertion.AssertionID) == "" {
			return TargetedReverification{}, errors.New("targeted reverification contains an assertion without assertion_id")
		}
		if seen[assertion.AssertionID] {
			return TargetedReverification{}, fmt.Errorf("targeted reverification contains duplicate assertion_id %q", assertion.AssertionID)
		}
		seen[assertion.AssertionID] = true
		if value.Result == "pass" && assertion.Result != "pass" {
			return TargetedReverification{}, fmt.Errorf("pass targeted reverification contains non-pass assertion %q", assertion.AssertionID)
		}
	}
	return value, nil
}

func sortedStrings(values []string) []string {
	set := map[string]struct{}{}
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			set[value] = struct{}{}
		}
	}
	result := make([]string, 0, len(set))
	for value := range set {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
