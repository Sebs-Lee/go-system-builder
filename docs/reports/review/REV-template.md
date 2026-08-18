# REV — Document Verification Record {runid}-{resp}

> 本模板两部分：§0 是必产物的**证据信封**（激活后第一件事就复制落盘——每个字段带一行"填什么"，写骨架的过程就是读懂自己要交什么）；§1 起是**条件产物**——只有出现 finding 才写的审查报告。双 PASS 不写报告（没有人读"都挺好"）。

## 0. 证据信封（必产物——document_review_record）

落盘到 `docs/reports/review/REV-{runid}-{resp}.json`，12 个字段全被机器校验：

```json
{
  "schema_version": "1.0.0",
  "evidence_id": "填写 REV-{runid}-{resp}（与文件名一致，机器互证）",
  "kind": "document_review",
  "runtime_id": "填写当前 runtime id（从 .claude/loop-state.json 顶部复制）",
  "baseline_generation": "填写当前 baseline generation（同上）",
  "producer_agent_id": "填写你的 agent id（你是谁就写谁——独立性机器核对两条证据互异）",
  "producer_responsibility": "填写 DV-SPEC-CONSISTENCY 或 DV-TASK-EXECUTABILITY（激活信封指定的职责，错值 gate 直接 Unknown）",
  "subject_refs": [
    "手动从 .claude/loop-state.json 的 documents[] 逐条复制 {path, version, sha256}——多一少一都拒。故意没有自动命令：逐条抄写就是'我签的是哪一版'的对峙，这一步的笨拙是审查的锚"
  ],
  "conclusion": "审查完成后回填，三选一：pass / fix_required / req_change_required（与 gate 同词，全流程没有第二套枚举）",
  "requested_event": "仅 fix_required 时填 document_fix_required（触发 TR-004 回 planning）；pass 留空；req_change_required 时填 req_change_required",
  "created_at": "填写 ISO 时间戳"
}
```

注：`review_round` 字段**不写**——S5 是轮 0，缺省即正确；误填反而静默失配。

## 1. Fingerprinted Inputs（以下仅在有 finding 时随报告填写）

| Kind | ID | Path | Version | SHA-256 |
|:---|:---|:---|:---|:---|
| REQ | REQ-{id} | `docs/requirements/REQ-{id}.md` | {version} | `{sha256}` |
| contract | {id} | `docs/contracts/{id}.md` | {version} | `{sha256}` |
| TASK | TASK-{id} | `docs/tasks/TASK-{id}.md` | {version} | `{sha256}` |
| module current truth | {module} | `docs/design/prototypes/{module}/` (scenario four-pack + stories/flows/index/*.html) | current | `{sha256}` |

## 2. Assigned Conclusion

| Scope | Expected behavior / criterion | Result | Evidence |
|:---|:--|:--|:--|
| {module/clause/path} | {criterion} | pass / fail / n/a（须记理由与证据） | {command/path/sample} |

N/A requires a recorded rationale and evidence.

## 3. Findings

| Finding ID | Severity | Location | Expected | Observed | Evidence | Canonical BUG |
|:--|:--|:--|:--|:--|:--|:--|
| REV-F001 | P0/P1/P2/P3 | `{path:line}` | {contract/REQ} | {fact} | {evidence} | BUG-{id} / pending / n/a |

Blocking findings cannot be repaired in this assignment. They enter
`.claude/skills/bug-resolution/SKILL.md`.

## 4. Checks

| Check | Command / Method | Result | Evidence ref |
|:--|:--|:--|:--|
| tasks check | `loop-harness tasks check`（覆盖/DAG 机器已判，消费其结论不重算） | pass / fail / blocked / not_run | `{ref}` |
| {check} | `{command}` | pass / fail / blocked / not_run | `{ref}` |

## 5. Result

```text
Conclusion: pass | fix_required | req_change_required
Requested lifecycle event: (none) | document_fix_required | req_change_required
```
