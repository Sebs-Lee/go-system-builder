# G01 — 上下文与技能的落地设计

> 全局设计；承接 [G01 guideline](../GLOBAL-OPTIMIZATION-GUIDELINES.md#g01--按任务选择技能与上下文) 与 [共用工作包](../GLOBAL-OPTIMIZATION-DESIGN.md#3-共用基础当前工作包)。新接口/配置均为拟实现。

## 1. 调查证据与结论

- [角色卡](../../agents/backend-builder.md) 声明 21 个 Skill，frontend 为 16 个。
- [team/launch.go](../../internal/team/launch.go) 的 `roleDefaultSkills` 另维护全角色列表；`GenerateReadbackRequests` 将 agent-dispatch、角色默认与 manifest.skill_refs 合并后解析。只删 frontmatter 不会减掉 teammate 的加载。
- [team/validator.go](../../internal/team/validator.go) 有 responsibility→mandatory skills 和 risk→required responsibility 映射；这些职责要求不能被成本优化裁掉。
- [activation_envelope.go](../../internal/assignment/activation_envelope.go) 用 SkillRefs 派生命令类别，未知或无有效映射时存在 test/lint/build 兼容回落。知识加载与能力授权当前存在耦合。
- [Bootstrap](../../internal/workspace/bootstrap.go) 复制所有已安装 agents/skills；它控制安装量，不等于全部技能均加载。已有 digest 验证不允许在 ready 后静默改资产。

结论：需要统一选择生产者，分别处理“哪些资料存在”“哪些技能加载”“哪些操作被授权”。第一版只优化加载与派发上下文，不用删安装资产来声称 token 改善。

## 2. 具体改动与数据来源

拟新增 `internal/agentcontext/selection.go` 与内嵌 `agent-context-profiles.json`。profile 每角色声明核心技能、按明确技术栈触发的技能、按责任/已登记风险要求的技能、版本。选择函数消费原 manifest、角色定义、批准的项目技术栈与责任全集；不扫描自然语言猜风险或发明新责任。

选择集合为核心技能 ∪ 明确 responsibility 必需技能 ∪ 经确认的技术栈技能 ∪ 本 Assignment 的已登记 skill_refs。重复去重，每项保留来源与加载原因。缺 core 或 responsibility 技能时不能发出可开工作业；未配置 profile 的角色保留 legacy 列表并标识。

主会话阶段只加载原 primary methodology；Worker 核心保留 agent-dispatch、对应工程技能及适用基础质量实践，具体集合通过固定案例验证。高风险技能由原风险责任带入，不以 token 阈值删除。

修改 `team.GenerateReadbackRequests`、角色生成/发布流程和工作区 launch，使它们消费同一 selection。profile 在 Harness 中内嵌，安装的角色 frontmatter 由相同 profile 生成最小核心列表；角色正文的“加载全部额外技能”改为“加载当前包已解析的技能”。

## 3. 权限分离与实际接线

保持原角色权限及批准的 manifest/activation 能力来源。新增 `SelectedContextSkills` 用于加载；原授权 SkillRefs 在切换期原样保留。不得把缩减后的集合直接传给 `commandClassesForSkills`，触发兼容默认而扩大或改变能力。

后续若要独立 `capability_profile`，须单独修改 activation、policy、角色一致性校验和 schema，证明有效权限等价或明确收窄；不与本次加载减负捆绑。

工作区Worker的installed role保留核心frontmatter；当前 `workspace.Launch --allowedTools`没有Skill，首版不为加载新增工具权限。包为每个额外技能给已安装、BootstrapReceipt固定的SKILL.md路径/digest/目的，Worker用已有Read工具消费，相关引用文件沿同安装闭包读；launch提示给明确读序。文件不可读或digest不符不能算加载完成。

teammate的readback request用同一selected set；宿主已验证支持Skill调用时使用原调用，不支持时给同样的显式Read路径。原生subagent核心由frontmatter预加载，额外技能仍按包通过已支持的Skill/Read入口消费；记录core已加载项以免重复要求。工作包仅写selection不足以证明消费，验收检查真实工具事件/宿主加载记录。第一版不把新增Skill放进workspace allowedTools，也不改变activation授权集合。

selected profile 与来源摘要在执行准备时冻结并进入 BootstrapReceipt/Execution 的版本化元数据。ready 后要变更技能必须保全原执行，并按原允许的重派/新代次路径处理；不能覆写原 bootstrap。

## 4. 上下文装配

`workpacket.Build` 将阅读路径、章节/anchor、purpose、skill selection、唯一动作和收尾放入当前包。仍读取原权威文档，必要段落可作带 source SHA 的摘要视图；摘要丢失不丢原资料。

Hook 原生消息按身份→本人状态→动作→包路径→等待 owner 排序，按 Unicode 有界截断；具体范围在包中。新启动与阶段/责任变化发完整包；普通工具只给当前变化。相关资料不变时不要求重新读全部链，发生漂移则指出哪些输入失效。

## 5. 失败与恢复

- 来源技能缺失：Main 修复已安装资产或 profile；不回落到 factory 的另一份文件。原 catalog 混合布局禁令继续有效。
- 技术栈/风险缺决定：发当前规划题并归 owner；不以最小加载伪装可开工。
- profile 变更：旧执行继续使用其固定版本或走允许的重派，新发行不得复用旧 digest。
- 消息被裁剪/包投影失败：保留短入口与具体系统错误，由同身份重建视图，不让 Worker重新猜资料集合。

## 6. 验收与实施切片

1. 在 `internal/team/launch_test.go` 加选择一致性测试：backend 无数据库/对象存储修改不加载对应技能；已登记数据库/安全责任必需技能仍在；无 profile 是明确 legacy。
2. 在 assignment activation 测试中比较缩减加载前后的有效工具、写入范围、命令类别完全一致，专测空集/未知 Skill 的兼容回落。
3. 在 bootstrap/launch 测试中验证 profile 固定、缺安装资产、digest 漂移、恢复与 ready 禁止静默修改。
4. 在 lifecycle context 测试中使用长 scope、中文与很长路径，确认 next 和工作包引用完整、UTF-8 合法。
5. native subagent/teammate/workspace三入口分别验证实际Skill或Read路径，workspace不带Skill也确实读到必要额外技能及引用；未知宿主能力给明确限制。记录实际加载与可取得的token，同案例比较首次有效操作、返工与遗漏风险。

先 shadow 生成 selection 对照旧列表，再启用一个 Builder 角色；不同时裁安装资产或调整模型。通过后扩展各角色并移除 `roleDefaultSkills` 重复权威。嵌入 profile/Schema 和角色生成测试一起发布，目标项目不引用 blueprint 或 Go 源码。

## 7. 完成与迁移边界

设计完成要求能定位所有三个入口的 selection 生产者，并明确授权不变。实现完成要求一致性/权限/恢复测试通过；行为改善另需实测。旧 manifest 和现有 ready Worker 不自动重写，兼容路径标 legacy，下一次合法发行使用新 profile。
