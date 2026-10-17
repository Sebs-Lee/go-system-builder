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
	if err := workspace.RequireMain(s.Root); err != nil {
		return Receipt{}, err
	}
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
		id = prior.EvidenceID
		found := false
		for _, ev := range evidenceRows(snap.State) {
			if ev["id"] == id {
				envelopePath = text(ev["path"])
				found = true
				break
			}
		}
		if !found {
			return Receipt{}, fmt.Errorf("original planning evidence missing")
		}
	} else {
		if err := runtime.ValidateIssuedTaskPlanningPacket(s.Root, snap.State, p, s.Files); err != nil {
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
		stored, err := workpacket.Publish(s.Root, envelopePath, append(encoded, '\n'))
		if err != nil {
			return Receipt{}, err
		}
		if err := runtime.ValidateTaskPlanningEnvelope(p, id, answer, stored); err != nil {
			return Receipt{}, err
		}
	}
	if err := ctx.Err(); err != nil {
		return s.rejectedTaskPlanning(snap, p, err), err
	}
	next, err := runtime.RecordEvidence(s.Root, s.StatePath, s.JournalPath, runtime.EvidenceRequest{ExpectedRevision: -1, ID: id, Kind: "planning_task", Path: envelopePath, ProducedBy: []string{"Main"}, ResponsibilityID: "Orchestrator", Validator: s.Validator, Submission: &runtime.SubmissionRequest{PacketID: p.ID, Owner: "Main", AnswerJSON: data, Files: s.Files, Authorize: func(root string, state map[string]any) error {
		if err := workspace.RequireMain(root); err != nil {
			return err
		}
		if err := runtime.ValidateIssuedTaskPlanningPacket(root, state, p, s.Files); err != nil {
			return err
		}
		return s.validateTaskPlanningChecks(state, p)
	}}})
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
	key := workpacket.Digest([]byte("submission:" + p.ID))
	if record.Owner != "Main" || record.Key != key || record.FollowupID != workpacket.Digest([]byte(key+":after-domain")) || record.OriginState != "planning" || record.OriginPhase != "tasks" || record.OriginEntry != p.Binding.StageEntry {
		return r, fmt.Errorf("S4 receipt original responsibility differs")
	}
	for _, ev := range evidenceRows(snap.State) {
		if ev["id"] != record.EvidenceID {
			continue
		}
		data, err := workpacket.ReadFile(s.Root, text(ev["path"]), 1<<20)
		if err != nil || workpacket.Digest(data) != ev["sha256"] || ev["kind"] != "planning_task" {
			return r, fmt.Errorf("original S4 evidence no longer matches")
		}
		var value map[string]json.RawMessage
		json.Unmarshal(data, &value)
		var answer workpacket.TaskPlanningAnswer
		if err := json.Unmarshal(value["task_planning_answer"], &answer); err != nil {
			return r, fmt.Errorf("original S4 professional answer malformed")
		}
		if err := runtime.ValidateTaskPlanningEnvelope(p, record.EvidenceID, answer, data); err != nil {
			return r, err
		}
		digest, err := workpacket.TaskPlanningAnswerDigest(answer)
		if err != nil || digest != record.ContentHash {
			return r, fmt.Errorf("S4 receipt professional digest differs")
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
		if _, err := runtime.SubmissionFollowup(snap.State, record.FollowupID); err == nil {
			r.NextAction = "S4 本人交付已接收；原 TR-002 门和 system followup 仍待处理，Main 按正常监听流程继续；无需重交或手动 recover。"
			r.WaitingOn = append(r.WaitingOn, Waiting{Reason: "system_followup", Owner: "Main", Ref: record.FollowupID})
		} else if record.ControlStatus == "advanced" {
			r.NextAction = "S4 本人交付已接收，TR-002 已推进至 S5；后续由 S5 独立审查。"
		}
		return r, nil
	}
	return r, fmt.Errorf("S4 receipt original domain reference missing")
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
