# Loop Wake-up Prompt (project `.claude/loop.md`)

> Template source. The release tarball installs this file at `.claude/loop.md`
> in the target project. It is the default Prompt that Claude Code's built-in
> `/loop` scheduler delivers on each wake-up.
>
> Source: `docs/control/agent-protocol.md`, `docs/control/loop-definition.json`, and
> `AGENTS-template.md`.

## What this file is — and is not

| Is | Is not |
|:---|:---|
| the Layer 2 Wake-up Prompt that recovers the Main-session Driver | the Engineering Loop itself |
| a short router back into `AGENTS.md`, the Runtime Bookmark, and the current stage of `agent-protocol.md` | a second copy of the Main Spine |
| consumed by Claude Code's built-in `/loop` scheduler | a REQ binder, a state machine, or a `/goal` replacement |

Hook events route actual saves and native returns to their original consumers.
Read-only planning observations show the current responsibility without
re-running its outlet. `/loop` is a fallback scheduler: it resumes the current
REQ responsibility and does not bind a REQ, grant approval or create another plan.

## Wake-up algorithm

If the latest Hook context already identifies the current responsibility,
accurate output path and next action, continue that work directly. A wake-up
alone is not a reason to re-read the library, reconcile Runtime or copy a REQ
hash. After a new session, compaction or contradictory context, recover as follows:

1. Start with the latest Hook responsibility. If it is unavailable, read the
   Runtime milestone and the relevant entry in `AGENTS.md`.
2. Follow the current stage anchor in `docs/control/agent-protocol.md` and its
   exact inputs/output references. Read only the direct upstream materials and
   primary Skill needed for the action. Do not reconstruct paths or machine
   fields from memory.
3. Continue the missing professional work through `DRIVE()` in
   `loop-orchestration`. Existing accepted facts remain valid unless the
   original consumer reports otherwise. Investigate source drift only when
   actual work or a diagnostic calls for it.
4. Follow the original completion boundary: Main saves its current answer;
   native Workers return original references for SubagentStop reception.
   Distinguish reception, stage transition and next-material preparation.
   A successful receipt with unfinished followup needs the named continuation,
   not a new answer or an unrelated tool call to trigger progress.
5. Continue across task completion, native return and recoverable failures.
   Stop at an actual human decision, explicit user stop, or a real external
   wait. Several valid implementations are not a reason to ask for a choice.

## Responsibility and recovery

- Reuse original responsibility and accepted results; do not duplicate TASKs,
  assignments or evidence after a wake-up.
- Pick a low-risk reversible implementation within the authorized requirement.
  An applicable existing human decision needs no second approval.
- A completed TASK or local repair does not complete the REQ. The stage contract
  and original consumers determine remaining review, integration and acceptance.
- Native work follows its original brief or questionnaire. Only an already
  registered frozen execution retains its named readback/approval/activation
  recovery contract; do not copy those steps into new native work.
- A Hook `warn`, `block` or missing item is not itself a human gateway. Follow
  its concrete cause and recovery. Ordinary tools remain available for repair;
  escalate only when the original domain actually requires a human decision.
- Normal work needs no `status`/`next`/`reconcile` ritual. Read-only diagnostics
  and supported original recovery commands remain available for a concrete
  technical gap. Do not refresh accepted sources or hand-push a transition.
- If Runtime is inactive with no bound REQ, follow `AGENTS.md`'s S0/S1 entry
  using actual authorization. This wake-up supplies no authorization of its own.

## Reference

- Main Spine: `docs/control/agent-protocol.md`
- Loop Definition: `docs/control/loop-definition.json`
- Project Entry: `AGENTS-template.md`


## 临时 worktree 与阶段交付

S1 原绑定答卷记录实际开发分支和最终发布上游；不猜默认分支。显式 `req bind` 保留为兼容与恢复入口，不是正常问卷之外的另一项手续。项目根目录保存当前 REQ 权威事实。

S2 在最终答卷前将正式上游设计提交到绑定开发分支，使 S3 能读取真实已提交来源。其他阶段同样在实际来源合同要求时提交相关产物；不要自动提交用户无关变更。新 worktree 不包含主会话未提交或仅暂存的文件。

新原生 S6/S9 写入责任由 Main 携带准确 brief 通过 Claude Code Agent 派发，并使用原生 worktree 隔离。子会话提交实际成果并返回原引用，由原消费者接收；Main 另行完成集成和检查。已有冻结执行保留其具名 `runtime worktree-create`、原登记及集成恢复流程，详见 `docs/workspace-execution-recovery.md`，不推广为新责任的启动前提。

正式文件、私人答卷、运行证据和 Runtime 各有来源合同，不能统一要求全部对象先入 Git。固定 Git 来源只接受已提交产物，明确磁盘来源读实际文件，Runtime 读权威快照。集成不等于 release；临时 worktree 不等于交付完成。清理提醒不新增 Stop 或普通工具硬门禁。
