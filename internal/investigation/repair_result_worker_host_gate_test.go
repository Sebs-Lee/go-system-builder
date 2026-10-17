package investigation_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/entroforge/go-system-builder/internal/cli"
	"github.com/entroforge/go-system-builder/internal/hookctx"
	"github.com/entroforge/go-system-builder/internal/repair"
	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/workpacket"
	"github.com/entroforge/go-system-builder/internal/workspace"
)

// This denial-only fixture uses the original S9 approval/dispatch/PlanReport
// and workspace prepare paths. Its authored red command is mechanical test
// input; it is not a claim that a model or Worker completed professional work.
func TestS9WorkerPrepareFailsClosedWhenOriginalIsolationProbeIsUnavailable(t *testing.T) {
	fixture, plan, owners := prepareS9RepairPlanReportHookFixtureWithSetup(t, installS9WorkerBootstrapInputs)
	if len(plan.Assignments) != 2 || len(owners) != 2 {
		t.Fatalf("original S9 fixture assignments/owners = %d/%d, want two", len(plan.Assignments), len(owners))
	}

	const redCommand = "test -f internal/api/payload.go"
	var issueOut, issueErr bytes.Buffer
	if code := cli.Run([]string{"runtime", "stage-answer", "issue", "--root", fixture.root}, nil, &issueOut, &issueErr); code != 0 {
		t.Fatalf("issue original sibling PlanReports: code=%d stderr=%s stdout=%s", code, issueErr.String(), issueOut.String())
	}
	var packets []workpacket.RepairPlanReportPacket
	if err := json.Unmarshal(issueOut.Bytes(), &packets); err != nil || len(packets) != 2 {
		t.Fatalf("decode original sibling PlanReports: packets=%d err=%v stdout=%s", len(packets), err, issueOut.String())
	}
	for _, packet := range packets {
		answerPath := writeS9RepairPlanReportForResult(t, fixture.root, packet, redCommand)
		receipt := runS9PostToolUse(t, fixture, packet.Binding.OwnerAgentID, fixture.root, answerPath)
		if !receipt.Accepted || receipt.PacketID != packet.ID || receipt.WorkComplete {
			t.Fatalf("original S9 PlanReport Hook receipt = %+v", receipt)
		}
	}

	snapshot := currentS9Snapshot(t, fixture)
	started, err := repair.BeginRepairExecution(fixture.root, fixture.statePath, fixture.journalPath, repair.BeginRepairExecutionRequest{
		RuntimeRequest: repair.RuntimeRequest{ExpectedRevision: snapshot.Revision, Actor: "main-session"},
	})
	if err != nil {
		t.Fatalf("original S9 Begin after complete sibling PlanReport coverage: %v", err)
	}
	if life, _ := started.State["lifecycle"].(map[string]any); life == nil || life["phase"] != "fixing" {
		t.Fatalf("original S9 Begin did not enter fixing: %v", started.State["lifecycle"])
	}

	owner := owners[0]
	if code := runS9Command(t, fixture.root, "runtime", "workspace", "bind", "--root", fixture.root); code != 0 {
		t.Fatal("original Main workspace bind failed")
	}

	// Use the same source authority, frozen Report command/profile and original
	// PreflightChecks API as workspace prepare. The probe has no Runtime/Git
	// side effect; when this host supports isolation, this denial-only fixture
	// is explicitly skipped rather than claimed as Worker-success coverage.
	bound := currentS9Snapshot(t, fixture)
	registry, err := workspace.Decode(bound.State)
	if err != nil || registry == nil {
		t.Fatalf("decode original bound Main workspace: registry=%v err=%v", registry, err)
	}
	authority, err := repair.ResolveWorkspaceAuthority(fixture.root, bound.State, owner.manifestAssignmentID, owner.agentID)
	if err != nil {
		t.Fatalf("resolve original S9 Worker authority: %v", err)
	}
	profiles, err := runtime.RepairResultCheckProfiles(authority.ReportRef.SHA256, authority.Checks, "worker")
	if err != nil {
		t.Fatalf("derive original PlanReport Worker profiles: %v", err)
	}
	probeExecution := workspace.Execution{CheckLocation: "worker", Checks: authority.Checks, CheckProfiles: profiles}
	probeErr := registry.PreflightChecks(context.Background(), probeExecution, fixture.root)
	if probeErr == nil {
		t.Skip("the original Worker isolation probe passes on this host; this denial-only test does not claim the positive Worker matrix")
	}
	if !strings.Contains(probeErr.Error(), "Worker check capability unavailable") {
		t.Fatalf("original Worker probe failed outside its named capability gate: %v", probeErr)
	}
	t.Logf("original isolation capability probe refused: %v", probeErr)

	// Public prepare may stage ignored Main-side assignment context before its
	// preflight. The contract starts its zero-write assertion at the Runtime
	// boundary: no preparing/ready CAS, Worker execution/worktree/bootstrap, or
	// product write may result from a denied isolation capability.
	stateBefore := mustRead(t, fixture.statePath)
	journalBefore := mustRead(t, fixture.journalPath)
	headBefore, err := workspace.Git(context.Background(), fixture.root, "rev-parse", "HEAD")
	if err != nil || headBefore == "" {
		t.Fatalf("read Main HEAD before denied prepare: head=%q err=%v", headBefore, err)
	}
	productStatusBefore, err := workspace.GitBytes(context.Background(), fixture.root,
		"status", "--porcelain=v1", "-z", "--untracked-files=all", "--", ".", ":(exclude).claude/**")
	if err != nil {
		t.Fatalf("capture product-tree status before denied prepare: %v", err)
	}
	productPath := filepath.Join(fixture.root, "internal", "api", "payload.go")
	if _, err := os.Lstat(productPath); !os.IsNotExist(err) {
		t.Fatalf("red-check product file should still be absent before prepare; stat err=%v", err)
	}
	if packets, err := runtime.RepairResultPackets(bound.State); err != nil || len(packets) != 0 {
		t.Fatalf("unexpected S9 candidate/result packet before prepare: count=%d err=%v", len(packets), err)
	}

	// Plan is local-only and gives the precise worker path the original CLI
	// would materialize after successful preflight. It does not register an
	// Execution or create a worktree.
	loaded, err := hookctx.LoadFull(fixture.root, owner.agentID)
	if err != nil {
		t.Fatalf("load original registered S9 owner context: %v", err)
	}
	var assignment *hookctx.AssignmentContext
	for i := range loaded.Assignments {
		candidate := &loaded.Assignments[i]
		if candidate.AssignmentID == owner.manifestAssignmentID && candidate.OwnerAgentID == owner.agentID {
			assignment = candidate
			break
		}
	}
	if assignment == nil {
		t.Fatalf("original Hook context omitted registered S9 owner %s", owner.agentID)
	}
	inputs := []string{"docs/control/loop-definition.json", "docs/control/hook-policy.json"}
	for _, ref := range []string{assignment.ManifestRef, assignment.AgentDefinitionRef} {
		path, _, _ := strings.Cut(ref, "#")
		if path != "" {
			inputs = append(inputs, path)
		}
	}
	inputs = append(inputs, authority.Inputs...)
	planned, err := registry.Plan(context.Background(), bound.State, owner.manifestAssignmentID, owner.agentID,
		authority.Assignment.Scope, authority.Checks, inputs)
	if err != nil {
		t.Fatalf("construct side-effect-free original S9 prepare plan: %v", err)
	}
	if planned.Path == "" {
		t.Fatal("original S9 prepare plan has no Worker worktree path")
	}
	if _, err := os.Lstat(planned.Path); !os.IsNotExist(err) {
		t.Fatalf("Worker worktree exists before the original preflight gate: path=%s err=%v", planned.Path, err)
	}

	var prepareOut, prepareErr bytes.Buffer
	code := cli.Run([]string{"runtime", "workspace", "prepare", "--root", fixture.root,
		"--assignment", owner.manifestAssignmentID, "--agent", owner.agentID, "--check-location", "worker"},
		nil, &prepareOut, &prepareErr)
	if code == 0 {
		t.Fatalf("original Worker prepare unexpectedly passed an unavailable-isolation host gate: stdout=%s stderr=%s", prepareOut.String(), prepareErr.String())
	}
	combined := prepareErr.String() + prepareOut.String()
	if !strings.Contains(combined, "Worker check capability unavailable") {
		t.Fatalf("original Worker prepare did not preserve its named capability error: code=%d stdout=%s stderr=%s; probe=%v", code, prepareOut.String(), prepareErr.String(), probeErr)
	}
	t.Logf("original S9 Worker prepare refused at capability gate: %s", combined)
	if !strings.Contains(combined, "Bubblewrap") && !strings.Contains(combined, "user namespaces") && !strings.Contains(combined, "isolated Worker checks are unavailable") {
		t.Fatalf("original prepare error is not an explicit isolation capability refusal: %s", combined)
	}
	if !bytes.Equal(stateBefore, mustRead(t, fixture.statePath)) || !bytes.Equal(journalBefore, mustRead(t, fixture.journalPath)) {
		t.Fatal("denied Worker prepare changed the bound Runtime state or journal")
	}
	if _, err := os.Lstat(planned.Path); !os.IsNotExist(err) {
		t.Fatalf("denied Worker prepare materialized a worktree: path=%s err=%v", planned.Path, err)
	}
	if _, err := os.Lstat(productPath); !os.IsNotExist(err) {
		t.Fatalf("denied Worker prepare changed red-check product content; stat err=%v", err)
	}
	productStatusAfter, err := workspace.GitBytes(context.Background(), fixture.root,
		"status", "--porcelain=v1", "-z", "--untracked-files=all", "--", ".", ":(exclude).claude/**")
	if err != nil || !bytes.Equal(productStatusBefore, productStatusAfter) {
		t.Fatalf("denied Worker prepare changed product-tree status: err=%v before=%q after=%q", err, productStatusBefore, productStatusAfter)
	}
	after := currentS9Snapshot(t, fixture)
	afterRegistry, err := workspace.Decode(after.State)
	if err != nil || afterRegistry == nil {
		t.Fatalf("decode workspace after denied prepare: registry=%v err=%v", afterRegistry, err)
	}
	if _, exists := afterRegistry.Executions[owner.manifestAssignmentID]; exists {
		t.Fatalf("denied Worker prepare published an Execution lease: %+v", afterRegistry.Executions[owner.manifestAssignmentID])
	}
	if packets, err := runtime.RepairResultPackets(after.State); err != nil || len(packets) != 0 {
		t.Fatalf("denied Worker prepare created a candidate/result packet: count=%d err=%v", len(packets), err)
	}
	if after.Revision != bound.Revision {
		t.Fatalf("denied Worker prepare Runtime revision=%d, want unchanged bound revision %d", after.Revision, bound.Revision)
	}
	if afterHead, err := workspace.Git(context.Background(), fixture.root, "rev-parse", "HEAD"); err != nil || afterHead != headBefore {
		t.Fatalf("denied Worker prepare changed Main HEAD: before=%q after=%q err=%v", headBefore, afterHead, err)
	}
	if after.State["runtime_id"] != bound.State["runtime_id"] {
		t.Fatal("denied Worker prepare changed Runtime identity")
	}
}
