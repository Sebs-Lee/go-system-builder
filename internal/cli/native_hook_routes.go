package cli

import (
	"context"
	"io"
	"path/filepath"
	"sync"

	"github.com/entroforge/go-system-builder/internal/metrics"
	"github.com/entroforge/go-system-builder/internal/policy"
	"github.com/entroforge/go-system-builder/internal/runtime"
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
