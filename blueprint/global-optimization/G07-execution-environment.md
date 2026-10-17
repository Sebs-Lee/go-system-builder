# G07 — 执行环境、检查入口与故障恢复的落地设计

> 全局设计；承接 [G07 guideline](../GLOBAL-OPTIMIZATION-GUIDELINES.md#g07--检查环境提供稳定执行入口)。Main 入口预检、typed failure/只读能力、显式冻结 CheckProfile 和实际指纹采集已落地；G05 可信接收/受限复用及真实环境对照待实施。

## 1. 调查证据与实际运行

[PreflightChecks](../../internal/workspace/capabilities.go) 在Worker发布前检查实际隔离能力；[check_linux.go](../../internal/workspace/check_linux.go) 使用Bubblewrap/seccomp，只支持声明架构，并禁止网络/Unix socket等。非Linux [check_other.go](../../internal/workspace/check_other.go) 明确不可用。

已有 `--check-location main` 是显式选新执行的runner，不是Worker内放开任意shell。[RunCheck](../../internal/workspace/check.go) 固定5分钟、隔离副本最大512MiB；[Main runner](../../internal/integration/check_runner.go) 固定5分钟且记录真实日志/退出/HEAD。预检和launch使用相同注册checks，工具已给精确check命令。

2026-10-13 macOS抽查：三个workspace测试首次在默认临时目录失败，报REQ绑定根目录不一致；使用规范化 `TMPDIR=/private/tmp` 后三个都通过。源码中 `inspectCheckout` 对root做Canonical，`ValidateAuthority` 用Clean比较已有authority，`Bind`使用Abs；夹具直接写入临时目录字符串。此结果证明路径别名是必须覆盖的边界，不等于已经证明所有生产bind入口均失败。

## 2. 执行profile与预检

拟在原Assignment/Execution增加可选 `check_profiles`引用，内嵌 `check-profile.schema.json`：id/version、原命令、runner/location、timeout、offline dependencies、environment facts、fixtures/resource refs、input closure、reuse_policy与副作用类别。旧required_checks仍作为兼容字段；同一来源投影，不能维护两份可分歧命令。

初版只覆盖local process：执行指定的已批准命令，不执行profile中的任意安装/下载脚本。依赖准备由Main/环境责任按明确动作完成，profile描述如何验证就绪。测试数据引用版本化fixture，动态服务有资源声明和reset/cleanup证据。

`Preflight(profile, execution, host) -> CapabilityReport`返回可用runner、缺工具/依赖/资源、环境指纹和具体恢复；必要checks无可用环境时不发布ready。读取配置、探测工具版本和无副作用能力检查有超时，不跑长业务suite作为预检。

Main模式也校验实际工具和配置；必须支持命令运行的环境才能称可开工。初始 PreflightChecks 对 main 直接返回的缺口已通过 §7 的入口预检修正；完整工具配置/依赖/profile 仍待接续，不因选择Main就声称环境已完整验证。

## 3. 平台与路径治理

所有新bind/prepare的路径在实际存在的checkout上Canonical并核对Git common directory、bound dev branch和目标；不更改Main cwd/checkout。旧记录保留原字符串用于审计，读时区分同实体路径别名与真实搬迁：前者通过统一规范化证明，后者必须走原显式rebind。

拟修改 `workspace.Bind`、`ValidateAuthority` 与测试夹具统一路径策略；不能用字符串放宽或共享仓库身份让任意checkout替代Main。原路径不存在、分支改变或另一个linked worktree仍拒绝/恢复。

Worker隔离不可用时，工作包给明确环境待办或供Main选择新执行的check_location；不自动切换原ready执行。不把macOS/Windows无隔离支持说成全部平台可运行Worker检查。

## 4. 统一故障分类与观察

拟新增typed `CheckFailure`：code、category、check/profile ref、owner、observed、repair、retry_condition、receipt/log refs。类别含 capability_missing、dependency_missing、resource_wait、process_exit、timeout/cancelled、input_drift、diagnostic_io和unclassified_failure。

退出码非0只能证明进程失败；只有特定runner的结构化测试结果能标为assertion_failed，不能凭字符串泛化为产品缺陷。未知错误保留原证据并归诊断责任。

长命令沿原process tree及有界观测，工作包提供run ref/状态入口。正常running不算无进展；timeout保持失败现场，修复环境后按原retry-preserved或重新检查。新timeout由profile明确配置并受Hook/Integration总时限约束，不能长检查卡住短Hook。

## 5. 恢复与凭据

预检失败不登记PASS、不修改冻结输入；修复依赖后再次探测。执行失败有完整回执、日志和适用性，即使产品未改也说明为何重跑。

Main集成检查失败保持checkpoint/候选；代码改变走原rework，新报告绑定新内容；纯环境恢复重验原候选。fixture cleanup失败保持资源待办，不强删外部资源或隐瞒状态。

回执按G05 v2补环境/fixture/runner字段；日志与最小输出示例随错误送达，完整日志路径可读。Schema错误用现有discriminator pruning为基础生成JSON Pointer，不丢原详细错误。

无法识别目标分支时，摘要按分支逐行给首个错误，不在标题再次枚举同一分支名；原完整错误保留在unwrap链和verbose入口。接手时在干净`ed61728`复现分支摘要长度用例失败，重复名称是可移除的反馈负担，不能通过降低测试门槛解决。

## 6. 改动位置与验收

修改 workspace/capabilities、check、check_other/linux、binding、原workspace prepare/launch、integration/check_runner与对应Schema；若profile入Execution，同步rebind/recovery/replace/bootstrap的复制和验证，不遗漏历史记录。

测试覆盖：worker Linux隔离可用/不可用、Main缺工具、网络依赖不可offline执行、超过副本上限、profile超时、命令退出、日志I/O、资源占用、输入漂移、path alias/真实relocation/错误checkout、已交付执行禁止替换、环境恢复不重复合并。

真实平台分别执行：Linux支持runner；macOS/Windows明确的不可用及Main路径；安装资产、路径空格/中文、临时目录别名和中断恢复。交叉编译不替代这些运行。

先补path/能力diagnostic与Main预检，再上线少量profile；完整环境指纹后才启用G05复用。旧命令profile默认never复用，缺字段保留legacy限制；本次没有修改环境或生产绑定。

## 7. 首个 Main 入口预检与真实故障契约

原 Main runner 使用继承环境中的 `sh -c <已登记命令>`。Main prepare/launch
新增只读入口预检，核验真实 shell 和可解析的简单字面命令工具是否存在，
不执行检查或版本探测/安装脚本。仅可证明入口可用，不能据此宣称依赖、fixture
或服务全部准备好。复杂 shell/变量/控制流保留 legacy 执行语义，投影明确
unverified，不从文本猜全部工具。required check 的明确入口缺失拒绝 ready；
`locked:` 仍非可执行检查，Worker 隔离模式与现有上界不变。

公共 typed Failure 仅表达 code/category/owner/check/observed/repair/
retry_condition/receipt/log 与原 error unwrap 链。process_exit 不等于产品
assertion_failed；timeout/cancelled、diagnostic_io、capability_missing 保留
明确来源。Main/Worker 的真实运行回执增加 failure 对象，保留旧字段；不用
调用者填写 status 或存在日志来证明 PASS。旧检查没有完整 profile/environment
闭包，复用仍 never；本切片不改变 Integrate 的 required/fresh 检查。

预检在原 prepare/launch 发布/预留 session 前执行；无副作用输出 Capability
Report 可只读查询并给实际 owner 修复动作，不写 Runtime、PASS 或新批准。
Main 的 shell/工具路径空格、中文和引用语法覆盖；未知语法不伪称 verified。
真实执行失败/中断/诊断 I/O 保持原 checkpoint/候选，不自动换 runner、合并、
修环境或放宽 Worker 隔离。显式 profile 与完整 G05 键保持后续必须完成项。

首个切片的源码/实际执行/安装范围见 [全局设计 §7.19](../GLOBAL-OPTIMIZATION-DESIGN.md#719-g07-main-检查入口预检只读能力与真实故障分类) 和 [检查能力安装证据](../evidence/global-optimization-check-capability-install-20261013.json)。

## 8. CheckProfile 冻结与回执采集契约

原 manifest Assignment 可选 check_profiles，每项 id/version/check_index/command/
location/timeout_seconds/environment_policy/environment/dependencies/fixtures/
input_closure/side_effects/reuse_policy。check_index 必须命中同一原 required_checks
且命令逐字相同；不得覆盖 locked: 或增加第二命令。location 与原新执行选择
一致，timeout 为1–300秒、外层 Hook/Integration 的更短截止仍优先。旧无字段
保留旧5分钟/继承环境，旧 ready execution不补 profile；replace/rebind/adopt
保留原 profile，而不是从当前文件抓取新命令。

profile 位于已登记 manifest 原批准边界，不建立新批准索引。依赖/fixture
为明确仓库相对路径+摘要，实际 source closure 经原 frozen inputs/Bootstrap
校验。environment_policy 为 inherited 或 declared_only；后者只传批准白名单
事实，记录 unset 与值的差异，不转储所有环境变量或秘密。未声明完整环境、
副作用unknown、旧回执、不完整输入一律没有复用资格。

第一次启用 profile 仍执行原检查；实际回执采集 profile canonical digest、
实际受测树/依赖/fixture内容、runner二进制及平台指纹、白名单环境与日志摘要、
原Runtime/generation/Assignment/owner/执行代次，保留原字段。输入在执行中
变化或诊断缺失不得写PASS。Main post integration与独立责任保留原fresh要求，
profile不把作者检查替代领域判断。G05只有原工具回执经原登记接收验证后，才
能按完整键考虑同责任复用；本阶段先完成采集，不假造命中或启用全局缓存。

显式 profile 采集链与真实集成/安装范围见 [全局设计 §7.20](../GLOBAL-OPTIMIZATION-DESIGN.md#720-g07-显式-checkprofile实际指纹与-v2-观测回执) 和 [CheckProfile安装证据](../evidence/global-optimization-check-profiles-install-20261013.json)。
