# 模板设计主线与全局优化

`blueprint/` 位于仓库根目录，保存本模板自身的设计与维护文档，与目标项目的 `docs/` 分开；发布包不包含此目录。层级定义以 [L1](l1-principles/L1-design-principles.md) 为准。

| 层 | 问题 | 入口 |
| --- | --- | --- |
| L1 | 为什么这样设计、守住哪些原则 | [原则](l1-principles/L1-design-principles.md) |
| L2 | 生命周期的目标与边界 | [生命周期](l2-lifecycle/L2-lifecycle-plan.md) |
| L3 | 每个阶段如何消费机制与收口 | [阶段索引](l3-stages/L3-README.md) |
| L4 | 跨阶段机制怎样定义与恢复 | [机制索引](l4-mechanisms/README.md) |

下层引用上层而不复制权威。

## 全局优化

[全局优化指南：确定交付、减负与执行流畅度](GLOBAL-OPTIMIZATION-GUIDELINES.md) 是独立于四层主线的全局设计文档，不属于 L4，也不新增层级。它汇总八类调优 guideline、交付优化策略、机制覆盖缺口及实战评价方法；具体生命周期和机制契约仍引用所属层，原则权威仍是 L1。

[全局优化落地设计](GLOBAL-OPTIMIZATION-DESIGN.md) 逐项调查上述要求，给出共用工作包/提交设计、十项交付实施卡、机制缺口覆盖、迁移与验收；[global-optimization/](global-optimization/) 保存八类调优各自的代码证据、具体改动、失败恢复和实施切片。

此目录保留 L1–L4 正式设计与独立的全局优化指南/落地设计，不保存游离于正式设计的迭代计划、调查、复审等中间材料。设计描述不代表实现已通过；实现与测试以代码和实际执行结果为准。安装资产不得依赖此目录。

实现资产由代码、Schema、Skill、Agent Definition、模板与测试承载；真实需求中的反馈按 L1 演化协议回写设计。
