package review

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// WS4c: the workspace digest binds a cold-start write surface end to end.
func TestWorkspaceDigestTracksDrift(t *testing.T) {
	root := t.TempDir()
	ws := "e2e-workspace/plan-1"
	empty, err := WorkspaceDigest(root, ws)
	if err != nil {
		t.Fatal(err)
	}
	abs := filepath.Join(root, ws)
	if err := os.MkdirAll(abs, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(abs, "flow.spec.ts"), []byte("spec v1"), 0o644); err != nil {
		t.Fatal(err)
	}
	d1, err := WorkspaceDigest(root, ws)
	if err != nil {
		t.Fatal(err)
	}
	if d1 == empty {
		t.Fatal("digest must change once a spec lands")
	}
	if err := os.WriteFile(filepath.Join(abs, "flow.spec.ts"), []byte("spec v2"), 0o644); err != nil {
		t.Fatal(err)
	}
	d2, _ := WorkspaceDigest(root, ws)
	if d2 == d1 {
		t.Fatal("digest must track content drift")
	}
}

func TestPrepareWorkspaceRejectsEscape(t *testing.T) {
	plan := &Plan{}
	out := "/tmp/outside"
	plan.VerificationArtifactWorkspace = &out
	if _, err := prepareVerificationWorkspace(t.TempDir(), plan); err == nil || !strings.Contains(err.Error(), "repository-relative") {
		t.Fatalf("outside workspace must fail, got %v", err)
	}
	// A relative path outside e2e-workspace/ fails the surface rule.
	out2 := "docs/reports/e2e"
	plan.VerificationArtifactWorkspace = &out2
	if _, err := prepareVerificationWorkspace(t.TempDir(), plan); err == nil || !strings.Contains(err.Error(), "e2e-workspace/") {
		t.Fatalf("off-surface workspace must fail, got %v", err)
	}
	ws := "e2e-workspace/plan-1"
	plan.VerificationArtifactWorkspace = &ws
	digest, err := prepareVerificationWorkspace(t.TempDir(), plan)
	if err != nil {
		t.Fatalf("in-repo workspace: %v", err)
	}
	if digest == "" {
		t.Fatal("registered workspace must carry a digest")
	}
}

// WS4d: the redaction gate and the buffer merge.
func TestSanitizeCaptureRejectsSecrets(t *testing.T) {
	cases := []CaptureStep{
		{Action: "login", Observed: "password: hunter2 accepted"},
		{Action: "call api", Observed: "Authorization: Bearer abcdef0123456789"},
		{Action: "set config", Observed: "api_key = sk-1234567890"},
	}
	for _, step := range cases {
		if err := SanitizeCapture(step); err == nil {
			t.Fatalf("secret-carrying capture must be rejected: %+v", step)
		}
	}
	if err := SanitizeCapture(CaptureStep{Action: "open page", Observed: "title visible"}); err != nil {
		t.Fatalf("clean capture must pass: %v", err)
	}
}

func TestCaptureBufferRoundTripAndMerge(t *testing.T) {
	root := t.TempDir()
	path := CaptureFile(root, "rt-x", 1, "assignment-qa-1")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	lines := `{"sequence":1,"action":"open /login","observed":"form rendered","evidence_refs":["shot-1.png"]}
{"sequence":2,"action":"submit empty form","observed":"validation error","evidence_refs":["shot-2.png"]}
`
	if err := os.WriteFile(path, []byte(lines), 0o644); err != nil {
		t.Fatal(err)
	}
	steps := LoadCaptureSteps(path)
	if len(steps) != 2 || steps[1].Observed != "validation error" {
		t.Fatalf("buffer round trip wrong: %+v", steps)
	}

	findings := []Finding{
		{FindingID: "finding-1", Encounter: Encounter{}},
		{FindingID: "finding-2", Encounter: Encounter{Timeline: []TimelineStep{{Sequence: 1, Action: "manual"}}}},
	}
	mergeCapturedTimeline(findings, steps)
	if len(findings[0].Encounter.Timeline) != 2 {
		t.Fatalf("empty timeline must absorb the buffer, got %d", len(findings[0].Encounter.Timeline))
	}
	if len(findings[1].Encounter.Timeline) != 1 {
		t.Fatal("a reviewer-written timeline must never be rewritten")
	}
}
