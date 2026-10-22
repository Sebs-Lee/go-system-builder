package submission

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/entroforge/go-system-builder/internal/controller"
	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/semantic"
	"github.com/entroforge/go-system-builder/internal/workpacket"
	"github.com/entroforge/go-system-builder/internal/workspace"
)

func (s *Service) IssueTaskPlanning(ctx context.Context) (workpacket.TaskPlanningPacket, error) {
	s = s.WithContext(ctx)
	if err := workspace.RequireMain(s.Root); err != nil {
		return workpacket.TaskPlanningPacket{}, err
	}
	snap, err := s.snapshot()
	if err != nil {
		return workpacket.TaskPlanningPacket{}, err
	}
	binding, err := runtime.CurrentTaskPlanningBinding(snap.State)
	if err != nil {
		return workpacket.TaskPlanningPacket{}, err
	}
	p, found, err := currentTaskPlanningPacket(snap.State, binding)
	if err != nil {
		return p, err
	}
	if !found {
		inputs, err := runtime.ResolveTaskPlanningInputs(s.Root, snap.State, s.Files)
		if err != nil {
			return p, err
		}
		p, err = workpacket.BuildTaskPlanning(binding, inputs)
		if err != nil {
			return p, err
		}
	}
	if err := ctx.Err(); err != nil {
		return p, err
	}
	if _, err := runtime.RecordTaskPlanningPacket(s.Root, s.StatePath, s.JournalPath, p, s.Files, s.Validator); err != nil {
		return p, err
	}
	view, _ := json.MarshalIndent(p, "", "  ")
	view = append(view, '\n')
	actual, err := workpacket.Publish(s.Root, p.ViewPath, view)
	if err != nil {
		return p, err
	}
	if !bytes.Equal(actual, view) {
		return p, fmt.Errorf("published S4 view differs from original issuance")
	}
	draft, _ := json.MarshalIndent(p.Draft, "", "  ")
	if _, err := workpacket.Publish(s.Root, p.AnswerPath, append(draft, '\n')); err != nil {
		return p, err
	}
	return p, nil
}

func (s *Service) TaskPlanningForPath(path string) (workpacket.TaskPlanningPacket, bool, error) {
	snap, err := s.snapshot()
	if err != nil {
		return workpacket.TaskPlanningPacket{}, false, err
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(s.Root, path)
	}
	entities, _ := snap.State["entities"].(map[string]any)
	rows, _ := entities["work_packets"].([]any)
	for _, raw := range rows {
		row, _ := raw.(map[string]any)
		if row["template_id"] != "s4-task-planning-v1" {
			continue
		}
		p, found, err := runtime.FindTaskPlanningPacket(snap.State, text(row["packet_id"]))
		if err != nil {
			return p, false, err
		}
		if found && filepath.Clean(path) == filepath.Join(s.Root, filepath.FromSlash(p.AnswerPath)) {
			return p, true, workpacket.CheckPath(s.Root, p.AnswerPath)
		}
	}
	return workpacket.TaskPlanningPacket{}, false, nil
}

func (s *Service) SaveTaskPlanning(ctx context.Context, path string) (Receipt, error) {
	s = s.WithContext(ctx)
	p, issued, err := s.TaskPlanningForPath(path)
	if err != nil || !issued {
		return Receipt{}, fmt.Errorf("S4 answer path has no original Main issuance: %v", err)
	}
	snap, err := s.snapshot()
	if err != nil {
		return Receipt{}, err
	}
	data, err := workpacket.ReadFile(s.Root, p.AnswerPath, 256<<10)
	if err != nil {
		return s.rejectedTaskPlanning(snap, p, err), err
	}
	answer, issues := workpacket.ValidateTaskPlanningAnswer(p, data)
	if len(issues) > 0 {
		err := fmt.Errorf("S4 task-planning answer %s: %s", issues[0].Pointer, issues[0].Expected)
		r := s.rejectedTaskPlanning(snap, p, err)
		r.FieldErrors = issues
		return r, err
	}
	digest, err := workpacket.TaskPlanningAnswerDigest(answer)
	if err != nil {
		return Receipt{}, err
	}
	prior, accepted, err := runtime.FindSubmission(snap.State, p.ID)
	if err != nil {
		return Receipt{}, err
	}
	id := "ev-submission-" + workpacket.Digest([]byte("submission:"+p.ID))
	envelopePath := ".claude/evidence/submissions/" + p.ID + "/" + digest + ".json"
	if accepted {
		if prior.ContentHash != digest {
			return s.rejectedTaskPlanning(snap, p, runtime.ErrSubmissionConflict), runtime.ErrSubmissionConflict
		}
		life, _ := snap.State["lifecycle"].(map[string]any)
		baseline, _ := snap.State["baseline"].(map[string]any)
		if (prior.ControlStatus == "pending" || prior.ControlStatus == "no_transition") && snap.State["runtime_id"] == p.Binding.RuntimeID && number(baseline["generation"]) == p.Binding.Generation && life["state"] == "planning" && life["phase"] == "tasks" && number(life["phase_revision"]) == p.Binding.StageEntry {
			result, controlErr := controller.RunControlCycle(ctx, controller.ControlRequest{Root: s.Root, CWD: s.Root, StatePath: s.StatePath, JournalPath: s.JournalPath, Event: "PostToolUse", ToolName: "Write", SubmissionFollowupID: prior.FollowupID, Files: s.Files})
			if controlErr == nil && result.Error == "" {
				snap = result.Snapshot
			}
		}
		return s.taskPlanningReceipt(snap, p)
	} else {
		life, _ := snap.State["lifecycle"].(map[string]any)
		baseline, _ := snap.State["baseline"].(map[string]any)
		if snap.State["runtime_id"] != p.Binding.RuntimeID || number(baseline["generation"]) != p.Binding.Generation || life["state"] != "planning" || life["phase"] != "tasks" || number(life["phase_revision"]) != p.Binding.StageEntry {
			err := fmt.Errorf("S4 answer is not for the current task-planning responsibility")
			return s.rejectedTaskPlanning(snap, p, err), err
		}
		if err := s.validateTaskPlanningChecks(snap.State, p); err != nil {
			return s.rejectedTaskPlanning(snap, p, err), err
		}
		arch, err := runtime.TaskPlanningArtifacts(s.Root, snap.State, p, answer, s.Files)
		if err != nil {
			return s.rejectedTaskPlanning(snap, p, err), err
		}
		envelope, err := workpacket.TaskPlanningEvidence(p, id, arch, answer, time.Now().UTC())
		if err != nil {
			return Receipt{}, err
		}
		encoded, _ := json.MarshalIndent(envelope, "", "  ")
		_, err = workpacket.Publish(s.Root, envelopePath, append(encoded, '\n'))
		if err != nil {
			return Receipt{}, err
		}
	}
	if err := ctx.Err(); err != nil {
		return s.rejectedTaskPlanning(snap, p, err), err
	}
	next, err := runtime.RecordEvidence(s.Root, s.StatePath, s.JournalPath, runtime.EvidenceRequest{Context: ctx, ExpectedRevision: -1, ID: id, Kind: "planning_task", Path: envelopePath, ProducedBy: []string{"Main"}, ResponsibilityID: "Orchestrator", Validator: s.Validator, Submission: &runtime.SubmissionRequest{PacketID: p.ID, Owner: "Main", AnswerJSON: data, Files: s.Files}})
	if err != nil {
		return rejected(snap, p.ID, err), err
	}
	record, found, err := runtime.FindSubmission(next.State, p.ID)
	if err != nil || !found {
		return Receipt{}, fmt.Errorf("S4 reception has no durable original record")
	}
	result, controlErr := controller.RunControlCycle(ctx, controller.ControlRequest{Root: s.Root, CWD: s.Root, StatePath: s.StatePath, JournalPath: s.JournalPath, Event: "PostToolUse", ToolName: "Write", SubmissionFollowupID: record.FollowupID, Files: s.Files})
	if controlErr == nil && result.Error == "" {
		next = result.Snapshot
	}
	r, err := s.taskPlanningReceipt(next, p)
	if err != nil {
		return Receipt{}, err
	}
	return r, nil
}

func currentTaskPlanningPacket(state map[string]any, binding workpacket.PlanningBinding) (workpacket.TaskPlanningPacket, bool, error) {
	entities, _ := state["entities"].(map[string]any)
	rows, _ := entities["work_packets"].([]any)
	var current workpacket.TaskPlanningPacket
	for _, raw := range rows {
		row, _ := raw.(map[string]any)
		if row["template_id"] != "s4-task-planning-v1" {
			continue
		}
		candidate, found, err := runtime.FindTaskPlanningPacket(state, text(row["packet_id"]))
		if err != nil {
			return current, false, err
		}
		if !found || candidate.Binding.RuntimeID != binding.RuntimeID || candidate.Binding.Generation != binding.Generation || candidate.Binding.StageEntry != binding.StageEntry {
			continue
		}
		if candidate.Binding != binding {
			return candidate, false, fmt.Errorf("S4 current stage has an original packet with a different Main or locked REQ binding")
		}
		if current.ID != "" {
			return current, false, fmt.Errorf("S4 current stage has more than one original questionnaire")
		}
		current = candidate
	}
	return current, current.ID != "", nil
}

func (s *Service) rejectedTaskPlanning(snapshot runtime.Snapshot, p workpacket.TaskPlanningPacket, err error) Receipt {
	r := rejected(snapshot, p.ID, err)
	answerPath := filepath.Join(s.Root, filepath.FromSlash(p.AnswerPath))
	r.NextAction = fmt.Sprintf("按具体错误修复 S4 交付或来源，再保存原答案 %s。冻结上游变化须按原阶段流程处理。", answerPath)
	var outputErr *runtime.TaskPlanningOutputSourceError
	if errors.As(err, &outputErr) {
		r.NextAction = outputErr.NextAction()
	}
	if errors.Is(err, runtime.ErrSubmissionConflict) {
		r.NextAction = fmt.Sprintf("原 S4 答案已接收；内容变更按原返工流程处理。当前答案路径为 %s。", answerPath)
	}
	return r
}

func (s *Service) validateTaskPlanningChecks(state map[string]any, p workpacket.TaskPlanningPacket) error {
	bound, _ := state["bound_req"].(map[string]any)
	if text(bound["id"]) != p.Binding.Requirement.ID {
		return fmt.Errorf("S4 original checker REQ binding differs")
	}
	result, err := semantic.TasksCheckWithFiles(s.Root, s.Files, p.Binding.Requirement.ID)
	if err != nil {
		return fmt.Errorf("tasks_checked: %w", err)
	}
	if result.Tasks == 0 || len(result.Problems) != 0 {
		return fmt.Errorf("tasks_checked: %d TASKs, problems: %v", result.Tasks, result.Problems)
	}
	return nil
}

func (s *Service) taskPlanningReceipt(snap runtime.Snapshot, p workpacket.TaskPlanningPacket) (Receipt, error) {
	r := emptyReceipt(snap, p.ID)
	record, found, err := runtime.FindSubmission(snap.State, p.ID)
	if err != nil {
		return r, err
	}
	if !found {
		return rejected(snap, p.ID, fmt.Errorf("S4 answer not received")), nil
	}
	r.SubmissionKey = record.Key
	r.Accepted = true
	r.WorkComplete = true
	r.ResponsibilityStatus = "accepted"
	r.DomainRefs = []string{record.EvidenceID}
	r.NextAction = "S4 本人任务拆分交付已接收；原 TR-002 门由系统继续处理。接收不代表 S5 审查通过。"
	if baseline, ok := snap.State["baseline"].(map[string]any); !ok || snap.State["runtime_id"] != p.Binding.RuntimeID || number(baseline["generation"]) != p.Binding.Generation {
		r.WorkComplete = false
		r.ResponsibilityStatus = "replaced"
	}
	life, _ := snap.State["lifecycle"].(map[string]any)
	if record.ControlStatus == "replaced" || ((record.ControlStatus == "pending" || record.ControlStatus == "no_transition") && (life["state"] != "planning" || life["phase"] != "tasks" || number(life["phase_revision"]) != p.Binding.StageEntry)) {
		r.WorkComplete = false
		r.ResponsibilityStatus = "replaced"
	}
	if !r.WorkComplete {
		r.NextAction = "这是旧代或已替换的 S4 接收记录；按当前监听给出的问卷继续，不重交此旧卷。"
	} else if r.WorkComplete && (record.ControlStatus == "pending" || record.ControlStatus == "no_transition") && life["state"] == "planning" && life["phase"] == "tasks" && number(life["phase_revision"]) == p.Binding.StageEntry {
		r.ConsumerOwner, r.ConsumerPending = "main", true
		r.NextAction = "S4 本人交付已接收，可以结束本责任；Main 检查原 TR-002 门与 system_followup " + record.FollowupID + " 的具体缺项。真实材料修复后由原 Post 接续；技术中断沿已有 runtime stage-answer recover 恢复，不重交已接收答案，也不制造工具事件等待推进。"
		r.WaitingOn = append(r.WaitingOn, Waiting{Reason: "system_followup", Owner: "Main", Ref: record.FollowupID})
	} else if record.ControlStatus == "advanced" {
		r.NextAction = "S4 本人交付已接收，TR-002 已推进至 S5；后续由 S5 独立审查。"
	}
	return r, nil
}

func (s *Service) TaskPlanningReceipt(path, packetID string) (Receipt, error) {
	snap, err := s.snapshot()
	if err != nil {
		return Receipt{}, err
	}
	var p workpacket.TaskPlanningPacket
	var found bool
	if packetID != "" {
		p, found, err = runtime.FindTaskPlanningPacket(snap.State, packetID)
	} else {
		p, found, err = s.TaskPlanningForPath(path)
	}
	if err != nil || !found {
		return Receipt{}, fmt.Errorf("no original S4 issuance")
	}
	return s.taskPlanningReceipt(snap, p)
}
