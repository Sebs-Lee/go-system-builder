// Package acceptance owns the small, machine-readable S10 audit manifest.
//
// The Markdown ACC and release-audit reports remain the human-readable
// records. This package validates only the finite completion ledger that the
// Quality Gate needs to consume; it does not attempt to parse prose tables.
package acceptance

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/entroforge/go-system-builder/internal/schema"
)

const (
	SchemaVersionAcceptance = "1.0.0"
	ManifestAcceptance      = "acceptance"
	ManifestReleaseAudit    = "release_audit"
)

// AuditAreaIDs is the finite release-audit denominator from L3-S10 §4.5.
var AuditAreaIDs = []string{
	"state_machine",
	"transaction_uow",
	"concurrency_idempotency",
	"data_migration",
	"call_sites_topology",
	"observability_errors",
	"verification_evidence",
	"docs_release_scope",
}

// Manifest is the structured completion ledger referenced by an S10
// acceptance or release-audit evidence envelope.
type Manifest struct {
	SchemaVersion      string                `json:"schema_version"`
	ManifestType       string                `json:"manifest_type"`
	RuntimeID          string                `json:"runtime_id"`
	BaselineGeneration int                   `json:"baseline_generation"`
	ReviewRound        int                   `json:"review_round"`
	CoverageInventory  []CoverageItem        `json:"coverage_inventory"`
	Counterevidence    []CounterevidenceItem `json:"counterevidence"`
	AuditAreas         []AuditArea           `json:"audit_areas,omitempty"`
	Risks              []RiskItem            `json:"risks"`
	TechnicalDebt      []DebtItem            `json:"technical_debt"`
	BlockingFindings   []BlockingFinding     `json:"blocking_findings"`
	Metrics            Metrics               `json:"metrics"`
}

// CoverageItem is one finite object in the frozen S10 review denominator.
type CoverageItem struct {
	ID           string   `json:"id"`
	Category     string   `json:"category"`
	SourceRefs   []string `json:"source_refs"`
	Expected     string   `json:"expected"`
	Oracle       string   `json:"oracle"`
	Owner        string   `json:"owner"`
	EvidenceRefs []string `json:"evidence_refs"`
	Disposition  string   `json:"disposition"`
	NAReason     string   `json:"na_reason,omitempty"`
}

// CounterevidenceItem records the deliberate attempt to disprove one
// coverage conclusion.
type CounterevidenceItem struct {
	ID           string   `json:"id"`
	InventoryID  string   `json:"inventory_id"`
	Question     string   `json:"question"`
	EvidenceRefs []string `json:"evidence_refs"`
	Outcome      string   `json:"outcome"`
}

// AuditArea is one of the eight release-architecture audit areas.
type AuditArea struct {
	ID           string   `json:"id"`
	Conclusion   string   `json:"conclusion"`
	Owner        string   `json:"owner"`
	EvidenceRefs []string `json:"evidence_refs"`
}

// RiskItem keeps every non-blocking risk accountable to an owner, tracking
// artifact, impact, and recovery point before it can be handed to S11.
type RiskItem struct {
	ID            string `json:"id"`
	Severity      string `json:"severity"`
	Impact        string `json:"impact"`
	Owner         string `json:"owner"`
	TrackingRef   string `json:"tracking_ref"`
	RecoveryPoint string `json:"recovery_point"`
}

// DebtItem prevents technical debt from disappearing merely because it does
// not block the release.
type DebtItem struct {
	ID          string `json:"id"`
	Impact      string `json:"impact"`
	Owner       string `json:"owner"`
	TrackingRef string `json:"tracking_ref"`
}

// BlockingFinding is an explicit route-bearing blocker. A non-empty list
// makes the S10 manifest fail; it cannot be hidden in a prose report.
type BlockingFinding struct {
	ID    string `json:"id"`
	Route string `json:"route"`
}

// Metrics are objective S10 completion indicators. Coverage values are
// checked against the manifest's declared category rows. The hard metric
// categories must be represented explicitly; optional categories may have a
// zero denominator only when they are genuinely outside this audit scope.
type Metrics struct {
	RequirementCoverage  float64 `json:"requirement_coverage"`
	ContractCoverage     float64 `json:"contract_coverage"`
	ChangedPathCoverage  float64 `json:"changed_path_coverage"`
	AuditAreaCoverage    float64 `json:"audit_area_coverage"`
	UnknownCount         int     `json:"unknown_count"`
	UnsupportedPassCount int     `json:"unsupported_pass_count"`
	UnownedRiskCount     int     `json:"unowned_risk_count"`
	UntrackedDebtCount   int     `json:"untracked_debt_count"`
	BlockingFindingCount int     `json:"blocking_finding_count"`
}

// Summary is the gate-facing, derived view of a valid manifest.
type Summary struct {
	ManifestType         string
	InventoryCount       int
	DispositionedCount   int
	CounterevidenceCount int
	AuditAreaCount       int
	BlockingFindingCount int
	UnknownCount         int
	UnsupportedPassCount int
	EvidenceRefs         []string
	Metrics              Metrics
}

// Validate decodes and validates one S10 manifest as a completion artifact.
// expectedType must be "acceptance" or "release_audit". Errors name the
// exact row or metric and include the recovery action an Agent should take.
func Validate(data []byte, expectedType string) (Summary, error) {
	return validate(data, expectedType, true)
}

// ValidateForOutcome validates the manifest mode appropriate for an evidence
// envelope. Passing outcomes require a clean ledger. A routed
// review_required/blocked outcome still needs a structurally complete,
// evidence-linked ledger, but may retain the unresolved rows that explain
// why it must return through S7/S8/S9 or pause.
func ValidateForOutcome(data []byte, expectedType, outcome string) (Summary, error) {
	requireClean := !allowsUnresolvedOutcome(expectedType, outcome)
	return validate(data, expectedType, requireClean)
}

func allowsUnresolvedOutcome(expectedType, outcome string) bool {
	return (expectedType == ManifestAcceptance && outcome == "review_required") ||
		(expectedType == ManifestReleaseAudit && outcome == "blocked")
}

func validate(data []byte, expectedType string, requireClean bool) (Summary, error) {
	var manifest Manifest
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return Summary{}, fmt.Errorf("S10 manifest JSON is invalid: %w; rewrite the file from the S10 manifest shape and run `loop-harness s10 manifest validate --file <path>`", err)
	}
	if err := schema.NewEmbeddedValidator().ValidateBytes("s10-audit-manifest.schema.json", data); err != nil {
		return Summary{}, fmt.Errorf("S10 manifest does not satisfy s10-audit-manifest.schema.json: %w; correct the named field and rerun `loop-harness s10 manifest validate --file <path>`", err)
	}

	summary := Summary{
		ManifestType:         manifest.ManifestType,
		InventoryCount:       len(manifest.CoverageInventory),
		CounterevidenceCount: len(manifest.Counterevidence),
		AuditAreaCount:       len(manifest.AuditAreas),
		BlockingFindingCount: len(manifest.BlockingFindings),
		Metrics:              manifest.Metrics,
	}
	var issues []string
	if manifest.SchemaVersion != SchemaVersionAcceptance {
		issues = append(issues, fmt.Sprintf("schema_version must be %q (got %q)", SchemaVersionAcceptance, manifest.SchemaVersion))
	}
	if expectedType != ManifestAcceptance && expectedType != ManifestReleaseAudit {
		issues = append(issues, fmt.Sprintf("expected manifest type must be acceptance or release_audit (got %q)", expectedType))
	} else if manifest.ManifestType != expectedType {
		issues = append(issues, fmt.Sprintf("manifest_type must be %q (got %q)", expectedType, manifest.ManifestType))
	}
	if strings.TrimSpace(manifest.RuntimeID) == "" {
		issues = append(issues, "runtime_id is required; copy it from the current Runtime")
	}
	if manifest.BaselineGeneration < 1 {
		issues = append(issues, "baseline_generation must be at least 1; bind the manifest to the current baseline")
	}
	if manifest.ReviewRound < 1 {
		issues = append(issues, "review_round must be at least 1; bind the manifest to the current S7 clean round")
	}
	if len(manifest.CoverageInventory) == 0 {
		issues = append(issues, "coverage_inventory is empty; freeze the finite S10 review denominator before writing PASS")
	}

	itemsByID := make(map[string]CoverageItem, len(manifest.CoverageInventory))
	evidenceRefs := make(map[string]struct{})
	categoryCounts := make(map[string]int)
	categoryDispositioned := make(map[string]int)
	for i, item := range manifest.CoverageInventory {
		row := fmt.Sprintf("coverage_inventory[%d] (%s)", i, item.ID)
		if strings.TrimSpace(item.ID) == "" {
			issues = append(issues, fmt.Sprintf("%s.id is required", row))
			continue
		}
		if _, exists := itemsByID[item.ID]; exists {
			issues = append(issues, fmt.Sprintf("%s is duplicate; keep one frozen coverage row per object", row))
			continue
		}
		itemsByID[item.ID] = item
		categoryCounts[item.Category]++
		if item.Disposition == "pass" || item.Disposition == "not_applicable" {
			categoryDispositioned[item.Category]++
			summary.DispositionedCount++
		} else if item.Disposition == "unknown" {
			summary.UnknownCount++
		}
		if strings.TrimSpace(item.Category) == "" {
			issues = append(issues, fmt.Sprintf("%s.category is required", row))
		}
		if len(nonEmpty(item.SourceRefs)) == 0 {
			issues = append(issues, fmt.Sprintf("%s.source_refs is required; every conclusion needs an authoritative source", row))
		}
		for _, ref := range nonEmpty(item.EvidenceRefs) {
			evidenceRefs[ref] = struct{}{}
		}
		if strings.TrimSpace(item.Expected) == "" {
			issues = append(issues, fmt.Sprintf("%s.expected is required", row))
		}
		if strings.TrimSpace(item.Oracle) == "" {
			issues = append(issues, fmt.Sprintf("%s.oracle is required; state how the expected fact was observed", row))
		}
		if strings.TrimSpace(item.Owner) == "" {
			issues = append(issues, fmt.Sprintf("%s.owner is required; assign one accountable responsibility", row))
		}
		switch item.Disposition {
		case "pass":
			if len(nonEmpty(item.EvidenceRefs)) == 0 {
				summary.UnsupportedPassCount++
				issues = append(issues, fmt.Sprintf("%s is PASS without evidence_refs; replace unsupported PASS with evidence or UNKNOWN", row))
			}
		case "not_applicable":
			if strings.TrimSpace(item.NAReason) == "" {
				issues = append(issues, fmt.Sprintf("%s is not_applicable without na_reason; cite the authoritative scope decision", row))
			}
		case "unknown", "fail":
			if requireClean {
				issues = append(issues, fmt.Sprintf("%s is %s; resolve it or route back through S7/S8/S9 before S10 can pass", row, item.Disposition))
			}
		default:
			issues = append(issues, fmt.Sprintf("%s.disposition must be pass, not_applicable, unknown, or fail", row))
		}
	}
	requiredCategories := []string{"requirement", "contract", "changed_path"}
	if expectedType == ManifestReleaseAudit {
		requiredCategories = append(requiredCategories, "audit_area")
	}
	for _, category := range requiredCategories {
		if categoryCounts[category] == 0 {
			issues = append(issues, fmt.Sprintf("coverage_inventory has no %s rows; add an explicit pass or not_applicable row with source_refs and evidence", category))
		}
	}

	seenCounterevidence := make(map[string]struct{}, len(manifest.Counterevidence))
	for i, item := range manifest.Counterevidence {
		row := fmt.Sprintf("counterevidence[%d] (%s)", i, item.ID)
		if strings.TrimSpace(item.ID) == "" {
			issues = append(issues, fmt.Sprintf("%s.id is required", row))
		}
		if _, exists := seenCounterevidence[item.InventoryID]; exists {
			issues = append(issues, fmt.Sprintf("%s duplicates inventory_id %q; provide exactly one counterevidence row per coverage item", row, item.InventoryID))
		}
		seenCounterevidence[item.InventoryID] = struct{}{}
		if _, exists := itemsByID[item.InventoryID]; !exists {
			issues = append(issues, fmt.Sprintf("%s.inventory_id %q does not reference coverage_inventory; link the question to a frozen row", row, item.InventoryID))
		}
		if strings.TrimSpace(item.Question) == "" {
			issues = append(issues, fmt.Sprintf("%s.question is required; state what would disprove the conclusion", row))
		}
		if len(nonEmpty(item.EvidenceRefs)) == 0 && item.Outcome != "unknown" {
			issues = append(issues, fmt.Sprintf("%s.evidence_refs is empty; record the negative-path check before marking %s", row, item.Outcome))
		}
		for _, ref := range nonEmpty(item.EvidenceRefs) {
			evidenceRefs[ref] = struct{}{}
		}
		switch item.Outcome {
		case "pass", "not_applicable":
		case "unknown", "fail":
			summary.UnknownCount++
			if requireClean {
				issues = append(issues, fmt.Sprintf("%s.outcome is %s; unresolved counterevidence cannot enter an S10 PASS", row, item.Outcome))
			}
		default:
			issues = append(issues, fmt.Sprintf("%s.outcome must be pass, not_applicable, unknown, or fail", row))
		}
	}
	if len(seenCounterevidence) != len(itemsByID) {
		missing := make([]string, 0)
		for id := range itemsByID {
			if _, ok := seenCounterevidence[id]; !ok {
				missing = append(missing, id)
			}
		}
		sort.Strings(missing)
		issues = append(issues, fmt.Sprintf("counterevidence is missing for coverage_inventory %s; one disproof question is required per item", strings.Join(missing, ", ")))
	}

	if expectedType == ManifestReleaseAudit {
		areas := make(map[string]AuditArea, len(manifest.AuditAreas))
		for i, area := range manifest.AuditAreas {
			row := fmt.Sprintf("audit_areas[%d] (%s)", i, area.ID)
			if _, exists := areas[area.ID]; exists {
				issues = append(issues, fmt.Sprintf("%s is duplicate; keep one row per audit area", row))
			}
			areas[area.ID] = area
			if strings.TrimSpace(area.Owner) == "" {
				issues = append(issues, fmt.Sprintf("%s.owner is required", row))
			}
			if len(nonEmpty(area.EvidenceRefs)) == 0 {
				issues = append(issues, fmt.Sprintf("%s.evidence_refs is required", row))
			}
			for _, ref := range nonEmpty(area.EvidenceRefs) {
				evidenceRefs[ref] = struct{}{}
			}
			if area.Conclusion != "pass" && area.Conclusion != "not_applicable" {
				issues = append(issues, fmt.Sprintf("%s.conclusion must be pass or not_applicable", row))
			}
		}
		missing := make([]string, 0)
		for _, id := range AuditAreaIDs {
			if _, ok := areas[id]; !ok {
				missing = append(missing, id)
			}
		}
		if len(missing) > 0 || len(areas) != len(AuditAreaIDs) {
			issues = append(issues, fmt.Sprintf("audit_areas must contain all 8 areas; missing %s", strings.Join(missing, ", ")))
		}
	}

	for i, risk := range manifest.Risks {
		row := fmt.Sprintf("risks[%d] (%s)", i, risk.ID)
		if strings.TrimSpace(risk.ID) == "" || strings.TrimSpace(risk.Impact) == "" || strings.TrimSpace(risk.Owner) == "" || strings.TrimSpace(risk.TrackingRef) == "" || strings.TrimSpace(risk.RecoveryPoint) == "" {
			issues = append(issues, fmt.Sprintf("%s requires id, impact, owner, tracking_ref, and recovery_point; unowned or untracked risks cannot enter S11", row))
		}
		if risk.Severity != "P0" && risk.Severity != "P1" && risk.Severity != "P2" && risk.Severity != "P3" {
			issues = append(issues, fmt.Sprintf("%s.severity must be P0, P1, P2, or P3", row))
		}
		// RC-02 (S10-10): a P0 risk is business-blocking by definition. It
		// must be routed through blocking_findings (S7/S8/S9 or pause) and
		// cannot be parked as a monitored non-blocking risk that silently
		// rides into S11.
		if risk.Severity == "P0" {
			issues = append(issues, fmt.Sprintf("%s.severity is P0; P0 risks are business-blocking and must be routed through blocking_findings with a resolution route (S7/S8/S9 or pause), not parked as a monitored risk entering S11", row))
		}
	}
	for i, debt := range manifest.TechnicalDebt {
		row := fmt.Sprintf("technical_debt[%d] (%s)", i, debt.ID)
		if strings.TrimSpace(debt.ID) == "" || strings.TrimSpace(debt.Impact) == "" || strings.TrimSpace(debt.Owner) == "" || strings.TrimSpace(debt.TrackingRef) == "" {
			issues = append(issues, fmt.Sprintf("%s requires id, impact, owner, and tracking_ref; untracked debt cannot enter S11", row))
		}
	}
	for i, finding := range manifest.BlockingFindings {
		row := fmt.Sprintf("blocking_findings[%d] (%s)", i, finding.ID)
		if strings.TrimSpace(finding.ID) == "" || strings.TrimSpace(finding.Route) == "" {
			issues = append(issues, fmt.Sprintf("%s requires id and route; resolve the blocker through S7/S8/S9 or pause", row))
		}
	}

	validateMetrics(&issues, manifest.Metrics, categoryCounts, categoryDispositioned, summary, manifest.ManifestType, requireClean)
	if len(issues) > 0 {
		return summary, fmt.Errorf("S10 manifest invalid: %s; next: correct the named rows, run `loop-harness s10 manifest validate --file <path>`, then register a new fingerprinted evidence envelope", strings.Join(issues, "; "))
	}
	for ref := range evidenceRefs {
		summary.EvidenceRefs = append(summary.EvidenceRefs, ref)
	}
	sort.Strings(summary.EvidenceRefs)
	return summary, nil
}

// ValidateEvidenceArtifact validates the S10-specific part of an evidence
// envelope before Runtime registers it. Generic envelope fields remain the
// responsibility of the existing Runtime/Quality Gate checks; this helper
// only closes the envelope -> immutable manifest edge.
func ValidateEvidenceArtifact(root, kind string, envelopeData []byte) error {
	expectedType := ""
	switch kind {
	case ManifestAcceptance, "acceptance_record":
		expectedType = ManifestAcceptance
	case ManifestReleaseAudit, "release_audit_record":
		expectedType = ManifestReleaseAudit
	default:
		return nil
	}
	var envelope struct {
		ManifestPath string `json:"audit_manifest_path"`
		ManifestSHA  string `json:"audit_manifest_sha256"`
		Conclusion   string `json:"conclusion"`
	}
	if err := json.Unmarshal(envelopeData, &envelope); err != nil {
		return fmt.Errorf("S10 %s evidence envelope is not valid JSON: %w", expectedType, err)
	}
	if strings.TrimSpace(envelope.ManifestPath) == "" || strings.TrimSpace(envelope.ManifestSHA) == "" {
		return fmt.Errorf("S10 %s evidence requires audit_manifest_path and audit_manifest_sha256; validate the manifest first with `loop-harness s10 manifest validate --file <path> --type %s`, then register a new envelope", expectedType, expectedType)
	}
	manifestPath, err := safeManifestPath(root, envelope.ManifestPath)
	if err != nil {
		return fmt.Errorf("S10 %s evidence manifest path: %w", expectedType, err)
	}
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("S10 %s evidence manifest %q is unreadable: %w; validate the referenced file and register a new envelope", expectedType, envelope.ManifestPath, err)
	}
	if sum := sha256.Sum256(manifestData); hex.EncodeToString(sum[:]) != envelope.ManifestSHA {
		return fmt.Errorf("S10 %s evidence audit_manifest_sha256 does not match %q; do not edit in place, regenerate the manifest and register a new envelope", expectedType, envelope.ManifestPath)
	}
	if _, err := ValidateForOutcome(manifestData, expectedType, envelope.Conclusion); err != nil {
		return fmt.Errorf("S10 %s evidence manifest %q is invalid: %w", expectedType, envelope.ManifestPath, err)
	}
	return nil
}

func safeManifestPath(root, value string) (string, error) {
	clean := filepath.Clean(value)
	if filepath.IsAbs(clean) || clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("audit_manifest_path must stay inside the repository: %q", value)
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve repository root: %w", err)
	}
	resolvedRoot, err := filepath.EvalSymlinks(rootAbs)
	if err != nil {
		return "", fmt.Errorf("resolve repository root symlinks: %w", err)
	}
	resolvedPath, err := filepath.EvalSymlinks(filepath.Join(rootAbs, clean))
	if err != nil {
		return "", fmt.Errorf("resolve manifest symlinks: %w", err)
	}
	relative, err := filepath.Rel(resolvedRoot, resolvedPath)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("audit_manifest_path must stay inside the repository: %q", value)
	}
	return filepath.Join(rootAbs, clean), nil
}

func validateMetrics(issues *[]string, metrics Metrics, counts, dispositioned map[string]int, summary Summary, manifestType string, requireClean bool) {
	coverage := []struct {
		name     string
		category string
		value    float64
	}{
		{"requirement_coverage", "requirement", metrics.RequirementCoverage},
		{"contract_coverage", "contract", metrics.ContractCoverage},
		{"changed_path_coverage", "changed_path", metrics.ChangedPathCoverage},
	}
	if manifestType == ManifestReleaseAudit {
		coverage = append(coverage, struct {
			name     string
			category string
			value    float64
		}{"audit_area_coverage", "audit_area", metrics.AuditAreaCoverage})
	}
	for _, item := range coverage {
		if item.value < 0 || item.value > 1 || math.IsNaN(item.value) {
			*issues = append(*issues, fmt.Sprintf("metrics.%s must be between 0 and 1", item.name))
			continue
		}
		want := 1.0
		if item.category == "audit_area" && manifestType == ManifestReleaseAudit {
			want = float64(summary.AuditAreaCount) / float64(len(AuditAreaIDs))
		} else if counts[item.category] > 0 {
			want = float64(dispositioned[item.category]) / float64(counts[item.category])
		}
		if math.Abs(item.value-want) > 0.000001 {
			*issues = append(*issues, fmt.Sprintf("metrics.%s=%g does not match %s coverage %g; derive it from the frozen rows", item.name, item.value, item.category, want))
		}
	}
	if metrics.UnknownCount != summary.UnknownCount {
		*issues = append(*issues, fmt.Sprintf("metrics.unknown_count=%d does not match derived unknown_count=%d", metrics.UnknownCount, summary.UnknownCount))
	}
	if metrics.UnsupportedPassCount != summary.UnsupportedPassCount {
		*issues = append(*issues, fmt.Sprintf("metrics.unsupported_pass_count=%d does not match derived unsupported_pass_count=%d", metrics.UnsupportedPassCount, summary.UnsupportedPassCount))
	}
	if metrics.UnownedRiskCount != 0 {
		*issues = append(*issues, fmt.Sprintf("metrics.unowned_risk_count=%d; every risk must have owner, tracking_ref, and recovery_point", metrics.UnownedRiskCount))
	}
	if metrics.UntrackedDebtCount != 0 {
		*issues = append(*issues, fmt.Sprintf("metrics.untracked_debt_count=%d; every technical_debt row must have owner and tracking_ref", metrics.UntrackedDebtCount))
	}
	if metrics.BlockingFindingCount != summary.BlockingFindingCount {
		*issues = append(*issues, fmt.Sprintf("metrics.blocking_finding_count=%d does not match blocking_findings=%d", metrics.BlockingFindingCount, summary.BlockingFindingCount))
	}
	if requireClean {
		for name, value := range map[string]int{
			"blocking_finding_count": metrics.BlockingFindingCount,
		} {
			if value != 0 {
				*issues = append(*issues, fmt.Sprintf("metrics.%s=%d; S10 PASS requires zero and an owner/tracking/route for every exception", name, value))
			}
		}
	}
}

func nonEmpty(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			result = append(result, value)
		}
	}
	return result
}
