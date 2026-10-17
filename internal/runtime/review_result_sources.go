package runtime

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/entroforge/go-system-builder/internal/reviewmeta"
	"github.com/entroforge/go-system-builder/internal/schema"
	"github.com/entroforge/go-system-builder/internal/workpacket"
)

// verifyReviewResultPacketSources re-reads the original pinned Plan through
// the Runtime authority root. It proves that the packet's Assignment digest,
// ordered Claim IDs, full Claim source objects, and frozen subject objects all
// came from those exact schema-validated Plan bytes.
func verifyReviewResultPacketSources(root string, packet workpacket.ReviewResultPacket) error {
	b := packet.Binding
	planBytes, err := workpacket.ReadFile(root, b.ReviewPlanRef, reviewResultArtifactLimit)
	if err != nil {
		return fmt.Errorf("read original pinned S7 ReviewPlan: %w", err)
	}
	if sha256Bytes(planBytes) != b.ReviewPlanSHA256 {
		return errors.New("original pinned S7 ReviewPlan bytes differ from the packet digest")
	}
	if err := workpacket.ValidateUniqueJSON(planBytes); err != nil {
		return fmt.Errorf("original pinned S7 ReviewPlan has duplicate JSON keys: %w", err)
	}
	if err := schema.NewEmbeddedValidator().ValidateBytes("review-plan.schema.json", planBytes); err != nil {
		return fmt.Errorf("original pinned S7 ReviewPlan schema: %w", err)
	}
	var plan reviewResultPlanSource
	if err := json.Unmarshal(planBytes, &plan); err != nil {
		return fmt.Errorf("decode original pinned S7 ReviewPlan source: %w", err)
	}
	if plan.SchemaVersion != "1.0.0" || plan.ReviewPlanID != b.ReviewPlanID || plan.ReviewRound != b.ReviewRound || plan.BaselineGeneration != b.BaselineGeneration {
		return errors.New("original pinned S7 ReviewPlan identity differs from the packet")
	}

	var original *reviewmeta.PlanAssignment
	for i := range plan.Assignments {
		if plan.Assignments[i].AssignmentID != b.AssignmentID {
			continue
		}
		if original != nil {
			return errors.New("original pinned S7 ReviewPlan has duplicate Assignment IDs")
		}
		assignment := plan.Assignments[i]
		original = &assignment
	}
	if original == nil || !sameStrings(original.ClaimIDs, b.ClaimIDs) || hasDuplicateStrings(original.ClaimIDs) {
		return errors.New("original pinned S7 Assignment Claim IDs differ from the packet")
	}
	assignmentBytes, err := json.Marshal(original)
	if err != nil || sha256Bytes(assignmentBytes) != b.AssignmentSHA256 {
		return errors.New("original pinned S7 Assignment JSON digest differs from the packet")
	}

	claimsByID := make(map[string]reviewmeta.Claim, len(plan.Claims))
	for _, claim := range plan.Claims {
		if claim.ClaimID == "" {
			return errors.New("original pinned S7 ReviewPlan contains a malformed Claim")
		}
		if _, duplicate := claimsByID[claim.ClaimID]; duplicate {
			return errors.New("original pinned S7 ReviewPlan has duplicate Claim IDs")
		}
		claimsByID[claim.ClaimID] = claim
	}
	selectedClaims := make([]reviewmeta.Claim, 0, len(original.ClaimIDs))
	for _, id := range original.ClaimIDs {
		claim, found := claimsByID[id]
		if !found {
			return fmt.Errorf("original pinned S7 Assignment Claim %s is absent from the Plan", id)
		}
		selectedClaims = append(selectedClaims, claim)
	}
	selectedClaimBytes, err := json.Marshal(selectedClaims)
	if err != nil {
		return fmt.Errorf("encode original pinned S7 Claim selection: %w", err)
	}
	claimSource, err := canonicalRuntimeJSON(selectedClaimBytes)
	if err != nil {
		return fmt.Errorf("canonicalize original pinned S7 Claims: %w", err)
	}
	packetClaims, err := canonicalRuntimeJSON(packet.Inputs.Claims)
	if err != nil || !bytes.Equal(claimSource, packetClaims) {
		return errors.New("S7 packet Claim inputs differ from the original pinned Plan")
	}

	subjectBytes, err := json.Marshal(plan.FrozenSubjects)
	if err != nil {
		return fmt.Errorf("encode original pinned S7 frozen subjects: %w", err)
	}
	subjectSource, err := canonicalRuntimeJSON(subjectBytes)
	if err != nil {
		return fmt.Errorf("canonicalize original pinned S7 frozen subjects: %w", err)
	}
	packetSubjects, err := canonicalRuntimeJSON(packet.Inputs.FrozenSubjects)
	if err != nil || !bytes.Equal(subjectSource, packetSubjects) {
		return errors.New("S7 packet frozen subject inputs differ from the original pinned Plan")
	}
	return nil
}

type reviewResultPlanSource struct {
	SchemaVersion      string                      `json:"schema_version"`
	ReviewPlanID       string                      `json:"review_plan_id"`
	ReviewRound        int                         `json:"review_round"`
	BaselineGeneration int                         `json:"baseline_generation"`
	FrozenSubjects     []reviewmeta.FrozenSubject  `json:"frozen_subjects"`
	Claims             []reviewmeta.Claim          `json:"claims"`
	Assignments        []reviewmeta.PlanAssignment `json:"assignments"`
}

func canonicalRuntimeJSON(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("empty JSON")
	}
	if err := workpacket.ValidateUniqueJSON(data); err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return nil, errors.New("trailing JSON value")
		}
		return nil, err
	}
	return json.Marshal(value)
}

func hasDuplicateStrings(values []string) bool {
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		if seen[value] {
			return true
		}
		seen[value] = true
	}
	return false
}
