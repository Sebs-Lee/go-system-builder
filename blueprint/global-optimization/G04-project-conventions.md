# G04 — 项目惯例与默认选择的落地设计

> 全局设计；承接 [G04 guideline](../GLOBAL-OPTIMIZATION-GUIDELINES.md#g04--项目惯例成为默认工程选择)。Schema/resolver/S4、新 S5 冻结链和 Builder/native 固定上下文入口已落地；真实模型/风格对照待验收。

## 1. 调查证据与结论

[docs/rules](../../docs/rules/README.md) 已提供错误、API、命名等规范；[ARCHITECTURE 模板](../../docs/architecture/ARCHITECTURE-template.md) 声明模块职责、模型、契约和取舍；[project-map](../../docs/project-map-template.md) 已有确认/unknown的技术栈事实。初始调查时没有统一、带适用范围和真实例子的派发默认选择视图；当前首个只读投影和新 S5 冻结链已建立。

不能从代码搜索出现频率或首个匹配推断批准惯例。目标是减少普通工程选择中的迟疑，并保持迁移、例外与业务决定可见。

## 2. 正式来源与投影

项目可选正式资产为 `docs/rules/project-conventions.json`，内嵌 `project-conventions.schema.json` 已实现；规则项为 id、topic、scope、default、source_refs、example_refs、exceptions、status。Architect按已批准规则/架构登记引用；它只索引决定，不重复写一份业务契约。安装只带空模板，不创建当前项目默认。

source_refs指向当前权威规则/架构/已批准决定；example_refs指向 bound development tree 的真实实现/测试。读取沿 file_sources，发包时冻结实际来源闭包。源码只是写法示例，不能覆盖合同不变量。

`internal/projectconventions` 已提供 `ResolveMany(root, files, requests)`；requests 携带原 TASK scope/显式读集/登记证明，匹配已声明 scope/topic，返回选择依据、例外与来源，不执行任意规则表达式或从 unknown stack 推断。新 S5 分析携带当前相关项，全文通过冻结路径展开；实际 Builder/native 入口接线见 §9–§10。

未建立 registry时，第一版可从明确的项目规则引用生成只读候选卡，由原 Architect确认一次再登记。没有匹配项可以继续作普通授权工程选择；缺决定若影响当前验收，就给具体上游题，不能把所有无惯例都变成开工闸。

## 3. 冲突与偏离

当 registry重复、默认值不同、处在迁移中或示例与规则矛盾时，显示相关来源并归 Architect/规则维护；不选择“最新文件”或更强措辞作权威。

Builder的偏离理由写在原计划/结果题，仅在偏离时展开，包含影响与验证。超出批准范围、锁定规格或人闸仍走原变更路径；说明理由本身不授予权限。

默认选择允许“参考相同模块/相同接口”的细化，语义差异由角色判断。新业务值域、权限效果或状态转移绝不靠惯例补成需求。

## 4. 失败与恢复

- example删除或改变：该项投影标 stale，重建或确认新示例；旧引用不继续呈现为当前标准。
- source未锁定/不可读：不封为默认；必要决定给明确责任，普通背景推荐可省略并说明。
- registry丢失：仍能从原规格完成必要工作；视图重建，不让运行时失去当前状态。
- 批准来源变更：只失效受影响规则引用和新派发包；已冻结执行依照原适用性/变更控制处理。

## 5. 改动位置与发布

新增 Schema、resolver、可选模板资产，并在 `packaging/include.txt`/releasegraph中列入正式安装资产；不得把 blueprint作为 target input。project-map只引用 registry，不再复制各条规范。

S2当前题引导 Architect确认已有惯例或登记新决定，S4/launch/workpacket消费相关投影，S5/S7检查来源与违规偏离。新建项目可以没有代码示例，包明确只有正式规范。

## 6. 验收与切片

Schema与resolver测试覆盖：无 registry、明确单项、多项冲突、迁移例外、跨模块不适用、旧例子、未批准来源、symlink/path escape、冻结漂移。

真实例行 API/错误处理/测试任务比较反复列方案、询问次数、风格返工和缺陷；新业务决定案例证明没有被默认补齐。

先做只读参考卡与两个常见 topic，不自动收集全仓库惯例；由同一来源建立 approved registry后再扩展。移除某项时保留正式决定历史及失效原因，不改 Runtime权限或已批准验收。

## 7. 首个切片的来源与批准契约

代码调查确认：原 Loop `file_sources` 默认从 bound development tree 读取正式
文档；安装的 Rules 列表与登记的 locked 架构文档已有各自权威入口。不能把
registry 自写的 `status=approved`、文件 mtime、代码频率或 `project-map` 的
unknown 栈事实当作批准证明。S4 的显式 authoring 命令可用 disk view；正式
S5/派发消费必须使用原 fileview 和原 REQ/文档登记快照，二者标明来源模式。

第一版候选卡只针对 API/错误处理的明确 Rules 引用，不从全文自动总结出默认
语义。卡包含 topic、适用 TASK scope、原来源 path/anchor/hash、当前可读状态、
待 Architect 填写/确认的选择和例外；`candidate` 不是当前批准默认。无 registry
返回明确 absent、无匹配返回 empty，不改变 TasksCheck Problems、DAG、权限或
原任务开工条件。不可读、越界、未知条件只输出有出处的专业问题。

registry 为可选正式索引，模板只能含空 entries，不安装已批准项目默认。每条
选择必须携带 source_refs：稳定项目规则沿原 Rules 权威入口验证；架构/决定
沿原登记文档的 locked/version/hash 验证，登记事实缺失就保留 candidate。
例子从同一个 bound tree 读取并固定 hash，不能用工作区最新文件替代。Resolver
只按明确的仓库相对 scope 前缀匹配，不支持脚本表达式、隐式技术栈推断或
“更具体者自动获胜”。同 topic 的不同当前默认全部列为 conflict，交原责任
处理；migration/superseded 不升级为 active。

参考卡与 registry 读取只读，不登记 Runtime 或批准；真正维护索引仍走原
Architect 计划、文档变更与登记边界。只有证明来源适用的 declared default
可进入新派发的冻结参考闭包，既有 execution 不补资料、不重算旧选择。源码
示例漂移须标 stale，正式来源变化走原失效/返工；任何默认都不补业务值域、
权限或状态转移，更不替代独立 S5/S7。

实施顺序：闭合 Schema/空模板 → 只读 resolver 与来源矩阵用例 → S4 候选输出
→ 新冻结包接线 → 原 Builder/审查 Skill 与正式安装资产。每步保留 absent/
candidate/conflict/stale 的测试和回执；真实风格/询问次数对照另记观测，代码
静态输出不声称已减少模型开销。
上述首个 resolver/S4/S5/安装切片已完成，完整代码与源码范围验收见
[全局设计 §7.17](../GLOBAL-OPTIMIZATION-DESIGN.md#717-g04-原来源惯例索引s4-投影与新-s5-冻结链)
和 [项目惯例安装证据](../evidence/global-optimization-project-conventions-install-20261013.json)。

## 8. 首个 resolver 与冻结输出的具体契约

registry schema_version=1.0.0，entries 每项包含 id/topic/scope/default/
source_refs/example_refs/exceptions/status；scope 是无通配符的仓库相对路径前缀，
`.` 表示项目范围，不支持表达式。source_refs 带 kind/id/path/anchor/sha256，
架构/决定另带 version；example_refs 带 path/sha256。status 为 candidate、
approved、migration、superseded；这只是声明，消费状态仍由 resolver 求证。

Rule 来源必须在原 locked Rules 索引的 File 表格中，引用的 rule_id/status/hash
须一致；架构/决定来源必须命中调用者提供的原登记 formal ref 的 id/path/
version/hash，源文档自身也须 locked。默认文字必须是某个已证明来源的引用
段落中实际存在的文字；缺失/歧义 anchor 不保守升级为完整文件的默认。机器
只证明索引、来源与字节一致性，语义适用性/例子冲突仍由原专业责任判断。

输出 task_conventions 带 resolver_version、TASK 出处、registry 状态和实际
sources(path/hash/source_mode)、相关 entries、具体 questions。只有来源证明
完整、status=approved、原文字/示例有效且未冲突的条目为 declared_default；
其他明确为 candidate/stale/conflict/migration/superseded。例子可为空；删除/
摘要改变不能继续显示当前默认。无 registry 时只为 TASK 明确引用的 API/
错误处理 Rule 生成候选，不自动猜 topic/stack 或选择默认。

scope 按目录边界相交筛选；同 topic、重叠适用 scope 的不同当前默认全部列
conflict；不同模块不互相覆盖，重复 id 不静默去重。来源包括实际读取的
registry、Rules 索引、规范/已登记文档和示例，不按 suffix 替换原 fileview。
disk 读取检查真实路径不能越界；git_tree 使用原提交字节，不以磁盘 symlink
代替。读取失败/解析上限只列问题，不修改原开工/覆盖/DAG/门的结果。
正式消费的 example 必须来自原 bound git_tree；显式 authoring 的 disk 示例
只表明当前草稿资料及摘要，不构成提交/批准证明。原 source contract 若将
某个 example 路径声明为 disk，正式投影列 stale，不能偷偷换读取来源。

S4 authoring 返回候选/来源模式；新 S5 issuer 用原 Packet.inputs 提供登记
证明，冻结同一输出及真实来源闭包。新增 artifact 字段可选，旧已发行包不
补齐；Worker 安装/核验、原接收/修订/恢复沿既有分析链。review-check 的
conventions 读取冻结字节，任何建议都不写 Runtime 或替代专业结论。

## 9. 原 Builder/native 上下文消费

初始 `workspaceContextInputs` 只冻结原任务/依赖链和 Rules 文件，缺少按本人
scope 求证的惯例投影；新 S5 冻结分析不能代替 S6 Builder 的实际入口。当前
新 v1.2 workspace/native assignment context 增加可选投影，复用本 resolver：
按原登记 TASK 路径解析，使用原当前 generation 的完整登记文档引用作证明，
正式读取继续原 bound development fileview；Rules 的磁盘资产摘要本身不能
证明 source_default。Scope 使用原 Assignment/TASK 相交，不改写入权限。

相关 registry/index/规范/示例进入原 FrozenInputs，发布前和原 Writer 登记锁内
重复核验真实摘要；只读 provider 不选择新的授权 Skill、不执行检查、不答专业
计划。Worker 使用自己的冻结根，native 使用原 authority root 的固定引用；
同一旧 ready/context binding 重放保持旧资料，不能用当前 registry 补字段。
有新专业决定走原维护/失效/重派；无索引仍可普通授权选择。

S5 有冻结分析时复用已发行 projection，避免另算两份默认；S6 实际 prepare/
launch/原 begin/owned checks/report/真实 Main integration 已验证新输入闭包
及原接收，native 固定登记/重放/漂移拒绝已覆盖；模型消费分别验收。不能
因当前入口/替身进程或打包测试通过便声称模型已消费 G04；其他领域 typed
adapter、G01–G08 和完整平台/实战验收仍保留在全局 goal 中。

## 10. 上下文接线实施契约

新 context 可选 conventions {resolver_version,task_conventions}，仅知识字段。
S5 直接取原已发行 analysis 的投影；无此字段的旧 analysis 不重新求值。S6
Builder 从原登记 TASK 路径分析，只用 TASK/Assignment 共同 scope；没有可
证明范围时列问题、不扩范围。架构/决定证明取原当前 generation 的登记集合，
正式读取沿原 catalog.file_sources 和 bound development branch。

所有实际可读来源须在原 frozen_inputs 中有相同摘要，Worker Bootstrap 安装，
native 的 input_read_order 列出；publish 与原 Writer 锁内复核原 fileview 与
真实字节，固定 binding 重放不重新选择。workspace/native 读取检查 bundle
版本、原 TASK 路径和 frozen_inputs 关系；legacy 缺字段保持兼容。任何 Context
都不替代 Task/Assignment/plan/activation、checks、报告或集成资格。

示例作为执行开始时的冻结参考；由本人的真实受控提交修改示例时，沿原
checks/report/Main verified integration 验收，不将成功集成误认成外部漂移。
外部改动在继续执行/接收前拒绝，恢复仍保留原来源。验收需实际 S5→S6、
prepare/launch/begin/owned checks/report/非 squash 集成，以及 native 固定
登记/重放/来源漂移；启动替身只证明进程/输入装配，不证明模型消费。

当前实现与测试范围见 [全局设计 §7.18](../GLOBAL-OPTIMIZATION-DESIGN.md#718-g04-实际-builderworkspace-与固定-native-惯例上下文) 和 [实际惯例上下文安装证据](../evidence/global-optimization-convention-context-install-20261013.json)。
