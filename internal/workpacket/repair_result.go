package workpacket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/entroforge/go-system-builder/internal/schema"
)

const repairResultTemplateID = "s9-repair-result-v1"

var (
	repairResultPacketLimit = 20 << 20
	repairResultSourceLimit = 4 << 20
	repairResultPacketIDRE  = regexp.MustCompile(`^packet-[a-f0-9]{64}$`)
	repairResultIDRE        = regexp.MustCompile(`^repair-result-[a-f0-9]{64}$`)
)

// RepairResultBinding freezes the original PlanReport authority and exact
// registered execution definition for one S9 result questionnaire. The
// neutral workpacket layer does not grant the authority represented here; the
// original repair/runtime consumer must reread and verify each source.
type RepairResultBinding struct {
	Sources          RepairPlanReportBinding `json:"sources"`
	PlanReportRef    string                  `json:"plan_report_ref"`
	PlanReportSHA256 string                  `json:"plan_report_sha256"`
	ExecutionSHA256  string                  `json:"execution_sha256"`
	// ReworkRef and ReworkSHA256 are deliberately present and empty on the
	// first attempt. A later attempt binds the exact immutable rework receipt.
	ReworkRef    string `json:"rework_ref"`
	ReworkSHA256 string `json:"rework_sha256"`
	// Resume binds a new answer attempt to one accepted Main blocker resolution.
	// It is omitted on original packets so their canonical bytes and identities
	// remain unchanged.
	Resume *RepairResultResume `json:"resume,omitempty"`
}

// RepairResultResume is the immutable, source-derived lineage from the prior
// packet's blocker through the accepted Main resolution. It contains facts,
// not permission; runtime readers must reconstruct and verify the original
// packet, evidence and journal event before treating it as current.
type RepairResultResume struct {
	PredecessorPacketID      string `json:"predecessor_packet_id"`
	PredecessorPacketSHA256  string `json:"predecessor_packet_sha256"`
	BlockerRef               string `json:"blocker_ref"`
	BlockerSHA256            string `json:"blocker_sha256"`
	ResolutionMessageRef     string `json:"resolution_message_ref"`
	ResolutionMessageSHA256  string `json:"resolution_message_sha256"`
	ResolutionEventID        string `json:"resolution_event_id"`
	ResolutionIdempotencyKey string `json:"resolution_idempotency_key"`
	ResolutionBeforeRevision int    `json:"resolution_before_revision"`
	ResolutionAfterRevision  int    `json:"resolution_after_revision"`
	ResolutionSequence       int    `json:"resolution_sequence"`
}

// RepairResultInputs retains canonical source documents and the neutral,
// closed projection of the registered workspace.Execution definition.
// Mutable execution progress (status, tool receipts, and delivery pointers)
// is excluded and must be reread by the original consumer at acceptance.
type RepairResultInputs struct {
	Sources       RepairPlanReportInputs `json:"sources"`
	PlanReport    json.RawMessage        `json:"plan_report"`
	Execution     json.RawMessage        `json:"execution"`
	ReworkReceipt json.RawMessage        `json:"rework_receipt"`
}

func (inputs *RepairResultInputs) UnmarshalJSON(data []byte) error {
	if err := ValidateUniqueJSON(data); err != nil {
		return err
	}
	type wire RepairResultInputs
	var decoded wire
	if err := decodeOneStrictJSON(data, &decoded); err != nil {
		return err
	}
	for _, source := range []struct {
		name string
		raw  *json.RawMessage
	}{
		{"PlanReport", &decoded.PlanReport},
		{"Execution", &decoded.Execution},
	} {
		canonical, err := canonicalRepairResultObject(*source.raw, source.name)
		if err != nil {
			return err
		}
		*source.raw = canonical
	}
	rework, err := canonicalRepairResultNullableObject(decoded.ReworkReceipt, "ReworkReceipt")
	if err != nil {
		return err
	}
	decoded.ReworkReceipt = rework
	*inputs = RepairResultInputs(decoded)
	return nil
}

// RepairResultBeforeFixResult records an author's direct professional
// observation for one exact frozen Contract assertion. It is not synthesized
// from PlanReport command output.
type RepairResultBeforeFixResult struct {
	AssertionID  string   `json:"assertion_id"`
	ObservedRed  string   `json:"observed_red"`
	EvidenceRefs []string `json:"evidence_refs"`
}

// RepairResultUnitResult records the professional implementation conclusion
// for one exact assigned unit. Machine-derived changes and checks are not part
// of this type.
type RepairResultUnitResult struct {
	UnitID                string   `json:"unit_id"`
	Status                string   `json:"status"`
	EvidenceRefs          []string `json:"evidence_refs"`
	ImplementationSummary string   `json:"implementation_summary"`
	RestoredInvariantRefs []string `json:"restored_invariant_refs"`
}

// RepairResultProfessionalAnswer contains only author-owned S9 professional
// conclusions and references. The original domain writer maps verdicts and
// performs candidate/Main acceptance; this neutral type preserves all four
// verdict values without implying receipt or completion.
type RepairResultProfessionalAnswer struct {
	Verdict               string                        `json:"verdict"`
	BeforeFixResults      []RepairResultBeforeFixResult `json:"before_fix_results"`
	UnitResults           []RepairResultUnitResult      `json:"unit_results"`
	DetectionGapArtifacts []string                      `json:"detection_gap_artifacts"`
	EvidenceRefs          []string                      `json:"evidence_refs"`
	ScopeDeviations       []string                      `json:"scope_deviations"`
	ResidualRisks         []string                      `json:"residual_risks"`
	MigrationRef          string                        `json:"migration_ref"`
	RollbackRef           string                        `json:"rollback_ref"`
}

type RepairResultDraft struct {
	Title       string              `json:"title"`
	Instruction string              `json:"instruction"`
	Read        []ReadItem          `json:"read"`
	Questions   map[string]Question `json:"questions"`
	Answer      json.RawMessage     `json:"answer"`
}

func (draft *RepairResultDraft) UnmarshalJSON(data []byte) error {
	if err := ValidateUniqueJSON(data); err != nil {
		return err
	}
	type wire RepairResultDraft
	var decoded wire
	if err := decodeOneStrictJSON(data, &decoded); err != nil {
		return err
	}
	answer, err := canonicalJSON(decoded.Answer)
	if err != nil {
		return fmt.Errorf("S9 RepairResult draft answer: %w", err)
	}
	decoded.Answer = answer
	*draft = RepairResultDraft(decoded)
	return nil
}

// RepairResultPacket is the immutable neutral S9 result questionnaire. The
// stable ResultID and original issuance event/path identities are derived
// from the packet's full source binding.
type RepairResultPacket struct {
	ID              string              `json:"packet_id"`
	TemplateID      string              `json:"template_id"`
	TemplateSHA     string              `json:"template_sha256"`
	Binding         RepairResultBinding `json:"binding"`
	Inputs          RepairResultInputs  `json:"inputs"`
	ResultID        string              `json:"result_id"`
	IssuanceEventID string              `json:"issuance_event_id"`
	ViewPath        string              `json:"view_path"`
	AnswerPath      string              `json:"answer_path"`
	Draft           RepairResultDraft   `json:"draft"`
}

func (packet *RepairResultPacket) UnmarshalJSON(data []byte) error {
	if err := ValidateUniqueJSON(data); err != nil {
		return err
	}
	type wire RepairResultPacket
	var decoded wire
	if err := decodeOneStrictJSON(data, &decoded); err != nil {
		return err
	}
	*packet = RepairResultPacket(decoded)
	return nil
}

// BuildRepairResult constructs a deterministic S9 authoring packet from the
// current, trusted PlanReport/Execution projection. It does not verify live
// source authority; the original S9 issuer and consumer do that.
func BuildRepairResult(binding RepairResultBinding, inputs RepairResultInputs) (RepairResultPacket, error) {
	if err := validateRepairResultBinding(binding); err != nil {
		return RepairResultPacket{}, err
	}
	canonicalInputs, err := canonicalRepairResultInputs(binding, inputs)
	if err != nil {
		return RepairResultPacket{}, err
	}
	templateBytes, err := schema.ReadAsset("stage-form-s9-result-template.json")
	if err != nil {
		return RepairResultPacket{}, err
	}
	var template struct {
		SchemaVersion string              `json:"schema_version"`
		Stage         string              `json:"stage"`
		Adapter       string              `json:"adapter"`
		Title         string              `json:"title"`
		Instruction   string              `json:"instruction"`
		Questions     map[string]Question `json:"questions"`
	}
	if err := decodeOneStrictJSON(templateBytes, &template); err != nil ||
		template.SchemaVersion != "1.0.0" || template.Stage != "S9" || template.Adapter != "repair_result" ||
		strings.TrimSpace(template.Title) == "" || strings.TrimSpace(template.Instruction) == "" ||
		!repairResultTemplateQuestions(template.Questions) {
		return RepairResultPacket{}, fmt.Errorf("unregistered S9 RepairResult template")
	}
	templateSHA := Digest(templateBytes)
	identityData, err := json.Marshal(struct {
		Binding     RepairResultBinding `json:"binding"`
		Inputs      RepairResultInputs  `json:"inputs"`
		TemplateSHA string              `json:"template_sha256"`
	}{binding, canonicalInputs, templateSHA})
	if err != nil {
		return RepairResultPacket{}, fmt.Errorf("encode S9 RepairResult packet identity: %w", err)
	}
	canonicalIdentity, err := canonicalJSON(identityData)
	if err != nil {
		return RepairResultPacket{}, err
	}
	packetHash := Digest(canonicalIdentity)
	packetID := "packet-" + packetHash
	resultID := "repair-result-" + packetHash
	if !repairResultPacketIDRE.MatchString(packetID) || !repairResultIDRE.MatchString(resultID) {
		return RepairResultPacket{}, fmt.Errorf("derived S9 RepairResult packet/result identity is invalid")
	}
	base := fmt.Sprintf("%s/g%d/S9-result/r%d/%s/%s", binding.Sources.RuntimeID, binding.Sources.BaselineGeneration, binding.Sources.ReviewRound, binding.Sources.RepairAssignmentID, packetID)
	answer, err := canonicalJSON([]byte(`{"verdict":null,"before_fix_results":null,"unit_results":null,"detection_gap_artifacts":null,"evidence_refs":null,"scope_deviations":null,"residual_risks":null,"migration_ref":null,"rollback_ref":null}`))
	if err != nil {
		return RepairResultPacket{}, err
	}
	read := makeRepairResultRead(binding, canonicalInputs)
	packet := RepairResultPacket{
		ID: packetID, TemplateID: repairResultTemplateID, TemplateSHA: templateSHA,
		Binding: binding, Inputs: canonicalInputs,
		ResultID:        resultID,
		IssuanceEventID: "evt-repair-result-packet-" + packetID,
		ViewPath:        ".claude/work-packets/" + base + ".json",
		AnswerPath:      ".claude/submissions/" + base + "/" + binding.Sources.OwnerAgentID + "/answer.json",
		Draft: RepairResultDraft{
			Title: template.Title, Instruction: template.Instruction,
			Read: read, Questions: template.Questions, Answer: answer,
		},
	}
	packetBytes, err := json.Marshal(packet)
	if err != nil {
		return RepairResultPacket{}, err
	}
	if len(packetBytes) > repairResultPacketLimit {
		return RepairResultPacket{}, fmt.Errorf("S9 RepairResult packet exceeds 20 MiB")
	}
	if err := schema.NewEmbeddedValidator().ValidateBytes("repair-result-work-packet.schema.json", packetBytes); err != nil {
		return RepairResultPacket{}, fmt.Errorf("generated S9 RepairResult packet violates its registered schema: %w", err)
	}
	return packet, nil
}

// RebuildRepairResult recomputes every machine-controlled field from the
// embedded template and frozen source documents.
func RebuildRepairResult(packet RepairResultPacket) (RepairResultPacket, error) {
	return BuildRepairResult(packet.Binding, packet.Inputs)
}

// RepairResultPacketDigest returns the canonical digest of a complete packet
// only when it exactly matches its deterministic rebuild.
func RepairResultPacketDigest(packet RepairResultPacket) (string, error) {
	rebuilt, err := RebuildRepairResult(packet)
	if err != nil {
		return "", err
	}
	if !reflect.DeepEqual(packet, rebuilt) {
		return "", fmt.Errorf("S9 RepairResult packet differs from its canonical rebuild")
	}
	return RepairResultPacketContentDigest(rebuilt)
}

// RepairResultPacketContentDigest hashes the canonical serialized packet without
// rebuilding its inputs. Callers must already have verified canonical packet
// identity through RepairResultPackets or RebuildRepairResult.
func RepairResultPacketContentDigest(packet RepairResultPacket) (string, error) {
	data, err := json.Marshal(packet)
	if err != nil {
		return "", err
	}
	canonical, err := canonicalJSON(data)
	if err != nil {
		return "", err
	}
	return Digest(canonical), nil
}

// ValidateRepairResultAnswer validates the full authoring file, preserves the
// immutable packet envelope, and returns only the canonical nine-field
// professional answer. Original S9 consumers verify authority, sources,
// execution, candidate, checks, and Result acceptance.
func ValidateRepairResultAnswer(packet RepairResultPacket, data []byte) (json.RawMessage, []FieldError) {
	issue := func(pointer, expected, observed string) (json.RawMessage, []FieldError) {
		return nil, []FieldError{{Pointer: pointer, Expected: expected, Observed: observed, Repair: "保留已发行题干，只填写 answer 中的专业观测、实施结论与资料引用。"}}
	}
	if len(data) > repairResultPacketLimit {
		return issue("/", "JSON <= 20 MiB", "oversized")
	}
	if err := ValidateUniqueJSON(data); err != nil {
		return issue("/", "unique JSON keys", err.Error())
	}
	if err := schema.NewEmbeddedValidator().ValidateBytes("stage-answer-s9-result.schema.json", data); err != nil {
		return issue("/", "registered closed S9 RepairResult answer", err.Error())
	}
	var draft RepairResultDraft
	if err := decodeOneStrictJSON(data, &draft); err != nil {
		return issue("/", "one closed S9 RepairResult questionnaire", err.Error())
	}
	expected, err := RebuildRepairResult(packet)
	if err != nil || !reflect.DeepEqual(packet, expected) {
		return issue("/", "canonical issued S9 RepairResult packet", "packet identity differs")
	}
	answer := draft.Answer
	draft.Answer = expected.Draft.Answer
	if !reflect.DeepEqual(draft, expected.Draft) {
		return issue("/", "unchanged issued machine fields", "questionnaire or binding was edited")
	}
	canonicalAnswer, err := canonicalJSON(answer)
	if err != nil {
		return issue("/answer", "one JSON professional answer object", err.Error())
	}
	if pointer, want, observed := validateCompleteRepairResultAnswer(packet.Inputs.Sources, canonicalAnswer); pointer != "" {
		return issue(pointer, want, observed)
	}
	return json.RawMessage(canonicalAnswer), nil
}

func validateCompleteRepairResultAnswer(sources RepairPlanReportInputs, data []byte) (pointer, expected, observed string) {
	var answer RepairResultProfessionalAnswer
	if err := decodeOneStrictJSON(data, &answer); err != nil {
		return "/answer", "valid S9 professional fields", err.Error()
	}
	var answerFields map[string]json.RawMessage
	if err := json.Unmarshal(data, &answerFields); err != nil {
		return "/answer", "valid S9 professional object", err.Error()
	}
	for _, name := range []string{"verdict", "before_fix_results", "unit_results", "detection_gap_artifacts", "evidence_refs", "scope_deviations", "residual_risks", "migration_ref", "rollback_ref"} {
		if bytes.Equal(bytes.TrimSpace(answerFields[name]), []byte("null")) {
			return "/answer/" + name, "explicit completed answer value, not null", "null"
		}
	}
	if answer.Verdict != "completed" && answer.Verdict != "blocked" && answer.Verdict != "contract_conflict" && answer.Verdict != "stale" {
		return "/answer/verdict", "completed | blocked | contract_conflict | stale", answer.Verdict
	}
	if answer.BeforeFixResults == nil || answer.UnitResults == nil || answer.DetectionGapArtifacts == nil || answer.EvidenceRefs == nil || answer.ScopeDeviations == nil || answer.ResidualRisks == nil {
		return "/answer", "all professional arrays explicitly supplied", "one or more arrays are null"
	}
	assignment, err := repairPlanReportObject(sources.RepairAssignment, "RepairAssignment")
	if err != nil {
		return "/inputs/sources/repair_assignment", "canonical exact RepairAssignment", err.Error()
	}
	assertions, ok := repairPlanReportStringSlice(assignment["assertion_ids"])
	if !ok || len(assertions) == 0 {
		return "/inputs/sources/repair_assignment/assertion_ids", "explicit exact assertion slots", "missing or empty"
	}
	if containsString(assertions, "all") {
		return "/inputs/sources/repair_assignment/assertion_ids", "concrete assertion IDs", "wildcard all is not an answer slot"
	}
	if len(answer.BeforeFixResults) != len(assertions) {
		return "/answer/before_fix_results", "one result for every exact Assignment assertion", fmt.Sprintf("got %d want %d", len(answer.BeforeFixResults), len(assertions))
	}
	for i, result := range answer.BeforeFixResults {
		prefix := fmt.Sprintf("/answer/before_fix_results/%d", i)
		if result.AssertionID != assertions[i] {
			return prefix + "/assertion_id", "exact frozen assertion_id in source order", result.AssertionID
		}
		if strings.TrimSpace(result.ObservedRed) == "" {
			return prefix + "/observed_red", "non-empty professional observation of the red condition", "empty"
		}
		if err := validateRepairResultRefs(result.EvidenceRefs, prefix+"/evidence_refs", true); err != nil {
			return err.pointer, "unique non-empty evidence references", err.observed
		}
	}
	unitIDs, ok := repairPlanReportStringSlice(assignment["unit_ids"])
	if !ok || len(unitIDs) == 0 {
		return "/inputs/sources/repair_assignment/unit_ids", "explicit exact assigned unit slots", "missing or empty"
	}
	if len(answer.UnitResults) != len(unitIDs) {
		return "/answer/unit_results", "one result for every exact assigned unit", fmt.Sprintf("got %d want %d", len(answer.UnitResults), len(unitIDs))
	}
	for i, result := range answer.UnitResults {
		prefix := fmt.Sprintf("/answer/unit_results/%d", i)
		if result.UnitID != unitIDs[i] {
			return prefix + "/unit_id", "exact frozen unit_id in source order", result.UnitID
		}
		if result.Status != "pass" && result.Status != "fail" && result.Status != "blocked" {
			return prefix + "/status", "pass | fail | blocked", result.Status
		}
		if strings.TrimSpace(result.ImplementationSummary) == "" {
			return prefix + "/implementation_summary", "non-empty professional implementation summary", "empty"
		}
		if err := validateRepairResultRefs(result.EvidenceRefs, prefix+"/evidence_refs", true); err != nil {
			return err.pointer, "unique non-empty evidence references", err.observed
		}
		if err := validateRepairResultRefs(result.RestoredInvariantRefs, prefix+"/restored_invariant_refs", false); err != nil {
			return err.pointer, "unique non-empty invariant references", err.observed
		}
	}
	for _, field := range []struct {
		name string
		refs []string
	}{
		{"detection_gap_artifacts", answer.DetectionGapArtifacts},
		{"evidence_refs", answer.EvidenceRefs},
		{"scope_deviations", answer.ScopeDeviations},
		{"residual_risks", answer.ResidualRisks},
	} {
		if err := validateRepairResultRefs(field.refs, "/answer/"+field.name, false); err != nil {
			return err.pointer, "unique non-empty strings", err.observed
		}
	}
	return "", "", ""
}

type repairResultValidationError struct {
	pointer  string
	observed string
}

func (e *repairResultValidationError) Error() string { return e.pointer + ": " + e.observed }

func validateRepairResultRefs(values []string, pointer string, requireOne bool) *repairResultValidationError {
	if requireOne && len(values) == 0 {
		return &repairResultValidationError{pointer: pointer, observed: "empty"}
	}
	seen := map[string]bool{}
	for _, value := range values {
		if strings.TrimSpace(value) == "" || value != strings.TrimSpace(value) || seen[value] {
			return &repairResultValidationError{pointer: pointer, observed: value}
		}
		seen[value] = true
	}
	return nil
}

func canonicalRepairResultInputs(binding RepairResultBinding, inputs RepairResultInputs) (RepairResultInputs, error) {
	canonicalSources, err := canonicalRepairPlanReportInputs(binding.Sources, inputs.Sources)
	if err != nil {
		return RepairResultInputs{}, fmt.Errorf("S9 RepairResult source closure: %w", err)
	}
	canonical := RepairResultInputs{Sources: canonicalSources}
	canonical.PlanReport, err = canonicalRepairResultObject(inputs.PlanReport, "PlanReport")
	if err != nil {
		return RepairResultInputs{}, err
	}
	canonical.Execution, err = canonicalRepairResultObject(inputs.Execution, "Execution")
	if err != nil {
		return RepairResultInputs{}, err
	}
	canonical.ReworkReceipt, err = canonicalRepairResultNullableObject(inputs.ReworkReceipt, "ReworkReceipt")
	if err != nil {
		return RepairResultInputs{}, err
	}
	if len(canonical.PlanReport)+len(canonical.Execution)+len(canonical.ReworkReceipt) > repairResultPacketLimit {
		return RepairResultInputs{}, fmt.Errorf("S9 RepairResult frozen input closure exceeds 20 MiB")
	}
	if !digestPattern.MatchString(binding.ExecutionSHA256) || Digest(canonical.Execution) != binding.ExecutionSHA256 {
		return RepairResultInputs{}, fmt.Errorf("S9 registered Execution definition digest differs from binding")
	}
	if err := validateRepairResultExecution(binding, canonical.Execution); err != nil {
		return RepairResultInputs{}, err
	}
	if err := validateRepairResultPlanReport(binding, canonicalSources, canonical.PlanReport); err != nil {
		return RepairResultInputs{}, err
	}
	if binding.ReworkRef == "" || binding.ReworkSHA256 == "" {
		if binding.ReworkRef != "" || binding.ReworkSHA256 != "" || string(canonical.ReworkReceipt) != "null" {
			return RepairResultInputs{}, fmt.Errorf("S9 first-attempt Rework tuple must have empty ref/SHA and explicit null receipt")
		}
	} else {
		if !safeRepairPlanReportRef(binding.ReworkRef) || !digestPattern.MatchString(binding.ReworkSHA256) || string(canonical.ReworkReceipt) == "null" {
			return RepairResultInputs{}, fmt.Errorf("S9 rework attempt requires one safe pinned Rework receipt")
		}
	}
	encoded, err := json.Marshal(canonical)
	if err != nil {
		return RepairResultInputs{}, err
	}
	if len(encoded) > repairResultPacketLimit {
		return RepairResultInputs{}, fmt.Errorf("S9 RepairResult frozen inputs exceed 20 MiB")
	}
	return canonical, nil
}

func validateRepairResultBinding(binding RepairResultBinding) error {
	if err := validateRepairPlanReportBinding(binding.Sources); err != nil {
		return fmt.Errorf("S9 RepairResult original source binding: %w", err)
	}
	if !safeRepairPlanReportRef(binding.PlanReportRef) || !digestPattern.MatchString(binding.PlanReportSHA256) || !digestPattern.MatchString(binding.ExecutionSHA256) {
		return fmt.Errorf("S9 RepairResult requires exact PlanReport and Execution source digests")
	}
	if (binding.ReworkRef == "") != (binding.ReworkSHA256 == "") {
		return fmt.Errorf("S9 ReworkRef and ReworkSHA256 must be both empty or both present")
	}
	if binding.ReworkRef != "" && (!safeRepairPlanReportRef(binding.ReworkRef) || !digestPattern.MatchString(binding.ReworkSHA256)) {
		return fmt.Errorf("S9 rework attempt source reference or digest is invalid")
	}
	if err := ValidateRepairResultResume(binding.Sources.OwnerAgentID, binding.Resume); err != nil {
		return err
	}
	return nil
}

// ValidateRepairResultResume checks the closed tuple's value-level shape. Its
// authority is established only by the Runtime's original packet/evidence/
// journal reader.
func ValidateRepairResultResume(owner string, resume *RepairResultResume) error {
	if resume == nil {
		return nil
	}
	if owner == "" || !repairResultPacketIDRE.MatchString(resume.PredecessorPacketID) ||
		!digestPattern.MatchString(resume.PredecessorPacketSHA256) || !safeRepairPlanReportRef(resume.BlockerRef) ||
		!digestPattern.MatchString(resume.BlockerSHA256) || !safeRepairPlanReportRef(resume.ResolutionMessageRef) ||
		!digestPattern.MatchString(resume.ResolutionMessageSHA256) || resume.ResolutionEventID != fmt.Sprintf("evt-agent-%s-blocker_resolved-r%d", owner, resume.ResolutionAfterRevision) ||
		resume.ResolutionIdempotencyKey != fmt.Sprintf("runtime:agent:%s:blocker_resolved:%d", owner, resume.ResolutionBeforeRevision) ||
		resume.ResolutionBeforeRevision < 0 || resume.ResolutionAfterRevision != resume.ResolutionBeforeRevision+1 || resume.ResolutionSequence < 1 {
		return fmt.Errorf("S9 Resume tuple has invalid packet, artifact, event or acceptance coordinates")
	}
	return nil
}

// RepairResultResumeSHA256 returns the canonical digest used by candidate and
// final provenance. An absent Resume is represented by an empty digest.
func RepairResultResumeSHA256(resume *RepairResultResume) (string, error) {
	if resume == nil {
		return "", nil
	}
	data, err := json.Marshal(resume)
	if err != nil {
		return "", err
	}
	canonical, err := canonicalJSON(data)
	if err != nil {
		return "", err
	}
	return Digest(canonical), nil
}

func validateRepairResultPlanReport(binding RepairResultBinding, sources RepairPlanReportInputs, raw json.RawMessage) error {
	if err := schema.NewEmbeddedValidator().ValidateBytes("repair-plan-report.schema.json", raw); err != nil {
		return fmt.Errorf("S9 exact PlanReport source schema: %w", err)
	}
	report, err := repairPlanReportObject(raw, "accepted PlanReport")
	if err != nil {
		return err
	}
	if report["record_type"] != "repair_plan_report" || report["session_id"] != binding.Sources.SessionID ||
		report["plan_id"] != binding.Sources.PlanID || report["assignment_id"] != binding.Sources.RepairAssignmentID ||
		report["agent_id"] != binding.Sources.OwnerAgentID {
		return fmt.Errorf("S9 PlanReport identity differs from the original source/owner binding")
	}
	assignment, err := repairPlanReportObject(sources.RepairAssignment, "RepairAssignment")
	if err != nil {
		return fmt.Errorf("S9 exact RepairAssignment source: %w", err)
	}
	assertionIDs, ok := repairPlanReportStringSlice(assignment["assertion_ids"])
	if !ok || len(assertionIDs) == 0 || !repairPlanReportJSONEqual(report["assertion_ids"], assertionIDs) {
		return fmt.Errorf("S9 PlanReport assertion set differs from its exact RepairAssignment")
	}
	if report["receipt"] == nil {
		return fmt.Errorf("S9 RepairResult requires an originally accepted PlanReport receipt")
	}
	receiptData, err := json.Marshal(report["receipt"])
	if err != nil {
		return err
	}
	canonicalReceipt, err := canonicalJSON(receiptData)
	if err != nil {
		return err
	}
	var receipt map[string]any
	decoder := json.NewDecoder(bytes.NewReader(canonicalReceipt))
	decoder.UseNumber()
	if err := decoder.Decode(&receipt); err != nil || receipt == nil {
		return fmt.Errorf("S9 PlanReport receipt must be an object")
	}
	expected := map[string]any{
		"contract": "s9-plan-report-receipt-v1", "owner_agent_id": binding.Sources.OwnerAgentID,
		"runtime_id": binding.Sources.RuntimeID, "baseline_generation": binding.Sources.BaselineGeneration,
		"review_round": binding.Sources.ReviewRound, "case_id": binding.Sources.CaseID,
		"case_ref": binding.Sources.CaseRef, "case_sha256": binding.Sources.CaseSHA256,
		"contract_id": binding.Sources.ContractID, "contract_ref": binding.Sources.ContractRef,
		"contract_sha256": binding.Sources.ContractSHA256, "session_id": binding.Sources.SessionID,
		"session_ref": binding.Sources.SessionRef, "session_sha256": binding.Sources.SessionSHA256,
		"plan_id": binding.Sources.PlanID, "plan_ref": binding.Sources.PlanRef,
		"plan_sha256": binding.Sources.PlanSHA256, "repair_assignment_id": binding.Sources.RepairAssignmentID,
		"repair_assignment_sha256":   binding.Sources.RepairAssignmentSHA256,
		"manifest_assignment_id":     binding.Sources.ManifestAssignmentID,
		"manifest_assignment_sha256": binding.Sources.ManifestAssignmentSHA256,
		"team_id":                    binding.Sources.TeamID, "task_id": binding.Sources.TaskID,
		"task_ref": binding.Sources.TaskRef, "task_sha256": binding.Sources.TaskSHA256,
		"manifest_ref": binding.Sources.ManifestRef, "manifest_sha256": binding.Sources.ManifestSHA256,
		"report_id": report["report_id"], "report_ref": binding.PlanReportRef,
		"work_complete": false,
	}
	for key, want := range expected {
		if !repairPlanReportJSONEqual(receipt[key], want) {
			return fmt.Errorf("S9 accepted PlanReport receipt %s differs from the frozen binding", key)
		}
	}
	return nil
}

func validateRepairResultExecution(binding RepairResultBinding, raw json.RawMessage) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
		return fmt.Errorf("S9 Execution definition must be one JSON object")
	}
	allowed := map[string]bool{
		"assignment_record": true, "assignment_record_sha256": true, "worker_checks": true, "check_profiles": true,
		"purpose": true, "work_packet_id": true, "stage_entry": true, "check_location": true,
		"platform_session_id": true, "assignment_id": true, "runtime_id": true,
		"baseline_generation": true, "execution_generation": true, "agent_id": true,
		"worktree_path": true, "branch": true, "target_branch": true, "base_commit": true,
		"write_paths": true, "required_checks": true, "inputs": true, "bootstrap_sha256": true,
	}
	for name := range fields {
		if !allowed[name] {
			return fmt.Errorf("S9 Execution definition contains unapproved field %q", name)
		}
	}
	// workspace.Execution deliberately omits several S6-only or not-yet-
	// available fields for an ordinary S9 dispatch. Their absence means that
	// no such S6 source was registered; the neutral packet must preserve that
	// absence instead of inventing a Record, WorkerChecks, or session tuple.
	// The core execution identity and Main/Worker check definition remain
	// required by the closed schema below.
	for _, name := range []string{
		"check_profiles", "check_location", "assignment_id", "runtime_id",
		"baseline_generation", "execution_generation", "agent_id", "worktree_path",
		"branch", "target_branch", "base_commit", "write_paths", "required_checks",
		"inputs", "bootstrap_sha256",
	} {
		if _, ok := fields[name]; !ok {
			return fmt.Errorf("S9 Execution definition omits required field %q", name)
		}
	}
	_, hasRecord := fields["assignment_record"]
	_, hasRecordSHA := fields["assignment_record_sha256"]
	if hasRecord != hasRecordSHA {
		return fmt.Errorf("S9 Execution Assignment Record reference and digest must be present together")
	}
	if hasRecord {
		recordNull := bytes.Equal(bytes.TrimSpace(fields["assignment_record"]), []byte("null"))
		digestNull := bytes.Equal(bytes.TrimSpace(fields["assignment_record_sha256"]), []byte("null"))
		if recordNull != digestNull {
			return fmt.Errorf("S9 Execution Assignment Record reference and digest nullability differs")
		}
	}
	var execution struct {
		AssignmentRecord       json.RawMessage `json:"assignment_record"`
		AssignmentRecordSHA256 string          `json:"assignment_record_sha256"`
		RuntimeID              string          `json:"runtime_id"`
		BaselineGeneration     int             `json:"baseline_generation"`
		AssignmentID           string          `json:"assignment_id"`
		AgentID                string          `json:"agent_id"`
	}
	if err := json.Unmarshal(raw, &execution); err != nil {
		return fmt.Errorf("decode S9 Execution definition: %w", err)
	}
	if execution.RuntimeID != binding.Sources.RuntimeID || execution.BaselineGeneration != binding.Sources.BaselineGeneration ||
		execution.AssignmentID != binding.Sources.ManifestAssignmentID || execution.AgentID != binding.Sources.OwnerAgentID {
		return fmt.Errorf("S9 Execution owner/assignment/runtime/generation differs from the frozen source binding")
	}
	if len(execution.AssignmentRecord) != 0 && !bytes.Equal(bytes.TrimSpace(execution.AssignmentRecord), []byte("null")) {
		if !digestPattern.MatchString(execution.AssignmentRecordSHA256) {
			return fmt.Errorf("S9 Execution must bind the canonical Assignment Record bytes")
		}
		var recordRef struct {
			RuntimeID          string `json:"runtime_id"`
			BaselineGeneration int    `json:"baseline_generation"`
			AssignmentID       string `json:"assignment_id"`
		}
		if err := decodeOneStrictJSON(execution.AssignmentRecord, &recordRef); err != nil || recordRef.RuntimeID != execution.RuntimeID || recordRef.BaselineGeneration != execution.BaselineGeneration || recordRef.AssignmentID != execution.AssignmentID {
			return fmt.Errorf("S9 Execution Assignment Record reference differs from the execution tuple")
		}
	}
	return nil
}

func canonicalRepairResultObject(raw json.RawMessage, name string) (json.RawMessage, error) {
	if len(raw) == 0 || len(raw) > repairResultSourceLimit {
		return nil, fmt.Errorf("S9 %s source must be non-empty and no larger than 4 MiB", name)
	}
	if err := ValidateUniqueJSON(raw); err != nil {
		return nil, fmt.Errorf("S9 %s source: %w", name, err)
	}
	canonical, err := canonicalJSON(raw)
	if err != nil {
		return nil, fmt.Errorf("S9 %s source: %w", name, err)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &object); err != nil || object == nil {
		return nil, fmt.Errorf("S9 %s source must be a JSON object", name)
	}
	return json.RawMessage(canonical), nil
}

func canonicalRepairResultNullableObject(raw json.RawMessage, name string) (json.RawMessage, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("S9 %s source must be explicit (object or null)", name)
	}
	if err := ValidateUniqueJSON(raw); err != nil {
		return nil, fmt.Errorf("S9 %s source: %w", name, err)
	}
	canonical, err := canonicalJSON(raw)
	if err != nil {
		return nil, fmt.Errorf("S9 %s source: %w", name, err)
	}
	if bytes.Equal(canonical, []byte("null")) {
		return json.RawMessage("null"), nil
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &object); err != nil || object == nil {
		return nil, fmt.Errorf("S9 %s source must be an object or null", name)
	}
	return json.RawMessage(canonical), nil
}

func repairResultTemplateQuestions(questions map[string]Question) bool {
	fields := []string{"verdict", "before_fix_results", "unit_results", "detection_gap_artifacts", "evidence_refs", "scope_deviations", "residual_risks", "migration_ref", "rollback_ref"}
	if len(questions) != len(fields) {
		return false
	}
	for _, field := range fields {
		question, ok := questions[field]
		if !ok || strings.TrimSpace(question.Question) == "" {
			return false
		}
	}
	return true
}

func makeRepairResultRead(binding RepairResultBinding, inputs RepairResultInputs) []ReadItem {
	read := makeRepairPlanReportRead(binding.Sources)
	read = append(read, ReadItem{Path: binding.PlanReportRef, Purpose: "Read the exact PlanReport accepted by the original S9 writer."})
	var execution struct {
		Path string `json:"worktree_path"`
	}
	_ = decodeOneStrictJSON(inputs.Execution, &execution)
	if execution.Path != "" {
		read = append(read, ReadItem{Path: execution.Path, Purpose: "Work only in the originally registered Worker execution; current HEAD and changes are checked at candidate receipt."})
	}
	if binding.ReworkRef != "" {
		read = append(read, ReadItem{Path: binding.ReworkRef, Purpose: "Read the exact accepted integration rework receipt for this attempt."})
	}
	return read
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
