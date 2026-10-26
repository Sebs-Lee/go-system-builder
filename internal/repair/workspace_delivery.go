package repair

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	runtimepkg "github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/workpacket"
	"github.com/entroforge/go-system-builder/internal/workspace"
)

type WorkspaceAuthority struct {
	Assignment RepairAssignment
	Contract   ApprovedContract
	Session    RepairSession
	Plan       RepairPlan
	Inputs     []string
	Report     PlanReport
	ReportRef  ArtifactRef
}

// ResolveWorkspaceAuthority reads the current repair's recorded working material.
// It does not reauthorize the caller or replay acceptance of those documents.
func ResolveWorkspaceAuthority(root string, state map[string]any, id, owner string) (WorkspaceAuthority, error) {
	var out WorkspaceAuthority
	p := repairPointer(state)
	if p == nil {
		return out, fmt.Errorf("current S9 repair is missing")
	}
	read := func(path string, target any) error {
		data, err := workpacket.ReadFile(root, path, 20<<20)
		if err != nil {
			return err
		}
		return json.Unmarshal(data, target)
	}
	planRef := ArtifactRef{Path: stringField(p["plan_ref"]), SHA256: stringField(p["plan_sha256"])}
	if err := read(planRef.Path, &out.Plan); err != nil {
		return out, err
	}
	var ok bool
	out.Assignment, ok = assignmentByID(out.Plan.Assignments, canonicalPlanAssignmentID(out.Plan, id))
	if !ok {
		return out, fmt.Errorf("current S9 assignment %s is missing", id)
	}
	contractRef := ContractRef{Path: stringField(p["contract_ref"]), SHA256: stringField(p["contract_sha256"])}
	var contract struct {
		ContractID               string       `json:"repair_contract_id"`
		CaseID                   string       `json:"case_id"`
		Revision                 int          `json:"revision"`
		Status                   string       `json:"status"`
		SourceFindingIDs         []string     `json:"source_finding_ids"`
		Units                    []RepairUnit `json:"repair_units"`
		ProspectiveScope         []string     `json:"prospective_scope"`
		ForbiddenScope           []string     `json:"forbidden_scope"`
		CompatibilityMigration   string       `json:"compatibility_migration"`
		StopEscalationConditions []string     `json:"stop_escalation_conditions"`
	}
	if err := read(contractRef.Path, &contract); err != nil {
		return out, err
	}
	out.Contract = ApprovedContract{ContractID: contract.ContractID, CaseID: contract.CaseID,
		Revision: contract.Revision, Status: contract.Status, SourceFindingIDs: contract.SourceFindingIDs,
		Units: contract.Units, ProspectiveScope: contract.ProspectiveScope, ForbiddenScope: contract.ForbiddenScope,
		CompatibilityMigration: contract.CompatibilityMigration, StopEscalationConditions: append([]string(nil), contract.StopEscalationConditions...), Ref: contractRef}
	sessionPath := stringField(p["path"])
	if err := read(sessionPath, &out.Session); err != nil {
		return out, err
	}
	out.Inputs = []string{planRef.Path, sessionPath, contractRef.Path}
	// The existing index is a list of accepted refs, not an assignment map.
	// Read it directly; an absent Plan remains a missing task, not a work ACL.
	for _, ref := range existingArtifactRefs(p["plan_report_refs"]) {
		var report PlanReport
		if err := read(ref.Path, &report); err != nil {
			continue
		}
		if report.AssignmentID == out.Assignment.AssignmentID {
			out.Report, out.ReportRef = report, ref
			out.Inputs = append(out.Inputs, ref.Path)
			break
		}
	}
	return out, nil
}

type WorkspaceCandidate struct {
	Version             int                                         `json:"version"`
	RecordType          string                                      `json:"record_type"`
	RuntimeID           string                                      `json:"runtime_id"`
	BaselineGeneration  int                                         `json:"baseline_generation"`
	ExecutionGeneration int                                         `json:"execution_generation"`
	AssignmentID        string                                      `json:"assignment_id"`
	SourceHead          string                                      `json:"source_head"`
	BaseCommit          string                                      `json:"base_commit"`
	SessionID           string                                      `json:"session_id"`
	PlanID              string                                      `json:"plan_id"`
	ContractSHA256      string                                      `json:"contract_sha256"`
	Proposal            RepairResultRequest                         `json:"proposal"`
	Provenance          *runtimepkg.RepairResultCandidateProvenance `json:"provenance,omitempty"`
}

func CreateWorkspaceCandidate(ctx context.Context, root string, state map[string]any, e workspace.Execution, proposal RepairResultRequest) (WorkspaceCandidate, ArtifactRef, error) {
	candidate, err := prepareWorkspaceCandidate(ctx, root, state, e, proposal)
	if err != nil {
		return WorkspaceCandidate{}, ArtifactRef{}, err
	}
	data, err := canonicalJSON(candidate)
	if err != nil {
		return candidate, ArtifactRef{}, err
	}
	key := sha256Bytes(data)
	refPath := artifactRoot + "/candidates/" + key + ".json"
	// Identical proposals reuse immutable bytes; no overwrite on retry.
	ref := fileRef(refPath, data)
	if existing, readErr := workpacket.ReadFile(root, ref.Path, 20<<20); readErr == nil && len(existing) > 0 {
		return candidate, ref, nil
	}
	ref, err = writeImmutable(root, refPath, "repair-workspace-candidate.schema.json", candidate)
	return candidate, ref, err
}

// prepareWorkspaceCandidate captures the real committed source, including an
// empty diff, and records the new observations without requiring prior PASS.
func prepareWorkspaceCandidate(ctx context.Context, root string, state map[string]any, e workspace.Execution, proposal RepairResultRequest) (WorkspaceCandidate, error) {
	raw, err := json.Marshal(state["workspace"])
	if err != nil {
		return WorkspaceCandidate{}, err
	}
	var registry workspace.ExecutionRegistry
	if err := json.Unmarshal(raw, &registry); err != nil {
		return WorkspaceCandidate{}, err
	}
	current, ok := registry.Execution(e.AssignmentID, workspace.RuntimeID(state), workspace.Generation(state))
	if !ok || current.Generation != e.Generation || current.AgentID != e.AgentID ||
		current.ReworkRef != e.ReworkRef || current.BaseCommit != e.BaseCommit || !sameAbsolutePath(current.Path, e.Path) {
		return WorkspaceCandidate{}, fmt.Errorf("candidate source is not the current S9 execution attempt")
	}
	a, err := ResolveWorkspaceAuthority(root, state, e.AssignmentID, e.AgentID)
	if err != nil {
		return WorkspaceCandidate{}, err
	}
	// Current execution identity is system material, not an author's ACL hint.
	proposal.ProducerAgentID = e.AgentID
	proposal.AssignmentID = a.Assignment.AssignmentID
	typed := proposal.issued != nil
	head, err := workspace.Git(ctx, e.Path, "rev-parse", "HEAD")
	if err != nil {
		return WorkspaceCandidate{}, err
	}
	status, err := candidateProductStatus(ctx, root, e, a.Session, proposal)
	if err != nil {
		return WorkspaceCandidate{}, err
	}
	if status != "" {
		return WorkspaceCandidate{}, fmt.Errorf("candidate requires a clean committed Worker: %s", status)
	}
	paths, err := workspace.GitBytes(ctx, e.Path, "diff", "--no-renames", "--name-only", "-z", e.BaseCommit, head)
	if err != nil {
		return WorkspaceCandidate{}, err
	}
	changed := []ChangedArtifact{}
	for _, path := range strings.Split(string(paths), "\x00") {
		if path == "" {
			continue
		}
		kind := "modified"
		data, err := workspace.GitBytes(ctx, e.Path, "show", head+":"+path)
		if err != nil {
			kind = "deleted"
			data, err = workspace.GitBytes(ctx, e.Path, "show", e.BaseCommit+":"+path)
		} else if _, oldErr := workspace.GitBytes(ctx, e.Path, "show", e.BaseCommit+":"+path); oldErr != nil {
			kind = "added"
		}
		if err != nil {
			return WorkspaceCandidate{}, err
		}
		hash := sha256.Sum256(data)
		changed = append(changed, ChangedArtifact{Path: path, SHA256: hex.EncodeToString(hash[:]), Status: kind})
	}
	proposal.AssignmentID = a.Assignment.AssignmentID
	proposal.ChangedArtifacts = changed
	if err = validateCandidateProposal(a, &proposal, typed); err != nil {
		return WorkspaceCandidate{}, err
	}
	c := WorkspaceCandidate{Version: 1, RecordType: "repair_workspace_candidate", RuntimeID: e.RuntimeID, BaselineGeneration: e.BaselineGeneration, ExecutionGeneration: e.Generation, AssignmentID: e.AssignmentID, SourceHead: head, BaseCommit: e.BaseCommit, SessionID: a.Session.SessionID, PlanID: a.Plan.PlanID, ContractSHA256: a.Contract.Ref.SHA256, Proposal: proposal}
	if typed {
		provenance := proposal.issued.provenance
		provenance.SourceHead = head
		provenance.BaseCommit = e.BaseCommit
		c.Provenance = &provenance
	}
	return c, nil
}

// Only this original consumer's captured log outputs may coexist with a clean
// product commit. A directory, authored reference or filename alone grants no
// exemption, and tracked/baseline product files always remain protected.
func candidateProductStatus(ctx context.Context, root string, e workspace.Execution, session RepairSession, proposal RepairResultRequest) (string, error) {
	args := append([]string{"status", "--porcelain=v1", "-z", "--no-renames", "--untracked-files=all", "--", "."}, workspace.WorkerSubmissionStatusExcludes()...)
	status, err := workspace.GitBytes(ctx, e.Path, args...)
	if err != nil {
		return "", err
	}
	captured := map[string]runtimepkg.RepairResultLocalEvidence{}
	if proposal.issued != nil && e.Purpose == workspace.NativeDeliveryPurpose {
		for _, local := range proposal.issued.provenance.LocalEvidence {
			captured[local.SourcePath] = local
		}
	}
	protected := map[string]bool{}
	for _, artifact := range session.BaselineArtifacts {
		protected[artifact.Path] = true
	}
	var remaining []string
	for _, entry := range strings.Split(string(status), "\x00") {
		if entry == "" {
			continue
		}
		if err := ctx.Err(); err != nil {
			return "", err
		}
		if len(entry) < 4 || entry[2] != ' ' {
			return "", fmt.Errorf("invalid candidate Git status")
		}
		path := entry[3:]
		local, found := captured[path]
		if entry[:2] == "??" && found && candidateLogPath(path, e, protected) {
			data, err := workpacket.ReadFile(e.Path, path, repairResultAnswerLimit)
			if err != nil || workpacket.Digest(data) != local.SHA256 {
				return "", fmt.Errorf("S9 captured log source changed before candidate: %s", path)
			}
			if err := ctx.Err(); err != nil {
				return "", err
			}
			copy, err := workpacket.ReadFile(root, local.Path, repairResultAnswerLimit)
			if err != nil || workpacket.Digest(copy) != local.SHA256 {
				return "", fmt.Errorf("S9 captured log handoff unavailable before candidate: %s", path)
			}
			continue
		}
		remaining = append(remaining, entry)
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return strings.Join(remaining, "; "), nil
}

// WorkspaceCandidateLogProjection carries only immutable, originally captured
// log identities into Main's existing clean-tree check. It grants no product
// scope, acceptance or permission to ignore tracked changes.
func WorkspaceCandidateLogProjection(ctx context.Context, root string, state map[string]any, e workspace.Execution, candidate WorkspaceCandidate) (map[string]string, error) {
	out := map[string]string{}
	if e.Purpose != workspace.NativeDeliveryPurpose || candidate.Provenance == nil {
		return out, nil
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	authority, err := ResolveWorkspaceAuthority(root, state, e.AssignmentID, e.AgentID)
	if err != nil {
		return nil, err
	}
	return workspaceCandidateLogProjection(ctx, root, e, candidate, authority.Session)
}

// Typed delivery already carries the original frozen Session. Later ACK and
// cleanup retain that baseline protection without reopening mutable planning
// sources or repeating reception. Captured bytes keep the same checks.
func WorkspaceCandidateLogProjectionForPacket(ctx context.Context, root string, packet workpacket.RepairResultPacket, e workspace.Execution, candidate WorkspaceCandidate) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if candidate.Provenance == nil || candidate.Provenance.PacketID != packet.ID {
		return nil, fmt.Errorf("captured log projection requires the original typed candidate packet")
	}
	var session RepairSession
	if err := json.Unmarshal(packet.Inputs.Sources.Session, &session); err != nil {
		return nil, err
	}
	if session.SessionID != packet.Binding.Sources.SessionID || session.RuntimeID != e.RuntimeID || session.BaselineGeneration != e.BaselineGeneration {
		return nil, fmt.Errorf("captured log projection original frozen Session differs")
	}
	return workspaceCandidateLogProjection(ctx, root, e, candidate, session)
}
func workspaceCandidateLogProjection(ctx context.Context, root string, e workspace.Execution, candidate WorkspaceCandidate, session RepairSession) (map[string]string, error) {
	out := map[string]string{}
	if e.Purpose != workspace.NativeDeliveryPurpose || candidate.Provenance == nil {
		return out, nil
	}
	protected := map[string]bool{}
	for _, artifact := range session.BaselineArtifacts {
		protected[artifact.Path] = true
	}
	for _, local := range candidate.Provenance.LocalEvidence {
		if !candidateLogPath(local.SourcePath, e, protected) {
			continue
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		data, err := workpacket.ReadFile(root, local.Path, repairResultAnswerLimit)
		if err != nil || workpacket.Digest(data) != local.SHA256 {
			return nil, fmt.Errorf("S9 captured log handoff unavailable before integration: %s", local.SourcePath)
		}
		out[local.SourcePath] = local.SHA256
	}
	return out, nil
}

func candidateLogPath(path string, e workspace.Execution, protected map[string]bool) bool {
	if protected[path] || !(isRuntimeLogShape(path) || strings.HasPrefix(path, "logs/") && filepath.Ext(path) == ".txt") {
		return false
	}
	for _, scope := range e.WritePaths {
		if workspace.PathMatchesScope(path, scope) {
			return false
		}
	}
	return true
}

func ValidateWorkspaceCandidate(root string, state map[string]any, e workspace.Execution, ref ArtifactRef) (WorkspaceCandidate, error) {
	raw, err := json.Marshal(state["workspace"])
	if err != nil {
		return WorkspaceCandidate{}, err
	}
	var registry workspace.ExecutionRegistry
	if err := json.Unmarshal(raw, &registry); err != nil {
		return WorkspaceCandidate{}, err
	}
	current, ok := registry.Execution(e.AssignmentID, workspace.RuntimeID(state), workspace.Generation(state))
	if !ok || current.Generation != e.Generation || current.AgentID != e.AgentID ||
		current.ReworkRef != e.ReworkRef || current.BaseCommit != e.BaseCommit || !sameAbsolutePath(current.Path, e.Path) {
		return WorkspaceCandidate{}, fmt.Errorf("candidate receiver is not the current S9 execution attempt")
	}
	c, err := ReadWorkspaceCandidate(root, ref)
	if err != nil {
		return c, err
	}
	if c.RuntimeID != e.RuntimeID || c.BaselineGeneration != e.BaselineGeneration ||
		c.ExecutionGeneration != e.Generation || c.AssignmentID != e.AssignmentID ||
		c.BaseCommit != e.BaseCommit || c.Proposal.ProducerAgentID != e.AgentID ||
		(current.DeliveryRef != "" && current.DeliveryRef != ref.Path) {
		return c, fmt.Errorf("candidate belongs to a different current S9 execution attempt")
	}
	if p := c.Provenance; p != nil && (p.ExecutionAssignmentID != e.AssignmentID ||
		p.ExecutionGeneration != e.Generation || p.RuntimeID != e.RuntimeID ||
		p.BaselineGeneration != e.BaselineGeneration || p.OwnerAgentID != e.AgentID ||
		p.ReworkRef != e.ReworkRef || !sameAbsolutePath(p.ExecutionRoot, e.Path)) {
		return c, fmt.Errorf("candidate belongs to a different Worker attempt or checkout")
	}
	return c, nil
}

func ReadWorkspaceCandidate(root string, ref ArtifactRef) (WorkspaceCandidate, error) {
	var c WorkspaceCandidate
	data, err := workpacket.ReadFile(root, ref.Path, 20<<20)
	if err != nil {
		return c, err
	}
	err = json.Unmarshal(data, &c)
	return c, err
}

func RegisteredWorkspaceResult(root string, state map[string]any, c WorkspaceCandidate) (ArtifactRef, bool, error) {
	return RegisteredWorkspaceResultContext(context.Background(), root, state, c)
}
func RegisteredWorkspaceResultContext(ctx context.Context, root string, state map[string]any, c WorkspaceCandidate) (ArtifactRef, bool, error) {
	if err := ctx.Err(); err != nil {
		return ArtifactRef{}, false, err
	}
	for _, ref := range existingArtifactRefs(repairPointer(state)["result_refs"]) {
		if c.Provenance != nil && ref.Path != artifactRoot+"/results/"+c.Proposal.ResultID+".json" {
			continue
		}
		if err := ctx.Err(); err != nil {
			return ArtifactRef{}, false, err
		}
		data, err := workpacket.ReadFile(root, ref.Path, 20<<20)
		var result RepairResult
		if err == nil {
			err = json.Unmarshal(data, &result)
		}
		if err != nil {
			return ArtifactRef{}, false, err
		}
		if result.ResultID == c.Proposal.ResultID {
			if result.Result != "pass" {
				return ArtifactRef{}, false, fmt.Errorf("recorded Result is %s; actual repair/check work remains pending", result.Result)
			}
			if result.AssignmentID != c.Proposal.AssignmentID || result.SessionID != c.SessionID || result.ProducerAgentID != c.Proposal.ProducerAgentID {
				return ArtifactRef{}, false, fmt.Errorf("candidate result identity collision")
			}
			return ref, true, nil
		}
	}
	return ArtifactRef{}, false, nil
}

// New candidate shape is checked once here. Missing work and non-passing
// observations remain recorded facts; Main's actual integration/checks follow.
func validateCandidateProposal(a WorkspaceAuthority, p *RepairResultRequest, typed bool) error {
	if strings.TrimSpace(p.ResultID) == "" {
		return fmt.Errorf("candidate result identity is missing")
	}
	if p.Result == "" {
		p.Result = "blocked"
	}
	if p.Result != "pass" && p.Result != "fail" && p.Result != "blocked" {
		return fmt.Errorf("candidate result must be pass, fail, or blocked")
	}
	if len(p.BeforeFixChecks) == 0 {
		p.BeforeFixChecks = append([]RepairCheck{}, a.Report.RedChecks...)
	}
	p.PlanReport = a.ReportRef
	return nil
}
