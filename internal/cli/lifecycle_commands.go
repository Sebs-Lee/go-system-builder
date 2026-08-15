package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/semantic"
	"github.com/entroforge/go-system-builder/internal/transition"
)

// lifecycleSnapshot loads the runtime for a lifecycle verb command. The
// returned guidance string explains why the verb cannot run when the
// snapshot is unavailable.
func lifecycleSnapshot(root string) (runtime.Snapshot, string) {
	statePath := filepath.Join(root, ".claude", "loop-state.json")
	journalPath := filepath.Join(root, ".claude", "loop-events.jsonl")
	snapshot, err := runtime.NewWriter(statePath, journalPath, root, semantic.RuntimeCandidateValidator{}).Snapshot()
	if err != nil {
		return runtime.Snapshot{}, "no readable runtime — bind a REQ first (req bind --approved-by <you>)"
	}
	return snapshot, ""
}

func snapshotLifecycle(snapshot runtime.Snapshot) (state string, phase any) {
	lifecycle, _ := snapshot.State["lifecycle"].(map[string]any)
	state, _ = lifecycle["state"].(string)
	return state, lifecycle["phase"]
}

// writeDecisionArtifact persists the human decision record that backs the
// human_decision evidence for a lifecycle verb. The machine derives the
// envelope; the human supplies only the decision and identity.
func writeDecisionArtifact(root, name string, payload map[string]any) (string, error) {
	dir := filepath.Join(root, ".claude", "decisions")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create decisions dir: %w", err)
	}
	payload["occurred_at"] = time.Now().UTC().Format(time.RFC3339Nano)
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", err
	}
	rel := filepath.ToSlash(filepath.Join(".claude", "decisions", name+".json"))
	if err := os.WriteFile(filepath.Join(root, rel), append(data, '\n'), 0o644); err != nil {
		return "", err
	}
	return rel, nil
}

// registerDecisionEvidence writes the decision artifact and records it as
// current human_decision evidence, returning the evidence id and the
// revision the follow-up transition must expect.
func registerDecisionEvidence(root string, snapshot runtime.Snapshot, id, decision, reason, approvedBy, scopeRef string) (string, int, error) {
	runtimeID, _ := snapshot.State["runtime_id"].(string)
	payload := map[string]any{
		"decision":    decision,
		"reason":      reason,
		"approved_by": approvedBy,
		"runtime_id":  runtimeID,
		"revision":    snapshot.Revision,
	}
	if scopeRef != "" {
		payload["scope"] = scopeRef
	}
	rel, err := writeDecisionArtifact(root, id, payload)
	if err != nil {
		return "", 0, err
	}
	_, err = runtime.RecordEvidence(root,
		filepath.Join(root, ".claude", "loop-state.json"),
		filepath.Join(root, ".claude", "loop-events.jsonl"),
		runtime.EvidenceRequest{
			ExpectedRevision: snapshot.Revision,
			ID:               id,
			Kind:             "human_decision",
			Path:             rel,
			ProducedBy:       []string{approvedBy},
			ScopeRefs:        []string{scopeRef},
			Validator:        semantic.RuntimeCandidateValidator{},
		})
	if err != nil {
		return "", 0, fmt.Errorf("register decision evidence: %w", err)
	}
	return id, snapshot.Revision + 1, nil
}

func runRuntimePause(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("runtime pause", flag.ContinueOnError)
	flags.SetOutput(stderr)
	bindUsage(flags, "runtime pause")
	root := flags.String("root", ".", "repository root")
	reason := flags.String("reason", "", "why the loop is paused (recorded in the human decision artifact)")
	approvedBy := flags.String("approved-by", "", "human approver identity")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *approvedBy == "" {
		if identity := detectGitIdentity(*root); identity != "" {
			fmt.Fprintf(stderr, "runtime pause requires --approved-by (detected git identity %q; rerun with --approved-by %q)\n", identity, identity)
		} else {
			fmt.Fprintln(stderr, "runtime pause requires --approved-by <human identity>")
		}
		return 2
	}
	snapshot, guidance := lifecycleSnapshot(*root)
	if guidance != "" {
		fmt.Fprintln(stderr, "runtime pause: "+guidance)
		return 1
	}
	state, phase := snapshotLifecycle(snapshot)
	switch state {
	case "paused":
		fmt.Fprintln(stderr, "runtime pause: already paused — resolve it with `runtime resume`, or amend/abort from the paused state")
		return 1
	case "release_authorized", "aborted":
		fmt.Fprintf(stderr, "runtime pause: runtime is terminal (%s) — use `runtime rollover` to archive and start fresh\n", state)
		return 1
	case "inactive":
		fmt.Fprintln(stderr, "runtime pause: nothing is bound — bind a REQ first (req bind --approved-by <you>)")
		return 1
	}
	evID, nextRev, err := registerDecisionEvidence(*root, snapshot,
		fmt.Sprintf("hd-pause-r%d", snapshot.Revision+1), "user_pause_requested", *reason, *approvedBy,
		fmt.Sprintf("runtime_pause:%s@%d", snapshot.State["runtime_id"], snapshot.Revision+1))
	if err != nil {
		fmt.Fprintln(stderr, formatFailure("runtime pause", err))
		return 1
	}
	next, err := transition.Apply(*root,
		filepath.Join(*root, ".claude", "loop-state.json"),
		filepath.Join(*root, ".claude", "loop-events.jsonl"),
		transition.Request{
			TransitionID: "GTR-001", ExpectedRevision: nextRev, Actor: "user",
			Evidence: map[string]string{
				"human_decision_record": evID,
				"pause_record":          "generated:pause_checkpoint",
			},
			OccurredAt: time.Now().UTC(),
		})
	if err != nil {
		fmt.Fprintln(stderr, formatFailure("runtime pause", err))
		return 1
	}
	_ = next
	fmt.Fprintf(stdout, "paused from %v (reason recorded; approved-by %s)\n", cursorLabel(state, phase), *approvedBy)
	fmt.Fprintln(stdout, "resume: loop-harness runtime resume --approved-by <you>   (baseline drift on resume routes to amendment)")
	return 0
}

func runRuntimeResume(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("runtime resume", flag.ContinueOnError)
	flags.SetOutput(stderr)
	bindUsage(flags, "runtime resume")
	root := flags.String("root", ".", "repository root")
	approvedBy := flags.String("approved-by", "", "human approver identity")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *approvedBy == "" {
		if identity := detectGitIdentity(*root); identity != "" {
			fmt.Fprintf(stderr, "runtime resume requires --approved-by (detected git identity %q; rerun with --approved-by %q)\n", identity, identity)
		} else {
			fmt.Fprintln(stderr, "runtime resume requires --approved-by <human identity>")
		}
		return 2
	}
	snapshot, guidance := lifecycleSnapshot(*root)
	if guidance != "" {
		fmt.Fprintln(stderr, "runtime resume: "+guidance)
		return 1
	}
	state, _ := snapshotLifecycle(snapshot)
	if state != "paused" {
		fmt.Fprintf(stderr, "runtime resume: runtime is %q, not paused — nothing to resume\n", state)
		return 1
	}
	evID, nextRev, err := registerDecisionEvidence(*root, snapshot,
		fmt.Sprintf("hd-resume-r%d", snapshot.Revision+1), "human_resume_approved", "", *approvedBy,
		fmt.Sprintf("runtime_resume:%s@%d", snapshot.State["runtime_id"], snapshot.Revision+1))
	if err != nil {
		fmt.Fprintln(stderr, formatFailure("runtime resume", err))
		return 1
	}
	next, err := transition.Apply(*root,
		filepath.Join(*root, ".claude", "loop-state.json"),
		filepath.Join(*root, ".claude", "loop-events.jsonl"),
		transition.Request{
			TransitionID: "TR-019", ExpectedRevision: nextRev, Actor: "user",
			Evidence: map[string]string{
				"human_decision_record": evID,
				"pause_record":          "generated:pause_checkpoint",
			},
			OccurredAt: time.Now().UTC(),
		})
	if err != nil {
		if strings.Contains(err.Error(), "baselines_unchanged") {
			fmt.Fprintln(stderr, "runtime resume: baseline drifted while paused — resume is refused; amend the baseline instead (req amend)")
			return 1
		}
		fmt.Fprintln(stderr, formatFailure("runtime resume", err))
		return 1
	}
	fmt.Fprintf(stdout, "resumed to %v (pause checkpoint verified and cleared)\n", cursorLabel(next.State["lifecycle"].(map[string]any)["state"].(string), next.State["lifecycle"].(map[string]any)["phase"]))
	return 0
}

func cursorLabel(state string, phase any) string {
	if phase == nil || phase == "" {
		return state
	}
	return fmt.Sprintf("%s.%v", state, phase)
}
