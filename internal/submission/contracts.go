package submission

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/entroforge/go-system-builder/internal/controller"
	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/semantic"
	"github.com/entroforge/go-system-builder/internal/workpacket"
	"github.com/entroforge/go-system-builder/internal/workspace"
)

func (s *Service) IssueContracts(ctx context.Context) (workpacket.ContractPacket, error) {
	s = s.WithContext(ctx)
	if err := workspace.RequireMain(s.Root); err != nil {
		return workpacket.ContractPacket{}, err
	}
	snap, err := s.snapshot()
	if err != nil {
		return workpacket.ContractPacket{}, err
	}
	binding, _, err := runtime.CurrentContractBinding(snap.State)
	if err != nil {
		return workpacket.ContractPacket{}, err
	}
	var p workpacket.ContractPacket
	entities, _ := snap.State["entities"].(map[string]any)
	rows, _ := entities["work_packets"].([]any)
	for _, raw := range rows {
		row, _ := raw.(map[string]any)
		if row["template_id"] != "s3-contract-v1" {
			continue
		}
		candidate, found, err := runtime.FindContractPacket(snap.State, text(row["packet_id"]))
		if err != nil {
			return p, err
		}
		if !found || candidate.Binding.RuntimeID != binding.RuntimeID || candidate.Binding.Generation != binding.Generation || candidate.Binding.StageEntry != binding.StageEntry {
			continue
		}
		if p.ID != "" {
			return p, fmt.Errorf("S3 current stage has more than one original questionnaire")
		}
		p = candidate
	}
	if p.ID == "" {
		inputs, err := runtime.ResolveContractInputs(s.Root, snap.State, s.Files)
		if err != nil {
			return p, err
		}
		p, err = workpacket.BuildContracts(binding, inputs)
		if err != nil {
			return p, err
		}
	}
	if err := ctx.Err(); err != nil {
		return p, err
	}
	if _, err := runtime.RecordContractPacket(s.Root, s.StatePath, s.JournalPath, p, s.Files, s.Validator); err != nil {
		return p, err
	}
	view, _ := json.MarshalIndent(p, "", "  ")
	view = append(view, '\n')
	actual, err := workpacket.Publish(s.Root, p.ViewPath, view)
	if err != nil {
		return p, err
	}
	if !bytes.Equal(actual, view) {
		return p, fmt.Errorf("published contract view differs from original issuance")
	}
	draft, _ := json.MarshalIndent(p.Draft, "", "  ")
	if _, err := workpacket.Publish(s.Root, p.AnswerPath, append(draft, '\n')); err != nil {
		return p, err
	}
	return p, nil
}

func (s *Service) ContractForPath(path string) (workpacket.ContractPacket, bool, error) {
	snap, err := s.snapshot()
	if err != nil {
		return workpacket.ContractPacket{}, false, err
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(s.Root, path)
	}
	entities, _ := snap.State["entities"].(map[string]any)
	rows, _ := entities["work_packets"].([]any)
	for _, raw := range rows {
		row, _ := raw.(map[string]any)
		if row["template_id"] != "s3-contract-v1" {
			continue
		}
		p, found, err := runtime.FindContractPacket(snap.State, text(row["packet_id"]))
		if err != nil {
			return p, false, err
		}
		if found && filepath.Clean(path) == filepath.Join(s.Root, filepath.FromSlash(p.AnswerPath)) {
			return p, true, workpacket.CheckPath(s.Root, p.AnswerPath)
		}
	}
	return workpacket.ContractPacket{}, false, nil
}

func (s *Service) SaveContracts(ctx context.Context, path string) (Receipt, error) {
	s = s.WithContext(ctx)
	if err := workspace.RequireMain(s.Root); err != nil {
		return Receipt{}, err
	}
	p, issued, err := s.ContractForPath(path)
	if err != nil || !issued {
		return Receipt{}, fmt.Errorf("contract path has no original Main issuance: %v", err)
	}
	snap, err := s.snapshot()
	if err != nil {
		return Receipt{}, err
	}
	data, err := workpacket.ReadFile(s.Root, p.AnswerPath, 256<<10)
	if err != nil {
		return rejected(snap, p.ID, err), err
	}
	answer, issues := workpacket.ValidateContractAnswer(p, data)
	if len(issues) > 0 {
		err := fmt.Errorf("contract answer %s: %s", issues[0].Pointer, issues[0].Expected)
		r := rejected(snap, p.ID, err)
		r.FieldErrors = issues
		return r, err
	}
	digest, err := workpacket.ContractAnswerDigest(answer)
	if err != nil {
		return Receipt{}, err
	}
	prior, accepted, err := runtime.FindSubmission(snap.State, p.ID)
	if err != nil {
		return Receipt{}, err
	}
	id := "ev-submission-" + workpacket.Digest([]byte("submission:"+p.ID))
	envelopePath := ""
	if accepted {
		if prior.ContentHash != digest {
			return rejected(snap, p.ID, runtime.ErrSubmissionConflict), runtime.ErrSubmissionConflict
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
		arch, err := runtime.ContractArtifacts(s.Root, snap.State, p, answer, s.Files)
		if err != nil {
			return s.unreceivedContractReceipt(snap, p, err), err
		}
		envelope, err := workpacket.ContractEvidence(p, id, arch, answer, time.Now().UTC())
		if err != nil {
			return Receipt{}, err
		}
		encoded, _ := json.MarshalIndent(envelope, "", "  ")
		encoded = append(encoded, '\n')
		envelopePath = ".claude/evidence/submissions/" + p.ID + "/" + workpacket.Digest(encoded) + ".json"
		stored, err := workpacket.Publish(s.Root, envelopePath, encoded)
		if err != nil {
			return Receipt{}, err
		}
		if !bytes.Equal(stored, encoded) {
			err := fmt.Errorf("S3 published evidence differs from this reception: %s", envelopePath)
			return s.unreceivedContractReceipt(snap, p, err), err
		}
	}
	if err := ctx.Err(); err != nil {
		return rejected(snap, p.ID, err), err
	}
	next, err := runtime.RecordEvidence(s.Root, s.StatePath, s.JournalPath, runtime.EvidenceRequest{Context: ctx, ExpectedRevision: -1, ID: id, Kind: "planning_contract", Path: envelopePath, ProducedBy: []string{"Main"}, ResponsibilityID: "Orchestrator", Validator: s.Validator, Submission: &runtime.SubmissionRequest{PacketID: p.ID, Owner: "Main", AnswerJSON: data, Files: s.Files, Authorize: func(root string, state map[string]any) error {
		if err := s.validateContractChecks(state, p, answer); err != nil {
			return err
		}
		if verified, ok := s.Files.(interface{ Verify() error }); ok {
			return verified.Verify()
		}
		return nil
	}}})
	if err != nil {
		return s.unreceivedContractReceipt(snap, p, err), err
	}
	record, found, err := runtime.FindSubmission(next.State, p.ID)
	if err != nil || !found {
		return Receipt{}, fmt.Errorf("contract reception has no durable original record")
	}
	if record.ControlStatus != "pending" {
		if _, err := runtime.SubmissionFollowup(next.State, record.FollowupID); record.ControlStatus != "no_transition" || err != nil {
			return s.contractReceipt(next, p)
		}
	}
	result, controlErr := controller.RunControlCycle(ctx, controller.ControlRequest{Root: s.Root, CWD: s.Root, StatePath: s.StatePath, JournalPath: s.JournalPath, Event: "PostToolUse", ToolName: "Write", SubmissionFollowupID: record.FollowupID, Files: s.Files})
	preparation := ""
	var preparationWaiting []Waiting
	if controlErr == nil && result.Error == "" {
		preparation, preparationWaiting = s.preparePlanningOutlet(ctx, &result)
	}
	if controlErr == nil && result.Error == "" {
		next = result.Snapshot
	}
	r, err := s.contractReceipt(next, p)
	if err != nil {
		return Receipt{}, err
	}
	if controlErr != nil || result.Error != "" {
		r.NextAction = "S3 答卷已接收；Main/系统需恢复原阶段接续，无需重新提交合同判断。"
		if controlErr != nil {
			r.NextAction += " 原因：" + controlErr.Error()
		} else {
			r.NextAction += " 原因：" + result.Error
		}
		if len(r.WaitingOn) == 0 {
			r.WaitingOn = append(r.WaitingOn, Waiting{Reason: "system_followup", Owner: "Main", Ref: record.FollowupID})
		}
	} else if !result.QualityGate.TransitionCommitted && len(result.QualityGate.Missing)+len(result.QualityGate.Conflicts) > 0 {
		r.NextAction = fmt.Sprintf("S3 答卷已接收，阶段尚未推进。原门 %s：缺项 %v；冲突 %v。按具体条目修复文档或系统接续，监听器会继续检查；无需重交已接收答案。", result.QualityGate.GateID, result.QualityGate.Missing, result.QualityGate.Conflicts)
	}
	r.NextAction += preparation
	r.WaitingOn = append(r.WaitingOn, preparationWaiting...)
	return r, nil
}

func (s *Service) validateContractChecks(state map[string]any, p workpacket.ContractPacket, answer workpacket.ContractAnswer) error {
	bound, _ := state["bound_req"].(map[string]any)
	if text(bound["id"]) != p.Binding.Requirement.ID {
		return fmt.Errorf("S3 original checker REQ binding differs")
	}
	paths := append([]string{*answer.IndexPath}, answer.ContractPaths...)
	result, err := semantic.ContractsCheckBatch(s.Root, s.Files, p.Binding.Requirement, paths)
	if err != nil {
		return fmt.Errorf("contracts_checked: %w", err)
	}
	if result.Contracts == 0 || len(result.Problems) != 0 {
		return fmt.Errorf("contracts_checked: %d contracts, problems: %v", result.Contracts, result.Problems)
	}
	return nil
}

func (s *Service) contractReceipt(snap runtime.Snapshot, p workpacket.ContractPacket) (Receipt, error) {
	r := emptyReceipt(snap, p.ID)
	record, found, err := runtime.FindSubmission(snap.State, p.ID)
	if err != nil {
		return r, err
	}
	if !found {
		return rejected(snap, p.ID, fmt.Errorf("contract answer not received")), nil
	}
	key := workpacket.Digest([]byte("submission:" + p.ID))
	if record.Owner != "Main" || record.Key != key || record.FollowupID != workpacket.Digest([]byte(key+":after-domain")) || record.OriginState != "planning" || record.OriginPhase != "contracts" || record.OriginEntry != p.Binding.StageEntry {
		return r, fmt.Errorf("contract receipt original responsibility differs")
	}
	for _, ev := range evidenceRows(snap.State) {
		if ev["id"] != record.EvidenceID {
			continue
		}
		data, err := workpacket.ReadFile(s.Root, text(ev["path"]), 1<<20)
		if err != nil || workpacket.Digest(data) != ev["sha256"] || ev["kind"] != "planning_contract" {
			return r, fmt.Errorf("original contract evidence no longer matches")
		}
		var value map[string]json.RawMessage
		json.Unmarshal(data, &value)
		var answer workpacket.ContractAnswer
		if err := json.Unmarshal(value["contract_answer"], &answer); err != nil {
			return r, fmt.Errorf("original S3 professional answer malformed")
		}
		if err := runtime.ValidateContractEnvelope(p, record.EvidenceID, answer, data); err != nil {
			return r, err
		}
		digest, err := workpacket.ContractAnswerDigest(answer)
		if err != nil || digest != record.ContentHash {
			return r, fmt.Errorf("contract receipt professional digest differs")
		}
		r.SubmissionKey = record.Key
		r.Accepted = true
		r.WorkComplete = true
		r.ResponsibilityStatus = "accepted"
		r.DomainRefs = []string{record.EvidenceID}
		r.NextAction = "合同交付已被原消费者接收；Main 按原门继续阶段工作。"
		if baseline, ok := snap.State["baseline"].(map[string]any); !ok || snap.State["runtime_id"] != p.Binding.RuntimeID || number(baseline["generation"]) != p.Binding.Generation {
			r.WorkComplete = false
			r.ResponsibilityStatus = "replaced"
		}
		if _, err := runtime.SubmissionFollowup(snap.State, record.FollowupID); err == nil {
			r.NextAction = "S3 答卷已接收，阶段接续尚未完成；监听器会继续原阶段门检查。Main/系统处理具体缺项，无需重交答案。"
			r.WaitingOn = append(r.WaitingOn, Waiting{Reason: "system_followup", Owner: "Main", Ref: record.FollowupID})
		}
		return r, nil
	}
	return r, fmt.Errorf("contract receipt original domain reference missing")
}

func (s *Service) ContractReceipt(path, packetID string) (Receipt, error) {
	snap, err := s.snapshot()
	if err != nil {
		return Receipt{}, err
	}
	var p workpacket.ContractPacket
	var found bool
	if packetID != "" {
		p, found, err = runtime.FindContractPacket(snap.State, packetID)
	} else {
		p, found, err = s.ContractForPath(path)
	}
	if err != nil || !found {
		return Receipt{}, fmt.Errorf("no original contract issuance")
	}
	return s.contractReceipt(snap, p)
}

func (s *Service) MainPlanningReceipt(path, packetID string) (Receipt, error) {
	if packetID != "" {
		snap, err := s.snapshot()
		if err != nil {
			return Receipt{}, err
		}
		if _, found, err := runtime.FindTaskPlanningPacket(snap.State, packetID); err != nil {
			return Receipt{}, err
		} else if found {
			return s.TaskPlanningReceipt(path, packetID)
		}
		if _, found, err := runtime.FindContractPacket(snap.State, packetID); err != nil {
			return Receipt{}, err
		} else if found {
			return s.ContractReceipt(path, packetID)
		}
	} else if _, found, err := s.TaskPlanningForPath(path); err != nil {
		return Receipt{}, err
	} else if found {
		return s.TaskPlanningReceipt(path, "")
	} else if _, found, err := s.ContractForPath(path); err != nil {
		return Receipt{}, err
	} else if found {
		return s.ContractReceipt(path, "")
	}
	return s.DesignReceipt(path, packetID)
}

// unreceivedContractReceipt keeps output/source errors on the current answer.
// Named revision applies only to an already accepted judgment.
func (s *Service) unreceivedContractReceipt(snapshot runtime.Snapshot, p workpacket.ContractPacket, err error) Receipt {
	r := rejected(snapshot, p.ID, err)
	r.NextAction = "本次 S3 答卷尚未接收。按具体错误修正专业文件或答卷，再保存原路径 " + filepath.Join(s.Root, filepath.FromSlash(p.AnswerPath)) + "；保留本次所需的专业修复。"
	return r
}
