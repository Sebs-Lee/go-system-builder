package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/entroforge/go-system-builder/internal/cli"
)

const (
	s7HookSubprocessMarker = "GSB_S7_POST_TOOL_USE_CHILD"
	s7HookSubprocessRoot   = "GSB_S7_POST_TOOL_USE_ROOT"
	s7HookSubprocessResult = "GSB_S7_POST_TOOL_USE_RESULT:"
)

type s7HookSubprocessOutcome struct {
	ExitCode int    `json:"exit_code"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
}

// invokeS7HookSubprocess calls the actual test executable in a fresh process,
// then runs the same public CLI entry point used by the warm-path helper.
func invokeS7HookSubprocess(t *testing.T, root string, payload []byte) (int, string, string) {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatalf("resolve test executable for cold S7 hook: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)
	command := exec.CommandContext(ctx, executable, "-test.run=^TestS7NativeStopSubprocessHelper$")
	command.Dir = root
	command.Env = append(os.Environ(), s7HookSubprocessMarker+"=1", s7HookSubprocessRoot+"="+root)
	command.Stdin = bytes.NewReader(payload)
	var processStdout, processStderr bytes.Buffer
	command.Stdout = &processStdout
	command.Stderr = &processStderr
	if err := command.Run(); err != nil {
		t.Fatalf("cold S7 hook subprocess failed: %v; stdout=%s stderr=%s", err, processStdout.String(), processStderr.String())
	}
	for _, line := range strings.Split(processStdout.String(), "\n") {
		if !strings.HasPrefix(line, s7HookSubprocessResult) {
			continue
		}
		var outcome s7HookSubprocessOutcome
		if err := json.Unmarshal([]byte(strings.TrimPrefix(line, s7HookSubprocessResult)), &outcome); err != nil {
			t.Fatalf("decode cold S7 hook subprocess result: %v; stdout=%s", err, processStdout.String())
		}
		return outcome.ExitCode, outcome.Stdout, outcome.Stderr
	}
	t.Fatalf("cold S7 hook subprocess returned no CLI result; stdout=%s stderr=%s", processStdout.String(), processStderr.String())
	return 1, "", ""
}

// TestS7NativeStopSubprocessHelper is selected only by invokeS7HookSubprocess.
// An ordinary package test run skips it without invoking the CLI.
func TestS7NativeStopSubprocessHelper(t *testing.T) {
	if os.Getenv(s7HookSubprocessMarker) != "1" {
		t.Skip("only runs inside a cold S7 SubagentStop subprocess")
	}
	root := os.Getenv(s7HookSubprocessRoot)
	if root == "" {
		t.Fatal("cold S7 hook root is missing")
	}
	payload, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read cold S7 hook payload: %v", err)
	}
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"hook", "--root", root, "--event", "SubagentStop"}, bytes.NewReader(payload), &stdout, &stderr)
	result, err := json.Marshal(s7HookSubprocessOutcome{ExitCode: code, Stdout: stdout.String(), Stderr: stderr.String()})
	if err != nil {
		t.Fatalf("encode cold S7 hook result: %v", err)
	}
	if _, err := fmt.Fprintf(os.Stdout, "%s%s\n", s7HookSubprocessResult, result); err != nil {
		t.Fatalf("write cold S7 hook result: %v", err)
	}
}
