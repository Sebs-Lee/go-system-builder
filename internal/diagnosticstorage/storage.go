// Package diagnosticstorage recognizes bounded display storage, never receipts
// or source authority. Unknown files beside diagnostics remain product files.
package diagnosticstorage

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/entroforge/go-system-builder/internal/workpacket"
)

var finalPath = regexp.MustCompile(`^\.claude/diagnostics/(planning|native-s9)/[0-9a-f]{64}\.json$`)
var temporaryPath = regexp.MustCompile(`^\.claude/diagnostics/planning/\.diagnostic-[0-9a-f]{32}\.tmp$`)

func Recognizes(ctx context.Context, root, path string) bool {
	final, temporary := finalPath.MatchString(path), temporaryPath.MatchString(path)
	if (!final && !temporary) || ctx.Err() != nil || workpacket.CheckPath(root, path) != nil {
		return false
	}
	data, err := workpacket.ReadFile(root, path, 64<<10)
	if ctx.Err() != nil {
		return false
	}
	if temporary {
		// The producer may still be writing, or have renamed it after status.
		return err == nil || os.IsNotExist(err)
	}
	if err != nil {
		return false
	}
	// A closed top-level shape also rejects duplicate keys and trailing JSON.
	decoder := json.NewDecoder(bytes.NewReader(data))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return false
	}
	fields := map[string]json.RawMessage{}
	for decoder.More() {
		key, err := decoder.Token()
		name, ok := key.(string)
		if err != nil || !ok || fields[name] != nil {
			return false
		}
		var value json.RawMessage
		if decoder.Decode(&value) != nil {
			return false
		}
		fields[name] = value
	}
	if token, err = decoder.Token(); err != nil || token != json.Delim('}') {
		return false
	}
	if _, err = decoder.Token(); err != io.EOF {
		return false
	}
	text := func(key string) string { var s string; _ = json.Unmarshal(fields[key], &s); return s }
	if text("message") == "" {
		return false
	}
	if strings.Contains(path, "/planning/") {
		_, err := time.Parse(time.RFC3339Nano, text("at"))
		return len(fields) == 2 && err == nil
	}
	var files map[string]string
	return len(fields) == 4 && len(text("scope")) == 64 && text("runtime") != "" && json.Unmarshal(fields["files"], &files) == nil && files != nil
}
