package investigation_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/entroforge/go-system-builder/internal/investigation"
	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/schema"
	req039fixtures "github.com/entroforge/go-system-builder/tests/fixtures/req039"
)

func TestRegisterHypothesisCreatesImmutableCaseRevisionAndHistory(t *testing.T) {
	fixture := readyCaseFixture(t, []string{"finding-2", "finding-1"})
	pointer := investigationPointer(t, fixture)
	oldPath := filepath.Join(fixture.root, filepath.FromSlash(pointer["path"].(string)))
	oldBytes, err := os.ReadFile(oldPath)
	if err != nil {
		t.Fatal(err)
	}

	snapshot, err := investigation.RegisterHypothesis(fixture.root, fixture.statePath, fixture.journalPath, investigation.HypothesisRequest{
		ExpectedRevision:     1,
		ExpectedCaseRevision: 1,
		ExpectedCaseSHA256:   pointer["sha256"].(string),
		CaseID:               "investigation-case-observation-batch-r1",
		HypothesisID:         "hypothesis-payload-drift",
		Statement:            "FE and BE payload schemas drift at the serialization boundary",
		Invariant:            "one authoritative payload contract owns field shape",
		Discriminator:        "compare generated client payload with the server DTO",
		ExpectedOutcomes: map[string]any{
			"support": "the field differs at the boundary",
			"refute":  "the field is identical across the boundary",
		},
		SourceFindingIDs: []string{"finding-1", "finding-2"},
		AssignmentID:     "assignment-hypothesis-1",
	})
	if err != nil {
		t.Fatalf("RegisterHypothesis() error = %v", err)
	}
	if snapshot.Revision != 2 {
		t.Fatalf("runtime revision = %d, want 2", snapshot.Revision)
	}
	newPointer := investigationPointerFromState(t, snapshot.State)
	if newPointer["revision"] != float64(2) && newPointer["revision"] != 2 {
		t.Fatalf("case pointer revision = %v, want 2", newPointer["revision"])
	}
	if newPointer["sha256"] == pointer["sha256"] {
		t.Fatal("case revision must receive a new content hash")
	}
	if string(oldBytes) != string(mustRead(t, oldPath)) {
		t.Fatal("previous Case revision was mutated")
	}

	newPath := filepath.Join(fixture.root, filepath.FromSlash(newPointer["path"].(string)))
	caseBytes := mustRead(t, newPath)
	if err := schema.NewEmbeddedValidator().ValidateBytes("review-investigation-case.schema.json", caseBytes); err != nil {
		t.Fatalf("revised Case schema: %v", err)
	}
	var document map[string]any
	if err := json.Unmarshal(caseBytes, &document); err != nil {
		t.Fatal(err)
	}
	if len(document["hypotheses"].([]any)) != 1 {
		t.Fatalf("hypotheses = %#v, want one registered hypothesis", document["hypotheses"])
	}
	hypothesis := document["hypotheses"].([]any)[0].(map[string]any)
	if hypothesis["hypothesis_id"] != "hypothesis-payload-drift" || hypothesis["discriminator"] == "" {
		t.Fatalf("unexpected hypothesis = %#v", hypothesis)
	}
	history := document["revision_history"].([]any)
	if len(history) != 1 || history[0].(map[string]any)["revision"] != float64(1) {
		t.Fatalf("revision_history = %#v, want predecessor revision 1", history)
	}
}

func TestRegisterHypothesisRequiresBoundAssignment(t *testing.T) {
	fixture := readyCaseFixture(t, []string{"finding-1"})
	pointer := investigationPointer(t, fixture)
	_, err := investigation.RegisterHypothesis(fixture.root, fixture.statePath, fixture.journalPath, investigation.HypothesisRequest{
		ExpectedRevision:     1,
		ExpectedCaseRevision: 1,
		ExpectedCaseSHA256:   pointer["sha256"].(string),
		CaseID:               "investigation-case-observation-batch-r1",
		HypothesisID:         "hypothesis-unbound",
		Statement:            "the boundary contract is inconsistent",
		Invariant:            "the boundary has one owner",
		Discriminator:        "inspect both sides of the boundary",
		ExpectedOutcomes:     map[string]any{"support": "drift", "refute": "no drift"},
		SourceFindingIDs:     []string{"finding-1"},
	})
	if err == nil || !strings.Contains(err.Error(), "assignment_id") || !strings.Contains(err.Error(), "assignment- prefix") {
		t.Fatalf("unbound hypothesis error = %v, want Assignment binding guidance", err)
	}
}

func TestCaseWorkflowRejectsStaleRevisionAndHashDrift(t *testing.T) {
	fixture := readyCaseFixture(t, []string{"finding-1"})
	pointer := investigationPointer(t, fixture)
	request := investigation.HypothesisRequest{
		ExpectedRevision:     0,
		ExpectedCaseRevision: 1,
		ExpectedCaseSHA256:   pointer["sha256"].(string),
		CaseID:               "investigation-case-observation-batch-r1",
		HypothesisID:         "hypothesis-stale",
		Statement:            "the boundary contract is inconsistent",
		Invariant:            "the boundary has one owner",
		Discriminator:        "inspect both sides of the boundary",
		ExpectedOutcomes:     map[string]any{"support": "drift", "refute": "no drift"},
		SourceFindingIDs:     []string{"finding-1"},
		AssignmentID:         "assignment-hypothesis-stale",
	}
	_, err := investigation.RegisterHypothesis(fixture.root, fixture.statePath, fixture.journalPath, request)
	if !strings.Contains(err.Error(), "stale") {
		t.Fatalf("stale error = %v, want stale guidance", err)
	}

	request.ExpectedRevision = 1
	request.ExpectedCaseSHA256 = strings.Repeat("b", 64)
	_, err = investigation.RegisterHypothesis(fixture.root, fixture.statePath, fixture.journalPath, request)
	if err == nil || !strings.Contains(err.Error(), "hash") || !strings.Contains(err.Error(), "runtime investigation") {
		t.Fatalf("hash drift error = %v, want hash and recovery guidance", err)
	}
}

func TestSubmitHypothesisResultRequiresRegisteredHypothesisAndFindingSubset(t *testing.T) {
	fixture := readyCaseFixture(t, []string{"finding-1", "finding-2"})
	pointer := investigationPointer(t, fixture)
	register := investigation.HypothesisRequest{
		ExpectedRevision:     1,
		ExpectedCaseRevision: 1,
		ExpectedCaseSHA256:   pointer["sha256"].(string),
		CaseID:               "investigation-case-observation-batch-r1",
		HypothesisID:         "hypothesis-boundary",
		Statement:            "the payload contract drifts at the boundary",
		Invariant:            "one contract owns the payload shape",
		Discriminator:        "compare request and DTO fields",
		ExpectedOutcomes:     map[string]any{"support": "field drift", "refute": "no drift"},
		SourceFindingIDs:     []string{"finding-1", "finding-2"},
		AssignmentID:         "assignment-boundary",
	}
	if _, err := investigation.RegisterHypothesis(fixture.root, fixture.statePath, fixture.journalPath, register); err != nil {
		t.Fatalf("RegisterHypothesis() error = %v", err)
	}
	pointer = investigationPointer(t, fixture)
	bad := investigation.HypothesisResultRequest{
		ExpectedRevision:     2,
		ExpectedCaseRevision: 2,
		ExpectedCaseSHA256:   pointer["sha256"].(string),
		CaseID:               register.CaseID,
		HypothesisID:         "hypothesis-missing",
		AssignmentID:         "assignment-1",
		Method:               "read-only schema comparison",
		EvidenceRefs:         []string{"evidence-schema-diff"},
		SourceBoundaryRefs:   []string{"finding-1:boundary"},
		Observed:             "the field is absent from the DTO",
		Counterfactual:       "if the contract is aligned, both sides carry the field",
		Result:               "supported",
		ExplainsFindingIDs:   []string{"finding-1", "finding-2", "finding-outside-case"},
	}
	_, err := investigation.SubmitHypothesisResult(fixture.root, fixture.statePath, fixture.journalPath, bad)
	if err == nil || !strings.Contains(err.Error(), "registered") || !strings.Contains(err.Error(), "source Finding") {
		t.Fatalf("invalid result error = %v, want binding guidance", err)
	}
}

func TestSubmitHypothesisResultRejectsUnboundFollowUpHypothesis(t *testing.T) {
	fixture := readyCaseFixture(t, []string{"finding-1"})
	pointer := investigationPointer(t, fixture)
	register := investigation.HypothesisRequest{
		ExpectedRevision:     1,
		ExpectedCaseRevision: 1,
		ExpectedCaseSHA256:   pointer["sha256"].(string),
		CaseID:               "investigation-case-observation-batch-r1",
		HypothesisID:         "hypothesis-primary",
		Statement:            "the boundary drops the required value",
		Invariant:            "the value survives the boundary",
		Discriminator:        "trace the value across the boundary",
		ExpectedOutcomes:     map[string]any{"support": "the value is dropped", "refute": "the value survives"},
		SourceFindingIDs:     []string{"finding-1"},
		AssignmentID:         "assignment-primary",
	}
	if _, err := investigation.RegisterHypothesis(fixture.root, fixture.statePath, fixture.journalPath, register); err != nil {
		t.Fatalf("RegisterHypothesis() error = %v", err)
	}
	pointer = investigationPointer(t, fixture)
	_, err := investigation.SubmitHypothesisResult(fixture.root, fixture.statePath, fixture.journalPath, investigation.HypothesisResultRequest{
		ExpectedRevision:     2,
		ExpectedCaseRevision: 2,
		ExpectedCaseSHA256:   pointer["sha256"].(string),
		CaseID:               register.CaseID,
		HypothesisID:         register.HypothesisID,
		AssignmentID:         register.AssignmentID,
		Method:               "read-only boundary trace",
		EvidenceRefs:         []string{"evidence-boundary-trace"},
		SourceBoundaryRefs:   []string{"service.go:87"},
		Observed:             "the value is dropped at the decoder",
		Counterfactual:       "an aligned decoder preserves the value",
		Result:               "supported",
		ExplainsFindingIDs:   []string{"finding-1"},
		NewHypotheses: []map[string]any{{
			"hypothesis_id":      "hypothesis-follow-up",
			"statement":          "the decoder has a second field mapping",
			"invariant":          "one decoder mapping owns the field",
			"discriminator":      "compare generated and runtime mappings",
			"expected_outcomes":  map[string]any{"support": "mappings differ", "refute": "mappings match"},
			"source_finding_ids": []any{"finding-1"},
		}},
	})
	if err == nil || !strings.Contains(err.Error(), "assignment_id") || !strings.Contains(err.Error(), "assignment- prefix") {
		t.Fatalf("unbound follow-up hypothesis error = %v, want Assignment binding guidance", err)
	}
}

func TestUpdateCaseRouteIsDeterministicAndPreservesExactFindings(t *testing.T) {
	fixture := readyCaseFixture(t, []string{"finding-1", "finding-2"})
	pointer := investigationPointer(t, fixture)
	register := investigation.HypothesisRequest{
		ExpectedRevision:     1,
		ExpectedCaseRevision: 1,
		ExpectedCaseSHA256:   pointer["sha256"].(string),
		CaseID:               "investigation-case-observation-batch-r1",
		HypothesisID:         "hypothesis-root",
		Statement:            "one broken payload authority explains both findings",
		Invariant:            "the payload contract has one owner",
		Discriminator:        "compare generated schema and DTO",
		ExpectedOutcomes:     map[string]any{"support": "mismatch", "refute": "match"},
		SourceFindingIDs:     []string{"finding-1", "finding-2"},
		AssignmentID:         "assignment-root",
	}
	if _, err := investigation.RegisterHypothesis(fixture.root, fixture.statePath, fixture.journalPath, register); err != nil {
		t.Fatalf("RegisterHypothesis() error = %v", err)
	}
	pointer = investigationPointer(t, fixture)
	result := investigation.HypothesisResultRequest{
		ExpectedRevision:     2,
		ExpectedCaseRevision: 2,
		ExpectedCaseSHA256:   pointer["sha256"].(string),
		CaseID:               register.CaseID,
		HypothesisID:         register.HypothesisID,
		AssignmentID:         "assignment-root",
		Method:               "read-only schema comparison",
		EvidenceRefs:         []string{"evidence-schema-diff"},
		SourceBoundaryRefs:   []string{"finding-1:boundary", "finding-2:boundary"},
		Observed:             "the same schema drift is present in both paths",
		Counterfactual:       "aligned schema removes both boundary mismatches",
		Result:               "supported",
		ExplainsFindingIDs:   []string{"finding-1", "finding-2"},
	}
	if _, err := investigation.SubmitHypothesisResult(fixture.root, fixture.statePath, fixture.journalPath, result); err != nil {
		t.Fatalf("SubmitHypothesisResult() error = %v", err)
	}
	pointer = investigationPointer(t, fixture)
	snapshot, err := investigation.UpdateCaseRoute(fixture.root, fixture.statePath, fixture.journalPath, investigation.RouteRequest{
		ExpectedRevision:     3,
		ExpectedCaseRevision: 3,
		ExpectedCaseSHA256:   pointer["sha256"].(string),
		CaseID:               register.CaseID,
		Route:                "s9_repair",
		RouteReason:          "supported causal model identifies an implementation boundary repair",
		PrimaryRootCause:     "two incompatible payload authorities drift",
		CausalModel:          map[string]any{"trigger": "new field", "propagation": "decoder drops field"},
		BlastRadius:          map[string]any{"surfaces": []any{"request", "response"}},
		DetectionGap:         map[string]any{"missing": "contract assertion"},
	})
	if err != nil {
		t.Fatalf("UpdateCaseRoute() error = %v", err)
	}
	if snapshot.Revision != 4 {
		t.Fatalf("runtime revision = %d, want 4", snapshot.Revision)
	}
	finalPointer := investigationPointerFromState(t, snapshot.State)
	caseDocument := readCaseDocument(t, fixture.root, finalPointer["path"].(string))
	if caseDocument["route"] != "s9_repair" || caseDocument["status"] != "investigating" {
		t.Fatalf("route/status = %v/%v, want s9_repair/investigating", caseDocument["route"], caseDocument["status"])
	}
	if got := caseDocument["source_finding_ids"].([]any); len(got) != 2 || got[0] != "finding-1" || got[1] != "finding-2" {
		t.Fatalf("source Finding set changed: %#v", got)
	}
	if _, err := investigation.UpdateCaseRoute(fixture.root, fixture.statePath, fixture.journalPath, investigation.RouteRequest{
		ExpectedRevision:     4,
		ExpectedCaseRevision: 4,
		ExpectedCaseSHA256:   finalPointer["sha256"].(string),
		CaseID:               register.CaseID,
		Route:                "s9_repair",
		RouteReason:          "duplicate conflicting route attempt",
	}); err == nil || !strings.Contains(err.Error(), "deterministic") {
		t.Fatalf("inconsistent route update error = %v, want deterministic route guidance", err)
	}
}

func TestUpdateCaseRouteCanReopenInvestigateMoreAfterNewEvidence(t *testing.T) {
	fixture := readyCaseFixture(t, []string{"finding-1"})
	pointer := investigationPointer(t, fixture)
	register := investigation.HypothesisRequest{
		ExpectedRevision:     1,
		ExpectedCaseRevision: 1,
		ExpectedCaseSHA256:   pointer["sha256"].(string),
		CaseID:               "investigation-case-observation-batch-r1",
		HypothesisID:         "hypothesis-boundary",
		AssignmentID:         "assignment-boundary",
		Statement:            "the boundary drops the required value",
		Invariant:            "the value survives the boundary",
		Discriminator:        "trace the value through the boundary",
		ExpectedOutcomes:     map[string]any{"support": "the value is dropped", "refute": "the value survives"},
		SourceFindingIDs:     []string{"finding-1"},
	}
	if _, err := investigation.RegisterHypothesis(fixture.root, fixture.statePath, fixture.journalPath, register); err != nil {
		t.Fatalf("RegisterHypothesis() error = %v", err)
	}
	pointer = investigationPointer(t, fixture)
	snapshot, err := investigation.UpdateCaseRoute(fixture.root, fixture.statePath, fixture.journalPath, investigation.RouteRequest{
		ExpectedRevision:     2,
		ExpectedCaseRevision: 2,
		ExpectedCaseSHA256:   pointer["sha256"].(string),
		CaseID:               register.CaseID,
		Route:                "investigate_more",
		RouteReason:          "the source Finding is not yet explained",
	})
	if err != nil {
		t.Fatalf("initial investigate_more route error = %v", err)
	}
	pointer = investigationPointerFromState(t, snapshot.State)
	result := investigation.HypothesisResultRequest{
		ExpectedRevision:     3,
		ExpectedCaseRevision: 3,
		ExpectedCaseSHA256:   pointer["sha256"].(string),
		CaseID:               register.CaseID,
		HypothesisID:         register.HypothesisID,
		AssignmentID:         register.AssignmentID,
		Method:               "read-only boundary trace",
		EvidenceRefs:         []string{"evidence-boundary-trace"},
		SourceBoundaryRefs:   []string{"service.go:87"},
		Observed:             "the value is dropped at the decoder",
		Counterfactual:       "an aligned decoder preserves the value",
		Result:               "supported",
		ExplainsFindingIDs:   []string{"finding-1"},
	}
	snapshot, err = investigation.SubmitHypothesisResult(fixture.root, fixture.statePath, fixture.journalPath, result)
	if err != nil {
		t.Fatalf("SubmitHypothesisResult() after investigate_more error = %v", err)
	}
	pointer = investigationPointerFromState(t, snapshot.State)
	snapshot, err = investigation.UpdateCaseRoute(fixture.root, fixture.statePath, fixture.journalPath, investigation.RouteRequest{
		ExpectedRevision:     4,
		ExpectedCaseRevision: 4,
		ExpectedCaseSHA256:   pointer["sha256"].(string),
		CaseID:               register.CaseID,
		Route:                "s9_repair",
		RouteReason:          "new supported evidence closes the causal chain",
		PrimaryRootCause:     "the decoder drops the required value",
		CausalModel:          map[string]any{"trigger": "payload crosses boundary", "propagation": "decoder drops value"},
		BlastRadius:          map[string]any{"surfaces": []any{"service"}},
		DetectionGap:         map[string]any{"missing": "boundary contract assertion"},
	})
	if err != nil {
		t.Fatalf("re-route to s9_repair error = %v", err)
	}
	finalPointer := investigationPointerFromState(t, snapshot.State)
	caseDocument := readCaseDocument(t, fixture.root, finalPointer["path"].(string))
	if caseDocument["route"] != "s9_repair" {
		t.Fatalf("route = %v, want s9_repair", caseDocument["route"])
	}
	history, ok := caseDocument["route_history"].([]any)
	if !ok || len(history) != 2 {
		t.Fatalf("route_history = %#v, want initial and re-route entries", caseDocument["route_history"])
	}
	if history[0].(map[string]any)["to"] != "investigate_more" || history[1].(map[string]any)["from"] != "investigate_more" || history[1].(map[string]any)["to"] != "s9_repair" {
		t.Fatalf("route_history = %#v, want investigate_more -> s9_repair", history)
	}
}

func TestUpdateCaseRouteReopensApprovedCaseWithCausalReassessmentEvidence(t *testing.T) {
	fixture := readyCaseFixture(t, []string{"finding-1"})
	pointer := investigationPointer(t, fixture)
	contractRef := ".claude/review/investigation/contracts/repair-contract-r2.json"
	approvedSnapshot, err := investigation.UpdateCase(fixture.root, fixture.statePath, fixture.journalPath, investigation.CaseRevisionRequest{
		ExpectedRevision:     1,
		ExpectedCaseRevision: 1,
		ExpectedCaseSHA256:   pointer["sha256"].(string),
		CaseID:               "investigation-case-observation-batch-r1",
		Operation:            "test_contract_approved_case",
		Mutate: func(document map[string]any) error {
			document["status"] = "contract_approved"
			document["route"] = "s9_repair"
			document["repair_contract_ref"] = contractRef
			document["repair_contract_sha256"] = strings.Repeat("a", 64)
			return nil
		},
	})
	if err != nil {
		t.Fatalf("seed approved Case: %v", err)
	}
	approvedState := approvedSnapshot.State["review"].(map[string]any)
	approvedState["repair"] = map[string]any{
		"session_id": "repair-session-old", "case_id": "investigation-case-observation-batch-r1", "contract_id": "repair-contract-old",
		"contract_ref": contractRef, "contract_sha256": strings.Repeat("c", 64),
		"path": ".claude/review/repair/sessions/repair-session-old.json", "sha256": strings.Repeat("d", 64), "revision": 1,
		"status": "blocked", "failure_route": "fail_same_cause", "targeted_reverification_refs": []any{".claude/review/repair/reverification/reverify-failure.json"},
		"targeted_reverification_artifacts": []any{}, "updated_at": "2026-08-26T00:00:00Z",
		"next_action": "re-open the Case with causal reassessment",
	}
	req039fixtures.WriteState(t, fixture.root, approvedSnapshot.State)
	pointer = investigationPointerFromState(t, approvedSnapshot.State)
	evidenceRel := ".claude/review/repair/reverification/reverify-failure.json"
	evidence := []byte("{\"result\":\"fail\",\"failure_class\":\"fail_same_cause\"}\n")
	evidencePath := filepath.Join(fixture.root, filepath.FromSlash(evidenceRel))
	if err := os.MkdirAll(filepath.Dir(evidencePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(evidencePath, evidence, 0o644); err != nil {
		t.Fatal(err)
	}
	evidenceSHA := sha256.Sum256(evidence)

	snapshot, err := investigation.UpdateCaseRoute(fixture.root, fixture.statePath, fixture.journalPath, investigation.RouteRequest{
		ExpectedRevision:     2,
		ExpectedCaseRevision: 2,
		ExpectedCaseSHA256:   pointer["sha256"].(string),
		CaseID:               "investigation-case-observation-batch-r1",
		Route:                "investigate_more",
		RouteReason:          "targeted reverification shows the approved causal model needs reassessment",
		CausalReassessmentEvidenceRefs: []investigation.EvidenceReference{{
			Path: evidenceRel, SHA256: hex.EncodeToString(evidenceSHA[:]),
		}},
	})
	if err != nil {
		t.Fatalf("reopen approved Case: %v", err)
	}
	finalPointer := investigationPointerFromState(t, snapshot.State)
	document := readCaseDocument(t, fixture.root, finalPointer["path"].(string))
	if document["status"] != "investigating" || document["route"] != "investigate_more" {
		t.Fatalf("reopened Case status/route = %v/%v, want investigating/investigate_more", document["status"], document["route"])
	}
	if document["repair_contract_ref"] != nil || document["repair_contract_sha256"] != nil {
		t.Fatalf("reopened Case must clear the superseded Contract pointer: %#v", document)
	}
	currentReview := snapshot.State["review"].(map[string]any)
	if repair, ok := currentReview["repair"].(map[string]any); ok && repair != nil {
		t.Fatalf("reopening the Case must retire the superseded S9 pointer so a new Contract can open a new session: %#v", repair)
	}
	refs, ok := document["causal_reassessment_refs"].([]any)
	if !ok || len(refs) != 1 || refs[0].(map[string]any)["path"] != evidenceRel || refs[0].(map[string]any)["sha256"] != hex.EncodeToString(evidenceSHA[:]) {
		t.Fatalf("reopened Case must retain exact causal reassessment evidence refs: %#v", document["causal_reassessment_refs"])
	}
	history := document["route_history"].([]any)
	last := history[len(history)-1].(map[string]any)
	if last["from"] != "s9_repair" || last["to"] != "investigate_more" {
		t.Fatalf("route history = %#v, want s9_repair -> investigate_more", history)
	}
}

func TestDuplicateRoutePersistsCanonicalCaseReference(t *testing.T) {
	fixture := readyCaseFixture(t, []string{"finding-1"})
	pointer := investigationPointer(t, fixture)
	currentPath := filepath.Join(fixture.root, filepath.FromSlash(pointer["path"].(string)))
	canonicalBytes := mustRead(t, currentPath)
	var canonical map[string]any
	if err := json.Unmarshal(canonicalBytes, &canonical); err != nil {
		t.Fatal(err)
	}
	canonical["case_id"] = "investigation-case-canonical"
	canonicalBytes, err := json.MarshalIndent(canonical, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	canonicalBytes = append(canonicalBytes, '\n')
	canonicalRel := ".claude/review/investigation/cases/investigation-case-canonical-r1.json"
	canonicalPath := filepath.Join(fixture.root, filepath.FromSlash(canonicalRel))
	if err := os.WriteFile(canonicalPath, canonicalBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	canonicalSHA := sha256.Sum256(canonicalBytes)

	snapshot, err := investigation.UpdateCaseRoute(fixture.root, fixture.statePath, fixture.journalPath, investigation.RouteRequest{
		ExpectedRevision:     1,
		ExpectedCaseRevision: 1,
		ExpectedCaseSHA256:   pointer["sha256"].(string),
		CaseID:               "investigation-case-observation-batch-r1",
		Route:                "duplicate",
		RouteReason:          "the same causal incident is already tracked canonically",
		CanonicalCaseID:      "investigation-case-canonical",
	})
	if err != nil {
		t.Fatalf("duplicate route error = %v", err)
	}
	finalPointer := investigationPointerFromState(t, snapshot.State)
	caseDocument := readCaseDocument(t, fixture.root, finalPointer["path"].(string))
	if caseDocument["route"] != "duplicate" || caseDocument["canonical_case_id"] != "investigation-case-canonical" {
		t.Fatalf("duplicate routing fields = %#v", caseDocument)
	}
	if caseDocument["canonical_case_ref"] != canonicalRel || caseDocument["canonical_case_sha256"] != hex.EncodeToString(canonicalSHA[:]) {
		t.Fatalf("canonical reference = %v/%v, want %s/%s", caseDocument["canonical_case_ref"], caseDocument["canonical_case_sha256"], canonicalRel, hex.EncodeToString(canonicalSHA[:]))
	}
}

func readyCaseFixture(t *testing.T, findingIDs []string) *intakeFixture {
	t.Helper()
	fixture := newIntakeFixture(t, findingIDs)
	setContractLifecycle(t, fixture)
	if _, err := investigation.Ingest(fixture.root, fixture.statePath, fixture.journalPath, investigation.IngestRequest{
		ExpectedRevision:  0,
		GroupingRationale: "the sealed batch is the provisional grouping boundary",
	}); err != nil {
		t.Fatalf("Ingest() error = %v", err)
	}
	return fixture
}

func investigationPointer(t *testing.T, fixture *intakeFixture) map[string]any {
	t.Helper()
	snapshot, err := runtime.NewStore(fixture.statePath, fixture.journalPath).Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	return investigationPointerFromState(t, snapshot.State)
}

func investigationPointerFromState(t *testing.T, state map[string]any) map[string]any {
	t.Helper()
	return state["review"].(map[string]any)["investigation"].(map[string]any)
}

func readCaseDocument(t *testing.T, root, relative string) map[string]any {
	t.Helper()
	data := mustRead(t, filepath.Join(root, filepath.FromSlash(relative)))
	var document map[string]any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	return document
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
