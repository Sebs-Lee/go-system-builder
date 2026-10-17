package mainreceipt

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/entroforge/go-system-builder/internal/assignmentrecord"
)

func TestPreservedMergeIsUnverifiedWithoutBecomingSuccessfulProof(t *testing.T) {
	for _, name := range []string{"bound preserved merge", "wrong Result", "wrong execution", "wrong completion field", "wrong Main root", "wrong branch", "missing merge", "wrong merge base", "verified without tested head"} {
		t.Run(name, func(t *testing.T) {
			f := newReaderFixture(t)
			cp := f.checkpoint
			cp.State, cp.TestedHead, cp.VerifiedAt = "preserved", "", ""
			cp.ReportPath, cp.ReportSHA = "", ""
			cp.CanonicalMainFacts.TestedHead = ""
			cp.CanonicalMainFacts.Checks = nil
			switch name {
			case "wrong Result":
				cp.CanonicalMainFacts.Binding.ResultSHA256 = strings.Repeat("f", 64)
			case "wrong execution":
				cp.CanonicalMainFacts.ExecutionGeneration++
			case "wrong completion field":
				cp.ReportPath = "wrong-result.json"
			case "wrong Main root":
				cp.CanonicalMainFacts.MainRoot = t.TempDir()
			case "wrong branch":
				cp.TargetBranch, cp.CanonicalMainFacts.TargetBranch = "other", "other"
			case "missing merge":
				cp.MergeCommit, cp.CanonicalMainFacts.MergeCommit = strings.Repeat("f", 40), strings.Repeat("f", 40)
			case "wrong merge base":
				cp.MergeBase, cp.CanonicalMainFacts.MergeBase = f.sourceHead, f.sourceHead
			case "verified without tested head":
				cp.State = "verified"
			}
			writeCheckpointFixture(t, f, cp)
			path := assignmentrecord.IntegrationCheckpointPath(f.root, f.input.Record.Reference())
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			proof, err := Verify(context.Background(), f.input)
			if err == nil || proof.CheckpointSHA256() != "" || proof.MainFacts().TestedHead != "" {
				t.Fatalf("unverified attempt returned successful proof: %+v, err=%v", proof, err)
			}
			if got, want := errors.Is(err, ErrNotVerified), name == "bound preserved merge"; got != want {
				t.Fatalf("ErrNotVerified=%t, want %t; err=%v", got, want, err)
			}
			after, readErr := os.ReadFile(path)
			if readErr != nil || !bytes.Equal(before, after) {
				t.Fatalf("read-only verification changed preserved checkpoint: %v", readErr)
			}
		})
	}
}
