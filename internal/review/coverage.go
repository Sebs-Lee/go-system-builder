package review

import (
	"fmt"
	"sort"
	"strings"
)

// CoverageItem is a frozen planning fact, not a new Runtime state. It names an
// object S7 must account for before a ReviewPlan can be registered.
type CoverageItem struct {
	ID        string `json:"id"`
	Kind      string `json:"kind"`
	SourceRef string `json:"source_ref"`
	Target    string `json:"target"`
	Lens      string `json:"lens"`
}

// BuildCoverageInventory projects state-only S6 facts into a deterministic
// planner input. Production callers should use BuildCoverageInventoryForRoot
// so the immutable completion envelope, rather than a copied index field, is
// the authority for changed paths.
func BuildCoverageInventory(state map[string]any) []CoverageItem {
	return legacyBuildCoverageInventory(state)
}

// BuildCoverageInventoryForRoot reads the current-generation completion
// envelope through its indexed path+sha256 when root is available. The
// resulting changed-surface set is the denominator that S7 must cover.
func BuildCoverageInventoryForRoot(root string, state map[string]any) []CoverageItem {
	seen := map[string]bool{}
	var items []CoverageItem
	add := func(item CoverageItem) {
		if item.SourceRef == "" || seen[item.SourceRef] {
			return
		}
		seen[item.SourceRef] = true
		items = append(items, item)
	}
	projection := buildS7BaselineProjection(root, state)
	for _, ref := range projection.ChangedPaths {
		add(CoverageItem{
			ID:        "surface:" + ref,
			Kind:      "changed_surface",
			SourceRef: ref,
			Target:    ref,
			Lens:      "qa",
		})
	}
	return sortedCoverageItems(items)
}

func sortedCoverageItems(items []CoverageItem) []CoverageItem {
	sort.Slice(items, func(i, j int) bool { return items[i].SourceRef < items[j].SourceRef })
	return items
}

// legacyBuildCoverageInventory is retained only for tests and callers that
// intentionally pass a state snapshot without a repository root. It reads
// scope_refs as the compatibility projection produced by S6.
func legacyBuildCoverageInventory(state map[string]any) []CoverageItem {
	generation := baselineGeneration(state)
	seen := map[string]bool{}
	var items []CoverageItem
	add := func(item CoverageItem) {
		if item.SourceRef == "" || seen[item.SourceRef] {
			return
		}
		seen[item.SourceRef] = true
		items = append(items, item)
	}
	for _, raw := range evidenceEntries(state) {
		entry, _ := raw.(map[string]any)
		if entry == nil || entry["kind"] != "completion_report" || intField(entry["baseline_generation"]) != generation {
			continue
		}
		for _, ref := range normalizeBaselinePaths(stringSliceValue(entry["scope_refs"])) {
			add(CoverageItem{ID: "surface:" + ref, Kind: "changed_surface", SourceRef: ref, Target: ref, Lens: "qa"})
		}
	}
	return sortedCoverageItems(items)
}

func validateCoverageInventory(root string, state map[string]any, plan *Plan) error {
	if err := validateS7BaselineProjection(root, state); err != nil {
		return err
	}
	required := BuildCoverageInventoryForRoot(root, state)
	if len(required) == 0 {
		return nil
	}
	if len(plan.CoverageInventory) == 0 {
		missing := make([]string, 0, len(required))
		for _, item := range required {
			missing = append(missing, item.SourceRef+" has no Coverage Inventory item")
		}
		return s7GateError(
			"S7_PLAN_SURFACE_COVERAGE",
			"ReviewPlan does not declare the complete S6 changed-surface inventory",
			missing,
			[]string{"copy the missing changed surfaces into coverage_inventory and add each source_ref to at least one Claim"},
			"runtime review-plan --file plan.json --expected-revision <N>",
		)
	}
	bySource := make(map[string]CoverageItem, len(plan.CoverageInventory))
	for _, item := range plan.CoverageInventory {
		if err := validateCoverageItem(item); err != nil {
			return s7GateError(
				"S7_PLAN_SURFACE_COVERAGE",
				"Coverage Inventory contains an invalid item",
				[]string{err.Error()},
				[]string{"complete id, kind, source_ref, target and lens for every coverage_inventory item"},
				"runtime review-plan --file plan.json --expected-revision <N>",
			)
		}
		if item.SourceRef == "" {
			return s7GateError(
				"S7_PLAN_SURFACE_COVERAGE",
				"Coverage Inventory contains an item without source_ref",
				[]string{"coverage_inventory item " + item.ID},
				[]string{"set source_ref to the exact changed path or authoritative surface id"},
				"runtime review-plan --file plan.json --expected-revision <N>",
			)
		}
		bySource[item.SourceRef] = item
	}
	claimed := map[string]bool{}
	for _, claim := range plan.Claims {
		for _, sourceRef := range claim.SourceRefs {
			claimed[sourceRef] = true
		}
	}
	frozen := make(map[string]bool, len(plan.FrozenSubjects))
	for _, subject := range plan.FrozenSubjects {
		frozen[normalizeSurface(subject.Path)] = true
	}
	var missing []string
	for _, item := range required {
		if _, ok := bySource[item.SourceRef]; !ok {
			missing = append(missing, item.SourceRef+" is absent from coverage_inventory")
			continue
		}
		if !claimed[item.SourceRef] {
			missing = append(missing, item.SourceRef+" has no Claim source_ref")
		}
		if item.Kind == "changed_surface" && !frozen[normalizeSurface(item.SourceRef)] {
			missing = append(missing, item.SourceRef+" is a changed surface but is absent from frozen_subjects")
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return s7GateError(
			"S7_PLAN_SURFACE_COVERAGE",
			"ReviewPlan changed-surface coverage is incomplete",
			missing,
			[]string{"add each changed surface to frozen_subjects with its current SHA-256, then add the same source_ref to a focused QA or E2E Claim; keep target/oracle specific to that surface"},
			"runtime review-plan --file plan.json --expected-revision <N>",
		)
	}
	return nil
}

func validateCoverageItem(item CoverageItem) error {
	if strings.TrimSpace(item.ID) == "" || strings.TrimSpace(item.Kind) == "" || strings.TrimSpace(item.SourceRef) == "" || strings.TrimSpace(item.Target) == "" || strings.TrimSpace(item.Lens) == "" {
		return fmt.Errorf("coverage_inventory item must include id, kind, source_ref, target and lens")
	}
	return nil
}
