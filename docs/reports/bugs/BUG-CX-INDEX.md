# BUG-CX 索引：S0-S4 agent 视角复杂度审查（2026-08-17）

> 背景：owner 要求对 S0-S4 做复杂度审查——代入 agent 视角检查真实使用时的迷茫点、引导是否到位、机制是否晦涩、每轮产出要求是否清晰。三路 sub-agent 模拟行走 + 主会话对高危结论逐条复核（冷启动死路/S5.5 矛盾/primary_skill 三处分歧均经主会话亲证）。
> 基线：HEAD=9cd52fa。全部为模板仓库自审（无绑定 runtime，Runtime ref=N/A）。

## Canonical BUG 一览

| BUG | 严重度 | 主题 | 根因（一句话） |
|:--|:--|:--|:--|
| [BUG-CX-01](BUG-CX-01.md) | P1 | 冷启动死路 + S0/S1 引导三处矛盾 | 投影/引导把"runtime 未初始化"当异常态，而它是每个新项目的第一个常态；stage 引导信息三处居所无单一权威 |
| [BUG-CX-02](BUG-CX-02.md) | P1 | 人工手势协议缺失（落锁/层拍板/人闸白名单） | 人机交互手势从未被当作设计对象——意图有了（human-only/approved），动作序列与持久痕迹没有 |
| [BUG-CX-03](BUG-CX-03.md) | P1 | S2 引导层停留 v3 口径（protocol 无八文件包/桥）+ 死引用 + 字段谜语 | 机制左移到机器后，agent 可读层未被当作同步交付物 |
| [BUG-CX-04](BUG-CX-04.md) | P1（含 P0 死锁路径） | S3/S4 机器契约未传达 + S5.5 "locked" 词汇碰撞 → TR-002 死锁 | 同 CX-03 族 + 一个词复用两个正交概念（基线代际锁 vs 文件 Status） |
| [BUG-CX-05](BUG-CX-05.md) | P2 | S0 模板前向引用谜语 + UI impact 双声明位静默不一致 | 模板加法方向缺少"第一次填写者测试"纪律（减法有新居映射，加法无示例锚点）；控制点未覆盖全部声明位 |
| [BUG-CX-06](BUG-CX-06.md) | P2 | 报错不指路 + bridge skipped 假绿 + S-NNN 位数静默 | 公理五未建制化为文案标准——新代码靠自觉达标，旧文案无回清轮次 |

## 跨 BUG 的系统性根因（修复时须一并建制，否则按症状修完会再发）

1. **机制左移、引导未随**（CX-03/04/05 共因）：v4 各轮把判断左移为机检（cross-matrix join/桥 guard/词表收敛），但 protocol/skill/模板与代码的同步没有 checklist。建议建制：任何机检/词表落地，其 PR/commit 必须列出「protocol 对应段 / skill 步骤 / 模板字段」三处的同步说明（可并入 L1 演化协议 §7 的修订流程）。
2. **词汇无单一居所**（CX-01 的 stage 引导三处、CX-03 的 oracle 七字段三处、CX-04 的双 locked）：引导与术语分散多处必然漂移。建议：每个机器契约词汇（Status 三词/locked 双义/七维度/S-NNN 位数）在 rules 或模板定一个 normative 居所，其余位置只留锚点。
3. **报错即文档没有标准**（CX-06）：建议把 cross_matrix.go 的文案水准（为什么红+往哪修）写成一条 checklist 进 L3-README 或贡献纪律。

> **2026-08-17 修复状态**：owner 全部接受后同批修复完毕，六份 BUG 均为 fixed（机器项附测试钉住：冷启动投影/§n 漂移/非 TASK 依赖/UI impact 不一致/rebind 指路）；全量测试 + validate --all + doctor 绿。三个系统性根因的防复发建制（机检落地三处同步说明/词汇 normative 居所/文案 checklist）留待下一轮 L3-README 修订时并入。
>
> **2026-08-17 第二轮审查（complexity-review-2）**：上轮修复验证 CX-01/03/06 ✓、CX-02/04/05 ◐/✗；新立 BUG-CX-07（**P0**：planning 门的前置事实由被门转换自己产生——自动推进鸡生蛋，所有 E2E 靠禁用的手动 transition 绕过；TR-002 文案承诺不存在的重跑）、BUG-CX-08（S0/S1 信任链文案：`<you>` 占位符/AGENTS 自相矛盾/§C 一致性校验假安全网——上轮自修项的缺口）、BUG-CX-09（S2 口径：模板 risk 示例违自家指导/包清单三处矛盾/ADR package 无定义/步骤断档）、BUG-CX-10（§n 机检精度：§1 子串 §10/SYNC 无条款列/reviewed 词表谎言/CONTRACTS 注记漏落地）。全部经主会话亲证后落盘。系统性根因再发印证：词汇多居所无对账纪律（CX-08/09/10 三族同源）。

## 建议修复顺序（已执行完毕，留档）

1. BUG-CX-04 的死锁路径（①）+ BUG-CX-01 的冷启动死路——两个"卡死"先修；
2. BUG-CX-02（落锁手势——S0 进口）；
3. BUG-CX-03（protocol #s2 重写——S2 主干引导）；
4. CX-04 余项 + CX-05 + CX-06（文案与模板补注，可一批清完）。

## 与既有裁决的关系

- CX-02 的人闸执行权问题与 L3-S1 v4.5.1/v4.6.1 已入档的信任边界（仓库层无法区分 actor）同源——本 BUG 只要求补文档侧意图白名单，不要求新机器强制。
- CX-05 的 UI impact 一致性校验是 D2 补洞，与 v4.4.0 的 preflight 方向一致。

> **2026-08-17 二轮修复状态**：BUG-CX-07..10 同批修复完毕（fixed）。CX-07 的决定性验证：`TestS2ToS11_HookDrivenCleanPath` 首次真实 PASS——修复过程中发现该测试此前一直被 product-blocker skip 掩盖（skip-as-pass 的测试卫生问题另记），去除 fixture 手工 documents[] 播种后 S2→S11 全程 hook 自动推进。全量测试 + validate --all + doctor 绿。

> **2026-08-18 S5 落地（四批次，L3-S5 v4.2.1 §8 计划逐行执行）**：CX-13→批次 A（6749b1d，S2 design 出链+fixture 去播种，spine 真实 PASS）；CX-11→批次 B（daa7a07，别名槽删除/活锁失效/漂移前置筛/author 降级如实）；CX-12→批次 C（9a97e58，envelope-first 三步叙事/findings-only/词汇合一/模板即教师+C5 分叉锁）。三份 BUG 全部 fixed；全量 -count=1 绿、validate/doctor 绿。执行中测试现场抓出三处首版错误（裸字符串失效字段/ARCH- 前缀/模板占位类型）——先红后绿纪律持续兑现。
