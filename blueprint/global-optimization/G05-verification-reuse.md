# G05 — 验证回执、适用性与受限复用的落地设计

> 全局设计；承接 [G05 guideline](../GLOBAL-OPTIMIZATION-GUIDELINES.md#g05--验证证据可追踪可判断是否复用)。G07 profile/v2 已实现；以下可信登记与 shadow 适用性为当前实施切片，启用复用和并发恢复仍待验收。

## 1. 调查证据与结论

[workspace.RunCheck](../../internal/workspace/check.go) 已在隔离副本运行声明命令，回执有 snapshot SHA、命令、执行代次、耗时和日志。[CommandCheckRunner](../../internal/integration/check_runner.go) 有 before/after HEAD、exit、failure_kind及少量允许环境设置；[Integrate](../../internal/integration/integrate.go) 对 Result变更/确认前目标漂移重验，ack后纯清理保留原收据。

[verifiedDomainChecks](../../internal/cli/workspace_delivery_checks.go) 对 S9结果检查 Main实际回执与 tested_head。[review.SubmitResult](../../internal/review/submit.go) 保留独立性、完整Claim集合及当前冻结输入。

已有凭据适合做明确的适用性判断，但不足以对任意命令全面缓存：依赖环境、动态输入和副作用未完全描述。第一版仅复用批准为确定性的同职责工具检查，独立审查和必需合并后检查仍重新执行。

## 2. 检查 profile 与回执 v2

复用 G07的 `CheckProfile`：固定命令、runner/location、timeout、相关输入闭包、声明环境/依赖、副作用类别、reuse_policy。默认 `never`；维护者确认无外部/时间/随机/持久副作用且闭包完整时才可设 `same_input_same_responsibility`。

新增回执 v2保留旧字段，增加 check_profile_id/version/digest、actual subject digest、runner fingerprint、environment/fixture fingerprint、role/responsibility、日志摘要、输入验证范围和 receipt integrity。环境只记录白名单事实/摘要，不转储所有环境变量。

回执为工具捕获，存于原 evidence目录；内容摘要经原领域接收记录绑定。只有文件存在不能证明真实性，authoring answer只能引用，不能自己填写 pass代替实际执行。

### 2.1 首个可信登记契约（2026-10-14）

原 runner 在原子发布完整回执后，把**内存中实际生成的字节摘要**交给本次调用的接收器；不提供导入用户回执或重新封印任意文件的命令。登记含原 binding、命令、profile digest、受测 checkout、回执路径/摘要及日志摘要。Worker 在原 Writer/CAS 中复核当前 Runtime、generation、owner、execution、profile 和已激活职责后登记到该 Execution；Main 在原 integration checkpoint CAS 中登记。登记只证明工具执行事实，不写 gate PASS、专业结论或完成态。

核验必须同时满足原登记的摘要、回执内 integrity、日志摘要、原 binding 和 profile；重算 integrity 的篡改仍不可信。新记录通过原状态/检查点恢复，旧 v1/旧 checkpoint 路径仍保留原领域兼容性，不能作为新复用来源。每个 Execution 的诊断登记有界，淘汰旧引用只导致重跑。

### 2.2 原隔离权限与首个静态检查闭包（2026-10-14）

新 Worker 观测附可选 permissions：原 runner 名称/版本、实际 platform/kernel、
uid/gid/groups、实际 Bubblewrap 与 shell 字节摘要、原 seccomp 和 argv policy
摘要、Main/common/Worker/cwd 坐标、原 scope/checks/frozen inputs/Bootstrap 的
authority 摘要及实际同一路径预检结果。permissions_digest 绑定完整事实；
未知、不可用、旧回执缺字段或事实变化均不能命中，不能以调用者布尔值证明
权限已知。采集与原执行共享同一 argv/filter 生产者，不新增隔离模式或回退。

原只读 host root 仍能读受测树以外的文件，因此 arbitrary shell、外部程序和
解释器脚本不能仅凭声明 whole_tree 获得首版复用资格。首个闭合语法仅允许
shell builtin `true`/`false`、`test -f <path>`、`test -s <path>`；path 必须是当前
受测树内的规范相对文件路径，禁止 expansion/control/glob、隐蔽控制目录和
逃逸 symlink。权限通过真实预检且 actual disposable copy 内容/模式摘要匹配
时才有 shadow 命中；Main 必需检查与 fresh/独立责任仍实际执行。

只读投影使用与真实 RunCheck 相同的 disposable-copy 输入算法，不拿 authoring
Worker 的目录模式或 Git 元数据替代实际受测输入。权限采集失败保留 unavailable
及诊断，实际原 runner 的失败照常返回，不产生 PASS。

## 3. 适用性函数与缓存范围

共享纯函数位于 `internal/checkexec/reuse.go`，workspace 用 `check_receipt.go` 投影：`DecideReuse(profile, currentInputs, owner, existingReceipts) -> {run|reuse, reason, receiptRef}`。首切片实际仅返回 run/shadow；后续在 RunCheck 锁定/验证当前输入后接入启用路径，找到精确一致且已接收的 pass 回执才复用；默认不执行跨 owner 复用。

键包括当前 Runtime/generation/responsibility、profile digest、实际受测内容闭包、environment/fixtures、runner版本和需要的权限条件。第一版无可靠子集闭包时用整个受测树；与全仓 Git HEAD无关的额外提交不能自动失效，相关文件内容变化必须失效。

复用后不新造检查PASS，只返回原回执引用并记录 reuse诊断。index可由回执重建，缓存缺失则跑原检查，不能影响门可用性或授权。

首切片输出 `action=run` 的 shadow 决策和 `would_reuse`；profile、原责任/owner/代次、同 checkout、六类实际摘要必须完全一致，且回执确为 pass。`main_required_integration`、独立责任、fresh/new-hypothesis、未知权限条件总是 run。Worker 的隔离能力/权限指纹尚未闭合时不称为可启用命中；shadow 不跳过任何真实检查。后续启用还必须完成进行中锁、进程中断和真实纯检查对照验收。

首切片的新来源核验用于新复用资格和诊断，不把旧领域回执强制迁移，也不因历史失败日志过期改写原 S9 接收条件。原显式 rebind 仅映射 evidence 路径，登记保留原 checkout 与回执/日志摘要；核验读搬迁后同目录的原日志字节，不能由新 cwd 获得 fresh 事实。

### 3.1 显式启用与运行记录（2026-10-14）

原 manifest profile 新增可选 reuse_mode=shadow/enabled，缺省 shadow；enabled
仅允许原 worker location、same_input_same_responsibility、declared_only/none/
whole_tree 及已闭合静态语法。字段属于原冻结 profile digest，不改旧 ready 或
把旧 shadow 回执升级成 enabled；新 profile 的第一次检查仍实际运行。启用
路径只有当原已登记回执与当前实际副本/权限/身份完全匹配时返回原引用，不
新发工具 PASS，不给 Main 必需检查、S9/fresh 或新假设使用。

同责任/profile/源树的请求先用原 filelock 串行；锁文件不 unlink。原 Main
integration 锁继续覆盖实际检查与原 Writer 接收。进行中记录为 Main evidence
目录下的有界诊断，只有 waiting/running/finished/interrupted，不是接收索引或
第二调度状态。竞争请求给原 run 引用及有界等待；后续取得锁后重新读取原
Runtime，而非沿旧 snapshot 消费。进程死亡后的未完成记录保存 interrupted
历史，新请求重新执行或核验原已登记回执；诊断文件存在不能成为缓存来源。

## 4. 必须重跑的边界

下列情况返回 run及理由：新/变化输入、profile或runner变化、未知环境、测试中断、旧回执缺失/损坏/未接收、失败回执、新区分性假设、跨独立责任、post integration、原门明确要求fresh。

候选树与合并树的结果不互代，作者检查不代独立判断；S9后fresh S7和全模块回归要求照旧。不同路径即使内容相同也须证明命令定位、配置和副作用不受影响，第一版默认不跨checkout复用。

当前 `post_merge` 取舍是既有批准执行模式，不由缓存自动选。G05首版不改Integrate必需检查；仅继续利用其已有清理/ack恢复复用。

## 5. 故障与恢复

多个相同检查请求先按职责/profile/input键串行化，已有真实进行中执行给running引用与有界观测。进程死亡保留未完成回执，重跑不记为缓存命中。等待锁超时是环境/资源等待，不填pass。

受测输入在命令执行中变化，回执记录漂移并拒绝复用。Main目标在首次ack前变化沿现有重验；已ack历史结果不因无关后续集成反复失效。

新回执writer须原子发布完整记录，损坏回执隔离并从真实执行恢复。legacy v1不能满足新的完整环境键，只用于既有领域判断或提示重跑。

## 6. 验收与切片

扩展 workspace/check、integration/check_phase、delivery_checks测试：同输入命中；命令/配置/fixture/runner变更；dirty树变化；HEAD移动但闭包相同；v1、伪造、截断、日志漂移；并发只运行一次；中断不通过；作者与独立责任不能复用。

保持 `TestPostMergeChecksUseCandidateOnTargetExactlyOnce`、`TestVerifiedRejectsTargetDriftEvenWhenCommandReturnsZero`、ResultSHA变更、确认/清理恢复和fresh-S7测试。

先在一个纯本地静态检查profile上shadow输出reuse reason，仍执行；核对准确性后再启用该profile的同职责复用。验证失败/超时率和遗漏缺陷不增加，才扩到其他确定性检查。不得默认对任意shell声明可缓存。
