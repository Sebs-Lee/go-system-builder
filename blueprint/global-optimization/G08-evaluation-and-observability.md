# G08 — 实战评估、观测和模型实验的落地设计

> 全局设计；承接 [G08 guideline](../GLOBAL-OPTIMIZATION-GUIDELINES.md#g08--用实战案例和对照实验调优)。新实验工具、指标和通用stall检测为拟实现。

## 1. 调查证据与结论

[metrics](../../internal/metrics/metrics.go) 已有gate、转换、CAS、恢复、集成和S7指标，family budget为16并要求生产者/消费者；[observations](../../internal/metrics/observations.go) 与 [compact](../../internal/metrics/compact.go) 通过有界发布/恢复压缩避免Hook等待汇总锁。

[IdleRecoveryExhausted](../../internal/hook/idle_budget.go) 只约束指定TeammateIdle恢复提醒，并保持权限/未完成事实；[recovery replay](../../internal/controller/replay.go) 是恢复过程的no-progress防护，不覆盖Agent所有普通操作。

[Hook-driven系统测试](../../tests/system/req039/s2_to_s11_hook_driven_test.go) 使用fixtures和runner证明原转换链；[S5原型](../../tools/stage-form-prototype/README.md) 证明纯内存交互。它们不能单独证明陌生Agent无需读流程手册即可完成真实任务。

## 2. 实战案例与执行器

拟新增工厂工具 `tools/optimization-eval/`，不进入release include。案例目录包含case.json、项目fixture、可公开给执行者的业务输入、隐藏oracle与期望接收/路由；case声明ID/版本、入口、角色、输入基线、配置变量、外部资源、时限和合法结束条件。

固定12类起始案例：首次Builder、首次独立S5、缺必需输入、缺业务决定、长scope、权限冲突、检查失败、交后等Main、并行保存、重复/响应丢失、compact恢复、旧基线/环境恢复。按不同角色/平台分组，不能用同一S5演示代表全部责任。

执行器建立临时真实Git/安装目录，发行真实作业包，运行受支持宿主，不把测试注入的Runtime事实称为Agent自行完成。隐藏设计历史但保留本次必要规格、代码与可定位资料；隐藏oracle不进入prompt。

纯函数/CLI fixture模式负责机制回归；真实宿主模式记录执行与平台能力。没有登录/工具支持时如实标not_run，保留可执行实验说明，不模拟模型成功。

## 3. 自动可测和不可直接可测的指标

接收consumer提供attempt/accepted/rejected原因、保存与登记时间、domain refs；assignment/checkpoint提供责任完成、等待和集成；tool check提供实际命令耗时/重跑理由。cas内部并发重试不算Agent首交拒收。

定义首交分母为作业首次完整提交；草稿保存不算，合法Finding/BLOCKER的有效接收与阶段PASS分别统计。总设计O05按当前必答/条件题与结论计算完整性；不增加Agent手填ready字段，不能从每次Edit计一个失败attempt。

token/技能加载/读文档轨迹来自宿主实际usage/transcript/支持事件；不可得标unavailable。文件字节或字符估算单列proxy，不当真实token。Main现有Hook不监听全部Read/Skill，Worker bootstrap虽有matcher `*`也不等于已采集这类指标。

实验模式可增加只读PostToolUse observer或经支持的宿主trace，采集文件引用、Skill ID和类型，不对Read运行全部控制cycle。事件匹配、额外开销与隐私边界经实际wire测试；默认生产不新增全量读审计。

通用stall的生产接线与实验完整trace分开：生产增加有界work_progress采样，Main PostToolUse matcher显式包含Read/Glob/Grep及已支持操作；Worker Bootstrap现有matcher `*`继续使用。`runHook`解析真实owner后、Bash非指定checkpoint早返回之前调用采样器；正常Read/status和Write/Edit/受控提交均能观察。采样只读当前包/责任/回执与必要摘要，缓存有限计数，不记录全文transcript、不运行Controller/qualitygate或改Runtime。未支持这些宿主事件的入口标stall coverage不完整，不承诺捕获纯思考或无事件的空转。

## 4. 数据结构与观察通道

实验结果保存在 `.local/optimization-evals/<experiment>/<run>/`，包含版本/配置/平台能力、usage、时间线、接收事实、原检查证据、失败原因和oracle判定；工具不形成生产Runtime权威。

线上先新增少量有消费者的有限指标，如submission outcome、save_to_receipt、responsibility_wait、stall诊断类别。沿原observer spool追加，汇总在health/会话边界做，family budget/producer-consumer注释与测试同步更新。

实验的高基数assignment/path标签留run文件；生产聚合用有限stage/role/outcome。统计丢失不改变授权或gate，指标缓存故障不给业务Worker新任务。

## 5. 通用无进展诊断

拟新增 `internal/hook/work_progress.go`，在上述明确支持的成功Read/查状态/修改/提交返回边界观察原作业事实：binding、当前具体缺口、answer语义摘要、相关输入/证据摘要和实际执行状态。仅对当前可识别责任采样，无当前作业则跳过。更新时间、全局revision、重复通知和message_id变化不单独算进展。

资料收集本身可以是进展：缓存有界的本作业相关 `source ref + digest + 已读范围` 集合，首次成功读取新的必要资料/章节重置重复计数；同来源同范围重复读取或重复status不重置。来源关联来自当前包与明确证据引用，不把无限读取无关路径当进展；缓存达到上限只标coverage不完整，不把正常读资料强判空转。实际工具不能提供范围/引用时标观察不足，不能只因领域Runtime暂未变就提示stall。

同作业同失败重复到配置的诊断阈值且无running进程/明确外部等待时，发一次typed stall：缺口、最后有效进展、已尝试动作、恢复owner与具体原因。重复诊断去重，新的区分性证据重置；不靠wall-clock判长任务死亡。

诊断不改阶段、不授产品写入、不将未完成标完成。Main从原checkpoint/assignment决定补资料、修adapter、恢复环境或合法返工；不自动替换或生成第二writer。unknown/unmapped gap明确归维护，专业未决归对应责任。

## 6. 对照实验与模型/预算

每次只调整一个主变量，冻结案例、业务输入、平台、工具profile和质量oracle；重复运行并保留holdout案例。比较分布/失败类别与质量约束，不从一次更快推断整体改善。

模型/推理预算为实验配置，不由Worker自己选择；角色名不自动决定难度。先固定入口/资料再比较候选配置，记录真实可用模型ID、预算、成本、首交成功及恢复次数。没有实际usage就不声称成本节省。

声明复杂/高风险任务分组、升级触发和授权策略；升级是原Assignment的明确重派/续跑安排，保留已提交证据与状态，不两个模型同时写同一责任。first phase不改agents现有model默认。

## 7. 验收、迁移与切片

扩展metrics测试验证有限family、丢失/并发/压缩中断不重复计数；consumer测试验证草稿/首次完整交付/重复/并发CAS的统计。stall测试覆盖无变化重试、纯revision churn、新证据、长运行、等人、缓存损坏与同事未交。实际Main设置→普通Read与非dispatch Bash status→observer→additionalContext全链路测试必须证明没有控制转换、副作用或PostToolUse deny，测采样时延；不能只调用采样函数测通过。

案例runner自身测试：隔离目录、版本固定、oracle不可见、not_run不计成功、trace缺失标unavailable、失败现场保存和清理范围。新设计回归用例另列，不能把当前未实现receiver测成通过。

先将已有卡点收成fixture并运行原版本基线，再上线最小S5真实链，逐个入口做A/B。生产只接有限指标和已验证stall诊断；模型实验与大规模trace保留工厂维护工具，配置回退不抹历史报告或Runtime字段。
