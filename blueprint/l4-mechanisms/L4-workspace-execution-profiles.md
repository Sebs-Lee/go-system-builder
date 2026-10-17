# L4 Registered workspace execution and bounded repair authority

Version: 1.0 (2026-09-21). Design decision implemented; real Claude workflow
acceptance remains a separate release condition. This file is framework design,
not a target-project runtime authority or an installation asset.

## Evidence, alternatives and decision

PR #8 review 5265399074 identified two reproduced failures on a68950b:
relocating Main passed PlanRebind but failed the real CLI Writer authority check;
ready Workers could not run pwd, git diff --stat or git log. Earlier delivery
failures required preserving source/merge identity and distinguishing retry from
code rework. These justify the following limited mechanisms, not unrestricted
new lifecycle states. The review and machine outputs belong in PR/CI or ignored
maintenance evidence, not target-project reports.

| Decision | Existing mechanism insufficient because | Minimal response | Cost and rejected alternative |
| --- | --- | --- | --- |
| Relocation transaction | A normal Writer correctly rejects a different root, including recovery at the moved root | Exact source snapshot, path-only delta, existing commit marker/journal, idempotent pointers | One durable relocation intent; reject generic skip-authority, hand edits and a second Runtime |
| Worker observation | File tools do not expose Git history/diffs; blanket Bash denial interrupts diagnosis | Fixed observe views, argv execution and safe Git controls; exact pwd | No new approval or revision; arbitrary shell parser and git-prefix allowlists are rejected |
| Runner preflight | Linux isolation/dependency failures previously appeared after dispatch | Probe before publishing execution/launch; explicit check_location main alternative | Linux dependencies, offline checks and 512 MiB remain visible costs; never silently run an unsandboxed Worker check |
| Check timing | Worker and merge-tree checks may have different inputs and costs | Paired checks remain default; reviewed post_merge exception | Saves a Worker run but discovers failures after merge; no general performance gain is claimed without measurement |
| Bounded repair grant | Repeating owner approval for unchanged, module-local technical repairs interrupts work | Optional scoped, fingerprinted, finite grant plus review per exact contract | Grant/usage evidence and explicit revocation semantics; no new release or REQ authority |

The owner requested preserving Main and scoped Workers while reducing repeated
interruptions. This permits implementing the options; it does not establish that
all projects should select them. Registered Worker execution is an explicit
prepare/launch choice, not a global replacement of Claude's native subagent mode.
Existing native flows retain their own controls. Unmeasured benefits remain
hypotheses to evaluate in real projects, not release claims.

## Sources, consumers and failure recovery

| Fact | Single source | Consumer | Invalidated by / recovery |
| --- | --- | --- | --- |
| REQ root/dev/release binding | bound_req.workspace | Writer, registry validation, integration | explicit REQ change or narrow relocation; never checkout to guess authority |
| Execution owner, scope, generation, base, check_location | Runtime workspace registry | prepare, launch, adapters, policy | revoke/replace; preserve old generation in history |
| Frozen input hashes | execution inputs | bootstrap/check/commit/delivery | changed input blocks; replan, never silently refresh hashes |
| Bootstrap files/digest | immutable bootstrap receipt + registry digest | launch/adapters | changed installed assets; preserve and explicitly replace |
| Claude UUID | execution platform_session_id | launcher/Hook identity mapping | resume same UUID or explicit replacement; exit0 does not complete task |
| Launch attempts/lease | Main launch records and process-owned lock | launcher/replacement/rebind | process exit releases lease; receipt retained for diagnosis |
| Worker check result | tree-hashed check receipt | operator/Worker report | changed tree requires rerun; no substitute for Main verified checkpoint |
| Integration stage and tested HEAD | existing checkpoint | Main integrator, ack, successor scheduling, cleanup | failure preserved at exact stage; retry/rework, not new parallel state machine |
| Rework evidence | archived checkpoint/candidate + Runtime ref | rework/resubmission | new candidate must be verified; no Git rollback |
| Relocation identity | immutable source/request/before/after intent | narrow Runtime transaction and pointer updater | CAS drift, wrong Git identity or active checkpoint blocks; same request resumes |
| Repair authority | human grant or REQ-pinned project policy hash | contract approval | scope/REQ/generation/expiry/revocation/hash/budget; return to normal human approval |
| Grant usage and exact technical review | Runtime configuration.repair + registered review SHA | approval transaction | atomic budget consumption; repeat approved contract is idempotent |

No observation creates completion evidence or increments Runtime revision.
Receipts referenced by active execution/checkpoints must remain available; archive
with the corresponding Runtime after terminal handoff. Never delete evidence to
make a failed check appear fresh. Platform launch/check logs are diagnostic;
Runtime/journal and immutable evidence remain authority.

## Worker capability contract

### Fresh worktree inheritance during bootstrap (2026-10-15)

A normal installed project may commit its launcher and project settings. Its
new Git worktree inherits those files before the Worker bootstrap installs its
registered native executable and minimal Hook settings. Those inherited bytes
are not evidence of Worker edits. During an execution's original `preparing`
bootstrap only, with no published bootstrap receipt/digest, the installer may
replace the inherited launcher, settings, or explicitly frozen generated-role
target when the existing regular file is byte-identical to that path in the
execution's exact `base_commit` and has no staged/unstaged modification. Use raw
Git blob bytes, not whitespace-trimmed command output. Source proof unavailable,
untracked contents, symlinks, changed bytes or an already-published receipt keep
the original preserve-and-refuse behavior. Never delete inherited files merely
to evade that check.

The installed Worker bytes remain the existing generated/bootstrap assets and
are hashed into the original receipt, published last. Retry accepts already
correct bytes but never restores a base copy over a changed bootstrapped asset.
No Main permissions, Runtime, journal or credentials are copied by this rule;
Main and Worker identity, scope, frozen inputs and original PLAN remain intact.
Tests must include a Git-tracked installed launcher/settings, the equivalent
changed/untracked/symlink cases, and a retry after receipt publication.

### S5 document review execution

The opt-in v1.2 packet path also uses the original prepare/launch registry for a read-only document reviewer. `purpose=document_review`, `work_packet_id` and `stage_entry` bind the current S5 dispatch to one independent root and execution generation. No product write scope, product commit/delivery or executable check is granted. The receipt-covered `.claude/inputs/work-packet.json` contains the immutable packet and execution view; the view's exact `authoring_path` names the Worker-local editable answer. Original plan/activation and domain gates still apply. Preparing this execution revokes shared Main authoring; failed preparation never falls back to Main. Receipt or original-stage retirement ends the worker's responsibility; a process exit alone does not. See [stage form §13](L4-stage-form.md) for reception and recovery authority.

### Host-protected paths and editable Worker drafts (2026-10-15)

Claude Code's `dontAsk` host policy denies writes under protected `.claude/` paths, even when PreToolUse returns an exact allow. Do not change the global host mode, ask the Worker to use Bash for file writes, or weaken the application Hook. New execution-local authoring drafts use only the exact files `.loop-submissions/plan.json`, `.loop-submissions/stage-answer.json`, `.loop-submissions/completion.json`, `.loop-submissions/commit.json`, and `.loop-submissions/repair-result.json`; issued Main packets, Runtime state, journals, receipts, evidence and frozen inputs remain under their existing `.claude/` authority paths.

The original receipt-bound frozen assignment context/view records each exact authoring path and its digest. New contexts name the `.loop-submissions/` files. An old frozen context with no `plan_path`, or an old packet view naming `.claude/submissions/stage-answer.json`, keeps that recorded legacy path; recovery must not inspect the workspace to guess a newer path or fall back based on which file exists. Readers and save consumers accept only the exact legacy or new fixed path authorized by the same original context/view. This is a transport-path change only: it adds no product write scope, Runtime field, registry, packet, permission bypass, or change to Main's canonical authoring/evidence paths.

`runtime workspace observe --view cwd|status|diff|staged|log` is a fixed capability.
It resolves the registered Worker, validates owner/generation/bootstrap/inputs,
uses fixed Git argv, disables external diff/textconv, pager, fsmonitor and lazy
fetch, strips Git environment overrides, and bounds output/time. It does not
claim to sandbox arbitrary programs. File tools remain subject to normal scope.
Mutating commands use commit/check/report/deliver adapters; human-only actions
remain human-only. Unsupported raw Bash gets executable observation commands.

prepare --check-location worker (compatible default) probes actual isolation and
snapshot size; launch repeats the probe before reserving the Claude UUID. Main
mode is chosen explicitly at preparation, persisted on the execution and cannot
be toggled in flight. It removes Worker check commands but does not waive Main
required_checks or independent verification. An existing execution must be
replaced under existing guards to change runners. Network-dependent checks need
preinstalled offline dependencies or an explicitly selected Main runner. Native
Windows/macOS do not gain this project's Linux isolation because Claude offers
its own sandbox there; the two mechanisms are distinct.

## Check timing and convergence

Ordinary assignments retain Worker required_checks, normal merge, then Main
required_checks. post_merge requires an explicit reviewed dispatch plan stating
why a Worker run is unsuitable or its measured duplicate cost warrants deferral.
The immutable manifest selects the mode; no edit of active manifests. Static
identity/scope/locked-input/conflict checks stay in both modes. No result from a
Worker tree certifies the merged tree.

After a post-merge failure Main already contains the candidate. Preserve its
merge and source/target/check receipts. Environment failure retries the same
preserved stage; code failure reopens the registered owner through rework and
requires a new candidate/check result. No automatic reset/stash, ack, cleanup or
successor release before verified. Main is an integration workspace, not an
implicit deployment target. Work that cannot tolerate a failed merge in Main
must retain paired checks (which also cannot guarantee merge-tree success).

## Authorization and independence

L3 S8 defines one-shot human approval, per-REQ grant and REQ-pinned policy as
separate authority sources. No delegation means the prior human path. Grants do
not waive S2 design authority, REQ amendments, destructive-action approval,
review budgets, independent S9 verification, fresh S7 or S11. Technical reviewers
make semantic judgments; booleans are not machine proofs. Technical approval is
not an implementation acceptance report, and its reviewer is not automatically
the final verifier. S9 explicitly rejects a verifier who owns the repair.
Revocation prevents new approvals; stop existing approved work through explicit
pause/revoke, retaining history. Project policy has finite contract budget and
REQ/generation/hash lifetime even when no wall-clock expiry is selected.

## Acceptance and evolution

Required automation: real CLI relocation success/repeat and commit-window
recovery; conflicting old/new authority; observation outputs and helper/shell
injection rejection; no Main/Worker Git mutations; capability refusal before
launch; post-merge failure preservation; grant budget/revocation/identity tests;
full tests, race checks, release graph and installation checks.

Required real-platform acceptance: observe, edit, declared dependency/check path,
report, Main merge, failed-check rework, session interruption/resume and independent
verification using actual Claude Hooks. Record CLI version, platform, runner,
source/merge SHA, failures and next actions. CI/cross-build/native Hook smoke do
not replace it. Missing login/platform capability is NOT EXECUTED. Collect actual
costs and interruption counts before promoting opt-in modes to defaults.

Change record: 2026-09-21 — PR #8 reproduced R1/R2 plus L3/L4 conflict prompted
this minimal recovery/observation decision and restored paired-check guidance.
Revisit through L1's evidence→proposal→cost/consumer→versioned change protocol.

## Check profile, bounded reuse and environment contract (2026-10-13)

Global G07 precedes G05: profiles declare approved command, runner/location, timeout, offline dependencies, fixtures, environment/input closure and side effects. Main also probes required tools/configuration; selecting Main does not prove readiness. Canonicalize existing checkout paths while preserving repository/branch identity; aliases do not authorize relocation or another linked checkout.

Receipt v2 binds actual subject, profile, runner, environment/fixtures, responsibility and integrity to original domain records. Reuse defaults to never; only confirmed deterministic checks with identical complete closure and the same responsibility may reference an existing accepted PASS. Unknown closure, drift, interruption, a new hypothesis, independent review, post-integration and fresh requirements rerun. Do not mint a new PASS, change ready Worker assets, weaken isolation or bypass Main integration/fresh S7.

Typed environment/process/timeout/resource/IO failures carry concrete owner, log and retry condition. Exit failure alone is not a product assertion failure. G05/G07 acceptance includes concurrency, recovery, alias/relocation, real Linux isolation and explicit unsupported-platform/Main paths.

## S5 named revision execution contract (2026-10-13)

A received review can be revised only by Main against its exact submission key and reason in the unchanged S5 origin. Issuance suspends old evidence qualification atomically with the new immutable packet. Main holds the original Git scheduling lock and old launch lease; a running child prevents retirement. Retain the old execution as retired history without changing its files. The successor requires an independent source until a fresh higher execution generation is prepared and bootstrapped. Do not reuse the old root or fall back to Main authoring. Original plan/activation authority remains bounded to the unchanged responsibility and frozen inputs; the new answer is a new professional review. See L4-stage-form §13.3 for reception, deduplication and receipt semantics.

## 检查入口预检与真实故障范围

原 Main 模式不再直接把环境当 ready：prepare/launch 在发布/会话预留前核验
实际 shell 与简单字面检查命令的工具入口。复杂 shell 保留 legacy，明确
unverified；入口可用不证明依赖、配置、fixture 或服务就绪。旧命令没有完整
CheckProfile/environment 闭包，不能据此启用缓存或省略必需合并检查。

原 ObservationCommands 提供固定 `check-capabilities` 只读入口，同样校验
owner/assignment/路径/Bootstrap。Worker 预检沿原 Bubblewrap/seccomp 路径，
不可用时给原环境责任动作，不切换 ready runner 或放宽隔离。真实失败回执
附 code/category/owner/恢复条件/receipt/log，process_exit 不等于产品断言
失败；取消/超时/诊断 I/O 保留原候选和 checkpoint。具体契约见 G07 §7。

## 显式 CheckProfile 的冻结身份

新 manifest Assignment 可选 profile，与原 required_checks 的索引/命令严格
一致；新执行固定 location、1–300秒timeout、环境政策和来源摘要。profile
不会创建第二检查列表或改变激活/写入权限。旧 ready execution 不补字段，
原 replace/adopt/rebind 持同一冻结 profile。

实际执行捕获完整受测树/runner/白名单环境/依赖/fixture观测与原责任绑定；
内容变化、取消/超时、诊断失败不能通过。v2日志摘要/完整性用于恢复/核验，
不是批准或可信缓存凭据。当前复用不启用，Main必需检查和独立fresh S7仍运行。

## 工具事实登记与 shadow 适用性

原 runner 原子发布后将实际生成字节摘要交给进程内接收器，Worker 由原
Writer/CAS 复核 current owner/execution/profile 后登记，Main 由原 integration
checkpoint CAS 保留原工具事实。checkpoint 恢复核验 Worker 登记与原 journal
的关系，不从回执文件补接收。每 Execution 至多64引用，过期引用只导致重跑。

`check-reuse` 属原固定 ObservationCommands/Hook 消费契约，只读输出 run 和
原因。首版 shadow 不省略检查；未知权限/隔离条件不能启用。原必需 Main、
独立责任、新假设和 fresh 优先要求重跑。原 rebind 可移动 evidence 坐标，保留
原 bytes/digest/checkout，不能由搬迁刷新成新检查或获得跨 checkout 复用资格。

新 permissions 事实与 digest 来自原实际隔离预检，包含 Bubblewrap、shell
及 loader/shared libraries 字节摘要、原 seccomp/argv policy、kernel/identity、
Main/common/Worker/cwd 坐标和原 scope/inputs/Bootstrap authority。观察与执行
共享原参数和 filter 生产者。旧回执缺字段、摘要不一致、不可用及事实变化均
不能命中；不接受调用者 PermissionKnown 布尔值代替来源核验。

whole_tree 不能闭合可读 host root 的 arbitrary shell 输入，首版 shadow 仅
接受受测树内 `true/false/test -f/test -s` 静态语法。只读观察与实际检查采用
同一临时副本算法和归一化模式；继承环境与未证明的命令闭包继续重跑。

原 profile 可选 reuse_mode=shadow/enabled，缺省 shadow；enabled 只允许原
Worker 的闭合静态子集和完整 declared_only/none/whole_tree 政策。模式属于
原冻结 digest，旧 shadow 回执不升级。实际 RunCheck 在校验当前副本和权限后
返回原已接收引用；不新造 PASS。Main 必需检查、独立职责及 S9/fresh 保持。

同责任/profile/源树进程锁覆盖原执行及 Writer 接收，原 Main integration 锁
仍保留。竞争请求返回有界 run 引用，取得锁后重读原 Runtime/owner/execution/
S9 authority。run.json 只有 waiting/running/finished/interrupted 诊断，不作为
接收或存活证明；check-running 只读至多64项。进程死亡后保留 interrupted
记录，再核验原工具回执或重跑，不能从诊断推断 PASS 或给新权限。
