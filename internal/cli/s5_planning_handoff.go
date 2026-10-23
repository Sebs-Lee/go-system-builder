package cli

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

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
	if err == nil && s5ManifestQuestionnairePath(root, snapshot.State) == "" {
		return // The original cursor already moved beyond this handoff.
	}
	if err == nil {
		err = prepareS5ReviewResponsibilities(ctx, root, snapshot.State)
	}
	if err != nil {
		receipt.ConsumerOwner, receipt.ConsumerPending = "Main", true
		manifest := s5ManifestQuestionnairePath(root, snapshot.State)
		receipt.WaitingOn = append(receipt.WaitingOn, submission.Waiting{Owner: "Main", Reason: "S5 原分工与题目准备待恢复：" + err.Error(), Ref: manifest})
		receipt.NextAction = "S4 专业交付已接收。Main／系统恢复 S5 原准备：" + err.Error() + "；不重新填写已接收的 S4 答案，不手工补 SHA 或登记元数据。"
		return
	}
	snapshot, err = store.Snapshot()
	if err != nil {
		receipt.NextAction = "S4 已接收，S5 原责任准备已执行；Main 恢复当前 Runtime 读取以取得准确原卷：" + err.Error()
		return
	}
	var next strings.Builder
	next.WriteString("S4 已接收，S5 两项独立责任已由原入口准备；Main 真实派发原生 Agent，不填机器元数据或重复确认。")
	entities := mapFieldCLI(snapshot.State, "entities")
	for _, raw := range anySlice(entities["work_packets"]) {
		row, _ := raw.(map[string]any)
		if row["template_id"] != "s5-document-review-v1" {
			continue
		}
		binding := mapFieldCLI(row, "binding")
		if binding["runtime_id"] != snapshot.State["runtime_id"] ||
			integerValue(binding["generation"]) != integerValue(mapFieldCLI(snapshot.State, "baseline")["generation"]) ||
			integerValue(binding["stage_entry"]) != integerValue(mapFieldCLI(snapshot.State, "lifecycle")["phase_revision"]) {
			continue
		}
		fmt.Fprintf(&next, "\n%s/%s: view %s; private answer %s。", stringValue(binding["owner"]), stringValue(binding["responsibility"]), filepath.Join(root, stringValue(row["view_path"])), filepath.Join(root, stringValue(row["answer_path"])))
	}
	receipt.NextAction = next.String()
}
