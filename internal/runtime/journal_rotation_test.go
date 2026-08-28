package runtime_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/entroforge/go-system-builder/internal/runtime"
)

// RC-10 Step A: the journal rotation threshold is an observability probe only.
// Rotation itself is deliberately not implemented (it must go through a
// Rollover-style durable-marker transaction to preserve the sequence/tail
// invariants), so the helper must be read-only and must treat a missing
// journal as an empty tail.
func TestJournalNeedsRotation(t *testing.T) {
	t.Run("missing journal is empty tail", func(t *testing.T) {
		needs, count, err := runtime.JournalNeedsRotation(filepath.Join(t.TempDir(), "journal.jsonl"))
		if err != nil {
			t.Fatal(err)
		}
		if needs || count != 0 {
			t.Fatalf("missing journal must read as empty: needs=%v count=%d", needs, count)
		}
	})

	t.Run("under threshold", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "journal.jsonl")
		if err := os.WriteFile(path, []byte("{\"a\":1}\n{\"a\":2}\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		needs, count, err := runtime.JournalNeedsRotation(path)
		if err != nil {
			t.Fatal(err)
		}
		if needs || count != 2 {
			t.Fatalf("two-line journal: needs=%v count=%d", needs, count)
		}
	})

	t.Run("over threshold", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "journal.jsonl")
		var b strings.Builder
		for i := 0; i < 10001; i++ {
			b.WriteString("{\"a\":1}\n")
		}
		if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
			t.Fatal(err)
		}
		needs, count, err := runtime.JournalNeedsRotation(path)
		if err != nil {
			t.Fatal(err)
		}
		if !needs || count != 10001 {
			t.Fatalf("10001-line journal: needs=%v count=%d", needs, count)
		}
	})

	t.Run("threshold is 10000", func(t *testing.T) {
		if runtime.JournalRotationThreshold != 10000 {
			t.Fatalf("journal rotation threshold changed: %d", runtime.JournalRotationThreshold)
		}
	})
}
