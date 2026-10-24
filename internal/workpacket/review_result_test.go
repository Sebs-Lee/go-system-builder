package workpacket

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/entroforge/go-system-builder/internal/schema"
)

func reviewResultTestBinding() ReviewResultBinding {
	return ReviewResultBinding{
		RuntimeID: "loop-REQ-041", BaselineGeneration: 3,
		ReviewPlanID: "review-plan-041-r1", ReviewPlanRef: ".claude/review/plan.json",
		ReviewPlanSHA256: strings.Repeat("a", 64), ReviewPlanRevision: 2, ReviewRound: 1,
		AssignmentID: "assignment-qa-1", AssignmentSHA256: strings.Repeat("b", 64),
		OwnerAgentID: "reviewer-qa-1", ClaimIDs: []string{"claim-login", "claim-reset"},
		SubjectDigest: strings.Repeat("c", 64), VerificationArtifactDigest: nil,
	}
}

func reviewResultTestInputs() ReviewResultInputs {
	return ReviewResultInputs{
		Claims: json.RawMessage(`[
            {"claim_id":"claim-login","lens":"qa","target":"login","assertion":"login works","oracle":"visible home","method":"submit credentials","applicability":"required","source_refs":["REQ-041"]},
            {"claim_id":"claim-reset","lens":"qa","target":"reset","assertion":"reset works","oracle":"email sent","method":"request reset","applicability":"required","source_refs":["REQ-041"]}
        ]`),
		FrozenSubjects: json.RawMessage(`[{"path":"docs/requirements/REQ-041.md","sha256":"dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"}]`),
	}
}

func TestBuildReviewResultPacketIsStableAndSchemasCompile(t *testing.T) {
	p, err := BuildReviewResult(reviewResultTestBinding(), reviewResultTestInputs())
	if err != nil {
		t.Fatal(err)
	}
	if question := p.Draft.Questions["findings"].Question; !strings.Contains(question, "visible_impact 位于 Finding 顶层") || !strings.Contains(question, "encounter.side_effects 字符串数组") || !strings.Contains(question, "不能在 Finding 顶层添加 side_effects") {
		t.Fatalf("Finding question must identify the existing nested field locations: %s", question)
	}
	again, err := BuildReviewResult(reviewResultTestBinding(), reviewResultTestInputs())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(mustReviewPacketJSON(t, p), mustReviewPacketJSON(t, again)) {
		t.Fatal("same binding and frozen questions must produce the same packet")
	}
	if p.TemplateID != "s7-review-result-v1" || p.ResultID != "review-result-"+strings.TrimPrefix(p.ID, "packet-") || p.IssuanceEventID != "evt-review-result-packet-"+p.ID {
		t.Fatalf("derived S7 identity mismatch: %+v", p)
	}
	if !strings.Contains(p.ViewPath, "/S7/") || !strings.Contains(p.AnswerPath, "/S7/") || !strings.HasSuffix(p.AnswerPath, "/reviewer-qa-1/answer.json") {
		t.Fatalf("S7 paths do not isolate this owner: view=%s answer=%s", p.ViewPath, p.AnswerPath)
	}
	for _, test := range []struct {
		schema string
		data   any
	}{
		{"review-result-work-packet.schema.json", p},
		{"stage-answer-s7.schema.json", p.Draft},
	} {
		data, err := json.Marshal(test.data)
		if err != nil {
			t.Fatal(err)
		}
		if err := schema.NewEmbeddedValidator().ValidateBytes(test.schema, data); err != nil {
			t.Fatalf("%s: %v", test.schema, err)
		}
	}
	stateBytes, err := schema.ReadAsset("loop-state.example.json")
	if err != nil {
		t.Fatal(err)
	}
	var state map[string]any
	if err := json.Unmarshal(stateBytes, &state); err != nil {
		t.Fatal(err)
	}
	state["schema_version"] = "1.2.0"
	state["configuration"].(map[string]any)["optimization"] = map[string]any{"packet": true}
	entities := state["entities"].(map[string]any)
	entities["work_packets"] = []any{p}
	entities["submissions"] = []any{}
	stateBytes, err = json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := schema.NewEmbeddedValidator().ValidateBytes("loop-state.schema.json", stateBytes); err != nil {
		t.Fatalf("v1.2 state work_packets oneOf must resolve the typed S7 schema: %v", err)
	}
	digest, err := ReviewResultPacketDigest(p)
	if err != nil || !digestPattern.MatchString(digest) {
		t.Fatalf("packet digest = %q, %v", digest, err)
	}
	pretty, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	var restored ReviewResultPacket
	if err := json.Unmarshal(pretty, &restored); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(p, restored) {
		t.Fatal("packet raw JSON fields should decode to canonical stable bytes")
	}
	restoredDigest, err := ReviewResultPacketDigest(restored)
	if err != nil || restoredDigest != digest {
		t.Fatalf("restored packet digest = %q, want %q; err=%v", restoredDigest, digest, err)
	}
	otherInputs := reviewResultTestInputs()
	otherInputs.FrozenSubjects = json.RawMessage(`[ { "sha256" : "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd", "path" : "docs/requirements/REQ-041.md" } ]`)
	other, err := BuildReviewResult(reviewResultTestBinding(), otherInputs)
	if err != nil || other.ID != p.ID {
		t.Fatalf("JSON object order/spacing should not change packet ID: id=%s err=%v", other.ID, err)
	}
}

func TestHistoricalReviewResultTemplateOnlyRestoresKnownFrozenQuestions(t *testing.T) {
	data, err := schema.ReadAsset("stage-form-s7-template-before-safety-stop.json")
	if err != nil {
		t.Fatal(err)
	}
	old, err := buildReviewResultWithTemplate(reviewResultTestBinding(), reviewResultTestInputs(), data)
	if err != nil {
		t.Fatal(err)
	}
	rebuilt, err := RebuildReviewResult(old)
	if err != nil || !reflect.DeepEqual(rebuilt, old) {
		t.Fatalf("known frozen template: %v", err)
	}
	if _, err := ReviewResultPacketDigest(old); err != nil {
		t.Fatal(err)
	}
	current, err := BuildReviewResult(old.Binding, old.Inputs)
	if err != nil || current.ID == old.ID || current.TemplateSHA == old.TemplateSHA {
		t.Fatalf("new work did not use current template: %v", err)
	}
	old.Draft.Instruction = "replace the original questions"
	if _, err := ReviewResultPacketDigest(old); err == nil {
		t.Fatal("tampered historical questions accepted")
	}
	old.TemplateSHA = strings.Repeat("f", 64)
	if _, err := RebuildReviewResult(old); err == nil {
		t.Fatal("arbitrary author-supplied template accepted")
	}
}

func TestValidateReviewResultAnswerAcceptsOnlyCompleteProfessionalAnswer(t *testing.T) {
	p, err := BuildReviewResult(reviewResultTestBinding(), reviewResultTestInputs())
	if err != nil {
		t.Fatal(err)
	}
	draftBytes, _ := json.Marshal(p.Draft)
	if _, fields := ValidateReviewResultAnswer(p, draftBytes); len(fields) == 0 || fields[0].Pointer != "/answer/claim_results/0/conclusion" {
		t.Fatalf("draft should remain incomplete at the Claim conclusion: %#v", fields)
	}
	var draft map[string]any
	if err := json.Unmarshal(draftBytes, &draft); err != nil {
		t.Fatal(err)
	}
	answer := draft["answer"].(map[string]any)
	results := answer["claim_results"].([]any)
	for _, value := range results {
		row := value.(map[string]any)
		row["conclusion"] = "pass"
		row["observed"] = "Observed the expected behavior."
	}
	answer["verdict"] = "pass"
	complete, _ := json.Marshal(draft)
	got, fields := ValidateReviewResultAnswer(p, complete)
	if len(fields) != 0 {
		t.Fatalf("complete professional answer rejected: %#v", fields)
	}
	var projection map[string]any
	if err := json.Unmarshal(got, &projection); err != nil {
		t.Fatal(err)
	}
	if len(projection) != 8 || projection["verdict"] != "pass" || projection["verification_artifact_digest"] != nil {
		t.Fatalf("validator must return only eight professional fields and preserve null artifact digest: %#v", projection)
	}

	boundDigest := strings.Repeat("e", 64)
	bound := reviewResultTestBinding()
	bound.VerificationArtifactDigest = &boundDigest
	boundPacket, err := BuildReviewResult(bound, reviewResultTestInputs())
	if err != nil {
		t.Fatal(err)
	}
	boundBytes, _ := json.Marshal(boundPacket.Draft)
	var boundDraft map[string]any
	_ = json.Unmarshal(boundBytes, &boundDraft)
	boundAnswer := boundDraft["answer"].(map[string]any)
	if boundAnswer["verification_artifact_digest"] != boundDigest {
		t.Fatalf("known verification artifact digest was not prefilled: %#v", boundAnswer)
	}
	for _, value := range boundAnswer["claim_results"].([]any) {
		row := value.(map[string]any)
		row["conclusion"] = "pass"
		row["observed"] = "Observed the expected behavior."
	}
	boundAnswer["verdict"] = "pass"
	boundBytes, _ = json.Marshal(boundDraft)
	if _, fields := ValidateReviewResultAnswer(boundPacket, boundBytes); len(fields) != 0 {
		t.Fatalf("matching bound verification artifact digest rejected: %#v", fields)
	}
	boundAnswer["verification_artifact_digest"] = strings.Repeat("f", 64)
	boundBytes, _ = json.Marshal(boundDraft)
	if _, fields := ValidateReviewResultAnswer(boundPacket, boundBytes); len(fields) == 0 || fields[0].Pointer != "/answer/verification_artifact_digest" {
		t.Fatalf("answer digest differing from non-null issued digest accepted: %#v", fields)
	}

	mutated := map[string]any{}
	_ = json.Unmarshal(complete, &mutated)
	mutated["title"] = "edited machine title"
	mutatedBytes, _ := json.Marshal(mutated)
	if _, fields := ValidateReviewResultAnswer(p, mutatedBytes); len(fields) == 0 || fields[0].Pointer != "/" {
		t.Fatalf("edited machine question accepted: %#v", fields)
	}

	unknown := bytes.Replace(complete, []byte(`"verdict":"pass"`), []byte(`"verdict":"pass","result_id":"forged"`), 1)
	if _, fields := ValidateReviewResultAnswer(p, unknown); len(fields) == 0 {
		t.Fatal("unknown professional field accepted")
	}
	duplicate := bytes.Replace(complete, []byte(`"title":"S07 验证 · ReviewResult"`), []byte(`"title":"S07 验证 · ReviewResult","title":"forged"`), 1)
	if _, fields := ValidateReviewResultAnswer(p, duplicate); len(fields) == 0 {
		t.Fatal("duplicate machine field accepted")
	}
}

func TestBuildReviewResultRejectsClaimDriftAndInvalidBinding(t *testing.T) {
	binding := reviewResultTestBinding()
	inputs := reviewResultTestInputs()
	inputs.Claims = json.RawMessage(strings.Replace(string(inputs.Claims), "claim-reset", "claim-other", 1))
	if _, err := BuildReviewResult(binding, inputs); err == nil {
		t.Fatal("Claim question outside the Assignment binding accepted")
	}
	inputs = reviewResultTestInputs()
	binding.ClaimIDs = []string{"claim-reset", "claim-login"}
	if _, err := BuildReviewResult(binding, inputs); err == nil {
		t.Fatal("Claim order different from the frozen Assignment accepted")
	}
	binding = reviewResultTestBinding()
	binding.OwnerAgentID = "../another-owner"
	if _, err := BuildReviewResult(binding, reviewResultTestInputs()); err == nil {
		t.Fatal("unsafe reviewer identity accepted")
	}
	inputs = reviewResultTestInputs()
	inputs.Claims = json.RawMessage(`[{"claim_id":"claim-login","claim_id":"claim-reset"}]`)
	if _, err := BuildReviewResult(reviewResultTestBinding(), inputs); err == nil {
		t.Fatal("duplicate JSON key in frozen input accepted")
	}
}

func mustReviewPacketJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
