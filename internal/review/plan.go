package review

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ---------------------------------------------------------------------------
// Plan loading: state.review.plan pins path+sha256; the content is read from
// disk and hash-verified on every load so consumers never trust a stale copy.
// ---------------------------------------------------------------------------

// PlanPointer is the state.review.plan projection.
type PlanPointer struct {
	PlanID                        string
	Path                          string
	SHA256                        string
	Revision                      int
	ReviewRound                   int
	Status                        string
	E2ECoverageState              string
	VerificationArtifactWorkspace string
	VerificationArtifactDigest    string
	SubmittedAt                   string
}

// PlanPointerFromState extracts the registered plan pointer, or nil.
func PlanPointerFromState(state map[string]any) *PlanPointer {
	review, _ := state["review"].(map[string]any)
	if review == nil {
		return nil
	}
	raw, _ := review["plan"].(map[string]any)
	if raw == nil {
		return nil
	}
	ptr := &PlanPointer{
		PlanID:           stringField(raw["plan_id"]),
		Path:             stringField(raw["path"]),
		SHA256:           stringField(raw["sha256"]),
		Revision:         intField(raw["revision"]),
		ReviewRound:      intField(raw["review_round"]),
		Status:           stringField(raw["status"]),
		E2ECoverageState: stringField(raw["e2e_coverage_state"]),
		SubmittedAt:      stringField(raw["submitted_at"]),
	}
	if ws, ok := raw["verification_artifact_workspace"].(string); ok {
		ptr.VerificationArtifactWorkspace = ws
	}
	if digest, ok := raw["verification_artifact_digest"].(string); ok {
		ptr.VerificationArtifactDigest = digest
	}
	if ptr.PlanID == "" || ptr.Path == "" {
		return nil
	}
	return ptr
}

// LoadPlan reads and hash-verifies the pinned plan file.
func LoadPlan(root string, state map[string]any) (*Plan, *PlanPointer, error) {
	ptr := PlanPointerFromState(state)
	if ptr == nil {
		return nil, nil, fmt.Errorf("no ReviewPlan is registered for the current review round; create one with `loop-harness runtime review-plan --file <plan.json>`")
	}
	abs := ptr.Path
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(root, filepath.FromSlash(ptr.Path))
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return nil, nil, fmt.Errorf("read ReviewPlan %s: %w", ptr.Path, err)
	}
	if sha256Of(data) != ptr.SHA256 {
		return nil, nil, fmt.Errorf("ReviewPlan %s drifted: pinned sha256 %s does not match the file on disk", ptr.Path, ptr.SHA256)
	}
	var plan Plan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, nil, fmt.Errorf("decode ReviewPlan %s: %w", ptr.Path, err)
	}
	if plan.ReviewPlanID != ptr.PlanID {
		return nil, nil, fmt.Errorf("ReviewPlan id mismatch: state pins %s but the file declares %s", ptr.PlanID, plan.ReviewPlanID)
	}
	return &plan, ptr, nil
}

// SubjectDigest computes the frozen product baseline digest every
// ReviewResult must bind: sha256 over the sorted "path:sha256" lines of the
// plan's frozen_subjects (L3-S7 §3.5).
func SubjectDigest(plan *Plan) string {
	lines := make([]string, 0, len(plan.FrozenSubjects))
	for _, subject := range plan.FrozenSubjects {
		lines = append(lines, subject.Path+":"+subject.SHA256)
	}
	sort.Strings(lines)
	return sha256Of([]byte(strings.Join(lines, "\n")))
}

// ---------------------------------------------------------------------------
// Plan validation (L3-S7 §4.2/§4.4 plan validator). The schema validates
// shape; this validates the exact-set coverage semantics.
// ---------------------------------------------------------------------------

// ValidatePlan proves the plan's Claim/Assignment coverage is closed:
// every required Claim is owned by exactly one Assignment, N/A Claims are
// never dispatched, lenses match, dependencies resolve, and the DV/QA
// minimum holds (L3-S7 §4.2: a round with product impact never has a
// zero-Claim white-box lens).
func ValidatePlan(plan *Plan) error {
	if plan.DispatchCapacityPolicy != "coverage_complete" {
		return fmt.Errorf("dispatch_capacity_policy must be coverage_complete (L3-S7 §4.5); downgrading to bounded_flow is a human policy decision, not a plan field")
	}
	claims := make(map[string]Claim, len(plan.Claims))
	for _, claim := range plan.Claims {
		if _, dup := claims[claim.ClaimID]; dup {
			return fmt.Errorf("duplicate claim_id %s", claim.ClaimID)
		}
		claims[claim.ClaimID] = claim
	}
	for _, claim := range plan.Claims {
		for _, dep := range claim.DependsOn {
			if _, ok := claims[dep]; !ok {
				return fmt.Errorf("claim %s depends on unknown claim %s", claim.ClaimID, dep)
			}
		}
	}
	if cycle := findClaimCycle(plan.Claims); cycle != "" {
		return fmt.Errorf("claim dependency cycle involving %s", cycle)
	}

	assignmentSeen := make(map[string]bool, len(plan.Assignments))
	claimOwner := make(map[string]string, len(plan.Claims))
	for _, assignment := range plan.Assignments {
		if assignmentSeen[assignment.AssignmentID] {
			return fmt.Errorf("duplicate assignment_id %s", assignment.AssignmentID)
		}
		assignmentSeen[assignment.AssignmentID] = true
		if assignment.ExecutionWave != "static" && assignment.ExecutionWave != "behavior" {
			return fmt.Errorf("assignment %s has unknown execution_wave %q", assignment.AssignmentID, assignment.ExecutionWave)
		}
		for _, claimID := range assignment.ClaimIDs {
			claim, ok := claims[claimID]
			if !ok {
				return fmt.Errorf("assignment %s references unknown claim %s", assignment.AssignmentID, claimID)
			}
			if claim.Applicability == "not_applicable" {
				return fmt.Errorf("assignment %s dispatches not_applicable claim %s; N/A Claims are plan dispositions and are never dispatched (L3-S7 §9.3)", assignment.AssignmentID, claimID)
			}
			if claim.Lens != assignment.Lens {
				return fmt.Errorf("assignment %s (lens %s) covers claim %s of lens %s; merging lenses in one Assignment is forbidden (L3-S7 §3.4)", assignment.AssignmentID, assignment.Lens, claimID, claim.Lens)
			}
			if owner, dup := claimOwner[claimID]; dup {
				return fmt.Errorf("claim %s is assigned to both %s and %s; the required Claim set must be partitioned exactly", claimID, owner, assignment.AssignmentID)
			}
			claimOwner[claimID] = assignment.AssignmentID
		}
	}
	requiredByLens := map[string]int{"delivery": 0, "qa": 0, "e2e": 0}
	naByLens := map[string]int{"delivery": 0, "qa": 0, "e2e": 0}
	for _, claim := range plan.Claims {
		if claim.Applicability == "not_applicable" {
			naByLens[claim.Lens]++
			continue
		}
		requiredByLens[claim.Lens]++
		if _, assigned := claimOwner[claim.ClaimID]; !assigned {
			return fmt.Errorf("required claim %s (%s) has no owning Assignment; platform capacity may queue work but never delete coverage (L3-S7 §4.5)", claim.ClaimID, claim.Lens)
		}
	}
	justified := plan.CoverageJustification != nil && strings.TrimSpace(*plan.CoverageJustification) != ""
	for _, lens := range []string{"delivery", "qa"} {
		if requiredByLens[lens] == 0 && !justified {
			return fmt.Errorf("the round has zero required %s Claims; that is only legal for a pure docs/metadata change and needs a non-empty coverage_justification (L3-S7 §4.2)", lens)
		}
	}
	switch plan.E2ECoverageState {
	case "cold_start":
		if plan.VerificationArtifactWorkspace == nil || strings.TrimSpace(*plan.VerificationArtifactWorkspace) == "" {
			return fmt.Errorf("e2e_coverage_state=cold_start requires a verification_artifact_workspace; cold start never compresses the blank matrix into one generic Agent (L3-S7 §4.3)")
		}
		if requiredByLens["e2e"] == 0 {
			return fmt.Errorf("e2e_coverage_state=cold_start requires at least one required e2e Claim")
		}
	case "not_applicable":
		if naByLens["e2e"] == 0 {
			return fmt.Errorf("e2e_coverage_state=not_applicable requires at least one e2e Claim carrying applicability=not_applicable with source and rationale; ui_impact=none alone is not a conclusion (L3-S7 §4.3)")
		}
	case "regression_available":
	default:
		return fmt.Errorf("unknown e2e_coverage_state %q", plan.E2ECoverageState)
	}
	if err := validateAssignmentOverlap(plan, claims); err != nil {
		return err
	}
	if err := validateColdStartE2EOverload(plan, claims); err != nil {
		return err
	}
	return nil
}

// ---------------------------------------------------------------------------
// §14.1 overlap / cold-start overload validators (L3-S7 §3.4, §4.3, §4.4).
// Both are deterministic mechanical gates: they only fire on facts the plan
// itself declares, and every rejection names the gap plus the next action.
// ---------------------------------------------------------------------------

// validateAssignmentOverlap is the §14.1 overlap validator ("三个 QA Agents
// 都执行全量代码 review"). Deterministic suspicion rule — a pair of
// Assignments is a suspected duplicated generic review when ALL hold:
//
//  1. same lens (a different lens/persona is NEVER a merge reason, even when
//     the read set overlaps — L3-S7 §4.4);
//  2. identical non-empty target set: the sorted distinct claim.target
//     values of both Assignments are equal (the same read set);
//  3. identical method set: the sorted distinct claim.method values are
//     equal (the same inspection method).
//
// A suspected pair is still RELEASED (never force-merged) when either:
//
//   - the sorted distinct claim.oracle sets differ — a different oracle is
//     an independent perspective that must be preserved (§3.4); or
//   - both Assignments carry non-empty, mutually distinct
//     non_overlap_boundary values — the written independent-question /
//     non-overlap reason the blueprint demands.
//
// Otherwise registration is rejected: the pair is duplicated labor and must
// be merged into one Assignment, or each side must write a distinct
// non_overlap_boundary (and, where applicable, a distinct oracle).
func validateAssignmentOverlap(plan *Plan, claims map[string]Claim) error {
	type signature struct {
		targets []string
		methods []string
		oracles []string
	}
	sigOf := func(assignment PlanAssignment) signature {
		var sig signature
		for _, claimID := range assignment.ClaimIDs {
			claim, ok := claims[claimID]
			if !ok {
				continue
			}
			sig.targets = append(sig.targets, claim.Target)
			sig.methods = append(sig.methods, claim.Method)
			sig.oracles = append(sig.oracles, claim.Oracle)
		}
		sig.targets = sortedDistinct(sig.targets)
		sig.methods = sortedDistinct(sig.methods)
		sig.oracles = sortedDistinct(sig.oracles)
		return sig
	}
	equal := func(a, b []string) bool {
		if len(a) != len(b) {
			return false
		}
		for i := range a {
			if a[i] != b[i] {
				return false
			}
		}
		return true
	}
	for i := 0; i < len(plan.Assignments); i++ {
		for j := i + 1; j < len(plan.Assignments); j++ {
			a, b := plan.Assignments[i], plan.Assignments[j]
			if a.Lens != b.Lens {
				continue
			}
			sa, sb := sigOf(a), sigOf(b)
			if len(sa.targets) == 0 || !equal(sa.targets, sb.targets) || !equal(sa.methods, sb.methods) {
				continue
			}
			if !equal(sa.oracles, sb.oracles) {
				// Different oracle: an independent view over the same read
				// set is explicitly preserved, never merged (L3-S7 §4.4).
				continue
			}
			if strings.TrimSpace(a.NonOverlapBoundary) != "" && strings.TrimSpace(b.NonOverlapBoundary) != "" &&
				a.NonOverlapBoundary != b.NonOverlapBoundary {
				// Both sides wrote a distinct non-overlap boundary: the
				// required independent-question reason exists.
				continue
			}
			return fmt.Errorf("assignments %s and %s (lens %s) are a duplicated generic review: identical target set %s, identical methods %s and identical oracles %s (L3-S7 §3.4/§4.4). Merge them into one Assignment, or give each a mutually distinct non_overlap_boundary stating the independent question it answers — reading the same files is never a reason to merge different lenses/personas/oracles",
				a.AssignmentID, b.AssignmentID, a.Lens,
				strings.Join(sa.targets, ", "), strings.Join(sa.methods, ", "), strings.Join(sa.oracles, ", "))
		}
	}
	return nil
}

// validateColdStartE2EOverload is the §14.1 cold-start overload validator
// ("E2E cold_start，多个 persona/入口/状态/负向路径却只生成一个全需求
// Assignment"). Deterministic threshold — registration is rejected when ALL
// hold:
//
//  1. e2e_coverage_state == cold_start;
//  2. exactly one Assignment owns every required e2e Claim (the whole blank
//     matrix rides on a single E2E Agent);
//  3. that Assignment owns >= 2 required e2e Claims;
//  4. those Claims plus the Assignment's focus_keys express >= 2 distinct
//     non-empty focus dimensions — focus_key is the plan's declared carrier
//     for persona / entry / flow-cluster / negative-path dimensions.
//
// A small scope passes: a single required e2e Claim, or several Claims
// sharing one focus dimension (单一 persona 单一 flow cluster), or any plan
// that already split e2e coverage across 2+ Assignments. Plans whose e2e
// Claims declare no focus keys cannot be judged mechanically and pass —
// the validator only removes overload it can prove from declared facts.
func validateColdStartE2EOverload(plan *Plan, claims map[string]Claim) error {
	if plan.E2ECoverageState != "cold_start" {
		return nil
	}
	var owners []PlanAssignment
	for _, assignment := range plan.Assignments {
		if assignment.Lens != "e2e" {
			continue
		}
		for _, claimID := range assignment.ClaimIDs {
			if claim, ok := claims[claimID]; ok && claim.Applicability != "not_applicable" {
				owners = append(owners, assignment)
				break
			}
		}
	}
	if len(owners) != 1 {
		return nil
	}
	owner := owners[0]
	dimensions := map[string]bool{}
	required := 0
	for _, claimID := range owner.ClaimIDs {
		claim, ok := claims[claimID]
		if !ok || claim.Applicability == "not_applicable" {
			continue
		}
		required++
		if key := strings.TrimSpace(claim.FocusKey); key != "" {
			dimensions[key] = true
		}
	}
	for _, key := range owner.FocusKeys {
		if key = strings.TrimSpace(key); key != "" {
			dimensions[key] = true
		}
	}
	if required < 2 || len(dimensions) < 2 {
		return nil
	}
	return fmt.Errorf("e2e_coverage_state=cold_start but %s is the only E2E Assignment and spans %d discriminable focus dimensions (%s) across %d required e2e Claims: the blank coverage matrix is compressed into one generic Agent (L3-S7 §4.3/§4.4). Expand the coverage matrix first (persona/entry/flow cluster/negative path/state/side effect), then split into one behavior-wave E2E Assignment per recoverable flow context inside the verification_artifact_workspace",
		owner.AssignmentID, len(dimensions), strings.Join(sortedKeys(dimensions), ", "), required)
}

// sortedDistinct returns the sorted unique non-empty values.
func sortedDistinct(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

// sortedKeys returns the sorted keys of a string set.
func sortedKeys(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for key := range set {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

// findClaimCycle returns a claim id participating in a dependency cycle.
func findClaimCycle(claims []Claim) string {
	const (
		white = 0
		gray  = 1
		black = 2
	)
	deps := make(map[string][]string, len(claims))
	color := make(map[string]int, len(claims))
	for _, claim := range claims {
		deps[claim.ClaimID] = claim.DependsOn
	}
	var visit func(id string) bool
	visit = func(id string) bool {
		switch color[id] {
		case gray:
			return true
		case black:
			return false
		}
		color[id] = gray
		for _, dep := range deps[id] {
			if visit(dep) {
				return true
			}
		}
		color[id] = black
		return false
	}
	for _, claim := range claims {
		if color[claim.ClaimID] == white && visit(claim.ClaimID) {
			return claim.ClaimID
		}
	}
	return ""
}

// ---------------------------------------------------------------------------
// Disposition projection: state.review.claims is maintained inside the
// review-plan / review-result CAS transactions. These helpers read it.
// ---------------------------------------------------------------------------

// ClaimDisposition is one claim's round-level projection.
type ClaimDisposition struct {
	Lens          string
	Applicability string
	Disposition   string
	AssignmentID  string
	ResultID      string
	FindingIDs    []string
}

// Dispositions reads state.review.claims.
func Dispositions(state map[string]any) map[string]ClaimDisposition {
	out := map[string]ClaimDisposition{}
	review, _ := state["review"].(map[string]any)
	raw, _ := review["claims"].(map[string]any)
	for claimID, value := range raw {
		row, _ := value.(map[string]any)
		if row == nil {
			continue
		}
		disp := ClaimDisposition{
			Lens:          stringField(row["lens"]),
			Applicability: stringField(row["applicability"]),
			Disposition:   stringField(row["disposition"]),
			AssignmentID:  stringField(row["assignment_id"]),
			ResultID:      stringField(row["result_id"]),
		}
		findings, _ := row["finding_ids"].([]any)
		for _, f := range findings {
			if s, ok := f.(string); ok {
				disp.FindingIDs = append(disp.FindingIDs, s)
			}
		}
		out[claimID] = disp
	}
	return out
}

// RoundComplete reports whether every required claim has a final
// disposition (pass / finding / not_applicable / blocked). planned and
// running claims keep the round open.
func RoundComplete(state map[string]any) bool {
	for _, disp := range Dispositions(state) {
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

// UndispositionedRequired lists required claims still waiting for a Result.
func UndispositionedRequired(state map[string]any) []string {
	var pending []string
	for claimID, disp := range Dispositions(state) {
		if disp.Applicability == "not_applicable" {
			continue
		}
		switch disp.Disposition {
		case "pass", "finding", "blocked":
		default:
			pending = append(pending, claimID)
		}
	}
	sort.Strings(pending)
	return pending
}

// StaticClaimsSettled reports whether every required static-wave claim has a
// final disposition; behavior assignments dispatch only behind this fact
// (L3-S7 §5.2/§5.3).
func StaticClaimsSettled(state map[string]any, plan *Plan) bool {
	waveByClaim := map[string]string{}
	for _, assignment := range plan.Assignments {
		for _, claimID := range assignment.ClaimIDs {
			waveByClaim[claimID] = assignment.ExecutionWave
		}
	}
	dispositions := Dispositions(state)
	for _, claim := range plan.Claims {
		if claim.Applicability == "not_applicable" || waveByClaim[claim.ClaimID] != "static" {
			continue
		}
		switch dispositions[claim.ClaimID].Disposition {
		case "pass", "finding", "blocked":
		default:
			return false
		}
	}
	return true
}

// RoundFindings lists the finding entity rows for the current review round.
func RoundFindings(state map[string]any) []map[string]any {
	round := currentReviewRound(state)
	var out []map[string]any
	entities, _ := state["entities"].(map[string]any)
	findings, _ := entities["findings"].([]any)
	for _, raw := range findings {
		row, _ := raw.(map[string]any)
		if row == nil {
			continue
		}
		if intField(row["review_round"]) == round {
			out = append(out, row)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return stringField(out[i]["finding_id"]) < stringField(out[j]["finding_id"])
	})
	return out
}

// ---------------------------------------------------------------------------
// small shared helpers (kept local so review does not import qualitygate)
// ---------------------------------------------------------------------------

// artifactDigestOrNil maps an empty workspace digest to nil so the plan
// pointer omits the field when no workspace exists.
func artifactDigestOrNil(digest string) any {
	if digest == "" {
		return nil
	}
	return digest
}

func sha256Of(data []byte) string {
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum[:])
}

// evidenceEntries returns the evidence array as raw maps.
func evidenceEntries(state map[string]any) []any {
	entries, _ := state["evidence"].([]any)
	return entries
}

func stringField(value any) string {
	s, _ := value.(string)
	return s
}

func intField(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case json.Number:
		n, _ := v.Int64()
		return int(n)
	}
	return 0
}

func currentReviewRound(state map[string]any) int {
	review, _ := state["review"].(map[string]any)
	return intField(review["round"])
}

func baselineGeneration(state map[string]any) int {
	baseline, _ := state["baseline"].(map[string]any)
	return intField(baseline["generation"])
}
