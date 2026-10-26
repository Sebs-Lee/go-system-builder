package cli

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"path/filepath"
	"sync"

	"github.com/entroforge/go-system-builder/internal/metrics"
	"github.com/entroforge/go-system-builder/internal/policy"
	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/workspace"
)

type nativeRoutingKey struct{}

// Only the initial routing reads share this snapshot. Domain consumers and
// Writers retain their own current reads, including after a successful receipt.
type nativeRoutingRead struct {
	root string
	once sync.Once
	snap runtime.Snapshot
	err  error
}

func nativeHookRoutes(root string, in policy.Input, out, diagnostic io.Writer) (bool, int) {
	defer metrics.Span(contextForHook(in), "hook/native_routes")()
	// The derived input never escapes the routing chain. All false branches
	// precede business writes; a receiving branch always returns handled=true.
	in.Context = context.WithValue(contextForHook(in), nativeRoutingKey{}, &nativeRoutingRead{root: root})
	if nativePendingRuntimeMainContext(root, in, out) {
		return true, 0
	}
	if handled, code := nativeMainReturnHook(root, in, out); handled {
		return true, code
	}
	if handled, code := nativeS6Hook(root, in, out, diagnostic); handled {
		return true, code
	}
	if handled, code := nativeS9Hook(root, in, out, diagnostic); handled {
		return true, code
	}
	return nativeReviewHook(root, in, out, diagnostic)
}

func nativeRoutingSnapshot(root string, in policy.Input) (runtime.Snapshot, error) {
	ctx := contextForHook(in)
	if err := ctx.Err(); err != nil {
		return runtime.Snapshot{}, err
	}
	read := func() (runtime.Snapshot, error) {
		defer metrics.Span(ctx, "hook/native_routes/initial_snapshot")()
		return runtime.NewStore(filepath.Join(root, ".claude/loop-state.json"), filepath.Join(root, ".claude/loop-events.jsonl")).WithContext(ctx).Snapshot()
	}
	if cached, ok := ctx.Value(nativeRoutingKey{}).(*nativeRoutingRead); ok && cached.root == root {
		cached.once.Do(func() { cached.snap, cached.err = read() })
		if err := ctx.Err(); err != nil {
			return runtime.Snapshot{}, err
		}
		return cached.snap, cached.err
	}
	return read()
}

// Pending is a typed Runtime boundary, not a stage/owner/completion fact.
// This Main read names the existing Writer command and never invokes it.
func nativePendingRuntimeMainContext(root string, in policy.Input, out io.Writer) bool {
	if !mainQuestionnaireViewer(in) {
		return false
	}
	switch in.Event {
	case "PreToolUse", "SessionStart":
	case "PostToolUse":
		if in.ToolName != "Agent" && in.ToolName != "Task" {
			return false
		}
	default:
		return false
	}
	if workspace.RequireMain(root) != nil {
		return false
	}
	if in.CWD != "" {
		current, err := workspace.Canonical(in.CWD)
		if err != nil {
			return false
		}
		main, err := workspace.Canonical(root)
		if err != nil || current != main {
			return false
		}
	}
	if _, err := nativeRoutingSnapshot(root, in); !errors.Is(err, runtime.ErrPendingRuntimeOperation) {
		return false
	}
	command := s7CaptureShellQuote(filepath.Join(root, ".claude/bin/loop-harness")) + " runtime stage-answer recover --root " + s7CaptureShellQuote(root)
	message := "Runtime reports a pending durable write. Main resumes the original paired Writer operation with: " + command + ". This read leaves the pending pair unchanged. The command returns the original consumer feedback and current next action; pending alone does not establish a professional stage, owner, or completion."
	_ = json.NewEncoder(out).Encode(map[string]any{"hookSpecificOutput": map[string]any{"hookEventName": in.Event, "additionalContext": message}})
	return true
}
