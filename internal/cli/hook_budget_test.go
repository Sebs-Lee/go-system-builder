package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/entroforge/go-system-builder/internal/policy"
)

func TestHookBudgetTimeoutKeepsListenerNonBlocking(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX fake Hook process")
	}
	script := filepath.Join(t.TempDir(), "hook")
	os.WriteFile(script, []byte("#!/bin/sh\nprintf '{\"decision\":\"allow\"}'\nsleep 30\n"), 0700)
	for _, tool := range []string{"Write", "Bash", "Read"} {
		t.Run(tool, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
			defer cancel()
			var out, errout bytes.Buffer
			start := time.Now()
			code := runBudgetedHook(ctx, script, nil, strings.NewReader(`{"hook_event_name":"PreToolUse","tool_name":"`+tool+`"}`), &out, &errout)
			want := 0
			if code != want || out.Len() != 0 || !strings.Contains(errout.String(), "internal budget") {
				t.Fatalf("code=%d stdout=%s stderr=%s", code, &out, &errout)
			}
			if time.Since(start) > 2*time.Second {
				t.Fatal("deadline failed to bound process")
			}
		})
	}
}

func TestHookFailurePreservesOriginalDiagnosticAndDiscardsPartialSuccess(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX fake Hook process")
	}
	script := filepath.Join(t.TempDir(), "hook")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nprintf '{\"accepted\":true}'\nprintf 'Main owns recovery: original Record source is missing\\n' >&2\nexit 1\n"), 0700); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var out, diagnostic bytes.Buffer
	code := runBudgetedHook(ctx, script, nil, strings.NewReader(`{"hook_event_name":"PreToolUse","tool_name":"Read"}`), &out, &diagnostic)
	if code != 0 || out.Len() != 0 || !strings.Contains(diagnostic.String(), "original Record source is missing") || !strings.Contains(diagnostic.String(), "original receipt") || !strings.Contains(diagnostic.String(), "does not prove no commit") {
		t.Fatalf("incomplete Hook hid its cause or delivered partial success: code=%d stdout=%s stderr=%s", code, &out, &diagnostic)
	}
}

func TestHookContextCarriesCancellationWithoutAuthorBudgetFields(t *testing.T) {
	var request policy.Input
	if err := json.Unmarshal([]byte(`{"session_id":"actual-session","Context":{"timeout":999},"hook_event_name":"PreToolUse"}`), &request); err != nil {
		t.Fatal(err)
	}
	if request.Context != nil {
		t.Fatal("platform/author payload supplied an internal Hook lifetime")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	request.Context = ctx
	derived := contextForHook(request)
	want, _ := ctx.Deadline()
	if got, ok := derived.Deadline(); !ok || got != want {
		t.Fatal("adding session context discarded the Hook deadline")
	}
	data, err := json.Marshal(request)
	if err != nil || bytes.Contains(data, []byte("Context")) || bytes.Contains(data, []byte("timeout")) {
		t.Fatalf("internal budget became an author/persistence field: %s %v", data, err)
	}
	cancel()
	select {
	case <-derived.Done():
	default:
		t.Fatal("existing nested Hook calls lost parent cancellation")
	}
}
