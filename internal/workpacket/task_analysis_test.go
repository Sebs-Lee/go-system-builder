package workpacket

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestTaskAnalysisIdentityAndLegacyProjection(t *testing.T) {
	base, err := BuildS5(testBinding(), []Reference{{ID: "REQ-039", Path: "docs/requirements/REQ-039.md", Version: "v1", SHA256: strings.Repeat("b", 64)}})
	if err != nil {
		t.Fatal(err)
	}
	legacy, err := RebuildS5(base)
	if err != nil || !reflect.DeepEqual(base, legacy) {
		t.Fatal("legacy packet changed", err)
	}
	sha := strings.Repeat("c", 64)
	ref := &TaskAnalysisRef{Path: ".claude/work-packets/analysis/" + sha + ".json", SHA256: sha, ParserVersion: "1.0.0"}
	bound, err := BindTaskAnalysis(base, ref)
	if err != nil {
		t.Fatal(err)
	}
	if bound.ID == base.ID || !reflect.DeepEqual(bound.Inputs, base.Inputs) || !reflect.DeepEqual(bound.Draft.Answer, base.Draft.Answer) || len(bound.Draft.Read) != len(base.Draft.Read)+1 {
		t.Fatal("analysis changed authority or did not bind identity")
	}
	rebuilt, err := RebuildS5(bound)
	if err != nil || !reflect.DeepEqual(bound, rebuilt) {
		t.Fatal("frozen packet not deterministically rebuilt", err)
	}
	b := bound.Binding
	b.Attempt = 2
	b.PreviousSubmissionKey = strings.Repeat("d", 64)
	b.RevisionReason = "Clarify an original assertion"
	next, err := RebuildS5ForBinding(b, bound)
	if err != nil || next.ID == bound.ID || !reflect.DeepEqual(next.TaskAnalysis, bound.TaskAnalysis) {
		t.Fatal("revision did not retain frozen analysis", err)
	}
	ref.Path = "docs/unrelated.json"
	if _, err := BindTaskAnalysis(base, ref); err == nil {
		t.Fatal("arbitrary analysis path accepted")
	}
}

func TestReadTaskAnalysisAcceptsOptionalStructuralCheckAndLeavesLegacyBytesAlone(t *testing.T) {
	for _, withStructuralCheck := range []bool{false, true} {
		name := "legacy"
		if withStructuralCheck {
			name = "current"
		}
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			base, err := BuildS5(testBinding(), []Reference{{ID: "REQ-039", Path: "docs/requirements/REQ-039.md", Version: "v1", SHA256: strings.Repeat("b", 64)}})
			if err != nil {
				t.Fatal(err)
			}
			artifact := TaskAnalysisArtifact{
				SchemaVersion:     "1.0.0",
				ParserVersion:     "1.0.0",
				Inputs:            base.Inputs,
				TaskContexts:      json.RawMessage("[]"),
				GranularityAdvice: json.RawMessage("[]"),
			}
			want := &TaskStructuralCheck{
				Tasks:          3,
				Cancelled:      1,
				ClausesTotal:   8,
				ClausesCovered: 7,
				Problems:       []string{"TASK-003: missing clause"},
			}
			if withStructuralCheck {
				artifact.StructuralCheck = want
			}
			data, err := json.Marshal(artifact)
			if err != nil {
				t.Fatal(err)
			}
			digest := Digest(data)
			path := ".claude/work-packets/analysis/" + digest + ".json"
			if err := os.MkdirAll(filepath.Dir(filepath.Join(root, path)), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, path), data, 0o600); err != nil {
				t.Fatal(err)
			}
			bound, err := BindTaskAnalysis(base, &TaskAnalysisRef{Path: path, SHA256: digest, ParserVersion: "1.0.0"})
			if err != nil {
				t.Fatal(err)
			}
			got, err := ReadTaskAnalysis(root, bound)
			if err != nil {
				t.Fatal(err)
			}
			if withStructuralCheck && !reflect.DeepEqual(got.StructuralCheck, want) {
				t.Fatalf("structural check = %+v, want %+v", got.StructuralCheck, want)
			}
			if !withStructuralCheck && got.StructuralCheck != nil {
				t.Fatalf("legacy artifact unexpectedly gained a structural check: %+v", got.StructuralCheck)
			}
			after, err := os.ReadFile(filepath.Join(root, path))
			if err != nil || !reflect.DeepEqual(after, data) {
				t.Fatalf("reading analysis changed frozen bytes: err=%v", err)
			}
		})
	}
}
