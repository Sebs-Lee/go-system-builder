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
	"github.com/entroforge/go-system-builder/internal/workpacket"
	"github.com/entroforge/go-system-builder/internal/workspace"
)

func (s *Service) IssueDesign(ctx context.Context) (workpacket.DesignPacket, error) {
	if err := workspace.RequireMain(s.Root); err != nil {
		return workpacket.DesignPacket{}, err
	}
	snap, err := s.snapshot()
	if err != nil {
		return workpacket.DesignPacket{}, err
	}
	p, err := runtime.CurrentDesignPacket(snap.State)
	if err != nil {
		return p, err
	}
	if err := ctx.Err(); err != nil {
		return p, err
	}
	if _, err := runtime.RecordDesignPacket(s.Root, s.StatePath, s.JournalPath, p, s.Files, s.Validator); err != nil {
		return p, err
	}
	view, _ := json.MarshalIndent(p, "", "  ")
	view = append(view, '\n')
	actual, err := workpacket.Publish(s.Root, p.ViewPath, view)
	if err != nil {
		return p, err
	}
	if !bytes.Equal(actual, view) {
		return p, fmt.Errorf("published design view differs from original issuance")
	}
	draft, _ := json.MarshalIndent(p.Draft, "", "  ")
	if _, err := workpacket.Publish(s.Root, p.AnswerPath, append(draft, '\n')); err != nil {
		return p, err
	}
	return p, nil
}

func (s *Service) DesignForPath(path string) (workpacket.DesignPacket, bool, error) {
	snap, err := s.snapshot()
	if err != nil {
		return workpacket.DesignPacket{}, false, err
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(s.Root, path)
	}
	entities, _ := snap.State["entities"].(map[string]any)
	rows, _ := entities["work_packets"].([]any)
	for _, raw := range rows {
		row, _ := raw.(map[string]any)
		if row["template_id"] != "s2-design-v1" {
			continue
		}
		p, found, err := runtime.FindDesignPacket(snap.State, text(row["packet_id"]))
		if err != nil {
			return p, false, err
		}
		if found && filepath.Clean(path) == filepath.Join(s.Root, filepath.FromSlash(p.AnswerPath)) {
			return p, true, workpacket.CheckPath(s.Root, p.AnswerPath)
		}
	}
	return workpacket.DesignPacket{}, false, nil
}

func (s *Service) SaveDesign(ctx context.Context, path string) (Receipt, error) {
	if err := workspace.RequireMain(s.Root); err != nil {
		return Receipt{}, err
	}
	p, issued, err := s.DesignForPath(path)
	if err != nil || !issued {
		return Receipt{}, fmt.Errorf("design path has no original Main issuance: %v", err)
	}
	snap, err := s.snapshot()
	if err != nil {
		return Receipt{}, err
	}
	data, err := workpacket.ReadFile(s.Root, p.AnswerPath, 256<<10)
	if err != nil {
		return rejected(snap, p.ID, err), err
	}
	answer, issues := workpacket.ValidateDesignAnswer(p, data)
	if len(issues) > 0 {
		err := fmt.Errorf("design answer %s: %s", issues[0].Pointer, issues[0].Expected)
		r := rejected(snap, p.ID, err)
		r.FieldErrors = issues
		return r, err
	}
	digest, err := workpacket.DesignAnswerDigest(answer)
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
		if err := runtime.ValidateDesignPacket(s.Root, snap.State, p, s.Files); err != nil {
			return rejected(snap, p.ID, err), err
		}
		arch, err := runtime.DesignArchitecture(s.Root, *answer.ArchitecturePath, s.Files)
		if err != nil {
			return rejected(snap, p.ID, err), err
		}
		envelope, err := workpacket.DesignEvidence(p, id, arch, answer, time.Now().UTC())
		if err != nil {
			return Receipt{}, err
		}
		encoded, _ := json.MarshalIndent(envelope, "", "  ")
		stored, err := workpacket.Publish(s.Root, envelopePath, append(encoded, '\n'))
		if err != nil {
			return Receipt{}, err
		}
		if err := runtime.ValidateDesignEnvelope(p, id, answer, stored); err != nil {
			return Receipt{}, err
		}
	}
	if err := ctx.Err(); err != nil {
		return rejected(snap, p.ID, err), err
	}
	next, err := runtime.RecordEvidence(s.Root, s.StatePath, s.JournalPath, runtime.EvidenceRequest{ExpectedRevision: -1, ID: id, Kind: "planning_design", Path: envelopePath, ProducedBy: []string{"Main"}, ResponsibilityID: "Orchestrator", Validator: s.Validator, Submission: &runtime.SubmissionRequest{PacketID: p.ID, Owner: "Main", AnswerJSON: data, Files: s.Files, Authorize: func(root string, state map[string]any) error {
		if err := workspace.RequireMain(root); err != nil {
			return err
		}
		return runtime.ValidateDesignPacket(root, state, p, s.Files)
	}}})
	if err != nil {
		return rejected(snap, p.ID, err), err
	}
	record, found, err := runtime.FindSubmission(next.State, p.ID)
	if err != nil || !found {
		return Receipt{}, fmt.Errorf("design reception has no durable original record")
	}
	result, controlErr := controller.RunControlCycle(ctx, controller.ControlRequest{Root: s.Root, CWD: s.Root, StatePath: s.StatePath, JournalPath: s.JournalPath, Event: "PostToolUse", ToolName: "Write", SubmissionFollowupID: record.FollowupID, Files: s.Files})
	if controlErr == nil && result.Error == "" {
		next = result.Snapshot
	}
	r, err := s.designReceipt(next, p)
	if err != nil {
		return Receipt{}, err
	}
	if controlErr != nil || result.Error != "" {
		r.NextAction = "Main 运行 `.claude/bin/loop-harness runtime stage-answer recover --root . --agent Main` 接续原 system followup；无需重新提交设计判断。"
		if len(r.WaitingOn) == 0 {
			r.WaitingOn = append(r.WaitingOn, Waiting{Reason: "system_followup", Owner: "Main", Ref: record.FollowupID})
		}
	}
	return r, nil
}

func (s *Service) designReceipt(snap runtime.Snapshot, p workpacket.DesignPacket) (Receipt, error) {
	r := emptyReceipt(snap, p.ID)
	record, found, err := runtime.FindSubmission(snap.State, p.ID)
	if err != nil {
		return r, err
	}
	if !found {
		return rejected(snap, p.ID, fmt.Errorf("design answer not received")), nil
	}
	key := workpacket.Digest([]byte("submission:" + p.ID))
	if record.Owner != "Main" || record.Key != key || record.FollowupID != workpacket.Digest([]byte(key+":after-domain")) || record.OriginState != "planning" || record.OriginPhase != "design" || record.OriginEntry != p.Binding.StageEntry {
		return r, fmt.Errorf("design receipt original responsibility differs")
	}
	for _, ev := range evidenceRows(snap.State) {
		if ev["id"] != record.EvidenceID {
			continue
		}
		data, err := workpacket.ReadFile(s.Root, text(ev["path"]), 1<<20)
		if err != nil || workpacket.Digest(data) != ev["sha256"] || ev["kind"] != "planning_design" {
			return r, fmt.Errorf("original design evidence no longer matches")
		}
		var value map[string]json.RawMessage
		json.Unmarshal(data, &value)
		var conclusion string
		var subjects []workpacket.Subject
		if json.Unmarshal(value["conclusion"], &conclusion) != nil || json.Unmarshal(value["subject_refs"], &subjects) != nil || len(subjects) != 2 {
			return r, fmt.Errorf("original design evidence malformed")
		}
		path := subjects[1].Path
		answer := workpacket.DesignAnswer{Conclusion: &conclusion, ArchitecturePath: &path}
		if err := runtime.ValidateDesignEnvelope(p, record.EvidenceID, answer, data); err != nil {
			return r, err
		}
		digest, err := workpacket.DesignAnswerDigest(answer)
		if err != nil || digest != record.ContentHash {
			return r, fmt.Errorf("design receipt professional digest differs")
		}
		r.SubmissionKey = record.Key
		r.Accepted = true
		r.WorkComplete = true
		r.ResponsibilityStatus = "accepted"
		r.DomainRefs = []string{record.EvidenceID}
		r.NextAction = "设计交付已被原消费者接收；Main 按原门继续阶段工作。"
		if baseline, ok := snap.State["baseline"].(map[string]any); !ok || snap.State["runtime_id"] != p.Binding.RuntimeID || number(baseline["generation"]) != p.Binding.Generation {
			r.WorkComplete = false
			r.ResponsibilityStatus = "replaced"
		}
		if record.ControlStatus == "pending" {
			r.NextAction = "Main 运行 `.claude/bin/loop-harness runtime stage-answer recover --root . --agent Main` 接续原 system followup；无需重新提交设计判断。"
			r.WaitingOn = append(r.WaitingOn, Waiting{Reason: "system_followup", Owner: "Main", Ref: record.FollowupID})
		}
		return r, nil
	}
	return r, fmt.Errorf("design receipt original domain reference missing")
}

func (s *Service) DesignReceipt(path, packetID string) (Receipt, error) {
	snap, err := s.snapshot()
	if err != nil {
		return Receipt{}, err
	}
	var p workpacket.DesignPacket
	var found bool
	if packetID != "" {
		p, found, err = runtime.FindDesignPacket(snap.State, packetID)
	} else {
		p, found, err = s.DesignForPath(path)
	}
	if err != nil || !found {
		return Receipt{}, fmt.Errorf("no original design issuance")
	}
	return s.designReceipt(snap, p)
}
