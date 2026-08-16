package scenario

import (
	"encoding/json"
	"fmt"
	"regexp"
)

// CrossMatrix is the convergence-1 carrier: the fact×FR×story completeness
// hunt made fill-in (L3-S2 v4.0.1). Each entry either names the branch that
// covers the cell or records why no branch exists — the field is the
// question (D4), the reference check is the machine's part.
type CrossMatrix struct {
	Module  string             `json:"module"`
	Entries []CrossMatrixEntry `json:"entries"`
}

// CrossMatrixEntry is one hunted cell. Exactly one of Branch /
// NoBranchReason must be set (schema-enforced); the reference fields are
// machine-checked against the module package.
type CrossMatrixEntry struct {
	Fact           string `json:"fact"`
	ReqRef         string `json:"req_ref"`
	Story          string `json:"story"`
	Branch         string `json:"branch,omitempty"`
	NoBranchReason string `json:"no_branch_reason,omitempty"`
}

// crossMatrixReqRefPattern accepts REQ-level ("REQ-040") and FR-level
// ("REQ-040/FR-003") references — FR-level is what the AC bridge resolves.
var crossMatrixReqRefPattern = regexp.MustCompile(`^REQ-[A-Z0-9]+(-[A-Z0-9]+)*(?:/FR-[A-Z0-9]+(-[A-Z0-9]+)*)?$`)

func decodeCrossMatrix(data []byte) (CrossMatrix, error) {
	var matrix CrossMatrix
	if err := json.Unmarshal(data, &matrix); err != nil {
		return CrossMatrix{}, fmt.Errorf("decode cross-matrix.json: %w", err)
	}
	return matrix, nil
}

// validateCrossMatrix checks that every hunted cell points at real package
// facts, stories, and branches, and that a reason is recorded exactly when
// no branch covers the cell.
func validateCrossMatrix(source sourcePackage) error {
	matrix := source.crossMatrix
	if matrix.Module != source.model.Module {
		return fmt.Errorf("cross-matrix.json module %q does not match scenario-model module %q", matrix.Module, source.model.Module)
	}
	if len(matrix.Entries) == 0 {
		return fmt.Errorf("cross-matrix.json must enumerate at least one hunted cell — an empty hunt is no hunt")
	}
	factIDs := map[string]bool{}
	for _, fact := range source.model.Facts {
		factIDs[fact.ID] = true
	}
	branchIDs := map[string]bool{}
	for _, rule := range source.model.Rules {
		for _, branch := range rule.Branches {
			branchIDs[branch.ID] = true
		}
	}
	seen := map[string]bool{}
	for i, entry := range matrix.Entries {
		cell := fmt.Sprintf("entry %d (fact=%s req_ref=%s story=%s)", i, entry.Fact, entry.ReqRef, entry.Story)
		key := entry.Fact + "|" + entry.ReqRef + "|" + entry.Story
		if seen[key] {
			return fmt.Errorf("cross-matrix %s duplicates an earlier cell", cell)
		}
		seen[key] = true
		if !factIDs[entry.Fact] {
			return fmt.Errorf("cross-matrix %s references unknown fact %q", cell, entry.Fact)
		}
		if !crossMatrixReqRefPattern.MatchString(entry.ReqRef) {
			return fmt.Errorf("cross-matrix %s req_ref %q must be REQ-<id> or REQ-<id>/FR-<id>", cell, entry.ReqRef)
		}
		if !storyRefPattern.MatchString(entry.Story) || !markdownHeadingContainsID(source.stories, entry.Story) {
			return fmt.Errorf("cross-matrix %s references story %q missing from stories.md", cell, entry.Story)
		}
		switch {
		case entry.Branch == "" && entry.NoBranchReason == "":
			return fmt.Errorf("cross-matrix %s names neither a branch nor a no-branch reason — silence is not N/A", cell)
		case entry.Branch != "" && entry.NoBranchReason != "":
			return fmt.Errorf("cross-matrix %s sets both branch and no-branch reason", cell)
		case entry.Branch != "" && !branchIDs[entry.Branch]:
			return fmt.Errorf("cross-matrix %s references unknown branch %q", cell, entry.Branch)
		}
	}
	return nil
}
