package review

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// taskDoc is one current-generation TASK document fact.
type taskDoc struct {
	id     string
	path   string
	sha256 string
}

// DraftPlan scaffolds a ReviewPlan from the current runtime facts
// (L3-S7 §4.2 claim generation order, planner assist): every
// current-generation TASK yields a delivery traceability claim; the union of
// Builder-changed paths yields a QA static claim per focus cluster; the E2E
// coverage state derives from the bound REQ's ui_impact. The draft is a
// scaffold, not a plan — oracle/method fields carry TODO markers the Planner
// must replace; registration validates the result normally.
func DraftPlan(state map[string]any, round int) (*Plan, []string) {
	generation := baselineGeneration(state)
	var notes []string

	var tasks []taskDoc
	for _, raw := range stateDocuments(state) {
		doc, _ := raw.(map[string]any)
		if doc == nil || doc["kind"] != "task" {
			continue
		}
		if intField(doc["generation"]) != generation {
			continue
		}
		id, _ := doc["id"].(string)
		path, _ := doc["path"].(string)
		sha, _ := doc["sha256"].(string)
		if id == "" || path == "" {
			continue
		}
		tasks = append(tasks, taskDoc{id: id, path: path, sha256: sha})
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].id < tasks[j].id })

	frozen := []FrozenSubject{}
	for _, task := range tasks {
		frozen = append(frozen, FrozenSubject{Path: task.path, SHA256: task.sha256, Kind: "task"})
	}

	// Builder-changed paths from completion envelopes.
	changed := map[string]bool{}
	for _, raw := range evidenceEntries(state) {
		entry, _ := raw.(map[string]any)
		if entry == nil || entry["kind"] != "completion_report" {
			continue
		}
		if intField(entry["baseline_generation"]) != generation {
			continue
		}
		// changed paths are carried in the envelope body; the draft reads the
		// index scope_refs as the cheaper projection.
		if refs, ok := entry["scope_refs"].([]any); ok {
			for _, ref := range refs {
				if s, _ := ref.(string); s != "" {
					changed[s] = true
				}
			}
		}
	}

	claims := []Claim{}
	assignments := []PlanAssignment{}
	claimSeq := 0
	nextClaim := func(lens, focus string) string {
		claimSeq++
		return fmt.Sprintf("claim-%s-%s-%d", lens, focus, claimSeq)
	}

	// Delivery traceability: one claim per TASK (L3-S7 §4.2 step 2).
	var dvClaimIDs []string
	for _, task := range tasks {
		id := nextClaim("dv", "traceability")
		claims = append(claims, Claim{
			ClaimID: id, Lens: "delivery", Target: task.path,
			Assertion: task.id + " is delivered as locked: every acceptance obligation lands in the implementation",
			Oracle:    "TODO(planner): the observable fact proving delivery for " + task.id,
			Method:    "requirement traceability review",
			Applicability: "required",
			SourceRefs:    []string{task.id, task.path},
			FocusKey:      "requirement-traceability",
		})
		dvClaimIDs = append(dvClaimIDs, id)
	}
	if len(dvClaimIDs) > 0 {
		assignments = append(assignments, PlanAssignment{
			AssignmentID: "assignment-dv-traceability", Lens: "delivery", ClaimIDs: dvClaimIDs,
			NonOverlapBoundary: "owns REQ/TASK traceability; QA owns implementation quality",
			ExecutionWave:      "static",
		})
	}

	// QA static baseline claims (L3-S7 §4.2 step 5): the standard focus set
	// over the changed surface.
	changedList := make([]string, 0, len(changed))
	for path := range changed {
		changedList = append(changedList, path)
	}
	sort.Strings(changedList)
	qaSurface := "the current change surface"
	if len(changedList) > 0 {
		qaSurface = strings.Join(changedList, ", ")
	}
	var qaClaimIDs []string
	for _, focus := range []struct{ key, assertion string }{
		{"logic-state-error", "normal/edge/error paths are self-consistent; state transitions and error ownership are complete"},
		{"maintainability", "naming, abstraction level and cognitive complexity stay within the project's idiom"},
		{"test-oracle", "behavior (not implementation detail) is asserted; negative/boundary paths carry valid oracles"},
	} {
		id := nextClaim("qa", focus.key)
		claims = append(claims, Claim{
			ClaimID: id, Lens: "qa", Target: qaSurface,
			Assertion: focus.assertion,
			Oracle:    "TODO(planner): the observable fact proving " + focus.key,
			Method:    "static code review",
			Applicability: "required",
			SourceRefs:    taskIDs(tasks),
			FocusKey:      focus.key,
		})
		qaClaimIDs = append(qaClaimIDs, id)
	}
	assignments = append(assignments, PlanAssignment{
		AssignmentID: "assignment-qa-static", Lens: "qa", ClaimIDs: qaClaimIDs,
		NonOverlapBoundary: "owns static quality; does not duplicate DV traceability",
		ExecutionWave:      "static",
	})

	// E2E coverage state from the bound REQ's ui_impact (§4.2 step 6).
	uiImpact := boundREQUIImpact(state)
	e2eState := "regression_available"
	switch uiImpact {
	case "none", "":
		e2eState = "not_applicable"
		claims = append(claims, Claim{
			ClaimID: "claim-e2e-na-1", Lens: "e2e", Target: "n/a",
			Assertion: "no user-observable behavior changed",
			Oracle:    "impact analysis shows no user-visible surface",
			Method:    "impact analysis",
			Applicability: "not_applicable",
			NARationale:   "bound REQ declares no UI impact; no entry point or browser-observable behavior is in scope",
			SourceRefs:    []string{"bound_req"},
		})
		notes = append(notes, "E2E assessed as not_applicable from ui_impact=none; verify against the real required surfaces (§4.3) before registering")
	default:
		e2eState = "cold_start"
		notes = append(notes, "E2E cold start: decompose persona/entry/flow/state/negative/recovery into 1..N behavior Assignments; never compress the blank matrix into one generic Agent")
		id := nextClaim("e2e", "flows")
		claims = append(claims, Claim{
			ClaimID: id, Lens: "e2e", Target: "TODO(planner): persona/flow surface",
			Assertion: "declared entry points produce the expected user-observable behavior",
			Oracle:    "TODO(planner): the flow-level oracle with console/network evidence",
			Method:    "real-browser execution",
			Applicability: "required",
			SourceRefs:    taskIDs(tasks),
			FocusKey:      "user-flow",
		})
		assignments = append(assignments, PlanAssignment{
			AssignmentID: "assignment-e2e-flows", Lens: "e2e", ClaimIDs: []string{id},
			NonOverlapBoundary: "owns the declared flows; cold-start spec authoring stays inside the verification workspace",
			ExecutionWave:      "behavior",
		})
	}

	if len(tasks) == 0 {
		notes = append(notes, "no current-generation TASK documents found; the draft carries a placeholder DV/QA minimum — registration will reject an unjustified zero-coverage lens")
	}

	plan := &Plan{
		SchemaVersion:       "1.0.0",
		ReviewPlanID:        fmt.Sprintf("review-plan-draft-r%d", round),
		ReviewRound:         round,
		BaselineGeneration:  generation,
		FrozenSubjects:      frozen,
		Claims:              claims,
		Assignments:         assignments,
		E2ECoverageState:    e2eState,
		DispatchCapacityPolicy: "coverage_complete",
		CreatedBy:           "orchestrator",
		CreatedAt:           time.Now().UTC().Format(time.RFC3339Nano),
	}
	return plan, notes
}

func taskIDs(tasks []taskDoc) []string {
	ids := make([]string, 0, len(tasks))
	for _, task := range tasks {
		ids = append(ids, task.id)
	}
	return ids
}

func boundREQUIImpact(state map[string]any) string {
	bound, _ := state["bound_req"].(map[string]any)
	metadata, _ := bound["metadata"].(map[string]any)
	impact, _ := metadata["ui_impact"].(string)
	return impact
}

// ValidatePlanTaskCoverage is the registration-time coverage block
// (L3-S7 §4.4 coverage matrix): every current-generation TASK must appear
// in at least one Claim's source_refs, so a plan cannot silently drop part
// of the S6 batch.
func ValidatePlanTaskCoverage(state map[string]any, plan *Plan) error {
	generation := baselineGeneration(state)
	required := map[string]bool{}
	for _, raw := range stateDocuments(state) {
		doc, _ := raw.(map[string]any)
		if doc == nil || doc["kind"] != "task" {
			continue
		}
		if intField(doc["generation"]) != generation {
			continue
		}
		if id, _ := doc["id"].(string); id != "" {
			required[id] = true
		}
	}
	if len(required) == 0 {
		return nil
	}
	covered := map[string]bool{}
	for _, claim := range plan.Claims {
		for _, ref := range claim.SourceRefs {
			if required[ref] {
				covered[ref] = true
			}
		}
	}
	var missing []string
	for id := range required {
		if !covered[id] {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf("ReviewPlan drops current-generation TASK(s) %v from every Claim's source_refs; coverage diff is a registration-time error (L3-S7 §4.4)", missing)
	}
	return nil
}
