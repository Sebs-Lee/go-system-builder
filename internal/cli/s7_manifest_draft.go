// s7_manifest_draft.go implements `loop-harness s7 manifest-draft`: a
// read-only scaffold of the reviewer team-manifest for one ReviewPlan
// Assignment (L3-S7 §13 known-friction: the 19-field manifest was
// handwritten). The draft pre-fills every control-plane fact and marks the
// rest TODO(planner); `runtime register-workgroup` validates the result
// normally.
package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/entroforge/go-system-builder/internal/review"
	"github.com/entroforge/go-system-builder/internal/runtime"
)

// runS7ManifestDraft generates the manifest draft. It performs no writes to
// the control plane; --out (or stdout) is the only output.
func runS7ManifestDraft(root, assignmentID, out string, stdout io.Writer) int {
	statePath := filepath.Join(root, ".claude", "loop-state.json")
	journalPath := filepath.Join(root, ".claude", "loop-events.jsonl")
	snapshot, err := runtime.NewStore(statePath, journalPath).Snapshot()
	if err != nil {
		fmt.Fprintf(stdout, "read runtime: %v\n", err)
		return 1
	}
	draft, notes, err := review.DraftManifest(root, snapshot.State, assignmentID)
	if err != nil {
		fmt.Fprintf(stdout, "manifest draft: %v\n", err)
		return 1
	}
	data, err := json.MarshalIndent(draft, "", "  ")
	if err != nil {
		fmt.Fprintf(stdout, "encode draft: %v\n", err)
		return 1
	}
	if out != "" {
		if err := os.WriteFile(out, append(data, '\n'), 0o644); err != nil {
			fmt.Fprintf(stdout, "write draft: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "draft team manifest written to %s\n", out)
	} else {
		fmt.Fprintln(stdout, string(data))
	}
	for _, note := range notes {
		fmt.Fprintf(stdout, "note: %s\n", note)
	}
	fmt.Fprintln(stdout, "next: Main chooses agent_id (Reviewer dispatch name string), platform_team_id (logical dispatch group label string) and professional responsibility choices; no opaque platform identity lookup is required. registration rejects authoring placeholders. validation is draft preparation metadata, not evidence of a professional PASS.")
	if out == "" {
		fmt.Fprintln(stdout, "preparation gap: save the manifest with --out before selecting a registration command.")
		return 0
	}
	rootAbs, rootErr := filepath.Abs(root)
	manifestAbs, manifestErr := filepath.Abs(out)
	if rootErr != nil || manifestErr != nil {
		fmt.Fprintln(stdout, "preparation gap: cannot resolve the actual root or saved manifest path.")
		return 0
	}
	candidates := 0
	for _, doc := range draft.Documents {
		if !strings.HasPrefix(doc.ID, "TASK-") || doc.Path == "" {
			continue
		}
		candidates++
		fmt.Fprintf(stdout, "Main TASK choice %s (%s), after completing the draft: %s runtime register-workgroup --root %s --manifest %s --task-id %s --task %s\n", doc.ID, doc.Path, s7CaptureShellQuote(filepath.Join(rootAbs, ".claude/bin/loop-harness")), s7CaptureShellQuote(rootAbs), s7CaptureShellQuote(manifestAbs), s7CaptureShellQuote(doc.ID), s7CaptureShellQuote(doc.Path))
	}
	if candidates == 0 {
		fmt.Fprintln(stdout, "preparation gap: no current-generation registered TASK document with an ID/path; Main must restore the TASK source before registration.")
	}
	return 0
}
