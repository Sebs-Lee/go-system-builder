package fileview

import (
	"context"
	"strings"
	"sync"

	"github.com/entroforge/go-system-builder/internal/metrics"
)

type readCacheKey struct{}
type immutableReads struct {
	mu   sync.Mutex
	data map[string][]byte
}

// WithReadCache scopes reuse to one request and its derived contexts. Only
// successful reads of pinned Git objects enter it. Disk/evidence reads, current
// Runtime facts, source-rule resolution and Verify remain live checks.
func WithReadCache(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if _, present := ctx.Value(readCacheKey{}).(*immutableReads); present {
		return ctx
	}
	return context.WithValue(ctx, readCacheKey{}, &immutableReads{data: map[string][]byte{}})
}

// Only View's resolved commit/tree and blob reads call this function; resolving
// a mutable branch reference must always use gitContext directly.
func immutableGit(ctx context.Context, root string, args ...string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	cache, _ := ctx.Value(readCacheKey{}).(*immutableReads)
	if cache == nil {
		return gitContext(ctx, root, args...)
	}
	key := root + "\x00" + strings.Join(args, "\x00")
	cache.mu.Lock()
	data, found := cache.data[key]
	cache.mu.Unlock()
	if found {
		defer metrics.Span(ctx, "source/git_cache_hit")()
		return append([]byte(nil), data...), nil
	}
	data, err := gitContext(ctx, root, args...)
	if err != nil {
		return nil, err
	}
	cache.mu.Lock()
	cache.data[key] = append([]byte(nil), data...)
	cache.mu.Unlock()
	return data, nil
}
