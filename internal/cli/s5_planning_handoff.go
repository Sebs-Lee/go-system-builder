package cli

import (
	"context"
	"path/filepath"
	"time"

	"github.com/entroforge/go-system-builder/internal/fileview"
	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/submission"
)

// Continue the real S4 completion into the original S5 preparation, without
// asking Main to assemble a manifest or save an unchanged confirmation. A
// preparation failure does not revoke the durable S4 answer.
func prepareS5PlanningHandoff(ctx context.Context, root string, receipt *submission.Receipt) {
	if !receipt.Accepted || receipt.CurrentStage != "S5" {
		return
	}
	store := runtime.NewStore(filepath.Join(root, ".claude/loop-state.json"), filepath.Join(root, ".claude/loop-events.jsonl")).WithContext(ctx)
	snapshot, err := store.Snapshot()
	if err != nil {
		receipt.ConsumerOwner, receipt.ConsumerPending = "Main", true
		receipt.NextAction = "S4 已接收，最后已知阶段为 S5；无法读取最终 Runtime，不能确认 S5 准备状态：" + err.Error() + "。Main 使用 " + submission.PlanningRecoveryCommand(root, "issue") + " 恢复原准备，保留已有问卷和答案。"
		return
	}
	if _, found, lookupErr := runtime.FindTaskPlanningPacket(snapshot.State, receipt.PacketID); lookupErr != nil || !found {
		return
	}
	if err == nil && s5ManifestQuestionnairePath(root, snapshot.State) == "" {
		return // The original cursor already moved beyond this handoff.
	}
	if err == nil {
		err = prepareS5ReviewResponsibilities(ctx, root, snapshot.State)
	}
	if err != nil {
		// Preparation may have issued the first responsibility before failing.
		// Read only the durable publication facts with a small display budget;
		// never retry the source validation or Writer after its deadline.
		observeCtx, observeCancel := context.WithTimeout(context.WithoutCancel(ctx), 500*time.Millisecond)
		latest, observeErr := store.WithContext(observeCtx).Snapshot()
		observeCancel()
		published := "最后已知准备入口尚未完成；准确发行进度需沿原 issue 核实。"
		if observeErr == nil {
			snapshot = latest
			published = stageS5ObservedContext(root, snapshot.State)
		}
		receipt.ConsumerOwner, receipt.ConsumerPending = "Main", true
		manifest := s5ManifestQuestionnairePath(root, snapshot.State)
		receipt.WaitingOn = append(receipt.WaitingOn, submission.Waiting{Owner: "Main", Reason: "S5 原分工与题目准备待恢复：" + err.Error(), Ref: manifest})
		receipt.NextAction = "S4 专业交付已接收。Main／系统恢复 S5 原准备：" + err.Error() + "；恢复入口：" + submission.PlanningRecoveryCommand(root, "issue") + "；保留已发行问卷和作者答案，不重新填写已接收的 S4 答案。\n" + published
		return
	}
	snapshot, err = store.Snapshot()
	if err != nil {
		receipt.NextAction = "S4 已接收，S5 原责任准备已执行；Main 恢复当前 Runtime 读取以取得准确原卷：" + err.Error() + "；恢复入口：" + submission.PlanningRecoveryCommand(root, "issue")
		return
	}
	receipt.NextAction = "S4 已接收，TR-002 已提交；S5 原责任已准备。Main 按独立原生上下文派发；方法入口 " + filepath.Join(root, ".claude/skills/agent-dispatch/SKILL.md") + "。\n" + stageS5ObservedContext(root, snapshot.State)
}

// Give the successor its own bounded budget after the reception transaction.
// The parent cancellation still applies; a consumed reception deadline does not.
func continueS5PlanningHandoff(parent context.Context, root string, receipt *submission.Receipt) {
	ctx, cancel := context.WithTimeout(fileview.WithReadCache(parent), 12*time.Second)
	defer cancel()
	prepareS5PlanningHandoff(ctx, root, receipt)
}
