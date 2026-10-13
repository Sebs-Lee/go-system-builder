# G05 — 验证回执、适用性与受限复用的落地设计

> 全局设计；承接 [G05 guideline](../GLOBAL-OPTIMIZATION-GUIDELINES.md#g05--验证证据可追踪可判断是否复用)。新 profile、回执版本和复用决策为拟实现。

## 1. 调查证据与结论

[workspace.RunCheck](../../internal/workspace/check.go) 已在隔离副本运行声明命令，回执有 snapshot SHA、命令、执行代次、耗时和日志。[CommandCheckRunner](../../internal/integration/check_runner.go) 有 before/after HEAD、exit、failure_kind及少量允许环境设置；[Integrate](../../internal/integration/integrate.go) 对 Result变更/确认前目标漂移重验，ack后纯清理保留原收据。

[verifiedDomainChecks](../../internal/cli/workspace_delivery_checks.go) 对 S9结果检查 Main实际回执与 tested_head。[review.SubmitResult](../../internal/review/submit.go) 保留独立性、完整Claim集合及当前冻结输入。

已有凭据适合做明确的适用性判断，但不足以对任意命令全面缓存：依赖环境、动态输入和副作用未完全描述。第一版仅复用批准为确定性的同职责工具检查，独立审查和必需合并后检查仍重新执行。

## 2. 检查 profile 与回执 v2

复用 G07的 `CheckProfile`：固定命令、runner/location、timeout、相关输入闭包、声明环境/依赖、副作用类别、reuse_policy。默认 `never`；维护者确认无外部/时间/随机/持久副作用且闭包完整时才可设 `same_input_same_responsibility`。

新增回执 v2保留旧字段，增加 check_profile_id/version/digest、actual subject digest、runner fingerprint、environment/fixture fingerprint、role/responsibility、日志摘要、输入验证范围和 receipt integrity。环境只记录白名单事实/摘要，不转储所有环境变量。

回执为工具捕获，存于原 evidence目录；内容摘要经原领域接收记录绑定。只有文件存在不能证明真实性，authoring answer只能引用，不能自己填写 pass代替实际执行。

## 3. 适用性函数与缓存范围

在 `internal/workspace` 拟新增 `check_reuse.go`：`DecideReuse(profile, currentInputs, owner, existingReceipts) -> {run|reuse, reason, receiptRef}`。在RunCheck锁定/验证当前输入后执行，找到精确一致且已接收的pass回执才复用；新接口默认不执行跨 owner复用。

键包括当前 Runtime/generation/responsibility、profile digest、实际受测内容闭包、environment/fixtures、runner版本和需要的权限条件。第一版无可靠子集闭包时用整个受测树；与全仓 Git HEAD无关的额外提交不能自动失效，相关文件内容变化必须失效。

复用后不新造检查PASS，只返回原回执引用并记录 reuse诊断。index可由回执重建，缓存缺失则跑原检查，不能影响门可用性或授权。

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
