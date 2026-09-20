# 分支优化完整回审与统一修复

日期：2026-09-20。对象：`0.1.6` 分支 `fff2b1b` 及工作区修复；实际基线是 `main`（仓库无 `master`）。

状态：本轮范围内的完整审查、10 项统一修复及最终全量/race 验收均已完成。按用户要求使用三个 `gpt-5.6-luna` / `max` 子代理分别审查交付、维护来源、修复与捕获证据，主代理复核跨模块消费和实际 Hook 路由。

## 汇总问题与修复验收

| 优先级 | 已复现的问题 | 修复结果与对应验收 |
| --- | --- | --- |
| P1 | dirty worker 上的绿色测试可认证 authority 的冻结失败版本 | 捕获实际执行位置、计划/assignment 和执行前后冻结输入；提交 PASS 必须核对捕获输入与冻结基线，保留诊断用途 |
| P1 | 合并前在 authority 运行 required_checks，尚未合并的新产物检查失败 | 预检运行于 worker，合并后联合检查运行于 authority；正常 E2E 删除双目录检查 workaround |
| P1 | 已验证 merge 从目标分支消失后仍可确认并清理 worker | 重用 receipt、确认和清理前检查 merge 的祖先可达性；失效持久化为 preserved，后继资格同步撤销；目标正常前进仍有效 |
| P2 | capture producer 的 command_output 引用不能由 ReviewResult 直接消费 | 统一格式解析、digest/路径校验与 provenance 关联；验证正常捕获至完整 clean-round，拒绝篡改/不可验证输出 |
| P2 | 注册的非标准 manifest 路径无法被 Hook/worktree 消费 | 从已登记 Agent/Team 引用读取；保持 owner 隔离和正式引用失效时的关闭行为 |
| P2 | agent-begin/agent-event 等入口默认 Runtime 路径落在进程 cwd | state/journal 统一以 --root 解析，包含 bug-event 的升级分支；真实 CLI 默认参数回归 |
| P2 | bug-event 显式消息路径重复拼接，读取错误被忽略，非法消息仍推进状态 | 绝对路径原样读取、相对路径以 root 解析；显式输入不可读或 schema 非法时拒绝且不改变 revision |
| P2 | PostToolUse 有提示生成函数却未接入实际路由，安装 matcher 缺 Bash | 实际 Agent/Task 与指定 Harness Bash 检查点输出只读调度摘要；普通 Bash、异步启动、worker、S5 不输出 |
| P2 | dispatch 硬编码文件来源，与 Definition.FileSources 分叉 | 使用真实 catalog 配置；测试 nested custom/disk、根 git_tree、.claude disk，catalog 缺失/非法一致失败 |
| P2 | REV 模板提示登记磁盘报告后推进，遗漏 Git 来源的提交/集成步骤 | 模板与协议明确 docs/reports/review 报告需提交/集成至 REQ 绑定开发分支；Hook 实测未提交 unreadable，提交后 advanced |

上表 10 项均已统一修复。旧的 opt-in 反例日志保留作调查证据，最终验收以默认开启的回归测试为准，不把跳过视为通过。

## 完整审查覆盖

- **S3/S4/S5**：REQ 隔离、closure 与 amend/freeze、总表精确集合、共享 Schema/样例/传递依赖、Git/disk 混合来源、冻结对象漂移与越界写域。
- **S6 交付**：正式注册、真实 SendMessage 激活及 agent-begin 恢复、write/output 路径并集、canonical Result、worker 预检、合并后检查、确认/清理、后继解锁、重报撤销/重新验证、来源分支移动、目标历史重写、checkpoint 损坏、并发集成。
- **S7 审核**：捕获输出和被测输入关联、dirty worker、执行期间变更、输出引用及摘要校验、正常 clean-round；人工旧证据兼容不等于自动捕获可以跳过验证。
- **S9 修复**：真实 CLI 从派发/worktree 到 worker-local PlanReport 由 authority 收录，再进入 execution；`internal/repair` 的集成测试覆盖结果→影响面→定向验证→fresh S7 cursor。
- **运行时与 Hook**：authority 别名/复制 Runtime/替代 state-journal/符号链接、Writer 恢复、PostCompact 恢复、实际 PostToolUse 路由、matcher 安装、只读提示不推进状态、默认 CLI root 定位。

前轮已修复的冻结摘要维护、canonical Result、write/output 并集、S5 不提前派发、REQ 专属模板路径和 task-complete root 定位继续纳入回归。

## 验证记录

最终验收通过：

- 当前源码构建、`go vet ./...`、`make fmt-check` 和 `git diff --check` 通过。
- 初始化隔离副本 `/tmp/branch-final-unified-validation-qtkfe3h4`；`doctor`、`validate --all` 与 Hook 进程 smoke 通过。日志分别为 `/tmp/branch-final-doctor.log`、`/tmp/branch-final-validate.log`、`/tmp/branch-final-smoke.log`。
- `internal/review` 全包通过（77.425s）；自动输出正常提交并进入 clean-round、dirty 冻结输入拒绝、执行期间变化、未声明产品漂移、篡改输出、缺少 buffer、withheld 输出拒绝以及手工旧 buffer 兼容全部通过。日志 `/tmp/capture-fix-review-final.log`、`/tmp/capture-fix-cli-final.log`。
- `internal/hookctx`、`internal/integration`、`internal/qualitygate` 全包通过；真实激活/后继释放、严格 worker 预检、非标准 manifest、目标重写、缺失 receipt、来源移动、损坏 checkpoint、并发集成的默认回归通过。日志 `/tmp/audit-owned-packages-final.log`、`/tmp/recheck-delivery-default-final.log`、`/tmp/final-audit-delivery-all-final.log`。最后一份早期日志启用了调查变量，最终测试文件已删除这些开关。
- `internal/assignment` 全包及 bug-event CLI 通过；消息相对/绝对路径、外部绝对输入、缺失文件、目录、非法 schema 以及拒绝后 state/journal 不变均有默认测试。
- 第一次隔离全量仅失败于尚未同步的旧预检断言（修复成功后已自动清理 worker，旧断言仍要求保留），并非生产集成失败。已同步最终测试并强化 JSON integrated/blocked 断言。随后 `go test -p 2 ./...` 全部通过（exit 0，REQ-039 系统包 178.982s），日志 `/tmp/branch-final-full-verified.log`；所有代码、测试与配置逐一比对当前工作区一致。

- 新增 CLI/REQ-039 E2E 在 race 下通过：`go test -race -p 2 ./internal/cli ./tests/system/req039 -run '^(TestRecheck|TestFinalAudit|TestCaptureExec|TestReviewResultCaptures)' -count=1`，CLI 38.300s、系统包 86.426s；日志 `/tmp/branch-final-race-e2e.log`。

- 受影响核心包全包 race 通过：`go test -race -p 2 ./internal/assignment ./internal/hookctx ./internal/integration ./internal/qualitygate ./internal/dispatch ./internal/review ./internal/runtime`，exit 0，无数据竞争报告；review 296.994s、runtime 61.728s。日志 `/tmp/branch-final-race-core.log`。

已完成的独立审查验证：

- `/tmp/final-audit-authority-boundaries.log`：pathscope、policy、runtime、recovery、controller、hook、fileview、workspace 通过。
- `/tmp/final-audit-authority-cli.log`：authority 复制/别名/替代路径拒绝与普通 Git/squash/PostCompact 恢复通过。
- `/tmp/final-fix-cli-roots.log`：agent-begin、agent-event 从其他 cwd 使用 --root 和默认路径通过。
- `/tmp/final-fix-dispatch-wiring.log`：真实 PostToolUse 路由、安装 matcher、负例和 reminder delivery 相关回归通过。
- `/tmp/final-audit-internal-repair.log`：repair 全包通过，包括 fresh S7 handoff；该项是领域 API 集成级验证。

## 证据边界和排除的误报

- 本报告中的 E2E 指隔离项目的真实 Git、CLI/Hook JSON 入口。没有启动原生 Claude 会话，不能据此声称真实 Agent spawn、Worktree 平台事件或同会话 exit-2 续跑已验证。Worktree 平台接线在文档中是可选能力，缺少原生实测是验收边界，不是必需机制缺陷。
- S9 worker 无 authority 的 `.claude/review` 副本不代表缺失正式输入；正式 CLI 已证明能引用 authority 并收录 worker-local 报告。没有为了测试而复制整份控制面。
- 按 L4-runtime-control-plane §4.4，可变 evidence 索引允许重建，读取仍服从 file_sources；不得误把合法索引刷新当成冻结产品获批，也不增加额外批准门。冻结产品 SHA 保留由独立测试验证。
- 测试使用进程级 Git 临时身份；真实源码目录不初始化 Runtime，不修改全局 Git 配置。全量验证使用初始化的隔离副本，以避免既有测试对本地 Runtime 的假设和 schema 路径长度敏感断言。

修改保留在工作区，未提交或发布。
