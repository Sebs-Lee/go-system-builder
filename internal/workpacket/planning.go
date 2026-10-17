package workpacket

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/entroforge/go-system-builder/internal/schema"
)

// PlanningBinding identifies the actual Main stage responsibility. It never
// fabricates a dispatched Assignment or an independent professional agent.
type PlanningBinding struct {
	RuntimeID   string    `json:"runtime_id"`
	Generation  int       `json:"generation"`
	StageEntry  int       `json:"stage_entry"`
	Stage       string    `json:"stage"`
	Owner       string    `json:"owner"`
	Requirement Reference `json:"requirement"`
}
type DesignAnswer struct {
	Conclusion       *string `json:"conclusion"`
	ArchitecturePath *string `json:"architecture_path"`
}
type DesignDraft struct {
	Title       string              `json:"title"`
	Instruction string              `json:"instruction"`
	Read        []ReadItem          `json:"read"`
	Questions   map[string]Question `json:"questions"`
	Answer      DesignAnswer        `json:"answer"`
}
type DesignPacket struct {
	ID          string          `json:"packet_id"`
	TemplateID  string          `json:"template_id"`
	TemplateSHA string          `json:"template_sha256"`
	Binding     PlanningBinding `json:"binding"`
	ViewPath    string          `json:"view_path"`
	AnswerPath  string          `json:"answer_path"`
	Draft       DesignDraft     `json:"draft"`
}

func BuildDesign(binding PlanningBinding) (DesignPacket, error) {
	req := binding.Requirement
	if binding.Stage != "S2" || binding.Owner != "Main" || !safeComponent(binding.RuntimeID) || binding.Generation < 1 || binding.StageEntry < 0 || req.ID == "" || req.Version == "" || !safeRelative(req.Path) || !digestPattern.MatchString(req.SHA256) {
		return DesignPacket{}, fmt.Errorf("S2 requires the actual Main stage and frozen requirement binding")
	}
	data, err := schema.ReadAsset("stage-form-s2-template.json")
	if err != nil {
		return DesignPacket{}, err
	}
	var template struct {
		Stage, Adapter, Title, Instruction string
		Questions                          map[string]Question
	}
	if err := json.Unmarshal(data, &template); err != nil {
		return DesignPacket{}, err
	}
	if template.Stage != "S2" || template.Adapter != "planning_design" || len(template.Questions) != 2 {
		return DesignPacket{}, fmt.Errorf("unregistered S2 design template")
	}
	templateSHA := Digest(data)
	identity, _ := json.Marshal(struct {
		Binding     PlanningBinding
		TemplateSHA string
	}{binding, templateSHA})
	id := "packet-" + Digest(identity)
	base := fmt.Sprintf("%s/g%d/S2/e%d/%s", binding.RuntimeID, binding.Generation, binding.StageEntry, id)
	p := DesignPacket{ID: id, TemplateID: "s2-design-v1", TemplateSHA: templateSHA, Binding: binding, ViewPath: ".claude/work-packets/" + base + ".json", AnswerPath: ".claude/submissions/" + base + "/Main/answer.json", Draft: DesignDraft{Title: template.Title, Instruction: template.Instruction, Read: []ReadItem{{Path: req.Path, Purpose: "Use this locked requirement to complete the original architecture and design self-review."}}, Questions: template.Questions}}
	return p, nil
}

func ValidateDesignAnswer(p DesignPacket, data []byte) (DesignAnswer, []FieldError) {
	issue := func(pointer, expected, observed, repair string) (DesignAnswer, []FieldError) {
		return DesignAnswer{}, []FieldError{{Pointer: pointer, Expected: expected, Observed: observed, Repair: repair}}
	}
	if len(data) > 256<<10 {
		return issue("/", "JSON <= 256KiB", "oversized", "保留当前题干，只填写专业答案。")
	}
	if err := rejectDuplicateKeys(data); err != nil {
		return issue("/", "unique JSON keys", err.Error(), "删除重复字段。")
	}
	if err := schema.NewEmbeddedValidator().ValidateBytes("stage-answer-s2.schema.json", data); err != nil {
		return issue("/", "registered S2 answer", err.Error(), "保留题干，只填写 answer。")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var draft DesignDraft
	if err := decoder.Decode(&draft); err != nil {
		return issue("/", "one design questionnaire", err.Error(), "恢复已发行问卷。")
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return issue("/", "one JSON object", "trailing data", "删除后续内容。")
	}
	answer := draft.Answer
	draft.Answer = p.Draft.Answer
	if !reflect.DeepEqual(draft, p.Draft) {
		return issue("/", "original issued questionnaire", "machine fields changed", "恢复题干与冻结来源。")
	}
	if answer.Conclusion == nil || *answer.Conclusion != "pass" {
		return issue("/answer/conclusion", "explicit professional design self-review pass", "unanswered", "完成原设计自审后填写；不能用文件存在推断通过。")
	}
	if answer.ArchitecturePath == nil || !safeRelative(*answer.ArchitecturePath) {
		return issue("/answer/architecture_path", "actual locked ARCHITECTURE path", "missing or escaping", "填写本次原设计产物的真实项目路径。")
	}
	return answer, nil
}

func DesignAnswerDigest(answer DesignAnswer) (string, error) {
	data, err := json.Marshal(answer)
	if err != nil {
		return "", err
	}
	return Digest(data), nil
}

// ValidateUniqueJSON shares the authoring duplicate-key boundary with adapters.
func ValidateUniqueJSON(data []byte) error { return rejectDuplicateKeys(data) }

// DesignEvidence is the original planning_design envelope, not a transition
// or a quality certificate. The adapter must validate the actual artifact first.
func DesignEvidence(p DesignPacket, id string, architecture Reference, answer DesignAnswer, at time.Time) (map[string]any, error) {
	if answer.Conclusion == nil || *answer.Conclusion != "pass" || answer.ArchitecturePath == nil || architecture.Path != *answer.ArchitecturePath || architecture.Version == "" || !digestPattern.MatchString(architecture.SHA256) {
		return nil, fmt.Errorf("planning evidence differs from professional answer or actual architecture")
	}
	req := p.Binding.Requirement
	return map[string]any{"schema_version": "1.0.0", "evidence_id": id, "kind": "planning_design", "runtime_id": p.Binding.RuntimeID, "baseline_generation": p.Binding.Generation, "producer_agent_id": p.Binding.Owner, "producer_responsibility": "Orchestrator", "subject_refs": []Subject{{Path: req.Path, Version: req.Version, SHA256: req.SHA256}, {Path: architecture.Path, Version: architecture.Version, SHA256: architecture.SHA256}}, "conclusion": *answer.Conclusion, "created_at": at.UTC().Format(time.RFC3339Nano)}, nil
}
