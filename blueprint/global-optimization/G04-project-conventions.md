# G04 — 项目惯例与默认选择的落地设计

> 全局设计；承接 [G04 guideline](../GLOBAL-OPTIMIZATION-GUIDELINES.md#g04--项目惯例成为默认工程选择)。候选 registry、投影与测试为拟实现。

## 1. 调查证据与结论

[docs/rules](../../docs/rules/README.md) 已提供错误、API、命名等规范；[ARCHITECTURE 模板](../../docs/architecture/ARCHITECTURE-template.md) 声明模块职责、模型、契约和取舍；[project-map](../../docs/project-map-template.md) 已有确认/unknown的技术栈事实。当前没有一套统一、带适用范围和真实例子的派发默认选择视图。

不能从代码搜索出现频率或首个匹配推断批准惯例。目标是减少普通工程选择中的迟疑，并保持迁移、例外与业务决定可见。

## 2. 正式来源与投影

拟新增项目可选正式资产 `docs/rules/project-conventions.json` 和内嵌 `project-conventions.schema.json`，规则项为 id、topic、scope、default、source_refs、example_refs、exceptions、status。Architect按已批准规则/架构登记引用；它只索引决定，不重复写一份业务契约。

source_refs指向当前权威规则/架构/已批准决定；example_refs指向 bound development tree 的真实实现/测试。读取沿 file_sources，发包时冻结实际来源闭包。源码只是写法示例，不能覆盖合同不变量。

拟新增 `internal/projectconventions` 的 `Resolve(files, taskScope, stackFacts) -> RelevantConventions`；匹配已登记 scope/topic，返回选择依据、例外与来源，不执行任意规则表达式。工作包携带少量当前相关项，全文通过路径展开。

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
