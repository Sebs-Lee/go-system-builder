package submission

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/entroforge/go-system-builder/internal/controller"
	"github.com/entroforge/go-system-builder/internal/fileview"
	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/transition"
	"github.com/entroforge/go-system-builder/internal/workpacket"
)

func (s *Service) IssueDesign(ctx context.Context) (workpacket.DesignPacket, error) {
	s = s.WithContext(ctx)
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
	return s.designForPath(snap, path)
}

func (s *Service) designForPath(snap runtime.Snapshot, path string) (workpacket.DesignPacket, bool, error) {
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
		if filepath.Clean(path) != filepath.Join(s.Root, filepath.FromSlash(text(row["answer_path"]))) {
			continue
		}
		p, found, err := runtime.FindDesignPacket(snap.State, text(row["packet_id"]))
		return p, found, err
	}
	return workpacket.DesignPacket{}, false, nil
}

func (s *Service) SaveDesign(ctx context.Context, path string) (Receipt, error) {
	s = s.WithContext(ctx)
	snap, err := s.snapshot()
	if err != nil {
		return Receipt{}, err
	}
	p, issued, err := s.designForPath(snap, path)
	if err != nil || !issued {
		return Receipt{}, fmt.Errorf("design path does not locate its registered responsibility: %v", err)
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
	prior, accepted, err := runtime.FindSubmission(snap.State, p.ID)
	if err != nil {
		return Receipt{}, err
	}
	digest, err := workpacket.DesignAnswerDigest(answer)
	if err != nil {
		return Receipt{}, err
	}
	if accepted && prior.ContentHash != digest {
		return rejected(snap, p.ID, runtime.ErrSubmissionConflict), runtime.ErrSubmissionConflict
	}
	next := snap
	if !accepted {
		if err := runtime.ValidateDesignPacket(s.Root, snap.State, p, s.Files); err != nil {
			return rejected(snap, p.ID, err), err
		}
		var missing []string
		if answer.Conclusion == nil || strings.TrimSpace(*answer.Conclusion) == "" {
			missing = append(missing, "complete the actual design self-review; leave conclusion empty while unfinished")
		}
		if answer.ArchitecturePath == nil || *answer.ArchitecturePath == "" {
			missing = append(missing, "provide the actual architecture_path")
		}
		var arch workpacket.Reference
		if answer.ArchitecturePath != nil && *answer.ArchitecturePath != "" {
			arch, err = runtime.DesignArchitecture(s.Root, *answer.ArchitecturePath, nil)
			if err != nil {
				missing = append(missing, err.Error())
			}
		}
		if len(missing) > 0 {
			r := emptyReceipt(snap, p.ID)
			r.NextAction = "Continue the current design answer at " + filepath.Join(s.Root, p.AnswerPath) + ": " + strings.Join(missing, "; ") + ". No design decision was recorded."
			return r, nil
		}
		id := "ev-submission-" + workpacket.Digest([]byte("submission:"+p.ID))
		envelopePath := ".claude/evidence/submissions/" + p.ID + "/" + digest + ".json"
		envelope, err := workpacket.DesignEvidence(p, id, arch, answer, time.Now().UTC())
		if err != nil {
			return Receipt{}, err
		}
		encoded, err := json.MarshalIndent(envelope, "", "  ")
		if err != nil {
			return Receipt{}, err
		}
		if _, err := workpacket.Publish(s.Root, envelopePath, append(encoded, '\n')); err != nil {
			return Receipt{}, err
		}
		if err := ctx.Err(); err != nil {
			return rejected(snap, p.ID, err), err
		}
		next, err = runtime.RecordEvidence(s.Root, s.StatePath, s.JournalPath, runtime.EvidenceRequest{ExpectedRevision: -1, ID: id, Kind: "planning_design", Path: envelopePath, ProducedBy: []string{"Main"}, ResponsibilityID: "Orchestrator", Validator: s.Validator, Submission: &runtime.SubmissionRequest{PacketID: p.ID, Owner: "Main", AnswerJSON: data}})
		if err != nil {
			return rejected(snap, p.ID, err), err
		}
		prior, accepted, err = runtime.FindSubmission(next.State, p.ID)
		if err != nil || !accepted {
			return Receipt{}, fmt.Errorf("design reception has no durable record")
		}
	}
	if (prior.ControlStatus != "pending" && prior.ControlStatus != "no_transition") || runtime.ValidateDesignPacket(s.Root, next.State, p, nil) != nil {
		return s.designReceipt(next, p)
	}
	result, controlErr := controller.RunControlCycle(ctx, controller.ControlRequest{Root: s.Root, CWD: s.Root, StatePath: s.StatePath, JournalPath: s.JournalPath, Event: "PostToolUse", ToolName: "Write", SubmissionFollowupID: prior.FollowupID, Files: s.Files})
	preparation := ""
	var preparationWaiting []Waiting
	if controlErr == nil && result.Error == "" {
		preparation, preparationWaiting = s.preparePlanningOutlet(ctx, &result)
	}
	if controlErr == nil && result.Error == "" {
		next = result.Snapshot
	}
	r, err := s.designReceipt(next, p)
	if err != nil {
		return Receipt{}, err
	}
	if controlErr != nil || result.Error != "" {
		r.NextAction = "Design was recorded; the original stage outlet remains pending. Save this same answer to continue; do not repeat the design report."
	}
	if len(result.QualityGate.Missing) > 0 || len(result.QualityGate.Conflicts) > 0 {
		r.NextAction = "Design report was recorded. Complete the actual current design/material gaps before progressing: " + strings.Join(append(result.QualityGate.Missing, result.QualityGate.Conflicts...), "; ") + ". Keep the same answer and its recorded observations; save it to continue the original outlet when ready."
	}
	r.NextAction += preparation
	r.WaitingOn = append(r.WaitingOn, preparationWaiting...)
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
	r.SubmissionKey = record.Key
	r.Accepted = true
	r.DomainRefs = []string{record.EvidenceID}
	r.ResponsibilityStatus = "accepted"
	r.WorkComplete = record.ControlStatus != "pending" && record.ControlStatus != "no_transition"
	r.NextAction = "Design report was recorded; follow the current stage's actual professional work. Acceptance is not independent certification of every design task."
	if (record.ControlStatus == "pending" || record.ControlStatus == "no_transition") && runtime.ValidateDesignPacket(s.Root, snap.State, p, nil) != nil {
		r.ResponsibilityStatus = "replaced"
		r.NextAction = "This design reception remains recorded, but the current stage responsibility has changed. Follow its current questionnaire; do not replay the old stage outlet."
		return r, nil
	}
	if record.ControlStatus == "pending" || record.ControlStatus == "no_transition" {
		r.NextAction = "Design report was recorded; the original stage outlet is pending. Save this same answer to continue the original outlet without another report."
		r.WaitingOn = append(r.WaitingOn, Waiting{Reason: "system_followup", Owner: "Main", Ref: record.FollowupID})
	}
	for _, ev := range evidenceRows(snap.State) {
		if ev["id"] == record.EvidenceID {
			materialPath := filepath.Join(s.Root, text(ev["path"]))
			data, readErr := os.ReadFile(materialPath)
			var material struct {
				Subjects []workpacket.Subject `json:"subject_refs"`
			}
			if readErr == nil {
				readErr = json.Unmarshal(data, &material)
			}
			if readErr != nil {
				r.NextAction = "The accepted design material is unavailable at " + materialPath + ": " + readErr.Error() + ". Restore the recorded material; do not submit a replacement proof."
				return r, nil
			}
			// Report missing actual reading material without re-certifying its
			// historical bytes, subject count, version or producer.
			for _, subject := range material.Subjects {
				if _, err := workpacket.ReadFile(s.Root, subject.Path, 2<<20); err != nil {
					r.NextAction = "Accepted design reading material is unavailable at " + filepath.Join(s.Root, subject.Path) + ": " + err.Error() + ". Main must restore that material; the original reception remains recorded."
					break
				}
			}
			return r, nil
		}
	}
	r.NextAction = "The accepted design record has no material reference; Main must restore that recorded material without repeating the professional answer."
	return r, nil
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

// preparePlanningOutlet completes only the next responsibility of an actual
// committed planning transition. Old-answer replay never calls this outlet.
func (s *Service) preparePlanningOutlet(ctx context.Context, result *controller.ControlResult) (string, []Waiting) {
	s = s.WithContext(ctx)
	if !result.QualityGate.TransitionCommitted || !runtime.PacketEnabled(result.Snapshot.State) {
		return "", nil
	}
	id := result.QualityGate.CandidateTransition
	if id != "PTR-PLAN-01" && id != "PTR-PLAN-02" {
		return "", nil
	}
	preparationFailed := func(err error) (string, []Waiting) {
		return " The preceding report and stage transition are recorded; the original next-stage questionnaire preparation failed: " + err.Error() + ". Main/system must continue the existing questionnaire preparation through the original listener after resolving this actual error. Do not resubmit the completed report, invent a questionnaire path, or fabricate inapplicable materials.", []Waiting{{Reason: "system_followup", Owner: "Main", Ref: id}}
	}
	service := *s
	if service.Files == nil {
		catalog, err := transition.LoadCatalog(s.Root)
		if err != nil {
			return preparationFailed(err)
		}
		ref, err := fileview.DevelopmentRef(result.Snapshot.State)
		if err != nil {
			return preparationFailed(err)
		}
		service.Files, err = fileview.New(s.Root, "refs/heads/"+strings.TrimPrefix(ref, "refs/heads/"), catalog.Definition.FileSources)
		if err != nil {
			return preparationFailed(err)
		}
	}
	var view, answer string
	var err error
	if id == "PTR-PLAN-01" {
		var packet workpacket.ContractPacket
		packet, err = service.IssueContracts(ctx)
		view, answer = packet.ViewPath, packet.AnswerPath
	} else {
		var packet workpacket.TaskPlanningPacket
		packet, err = service.IssueTaskPlanning(ctx)
		view, answer = packet.ViewPath, packet.AnswerPath
	}
	if err != nil {
		return preparationFailed(err)
	}
	current, err := s.snapshot()
	if err == nil {
		result.Snapshot = current
	}
	message := fmt.Sprintf(" Current next-stage questionnaire: read %s; answer %s. Fill its professional answers and save normally.", filepath.Join(s.Root, view), filepath.Join(s.Root, answer))
	if err != nil {
		message += " Preparation recorded, but current Runtime could not be read: " + err.Error()
	}
	return message, nil
}
