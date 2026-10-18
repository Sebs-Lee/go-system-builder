package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/entroforge/go-system-builder/internal/investigation"
	"github.com/entroforge/go-system-builder/internal/policy"
	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/workpacket"
)

// These are authoring projections of IngestRequest and HypothesisRequest.
// Case/Assignment versions and dispatch metadata come from their original
// consumers. This file owns no Runtime state, issuance, or acceptance ledger.
type investigationLeadIntakeAnswer struct {
	GroupingRationale string `json:"grouping_rationale"`
}

type investigationLeadQuestion struct {
	HypothesisID     string            `json:"hypothesis_id,omitempty"`
	AgentID          string            `json:"agent_id,omitempty"`
	Statement        string            `json:"statement"`
	Invariant        string            `json:"invariant"`
	Discriminator    string            `json:"discriminator"`
	ExpectedOutcomes map[string]string `json:"expected_outcomes"`
	SourceFindingIDs []string          `json:"source_finding_ids"`
	EvidenceRefs     []string          `json:"evidence_refs"`
}

func stageInvestigationLeadQuestionnaireContext(root string, request policy.Input, snapshot runtime.Snapshot) string {
	if context := stageInvestigationLeadContractContext(root, request, snapshot); context != "" {
		return context
	}
	if context := stageInvestigationLeadDispositionContext(root, request, snapshot); context != "" {
		return context + " Optional registered evidence examples (the route keeps its original reference semantics): " + investigationLeadEvidenceIndexContext(root, snapshot.State)
	}
	if !mainQuestionnaireViewer(request) || !investigationLeadCurrent(snapshot.State) {
		return ""
	}
	paths := investigationLeadQuestionnairePaths(root, snapshot.State)
	if len(paths) == 0 {
		return "S8 Main questionnaire cannot resolve the current sealed ObservationBatch source. Main must restore the original S7 handoff reference; the listener did not guess an answer path."
	}
	batch, document, batchRef, err := investigationLeadSource(root, snapshot.State)
	if err != nil {
		return "S8 Main questionnaire: " + paths[0] + ". Original source could not be read: " + err.Error()
	}
	path := paths[0]
	var draft any = investigationLeadIntakeAnswer{}
	questions := "Read the sealed Batch and immutable Findings, including encounter boundaries and capture gaps. Fill grouping_rationale: why this exact Finding set is a provisional causal investigation group, then save this file. Case identity, source set and versions are supplied by the original intake consumer."
	if document != nil {
		rows := []investigationLeadQuestion{}
		for _, h := range objectSliceAny(document["hypotheses"]) {
			q := investigationLeadQuestionFromCase(h)
			q.AgentID = investigationLeadRegisteredOwner(snapshot.State, stringValue(h["assignment_id"]))
			rows = append(rows, q)
		}
		if len(rows) == 0 {
			rows = append(rows, investigationLeadQuestion{ExpectedOutcomes: map[string]string{"support": "", "refute": ""}, SourceFindingIDs: []string{}, EvidenceRefs: []string{}})
		}
		draft = rows
		questions = "Each array row is one falsifiable causal question: fill statement, invariant, discriminator, expected_outcomes as an object with support/refute strings, source_finding_ids as Finding IDs, and evidence_refs as the registered evidence IDs or real execution anchors described below. Both reference fields are string arrays. Explain the read-only investigation bounds and when the discriminator must stop or return inconclusive. Consider credible competing mechanisms; do not manufacture an alternative merely to fill rows. Select the professional division of work; agent_id is optional when no owner has been selected. Existing hypothesis_id values are supplied from the Case; new rows need no ID, and rows may be reordered. Save this file to register the questions and their original Investigator dispatches. Main must then use the platform to actually start/resume the named investigators; registration does not prove a platform launch."
	}
	if repairLeadPreparationEvent(request.Event) {
		var err error
		if rows, questions := draft.([]investigationLeadQuestion); questions {
			err = prepareInvestigationLeadQuestions(path, rows)
		} else {
			err = writeRepairLeadDraft(path, draft)
		}
		if err != nil {
			return fmt.Sprintf("S8 Main questionnaire could not be prepared at %s: %v", path, err)
		}
	}
	if !repairLeadPreparationEvent(request.Event) && questionnaireMissing(path) {
		return "S8 Main questionnaire: " + path + ". Not prepared yet; the next Main SessionStart or PreToolUse will create the draft. This event is read-only. " + questions
	}
	var sources strings.Builder
	fmt.Fprintf(&sources, "Batch: %s. ", resolveRootPath(root, batchRef))
	ids := stringSliceAny(batch["finding_ids"])
	for _, row := range objectSliceAny(mapFieldCLI(snapshot.State, "entities")["findings"]) {
		for _, id := range ids {
			if stringValue(row["finding_id"]) == id {
				fmt.Fprintf(&sources, "Finding %s: %s. ", id, resolveRootPath(root, stringValue(row["path"])))
			}
		}
	}
	if document != nil {
		fmt.Fprintf(&sources, "Current Case: %s. %s ", resolveRootPath(root, stringValue(investigationPointer(snapshot.State)["path"])), investigationLeadEvidenceReferenceContext(root, snapshot.State))
	}
	routeHint := ""
	if routePaths := investigationLeadDispositionPaths(root, snapshot.State); len(routePaths) != 0 {
		routeHint = " Current professional route entry: " + routePaths[0] + ". Main may choose an earlier specification/REQ/duplicate/investigate_more disposition there; the original route consumer checks its actual conditions."
	}
	return "S8 Main questionnaire: " + path + ". " + sources.String() + questions + routeHint
}

// Display the original evidence index as authoring references. This performs
// no attestation, registration or new qualification check; the original save
// consumer still checks the referenced evidence and physical source bytes.
func investigationLeadEvidenceReferenceContext(root string, state map[string]any) string {
	return "For Hypothesis.evidence_refs only: values are registered evidence IDs or real existing execution anchors (test://<actual-run-or-check> only for an execution that actually exists). Finding/evidence file paths are reading sources, not evidence_refs values. Do not invent an execution anchor to fill a question. The original Hypothesis consumer verifies applicability and source SHA on save. " + investigationLeadEvidenceIndexContext(root, state)
}

func investigationLeadEvidenceIndexContext(root string, state map[string]any) string {
	var out strings.Builder
	generation := tolerantInt(mapFieldCLI(state, "baseline")["generation"])
	round := tolerantInt(mapFieldCLI(state, "review")["round"])
	var ids []string
	for _, row := range objectSliceAny(state["evidence"]) {
		id, path := stringValue(row["id"]), stringValue(row["path"])
		if id == "" || path == "" || strings.TrimSpace(stringValue(row["status"])) != "valid" || tolerantInt(row["baseline_generation"]) != generation || (round > 0 && tolerantInt(row["review_round"]) != round) {
			continue
		}
		if invalidated := row["invalidated_by"]; invalidated != nil && invalidated != "" {
			continue
		}
		if len(ids) == 0 {
			fmt.Fprintf(&out, "Recorded valid evidence for current generation %d / round %d (existing Runtime metadata for source lookup): ", generation, round)
		}
		fmt.Fprintf(&out, "%s -> %s; ", id, resolveRootPath(root, path))
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		out.WriteString("No valid current-generation/current-round evidence IDs are recorded. Main/system owns any missing registered source evidence. The original field-specific consumer retains its reference requirements.")
	} else {
		example, _ := json.Marshal([]string{ids[0]})
		fmt.Fprintf(&out, "Available registered-ID array example: evidence_refs=%s; read the corresponding path when assessing that evidence.", example)
	}
	return out.String()
}

// Append questions that already entered the original Case (for example a
// Worker's new_hypotheses). Existing author rows and edits stay intact.
func prepareInvestigationLeadQuestions(path string, originals []investigationLeadQuestion) error {
	data, err := workpacket.ReadFile(filepath.Dir(path), filepath.Base(path), 20<<20)
	if os.IsNotExist(err) {
		return writeRepairLeadDraft(path, originals)
	}
	if err != nil {
		return err
	}
	var rows []investigationLeadQuestion
	if workpacket.ValidateUniqueJSON(data) != nil || investigationLeadDecode(data, &rows) != nil {
		return nil // Preserve an in-progress or invalid author draft for feedback on save.
	}
	changed := false
	for _, original := range originals {
		if original.HypothesisID == "" {
			continue
		}
		found := false
		for _, row := range rows {
			if row.HypothesisID == original.HypothesisID || (row.HypothesisID == "" && investigationLeadSameQuestion(row, original)) {
				found = true
				break
			}
		}
		if !found {
			rows = append(rows, original)
			changed = true
		}
	}
	if changed {
		return persistRepairLeadQuestionnaire(path, rows)
	}
	return nil
}

func investigationLeadCurrent(state map[string]any) bool {
	life := mapFieldCLI(state, "lifecycle")
	pointer := investigationPointer(state)
	return stringValue(life["state"]) == "bug_resolution" && stringValue(life["phase"]) == "investigation" &&
		(pointer == nil || (stringValue(pointer["status"]) == "investigating" && stringValue(pointer["route_consumed_at"]) == "" &&
			(stringValue(pointer["route"]) == "" || stringValue(pointer["route"]) == "investigate_more")))
}

// Select only existing authority references. The original domain consumers
// perform their full baseline, exact-set, evidence and CAS validation on save.
func investigationLeadBatchReference(state map[string]any) (string, string) {
	review := mapFieldCLI(state, "review")
	if batch, _ := review["observation_batch"].(map[string]any); batch != nil {
		return stringValue(batch["path"]), stringValue(batch["sha256"])
	}
	rows := objectSliceAny(state["evidence"])
	for i := len(rows) - 1; i >= 0; i-- {
		if stringValue(rows[i]["kind"]) == "observation_batch" && stringValue(rows[i]["status"]) == "valid" {
			return stringValue(rows[i]["path"]), stringValue(rows[i]["sha256"])
		}
	}
	return "", ""
}

func investigationLeadQuestionnairePaths(root string, state map[string]any) []string {
	if paths := investigationLeadContractPaths(root, state); len(paths) != 0 {
		return paths
	}
	return append(investigationLeadQuestionnairePathsForQuestions(root, state), investigationLeadDispositionPaths(root, state)...)
}

func investigationLeadQuestionnairePathsForQuestions(root string, state map[string]any) []string {
	if !investigationLeadCurrent(state) {
		return nil
	}
	ref, sha := investigationLeadBatchReference(state)
	id := stringValue(state["runtime_id"])
	generation := integerValue(mapFieldCLI(state, "baseline")["generation"])
	round := integerValue(mapFieldCLI(state, "review")["round"])
	if ref == "" || len(sha) != 64 || id == "" || generation < 1 || round < 1 {
		return nil
	}
	name := "intake.json"
	if investigationPointer(state) != nil {
		name = "hypotheses.json"
	}
	path := filepath.Join(root, ".claude", "submissions", "s8", dispatchSlug(id), fmt.Sprintf("g%d", generation), fmt.Sprintf("r%d", round), sha, name)
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil
	}
	return []string{absolute}
}

func investigationLeadReadObject(root, ref, sha string) (map[string]any, error) {
	data, err := workpacket.ReadFile(root, ref, 20<<20)
	if err != nil {
		return nil, err
	}
	if workpacket.Digest(data) != sha {
		return nil, fmt.Errorf("original source %s SHA differs from its Runtime reference", ref)
	}
	if err := workpacket.ValidateUniqueJSON(data); err != nil {
		return nil, err
	}
	var document map[string]any
	err = json.Unmarshal(data, &document)
	return document, err
}

func investigationLeadSource(root string, state map[string]any) (map[string]any, map[string]any, string, error) {
	ref, sha := investigationLeadBatchReference(state)
	batch, err := investigationLeadReadObject(root, ref, sha)
	if err != nil {
		return nil, nil, ref, err
	}
	var document map[string]any
	if pointer := investigationPointer(state); pointer != nil {
		document, err = investigationLeadReadObject(root, stringValue(pointer["path"]), stringValue(pointer["sha256"]))
		if err == nil && (document["case_id"] != pointer["case_id"] || document["observation_batch_id"] != batch["observation_batch_id"] || document["observation_batch_sha256"] != sha) {
			err = fmt.Errorf("current Case and sealed Batch source differ")
		}
	}
	return batch, document, ref, err
}

func runSavedInvestigationLeadQuestionnaire(root string, request policy.Input, stdout, stderr io.Writer) (bool, int) {
	if handled, code := runSavedInvestigationLeadContract(root, request, stdout); handled {
		return true, code
	}
	if handled, code := runSavedInvestigationLeadDisposition(root, request, stdout); handled {
		return true, code
	}
	path, saved := savedQuestionnairePath(request)
	if !saved {
		return false, 0
	}
	statePath, journalPath := filepath.Join(root, ".claude", "loop-state.json"), filepath.Join(root, ".claude", "loop-events.jsonl")
	snapshot, err := runtime.NewStore(statePath, journalPath).Snapshot()
	if err != nil || !investigationLeadCurrent(snapshot.State) {
		return false, 0
	}
	paths := investigationLeadQuestionnairePathsForQuestions(root, snapshot.State)
	if len(paths) == 0 {
		return false, 0
	}
	intakePath := filepath.Join(filepath.Dir(paths[0]), "intake.json")
	if !samePath(path, paths[0]) && !samePath(path, intakePath) {
		return false, 0
	}
	feedback := func(message string) (bool, int) { return true, repairLeadFeedback(stdout, path, message) }
	batch, document, _, err := investigationLeadSource(root, snapshot.State)
	if err != nil {
		return feedback("S8 original source could not be read: " + err.Error())
	}
	data, err := workpacket.ReadFile(root, filepath.ToSlash(strings.TrimPrefix(path, filepath.Clean(root)+string(filepath.Separator))), 20<<20)
	if err != nil {
		return feedback("S8 answer could not be read: " + err.Error())
	}
	if err := workpacket.ValidateUniqueJSON(data); err != nil {
		return feedback("S8 answer is not a valid original request view: " + err.Error())
	}
	if samePath(path, intakePath) {
		var answer investigationLeadIntakeAnswer
		if err := investigationLeadDecode(data, &answer); err != nil {
			return feedback("S8 grouping answer: " + err.Error())
		}
		if strings.TrimSpace(answer.GroupingRationale) == "" {
			return feedback("S8 grouping_rationale is incomplete: explain why the sealed Finding set is a provisional investigation group, then save.")
		}
		if document != nil {
			if stringValue(document["grouping_rationale"]) != strings.TrimSpace(answer.GroupingRationale) {
				return feedback("S8 grouping answer conflicts with the original Case; no accepted content was overwritten.")
			}
			return feedback("S8 grouping was already received in the original Case. Repeated save did not write Runtime. Next Main questionnaire: " + paths[0])
		}
		next, err := investigation.Ingest(root, statePath, journalPath, investigation.IngestRequest{ExpectedRevision: snapshot.Revision, GroupingRationale: answer.GroupingRationale, OccurredAt: time.Now().UTC()})
		if err != nil {
			return feedback("S8 original intake validation: " + err.Error())
		}
		context := stageInvestigationLeadQuestionnaireContext(root, policy.Input{Event: "PreToolUse"}, next)
		return feedback("S8 original intake committed Case " + stringValue(investigationPointer(next.State)["case_id"]) + ". " + context)
	}
	var questions []investigationLeadQuestion
	if err := investigationLeadDecode(data, &questions); err != nil {
		return feedback("S8 hypotheses answer: " + err.Error())
	}
	if len(questions) == 0 {
		return feedback("S8 hypotheses answer has no professional question. Fill a falsifiable statement, invariant, discriminator, distinguishing outcomes and source evidence.")
	}
	// Resolve references before registering anything. IDs persisted into this
	// same DTO view distinguish an edit from a new question on later saves;
	// they are never a second acceptance record or an authority to overwrite.
	referencesChanged := false
	for index := range questions {
		q := &questions[index]
		if q.HypothesisID == "" {
			referencesChanged = true
			for _, h := range objectSliceAny(document["hypotheses"]) {
				if investigationLeadSameQuestion(*q, investigationLeadQuestionFromCase(h)) {
					q.HypothesisID = stringValue(h["hypothesis_id"])
					break
				}
			}
		}
		if q.HypothesisID == "" {
			professional := *q
			professional.AgentID = ""
			encoded, _ := json.Marshal(professional)
			digest := sha256.Sum256(append([]byte(stringValue(batch["observation_batch_id"])+"\x00"), encoded...))
			q.HypothesisID = "hypothesis-" + hex.EncodeToString(digest[:])
		}
	}
	if referencesChanged {
		if err := persistRepairLeadQuestionnaire(path, questions); err != nil {
			return feedback("S8 question references could not be retained in the same answer: " + err.Error())
		}
	}
	var dispatched []string
	for index, q := range questions {
		current, err := runtime.NewStore(statePath, journalPath).Snapshot()
		if err != nil {
			return feedback(err.Error())
		}
		_, document, _, err := investigationLeadSource(root, current.State)
		if err != nil || document == nil {
			return feedback(fmt.Sprintf("S8 question %d has no readable original Case: %v", index+1, err))
		}
		var registered map[string]any
		for _, h := range objectSliceAny(document["hypotheses"]) {
			if stringValue(h["hypothesis_id"]) == q.HypothesisID {
				registered = h
				break
			}
		}
		if registered != nil && !investigationLeadSameQuestion(q, investigationLeadQuestionFromCase(registered)) {
			return feedback(fmt.Sprintf("S8 question %d conflicts with registered Hypothesis %s; no accepted content was overwritten. Add a genuinely new question separately.", index+1, q.HypothesisID))
		}
		if registered == nil {
			outcomes := map[string]any{"support": q.ExpectedOutcomes["support"], "refute": q.ExpectedOutcomes["refute"]}
			current, err = investigation.RegisterHypothesis(root, statePath, journalPath, investigation.HypothesisRequest{ExpectedRevision: current.Revision, CaseID: stringValue(document["case_id"]), HypothesisID: q.HypothesisID, AssignmentID: "assignment-" + q.HypothesisID, Statement: q.Statement, Invariant: q.Invariant, Discriminator: q.Discriminator, ExpectedOutcomes: outcomes, SourceFindingIDs: q.SourceFindingIDs, EvidenceRefs: q.EvidenceRefs})
			if err != nil {
				return feedback(fmt.Sprintf("S8 question %d original Hypothesis validation: %v. Earlier registered questions remain in the Case; save again to finish only outstanding work.", index+1, err))
			}
			registered = map[string]any{"hypothesis_id": q.HypothesisID, "assignment_id": "assignment-" + q.HypothesisID}
		}
		owner, details, exists, err := investigationLeadDispatch(root, current.State, registered)
		if err != nil {
			return feedback(fmt.Sprintf("S8 question %d original dispatch: %v", index+1, err))
		}
		if exists && q.AgentID != "" && q.AgentID != owner {
			return feedback(fmt.Sprintf("S8 question %d owner conflicts with the original dispatch to %s; no dispatch was overwritten.", index+1, owner))
		}
		if !exists {
			owner = q.AgentID
			if owner == "" {
				owner = "investigator-" + strings.TrimPrefix(q.HypothesisID, "hypothesis-")
			}
			var out, diagnostic bytes.Buffer
			code := runRuntimeInvestigationDispatch([]string{"--root", root, "--case-id", stringValue(document["case_id"]), "--hypothesis-id", q.HypothesisID, "--agent-id", owner}, &out, &diagnostic)
			if code != 0 {
				return feedback(fmt.Sprintf("S8 question %d is registered; its original dispatch is still outstanding: %s. Preserve this answer and save again after resolving that source; completed registrations will be reused.", index+1, strings.TrimSpace(diagnostic.String())))
			}
			var result map[string]any
			if err := json.Unmarshal(out.Bytes(), &result); err != nil {
				return feedback("S8 original dispatch committed, but its response could not be read; save again to read the original registration.")
			}
			details = fmt.Sprintf("owner=%s TASK=%s task=%s manifest=%s", owner, stringValue(result["task_id"]), resolveRootPath(root, stringValue(result["task_path"])), resolveRootPath(root, stringValue(result["manifest_path"])))
		}
		dispatched = append(dispatched, q.HypothesisID+": "+details)
	}
	return feedback("S8 questions and original dispatch registrations are current; unchanged registered work was reused. " + strings.Join(dispatched, "; ") + ". Main must use the platform to start/resume these investigators with their registered TASK/manifest. Registration is not platform launch evidence. The original listener supplies each Investigator's exact HypothesisResult answer path; no generic PLAN_REPORT is required for plan_checkpoint.")
}

func investigationLeadDecode(data []byte, value any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode(value)
}

// The script-save observer only needs to distinguish untouched machine drafts
// from professional edits. Invalid JSON remains the original consumer's work.
func investigationLeadQuestionnaireHasProfessionalContent(path string, data []byte) bool {
	if err := workpacket.ValidateUniqueJSON(data); err != nil {
		return true
	}
	if strings.HasPrefix(filepath.Base(path), "contract-") {
		return investigationLeadContractHasProfessionalContent(path, data)
	}
	if strings.HasPrefix(filepath.Base(path), "route-") {
		return investigationLeadRouteHasProfessionalContent(data)
	}
	switch filepath.Base(path) {
	case "intake.json":
		var answer investigationLeadIntakeAnswer
		return investigationLeadDecode(data, &answer) != nil || strings.TrimSpace(answer.GroupingRationale) != ""
	case "hypotheses.json":
		var questions []investigationLeadQuestion
		if investigationLeadDecode(data, &questions) != nil {
			return true
		}
		for _, q := range questions {
			values := []string{q.Statement, q.Invariant, q.Discriminator, q.ExpectedOutcomes["support"], q.ExpectedOutcomes["refute"]}
			values = append(values, q.SourceFindingIDs...)
			values = append(values, q.EvidenceRefs...)
			for _, value := range values {
				if strings.TrimSpace(value) != "" {
					return true
				}
			}
		}
		return false
	}
	return true
}

func investigationLeadQuestionFromCase(h map[string]any) investigationLeadQuestion {
	return investigationLeadQuestion{HypothesisID: stringValue(h["hypothesis_id"]), Statement: stringValue(h["statement"]), Invariant: stringValue(h["invariant"]), Discriminator: stringValue(h["discriminator"]), ExpectedOutcomes: stringMapCLI(h["expected_outcomes"]), SourceFindingIDs: stringSliceAny(h["source_finding_ids"]), EvidenceRefs: stringSliceAny(h["evidence_refs"])}
}

// Reminder projection uses Runtime's existing activation binding. Reading a
// questionnaire never revalidates all manifests or TASK artifacts.
func investigationLeadRegisteredOwner(state map[string]any, assignment string) string {
	for _, agent := range objectSliceAny(mapFieldCLI(state, "entities")["agents"]) {
		if stringValue(agent["role"]) == "investigator" && strings.HasSuffix(stringValue(agent["prompt_ref"]), "#"+assignment) {
			return stringValue(agent["id"])
		}
	}
	return ""
}

func investigationLeadSameQuestion(a, b investigationLeadQuestion) bool {
	a.HypothesisID, b.HypothesisID, a.AgentID, b.AgentID = "", "", "", ""
	a.Statement, b.Statement = strings.TrimSpace(a.Statement), strings.TrimSpace(b.Statement)
	a.Invariant, b.Invariant = strings.TrimSpace(a.Invariant), strings.TrimSpace(b.Invariant)
	a.Discriminator, b.Discriminator = strings.TrimSpace(a.Discriminator), strings.TrimSpace(b.Discriminator)
	a.SourceFindingIDs, b.SourceFindingIDs = append([]string{}, a.SourceFindingIDs...), append([]string{}, b.SourceFindingIDs...)
	a.EvidenceRefs, b.EvidenceRefs = append([]string{}, a.EvidenceRefs...), append([]string{}, b.EvidenceRefs...)
	sort.Strings(a.SourceFindingIDs)
	sort.Strings(b.SourceFindingIDs)
	sort.Strings(a.EvidenceRefs)
	sort.Strings(b.EvidenceRefs)
	return reflect.DeepEqual(a, b)
}

// Read actual registered manifest/TASK bindings to recover a partial save.
// A workgroup name alone is never evidence that dispatch succeeded.
func investigationLeadDispatch(root string, state map[string]any, hypothesis map[string]any) (string, string, bool, error) {
	assignment := stringValue(hypothesis["assignment_id"])
	entities := mapFieldCLI(state, "entities")
	for _, team := range objectSliceAny(entities["teams"]) {
		if stringValue(team["kind"]) != "investigator" {
			continue
		}
		manifest, err := investigationLeadReadObject(root, stringValue(team["manifest_ref"]), stringValue(team["manifest_sha256"]))
		if err != nil {
			return "", "", false, err
		}
		for _, a := range objectSliceAny(manifest["assignments"]) {
			if stringValue(a["assignment_id"]) != assignment {
				continue
			}
			owner := stringValue(a["agent_id"])
			for _, task := range objectSliceAny(entities["tasks"]) {
				if !slices.Contains(stringSliceAny(task["owner_agent_ids"]), owner) {
					continue
				}
				doc, err := investigationLeadReadObject(root, stringValue(task["path"]), stringValue(task["sha256"]))
				if err != nil {
					return "", "", false, err
				}
				if stringValue(doc["assignment_id"]) == assignment && stringValue(doc["hypothesis_id"]) == stringValue(hypothesis["hypothesis_id"]) && stringValue(doc["case_id"]) == stringValue(investigationPointer(state)["case_id"]) {
					return owner, fmt.Sprintf("owner=%s TASK=%s task=%s manifest=%s", owner, stringValue(task["id"]), resolveRootPath(root, stringValue(task["path"])), resolveRootPath(root, stringValue(team["manifest_ref"]))), true, nil
				}
			}
			return "", "", false, fmt.Errorf("registered Assignment %s has no matching original TASK", assignment)
		}
	}
	return "", "", false, nil
}
