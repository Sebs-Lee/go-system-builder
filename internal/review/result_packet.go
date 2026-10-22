package review

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/entroforge/go-system-builder/internal/agentauth"
	loopruntime "github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/schema"
	"github.com/entroforge/go-system-builder/internal/semantic"
	"github.com/entroforge/go-system-builder/internal/workpacket"
	"github.com/entroforge/go-system-builder/internal/workspace"
	"github.com/entroforge/go-system-builder/internal/workstate"
)

// IssuedSubmitRequest is an internal application request. OwnerAgentID and
// source root are checked against the original packet and registered Runtime
// execution. AuthorizeSource is an optional compatibility callback; it is not
// required by the normal Hook questionnaire save path.
type IssuedSubmitRequest struct {
	Context                            context.Context `json:"-"`
	ExpectedRevision                   int
	PacketID, AnswerPath, OwnerAgentID string
	SourceRoot                         string
	AuthorizeSource                    func(root string, state map[string]any) error
	OccurredAt                         time.Time
}

// IssueResultPackets appends only questions to the original Runtime registry.
// The Main adapter supplies issuance authority; no Reviewer verdict is filled.
func IssueResultPackets(root, statePath, journalPath string, expectedRevision int, ownerFilter string) ([]workpacket.ReviewResultPacket, error) {
	return IssueResultPacketsContext(context.Background(), root, statePath, journalPath, expectedRevision, ownerFilter)
}

func IssueResultPacketsContext(ctx context.Context, root, statePath, journalPath string, expectedRevision int, ownerFilter string) ([]workpacket.ReviewResultPacket, error) {
	writer := loopruntime.NewWriter(statePath, journalPath, root, semantic.RuntimeCandidateValidator{}).WithContext(ctx)
	snapshot, err := writer.Snapshot()
	if err != nil {
		return nil, err
	}
	if expectedRevision >= 0 && snapshot.Revision != expectedRevision {
		return nil, loopruntime.ErrStaleRevision
	}
	if !typedResultReceiptsEnabled(snapshot.State) {
		return nil, errors.New("S7 issuance requires explicit Runtime v1.2/v1.3 packet mode")
	}
	plan, ptr, err := loadResultPacketPlan(root, snapshot.State, false)
	if err != nil {
		return nil, err
	}
	assignments := append([]PlanAssignment(nil), plan.Assignments...)
	sort.Slice(assignments, func(i, j int) bool { return assignments[i].AssignmentID < assignments[j].AssignmentID })
	packets := []workpacket.ReviewResultPacket{}
	for i := range assignments {
		a := &assignments[i]
		row := resultPacketAssignmentRow(snapshot.State, a.AssignmentID)
		owner := stringField(row["agent_id"])
		if workstate.ReviewAssignmentActivity(row) != "dispatched" || owner == "" || (ownerFilter != "" && owner != ownerFilter) {
			continue
		}
		packet, err := buildCurrentResultPacket(root, snapshot.State, plan, ptr, a, owner)
		if err != nil {
			return nil, err
		}
		if old, found, err := loopruntime.FindReviewResultPacket(snapshot.State, packet.ID); err != nil {
			return nil, err
		} else if found {
			if !reflect.DeepEqual(old, packet) {
				return nil, errors.New("S7 frozen packet identity conflict")
			}
			if err := verifyFrozenSubjects(root, plan, snapshot.State); err != nil {
				return nil, err
			}
			if err := verifyResultPacketIssue(root, snapshot.State, packet); err != nil {
				return nil, err
			}
		} else {
			encoded, err := json.Marshal(packet)
			if err != nil {
				return nil, err
			}
			var value map[string]any
			if err := json.Unmarshal(encoded, &value); err != nil {
				return nil, err
			}
			life, _ := snapshot.State["lifecycle"].(map[string]any)
			cursor := map[string]any{"state": life["state"], "phase": life["phase"]}
			snapshot, err = writer.UpdateCurrent(loopruntime.Mutation{
				EventID: packet.IssuanceEventID, TransitionID: "PACKET-ISSUE", Event: "work_packet_issued",
				Actor: "orchestrator", IdempotencyKey: "packet:" + packet.ID, RuntimeID: packet.Binding.RuntimeID,
				BaselineGeneration: packet.Binding.BaselineGeneration, From: cursor, To: cursor, RetainLastTransition: true, OccurredAt: time.Now().UTC(),
				Apply: func(state map[string]any) error {
					if err := verifyCurrentResultPacket(root, state, packet, false); err != nil {
						return err
					}
					if old, found, err := loopruntime.FindReviewResultPacket(state, packet.ID); err != nil {
						return err
					} else if found {
						if !reflect.DeepEqual(old, packet) {
							return errors.New("S7 packet identity conflict")
						}
						return errExactResultReplay
					}
					all, err := loopruntime.ReviewResultPackets(state)
					if err != nil {
						return err
					}
					for _, old := range all {
						b := old.Binding
						if b.RuntimeID == packet.Binding.RuntimeID && b.BaselineGeneration == packet.Binding.BaselineGeneration && b.ReviewPlanID == packet.Binding.ReviewPlanID && b.ReviewPlanRevision == packet.Binding.ReviewPlanRevision && b.ReviewRound == packet.Binding.ReviewRound && b.AssignmentID == packet.Binding.AssignmentID {
							return errors.New("S7 Assignment already has a different frozen issuance")
						}
					}
					entities, _ := state["entities"].(map[string]any)
					rows, _ := entities["work_packets"].([]any)
					entities["work_packets"] = append(rows, value)
					// Preserve the original 16 MiB checkpoint/recovery budget.
					// Refuse cumulative packet growth before accepting the CAS,
					// leaving room for the Writer's cursor and event metadata.
					candidate, err := json.MarshalIndent(state, "", "  ")
					if err != nil {
						return err
					}
					if len(candidate) > (16<<20)-(8<<10) {
						return errors.New("S7 cumulative issuance exceeds the original Runtime checkpoint budget; preserve the full Claim denominator when revising source representation")
					}
					return nil
				},
			})
			if err != nil {
				latest, readErr := writer.Snapshot()
				if readErr != nil {
					return nil, err
				}
				old, found, findErr := loopruntime.FindReviewResultPacket(latest.State, packet.ID)
				if findErr != nil || !found || !reflect.DeepEqual(old, packet) {
					return nil, err
				}
				if proofErr := verifyCurrentResultPacket(root, latest.State, packet, true); proofErr != nil {
					return nil, proofErr
				}
				snapshot = latest
			}
		}
		view, err := json.MarshalIndent(packet, "", "  ")
		if err != nil {
			return nil, err
		}
		view = append(view, '\n')
		if existing, err := workpacket.PublishBounded(root, packet.ViewPath, view, resultReceiptReadLimit); err != nil {
			return nil, err
		} else if !bytes.Equal(existing, view) {
			return nil, errors.New("S7 issued view differs from original registry")
		}
		draft, err := json.MarshalIndent(packet.Draft, "", "  ")
		if err != nil {
			return nil, err
		}
		_, independent, err := resultPacketSourceRoot(root, snapshot.State, packet)
		if err != nil {
			return nil, err
		}
		if !independent {
			if _, err := workpacket.PublishBounded(root, packet.AnswerPath, append(draft, '\n'), resultReceiptReadLimit); err != nil {
				return nil, err
			}
		}
		packets = append(packets, packet)
	}
	return packets, nil
}

func resultPacketAssignmentRow(state map[string]any, id string) map[string]any {
	review, _ := state["review"].(map[string]any)
	rows, _ := review["assignments"].(map[string]any)
	row, _ := rows[id].(map[string]any)
	return row
}

func loadResultPacketPlan(root string, state map[string]any, verifyBaseline bool) (*Plan, *PlanPointer, error) {
	if !typedResultReceiptsEnabled(state) {
		return nil, nil, errors.New("S7 packet mode is disabled")
	}
	life, _ := state["lifecycle"].(map[string]any)
	if life["state"] != "verification" {
		return nil, nil, errors.New("S7 issuance requires the original verification stage")
	}
	plan, ptr, err := loadReceiptPlan(root, state)
	if err != nil {
		return nil, nil, err
	}
	data, err := json.Marshal(plan)
	if err != nil {
		return nil, nil, err
	}
	if err := schema.NewValidator(root).ValidateBytes("review-plan.schema.json", data); err != nil {
		return nil, nil, err
	}
	if err := ValidatePlan(plan); err != nil {
		return nil, nil, err
	}
	if plan.BaselineGeneration != baselineGeneration(state) || plan.ReviewRound != currentReviewRound(state) {
		return nil, nil, errors.New("S7 Plan is not the current generation/round")
	}
	if ptr.Status != "running" && ptr.Status != "cannot_clean" && ptr.Status != "discovery_draining" {
		return nil, nil, errors.New("S7 Plan is not accepting assignments")
	}
	// New issuance validates the baseline once under the original Writer lock.
	// The pre-load retains pinned plan and eligibility checks. An existing
	// packet's restore branch keeps its original baseline check separately.
	if verifyBaseline {
		if err := verifyFrozenSubjects(root, plan, state); err != nil {
			return nil, nil, err
		}
	}
	return plan, ptr, nil
}

func buildCurrentResultPacket(root string, state map[string]any, plan *Plan, ptr *PlanPointer, a *PlanAssignment, owner string) (workpacket.ReviewResultPacket, error) {
	var zero workpacket.ReviewResultPacket
	row := resultPacketAssignmentRow(state, a.AssignmentID)
	if row == nil || row["agent_id"] != owner || (workstate.ReviewAssignmentActivity(row) != "dispatched" && workstate.ReviewAssignmentActivity(row) != "blocked") {
		return zero, errors.New("S7 Assignment has no current dispatched owner")
	}
	if row["lens"] != a.Lens || !jsonValueEqual(row["claim_ids"], a.ClaimIDs) {
		return zero, errors.New("S7 Assignment projection differs from original Claim/lens set")
	}
	if err := validateResultPacketOwner(state, a, owner); err != nil {
		return zero, err
	}
	assignmentSHA, err := planAssignmentDigest(a)
	if err != nil {
		return zero, err
	}
	claims := []Claim{}
	for _, id := range a.ClaimIDs {
		found := false
		for _, claim := range plan.Claims {
			if claim.ClaimID == id {
				claims = append(claims, claim)
				found = true
				break
			}
		}
		if !found {
			return zero, fmt.Errorf("S7 packet Claim %s is absent from original Plan", id)
		}
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return zero, err
	}
	subjectsJSON, err := json.Marshal(plan.FrozenSubjects)
	if err != nil {
		return zero, err
	}
	planRef := ptr.Path
	if filepath.IsAbs(filepath.FromSlash(planRef)) {
		planRef = repositoryPath(root, planRef)
	}
	if _, err := repositoryContainedPath(root, planRef); err != nil {
		return zero, err
	}
	return workpacket.BuildReviewResult(workpacket.ReviewResultBinding{
		RuntimeID: stringField(state["runtime_id"]), BaselineGeneration: plan.BaselineGeneration,
		ReviewPlanID: plan.ReviewPlanID, ReviewPlanRef: planRef, ReviewPlanSHA256: ptr.SHA256, ReviewPlanRevision: ptr.Revision,
		ReviewRound: plan.ReviewRound, AssignmentID: a.AssignmentID, AssignmentSHA256: assignmentSHA,
		OwnerAgentID: owner, ClaimIDs: append([]string(nil), a.ClaimIDs...), SubjectDigest: SubjectDigest(plan),
		VerificationArtifactDigest: nil,
	}, workpacket.ReviewResultInputs{Claims: claimsJSON, FrozenSubjects: subjectsJSON})
}

func validateResultPacketOwner(state map[string]any, a *PlanAssignment, owner string) error {
	entities, _ := state["entities"].(map[string]any)
	agents, _ := entities["agents"].([]any)
	count := 0
	role := map[string]string{"delivery": "delivery-verifier", "qa": "qa", "e2e": "e2e-tester"}[a.Lens]
	for _, raw := range agents {
		agent, _ := raw.(map[string]any)
		if agent["id"] != owner {
			continue
		}
		count++
		if agent["role"] != role {
			return errors.New("S7 issued owner role differs from original Assignment lens")
		}
	}
	if count != 1 {
		return errors.New("S7 owner must be a unique registered Reviewer")
	}
	return validateProducerIndependence(state, &Result{ProducerAgentID: owner})
}

func verifyCurrentResultPacket(root string, state map[string]any, p workpacket.ReviewResultPacket, issued bool) error {
	plan, ptr, err := loadResultPacketPlan(root, state, true)
	if err != nil {
		return err
	}
	a := findPlanAssignment(plan, p.Binding.AssignmentID)
	if a == nil {
		return errors.New("S7 issued Assignment left current Plan")
	}
	expected, err := buildCurrentResultPacket(root, state, plan, ptr, a, p.Binding.OwnerAgentID)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(expected, p) {
		return errors.New("S7 issuance no longer matches current Plan/Assignment/owner/subjects")
	}
	if issued {
		return verifyResultPacketIssue(root, state, p)
	}
	return nil
}

func verifyResultPacketIssue(root string, state map[string]any, p workpacket.ReviewResultPacket) error {
	return loopruntime.VerifyReviewResultPacketIssue(root, state, p)
}

// SubmitIssuedResult projects professional JSON into the existing Result and
// calls the same original consumer. It introduces no second accepting Writer.
func SubmitIssuedResult(root, statePath, journalPath string, request IssuedSubmitRequest) (loopruntime.Snapshot, error) {
	writer := loopruntime.NewWriter(statePath, journalPath, root, semantic.RuntimeCandidateValidator{}).WithContext(request.Context)
	snapshot, err := writer.Snapshot()
	if err != nil {
		return loopruntime.Snapshot{}, err
	}
	p, found, err := loopruntime.FindReviewResultPacket(snapshot.State, request.PacketID)
	if err != nil {
		return snapshot, err
	}
	if !found || request.OwnerAgentID == "" || p.Binding.OwnerAgentID != request.OwnerAgentID {
		return snapshot, errors.New("S7 save has no issuance for the trusted owner")
	}
	path := request.AnswerPath
	if filepath.IsAbs(path) {
		path = repositoryPath(root, path)
	}
	path = filepath.ToSlash(filepath.Clean(filepath.FromSlash(path)))
	if path != p.AnswerPath {
		return snapshot, errors.New("S7 save path differs from exact original issuance")
	}
	row := resultPacketAssignmentRow(snapshot.State, p.Binding.AssignmentID)
	isReplay := workstate.ReviewAssignmentActivity(row) == "consumed" || workstate.ReviewAssignmentActivity(row) == "result_submitted"
	// Exact Save replay still reads the current answer bytes so the original
	// SubmitResult idempotency check can compare AnswerDigest and the full
	// machine tuple. Replay only skips revalidating mutable checkout/Git state;
	// receipt-only recovery is available through IssuedResultReceipt.
	sourceRoot, _, err := registeredResultPacketSourceRoot(root, snapshot.State, p, !isReplay, request.Context)
	if err != nil {
		return snapshot, err
	}
	if request.SourceRoot != "" {
		requestedRoot, err := workspace.Canonical(request.SourceRoot)
		if err != nil {
			return snapshot, err
		}
		if requestedRoot != sourceRoot {
			return snapshot, errors.New("S7 answer source differs from original execution; shared Main entry is retired for independent workers")
		}
	}
	if request.AuthorizeSource != nil {
		if err := request.AuthorizeSource(root, snapshot.State); err != nil {
			return snapshot, err
		}
	}
	data, err := workpacket.ReadFile(sourceRoot, path, resultReceiptReadLimit)
	if err != nil {
		return snapshot, err
	}
	professional, issues := workpacket.ValidateReviewResultAnswer(p, data)
	if len(issues) > 0 {
		return snapshot, fmt.Errorf("S7 answer %s: %s (%s)", issues[0].Pointer, issues[0].Expected, issues[0].Observed)
	}
	var result Result
	if err := json.Unmarshal(professional, &result); err != nil {
		return snapshot, err
	}
	b := p.Binding
	result.SchemaVersion = "1.0.0"
	result.ResultID = p.ResultID
	result.AssignmentID = b.AssignmentID
	result.AssignmentRevision = b.ReviewPlanRevision
	result.ReviewPlanID = b.ReviewPlanID
	result.ReviewRound = b.ReviewRound
	result.BaselineGeneration = b.BaselineGeneration
	result.ProducerAgentID = b.OwnerAgentID
	result.SubjectDigest = b.SubjectDigest
	if result.VerificationArtifactDigest == nil {
		if b.VerificationArtifactDigest != nil {
			result.VerificationArtifactDigest = b.VerificationArtifactDigest
		} else if isReplay {
			// Reconstruct accepted machine metadata without reading mutable artifacts.
			// The original replay consumer still validates the complete receipt.
			if receipt, ok := row["result_receipt"].(map[string]any); ok {
				if digest, ok := receipt["verification_artifact_digest"].(string); ok {
					result.VerificationArtifactDigest = &digest
				}
			}
		} else if ptr := PlanPointerFromState(snapshot.State); row["lens"] == "e2e" && ptr != nil && ptr.VerificationArtifactWorkspace != "" {
			// This binds saved artifact bytes, never a professional execution verdict.
			// SubmitResult rechecks the actual workspace and all original evidence.
			digest, err := WorkspaceDigest(sourceRoot, ptr.VerificationArtifactWorkspace)
			if err != nil {
				return snapshot, err
			}
			result.VerificationArtifactDigest = &digest
		}
	}
	if b.VerificationArtifactDigest != nil && !equalOptionalString(result.VerificationArtifactDigest, b.VerificationArtifactDigest) {
		return snapshot, errors.New("S7 professional execution digest differs from frozen issuance")
	}
	if err := bindIssuedPathEvidence(root, sourceRoot, row, &result, isReplay); err != nil {
		return snapshot, err
	}
	resultData, err := json.Marshal(result)
	if err != nil {
		return snapshot, err
	}
	return SubmitResult(root, statePath, journalPath, SubmitRequest{Context: request.Context, ExpectedRevision: request.ExpectedRevision, AssignmentID: b.AssignmentID, ResultPath: filepath.Join(sourceRoot, filepath.FromSlash(path)), CaptureDir: CaptureFile(root, b.RuntimeID, b.BaselineGeneration, b.AssignmentID), OccurredAt: request.OccurredAt, resultData: resultData, issuedPacket: &p, packetOwner: request.OwnerAgentID, issuedSourceRoot: sourceRoot})
}

// Bind only the local path references consumed by the original evidence validator.
// Professional content and the author's questionnaire bytes remain unchanged.
func bindIssuedPathEvidence(root, sourceRoot string, row map[string]any, result *Result, replay bool) error {
	groups := func(value *Result) [][]string {
		var refs [][]string
		for _, claim := range value.ClaimResults {
			refs = append(refs, claim.EvidenceRefs)
		}
		for _, check := range value.Checks {
			refs = append(refs, check.EvidenceRefs)
		}
		for _, finding := range value.Findings {
			refs = append(refs, finding.EvidenceRefs)
			for _, step := range finding.Encounter.Timeline {
				refs = append(refs, step.EvidenceRefs)
			}
		}
		for _, blocked := range value.BlockedClaims {
			refs = append(refs, blocked.EvidenceRefs)
		}
		return refs
	}
	refs := groups(result)
	missing := false
	for _, group := range refs {
		for _, rawRef := range group {
			ref := strings.TrimSpace(rawRef)
			if strings.HasPrefix(ref, "path:") && !strings.Contains(ref, "#sha256=") {
				missing = true
			}
		}
	}
	if !missing {
		return nil
	}
	acceptedRefs := map[string]string{}
	if replay {
		receipt, _ := row["result_receipt"].(map[string]any)
		data, err := workpacket.ReadFile(root, stringField(receipt["result_ref"]), resultReceiptReadLimit)
		if err != nil {
			return err
		}
		var envelope struct {
			Submitted Result `json:"submitted_answer"`
		}
		if err := json.Unmarshal(data, &envelope); err != nil {
			return err
		}
		for _, group := range groups(&envelope.Submitted) {
			for _, rawRef := range group {
				ref := strings.TrimSpace(rawRef)
				if !strings.HasPrefix(ref, "path:") {
					continue
				}
				rel, digest, err := parsePathEvidenceRef(ref)
				if err != nil {
					return err
				}
				if digest != "" {
					acceptedRefs["path:"+rel] = ref
				}
			}
		}
	}
	for _, group := range refs {
		for i, rawRef := range group {
			ref := strings.TrimSpace(rawRef)
			if !strings.HasPrefix(ref, "path:") {
				continue
			}
			rel, digest, err := parsePathEvidenceRef(ref)
			if err != nil {
				return err
			}
			if digest != "" {
				continue
			}
			if replay {
				bound, ok := acceptedRefs[ref]
				if !ok {
					return fmt.Errorf("S7 replay has no original bound path evidence for %q", ref)
				}
				group[i] = bound
				continue
			}
			// Match the original local evidence resolver's containment and file rules.
			path, err := repositoryContainedPath(sourceRoot, rel)
			if err != nil {
				return err
			}
			info, err := os.Stat(path)
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() {
				return fmt.Errorf("S7 path evidence %q is not a regular file", ref)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			group[i] = ref + "#sha256=" + sha256Of(data)
		}
	}
	return nil
}

func validateIssuedResultRequest(root string, state map[string]any, request SubmitRequest, result *Result, replay bool, currentPlans ...*Plan) error {
	if request.issuedPacket == nil {
		row := resultPacketAssignmentRow(state, result.AssignmentID)
		if receipt, ok := row["result_receipt"].(map[string]any); ok && receipt["packet_id"] != nil && receipt["packet_id"] != "" {
			return errors.New("S7 Assignment has an issued answer receipt; use its exact original packet replay")
		}
		packets, err := loopruntime.CurrentReviewResultPackets(state, result.ProducerAgentID)
		if err != nil {
			return err
		}
		for _, p := range packets {
			if p.Binding.AssignmentID == result.AssignmentID {
				return errors.New("S7 Assignment has an issued answer; save its exact original packet instead of direct Result submission")
			}
		}
		return nil
	}
	p := *request.issuedPacket
	if request.packetOwner == "" || request.packetOwner != p.Binding.OwnerAgentID || request.packetOwner != result.ProducerAgentID {
		return errors.New("S7 saved owner differs from the issued packet producer")
	}
	if request.issuedSourceRoot == "" {
		return errors.New("S7 packet save has no registered execution root")
	}
	if err := verifyResultPacketIssue(root, state, p); err != nil {
		return err
	}
	if result.ResultID != p.ResultID || result.AssignmentID != p.Binding.AssignmentID || result.AssignmentRevision != p.Binding.ReviewPlanRevision || result.ReviewPlanID != p.Binding.ReviewPlanID || result.ReviewRound != p.Binding.ReviewRound || result.BaselineGeneration != p.Binding.BaselineGeneration || result.SubjectDigest != p.Binding.SubjectDigest || (p.Binding.VerificationArtifactDigest != nil && !equalOptionalString(result.VerificationArtifactDigest, p.Binding.VerificationArtifactDigest)) {
		return errors.New("S7 Result tuple differs from original packet")
	}
	row := resultPacketAssignmentRow(state, p.Binding.AssignmentID)
	if row["agent_id"] != request.packetOwner {
		return errors.New("S7 dispatched owner changed before reception")
	}
	ptr := PlanPointerFromState(state)
	if ptr == nil || ptr.PlanID != p.Binding.ReviewPlanID || repositoryPath(root, ptr.Path) != p.Binding.ReviewPlanRef || ptr.SHA256 != p.Binding.ReviewPlanSHA256 || ptr.Revision != p.Binding.ReviewPlanRevision || currentReviewRound(state) != p.Binding.ReviewRound || baselineGeneration(state) != p.Binding.BaselineGeneration {
		return errors.New("S7 issued answer is stale after Plan/round/generation revision")
	}
	sourceRoot, _, err := registeredResultPacketSourceRoot(root, state, p, !replay, request.Context)
	if err != nil {
		return err
	}
	if request.issuedSourceRoot != "" && request.issuedSourceRoot != sourceRoot {
		return errors.New("S7 source execution changed before reception")
	}
	if replay {
		return nil
	}
	// Reception already verifies the original issue and pinned Plan tuple.
	// Reuse the consumer's Plan; rebuilding issuance would repeat the full
	// product baseline here and again inside the original SubmitResult CAS.
	var plan *Plan
	if len(currentPlans) > 0 {
		plan = currentPlans[0]
	} else if plan, _, err = loadReceiptPlan(root, state); err != nil {
		return err
	}
	a := findPlanAssignment(plan, p.Binding.AssignmentID)
	if a == nil {
		return errors.New("S7 issued Assignment left current Plan")
	}
	if workstate.ReviewAssignmentActivity(row) != "dispatched" && workstate.ReviewAssignmentActivity(row) != "blocked" {
		return errors.New("S7 Assignment has no current dispatched owner")
	}
	if row["lens"] != a.Lens || !jsonValueEqual(row["claim_ids"], a.ClaimIDs) {
		return errors.New("S7 Assignment projection differs from original Claim/lens set")
	}
	assignmentSHA, err := planAssignmentDigest(a)
	if err != nil || assignmentSHA != p.Binding.AssignmentSHA256 || SubjectDigest(plan) != p.Binding.SubjectDigest {
		return errors.New("S7 issuance no longer matches current Plan/Assignment/subjects")
	}
	if err := validateResultPacketOwner(state, a, request.packetOwner); err != nil {
		return err
	}
	return validateReviewApprovalMode(root, state, request.packetOwner)
}

func validateReviewApprovalMode(root string, state map[string]any, owner string) error {
	entities, _ := state["entities"].(map[string]any)
	agents, _ := entities["agents"].([]any)
	for _, raw := range agents {
		agent, _ := raw.(map[string]any)
		if agent["id"] != owner {
			continue
		}
		mode, _ := agent["dispatch_mode"].(string)
		if mode == "plan_checkpoint" || mode == "one_shot" {
			return nil
		}
		// Explicit approval-mode and legacy dispatches retain their original
		// approved activation/readback chain. Only generic continuous PLAN
		// checkpoints are removed from the S7 professional answer flow.
		return agentauth.ValidateWorkingAuthorization(root, state, owner)
	}
	return fmt.Errorf("S7 Reviewer %s is not registered", owner)
}

func resultPacketSourceRoot(root string, state map[string]any, p workpacket.ReviewResultPacket) (string, bool, error) {
	return registeredResultPacketSourceRoot(root, state, p, true)
}

// registeredResultPacketSourceRoot resolves the packet's Main or Worker
// checkout from the current original Runtime registry. validateLive is used
// for a new save. Receipt replay passes false so it checks the persisted tuple
// and path identity without touching a Worker's live checkout.
func registeredResultPacketSourceRoot(root string, state map[string]any, p workpacket.ReviewResultPacket, validateLive bool, contexts ...context.Context) (string, bool, error) {
	parent := context.Background()
	if len(contexts) > 0 && contexts[0] != nil {
		parent = contexts[0]
	}
	main, err := workspace.Canonical(root)
	if err != nil {
		return "", false, err
	}
	b, err := workspace.Decode(state)
	if err != nil {
		return "", false, err
	}
	if b == nil {
		return main, false, nil
	}
	if err := b.ValidateAuthority(state); err != nil {
		return "", false, err
	}
	e, declared := b.Executions[p.Binding.AssignmentID]
	if !declared {
		return main, false, nil
	}
	if e.AssignmentID != p.Binding.AssignmentID || e.RuntimeID != p.Binding.RuntimeID || e.BaselineGeneration != p.Binding.BaselineGeneration || e.AgentID != p.Binding.OwnerAgentID || e.Status != "ready" || b.MainRoot != main || e.Path == main {
		return "", true, errors.New("S7 original execution owner/status/root differs")
	}
	if validateLive {
		ctx, cancel := context.WithTimeout(parent, 3*time.Second)
		defer cancel()
		if err := b.ValidateExecution(ctx, e, false); err != nil {
			return "", true, err
		}
		path, err := workspace.Canonical(e.Path)
		if err != nil || path != e.Path {
			return "", true, errors.New("S7 registered execution path is not canonical")
		}
		return path, true, nil
	}
	if !filepath.IsAbs(e.Path) || filepath.Clean(e.Path) != e.Path {
		return "", true, errors.New("S7 registered execution path is not a canonical absolute path")
	}
	return e.Path, true, nil
}

// IssuedResultReceipt reconstructs acceptance from original immutable facts;
// an authoring file is neither required nor interpreted as reception.
func IssuedResultReceipt(root string, state map[string]any, packetID, owner string) (ResultReceipt, bool, error) {
	p, found, err := loopruntime.FindReviewResultPacket(state, packetID)
	if err != nil {
		return ResultReceipt{}, false, err
	}
	if !found || owner == "" || p.Binding.OwnerAgentID != owner {
		return ResultReceipt{}, false, errors.New("S7 receipt has no issuance for owner")
	}
	row := resultPacketAssignmentRow(state, p.Binding.AssignmentID)
	if row["result_receipt"] == nil {
		return ResultReceipt{}, false, nil
	}
	encoded, err := json.Marshal(row["result_receipt"])
	if err != nil {
		return ResultReceipt{}, false, err
	}
	var receipt ResultReceipt
	if err := json.Unmarshal(encoded, &receipt); err != nil {
		return receipt, false, err
	}
	if receipt.PacketID != packetID || receipt.ProducerAgentID != owner {
		return receipt, false, errors.New("S7 current receipt belongs to a different issuance")
	}
	data, err := workpacket.ReadFile(root, receipt.ResultRef, resultReceiptReadLimit)
	if err != nil {
		return receipt, false, err
	}
	var envelope struct {
		SubmittedAnswer json.RawMessage `json:"submitted_answer"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return receipt, false, err
	}
	var result Result
	if err := json.Unmarshal(envelope.SubmittedAnswer, &result); err != nil {
		return receipt, false, err
	}
	if err := verifyResultReceiptClosure(root, state, row, receipt, &result); err != nil {
		return receipt, false, err
	}
	return receipt, true, nil
}
