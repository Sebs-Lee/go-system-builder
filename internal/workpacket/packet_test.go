package workpacket

import (
	"encoding/json"
	"github.com/entroforge/go-system-builder/internal/schema"
	"strings"
	"testing"
)

func testBinding() Binding {
	return Binding{RuntimeID: "loop-REQ-039", Generation: 1, StageEntry: 7,
		Owner: "reviewer-spec", AssignmentID: "assignment-spec", AssignmentSHA: strings.Repeat("a", 64),
		Responsibility: "DV-SPEC-CONSISTENCY", Attempt: 1}
}

func TestS5PacketAndDraftSchemas(t *testing.T) {
	p, err := BuildS5(testBinding(), []Reference{{ID: "REQ-039", Path: "docs/requirements/REQ-039.md", Version: "v1", SHA256: strings.Repeat("b", 64)}})
	if err != nil {
		t.Fatal(err)
	}
	for name, value := range map[string]any{"work-packet.schema.json": p, "stage-answer.schema.json": p.Draft} {
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		if err := schema.NewEmbeddedValidator().ValidateBytes(name, data); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}
}

func TestS5PacketIsStableAndOwnerIsolated(t *testing.T) {
	inputs := []Reference{{ID: "REQ-039", Path: "docs/requirements/REQ-039.md", Version: "v1", SHA256: strings.Repeat("b", 64)}}
	p, err := BuildS5(testBinding(), inputs)
	if err != nil {
		t.Fatal(err)
	}
	again, err := BuildS5(testBinding(), inputs)
	if err != nil {
		t.Fatal(err)
	}
	if p.ID != again.ID || p.AnswerPath != again.AnswerPath {
		t.Fatal("unchanged responsibility must not reissue")
	}
	binding := testBinding()
	binding.Owner = "reviewer-task"
	binding.Responsibility = "DV-TASK-EXECUTABILITY"
	other, err := BuildS5(binding, inputs)
	if err != nil {
		t.Fatal(err)
	}
	if p.AnswerPath == other.AnswerPath {
		t.Fatal("shared checkout answers collide")
	}
	binding = testBinding()
	binding.StageEntry++
	reentry, err := BuildS5(binding, inputs)
	if err != nil {
		t.Fatal(err)
	}
	if p.ID == reentry.ID {
		t.Fatal("stage reentry must not reuse an old packet")
	}
	if p.Draft.Answer.Conclusion != nil || !strings.Contains(p.Draft.Instruction, "answer") {
		t.Fatal("draft must be answerable without protocol lookup")
	}
}

func TestS5RejectsUnsafeIdentityAndUnknownResponsibility(t *testing.T) {
	inputs := []Reference{{ID: "REQ-039", Path: "docs/requirements/REQ-039.md", Version: "v1", SHA256: strings.Repeat("b", 64)}}
	for _, owner := range []string{"../other", "TODO(planner):owner", "", "x/y"} {
		binding := testBinding()
		binding.Owner = owner
		if _, err := BuildS5(binding, inputs); err == nil {
			t.Fatalf("unsafe owner accepted: %q", owner)
		}
	}
	binding := testBinding()
	binding.Responsibility = "QA-UNIT"
	if _, err := BuildS5(binding, inputs); err == nil {
		t.Fatal("unregistered S5 responsibility accepted")
	}
}

func TestS5AnswerCompletenessAndFrozenQuestionnaire(t *testing.T) {
	p, err := BuildS5(testBinding(), []Reference{{ID: "REQ-039", Path: "docs/requirements/REQ-039.md", Version: "v1", SHA256: strings.Repeat("b", 64)}})
	if err != nil {
		t.Fatal(err)
	}
	pass := "pass"
	fix := "fix_required"
	for _, tc := range []struct {
		name     string
		answer   Answer
		complete bool
		pointer  string
	}{
		{"draft", Answer{Findings: []Finding{}}, false, "/answer/conclusion"},
		{"pass", Answer{Conclusion: &pass, Findings: []Finding{}}, true, ""},
		{"fix needs observation", Answer{Conclusion: &fix, Findings: []Finding{{Path: "docs/requirements/REQ-039.md", Expected: "deny"}}}, false, "/answer/findings/0/observed"},
		{"fix", Answer{Conclusion: &fix, Findings: []Finding{{Path: "docs/requirements/REQ-039.md#errors", Expected: "deny", Observed: "unspecified"}}}, true, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			draft := p.Draft
			draft.Answer = tc.answer
			data, err := json.Marshal(draft)
			if err != nil {
				t.Fatal(err)
			}
			_, issues := ValidateAnswer(p, data)
			if (len(issues) == 0) != tc.complete {
				t.Fatalf("issues: %+v", issues)
			}
			if tc.pointer != "" {
				found := false
				for _, issue := range issues {
					if issue.Pointer == tc.pointer {
						found = true
					}
				}
				if !found {
					t.Fatalf("missing pointer %s: %+v", tc.pointer, issues)
				}
			}
		})
	}
	for _, change := range []func(*Draft){
		func(d *Draft) { d.Title = "a different responsibility" },
		func(d *Draft) { d.Read[0].Path = "docs/foreign.md" },
	} {
		data, err := json.Marshal(p.Draft)
		if err != nil {
			t.Fatal(err)
		}
		var draft Draft
		if err := json.Unmarshal(data, &draft); err != nil {
			t.Fatal(err)
		}
		change(&draft)
		draft.Answer = Answer{Conclusion: &pass, Findings: []Finding{}}
		data, err = json.Marshal(draft)
		if err != nil {
			t.Fatal(err)
		}
		if _, issues := ValidateAnswer(p, data); len(issues) == 0 {
			t.Fatal("machine questionnaire changed without rejection")
		}
	}
}
