# E2E Tester Evidence: E2E-{id}

> Status: draft / pass / finding / req_change_required / release_blocked / invalidated (ReviewResult verdict vocabulary; submit via `runtime review-result submit`)
> Runtime ref: `{runtime-id}@{revision}`
> Review round: {n}
> Workgroup manifest: `{team-manifest-path}`
> Assignment: `{assignment-id}`
> Responsibility: E2E-USER-FLOW / E2E-CONSOLE-NETWORK
> Agent: `{agent-id}`
> Activation: `{activation-ref}`

## 1. Header

Module current-truth package + source REQ refs + review round + bound runtime revision + module spec path.

| Field | Value |
|:---|:---|
| module | {module} |
| source REQ refs | REQ-{id} / none |
| review round | {n} |
| runtime ref | `{runtime-id}@{revision}` |
| current scenario package | `docs/design/prototypes/{module}/` · `{sha256}` |
| spec chain | current module spec `web/e2e/{module}/` `{sha256}` / contracts `{sha256}` / TASK `{sha256}` |

## 2. Stack-State Table

Containers / ports / running versions — proves the environment the bundle was
executed against. **Observed** versions, not "expected".

| Component | Version | Port | Evidence |
|:---|:---|:---|:---|
| backend | {commit} | {port} | `{ref}` |
| frontend | {commit} | {port} | `{ref}` |
| db | {version} | {port} | `{ref}` |
| browser | {name/version} | n/a | `{ref}` |
| playwright | {version} | n/a | `{ref}` |

## 3. Test-Architecture Diagram

Agent assignments + journey sequence (one box per `test()`). Pair this diagram
with the module's current `flows.md`; every flow box must trace to a `F-NNN` /
`PATH-*` step and a CASE.

```text
[login.spec] → [fund-list.spec] → [fund-detail.spec] → [kyc-review.spec]
      │                │                    │                    │
   CASE-001 · F-001 · PATH-001  CASE-002 · F-002 · PATH-001  CASE-003 · F-003 · PATH-001  CASE-004 · F-004 · PATH-001
```

## 4. Real-Browser Flow Execution

Per-PATH walk table with step-level evidence (mirrors e2e-browser-testing
SKILL §Evidence Bundle and serves as the legal source for BLOCKED findings).

| Field | Value |
|:---|:---|
| executed path | PATH-{id}: {path title} |
| CASE / polarity | CASE-{id} / positive / negative |
| branch obligation | BR-{id} |
| entry point used | {declared PATH entry point} |
| direct URL used | no / yes: {declared reason from PATH} |
| flow file | `docs/design/prototypes/{module}/flows.md` |

| Step ID | User action performed | Expected visible result | Observed visible result | Result | Evidence |
|:---|:---|:---|:---|:---|:---|
| PATH-{id}-001 | {click/type/navigation from PATH} | {expected} | {observed} | PASS / FAIL / BLOCKED | `{screenshot/trace/video/log}` |

## 5. API CRUD Walk

`# · Method · Path · Status · Notes` with **observed** HTTP codes (not
expected). Every required CRUD operation the affected flows depend on must
appear here.

| # | Method | Path | Status | Notes |
|:---|:---|:---|:---|:---|
| 1 | GET | `/api/funds` | 200 | observed at {ts} |
| 2 | POST | `/api/funds` | 201 | observed at {ts} |

## 6. CDP Findings

Bugs caught only at browser level (silent 4xx, uncaught exceptions, console
errors). Each finding cites the JSONL line so a reviewer can replay.

| Finding ID | CDP event | URL / Route | JSONL line | Notes |
|:---|:---|:---|:---|:---|
| E2E-CDP-001 | `Runtime.exceptionThrown` | `/funds` | evidence/run-001.jsonl#L42 | {description} |

## 7. Console And Network

| Check | Expected | Observed | Result | Evidence |
|:---|:---|:---|:---|:---|
| console errors | none except documented benign warnings | {observed} | PASS / FAIL | `{log-ref}` |
| failed network requests | none for required flow | {observed} | PASS / FAIL | `{trace-ref}` |
| request/response contract | matches FE/BE/SYNC contract | {observed} | PASS / FAIL | `{trace-ref}` |
| required branch | allow/reject branch is covered | {observed} | PASS / FAIL | `scenario-coverage.json` |
| `visible` | every declared visible checkpoint observed | {observed} | PASS / FAIL | `{screenshot/trace-ref}` |
| `terminal_state` | declared terminal state observed | {observed} | PASS / FAIL | `{evidence-ref}` |
| `persisted_effects` | every declared persistence assertion holds | {observed} | PASS / FAIL | `{evidence-ref}` |
| `forbidden_side_effects` | every declared forbidden effect remains absent | {observed} | PASS / FAIL | `{trace-ref}` |
| negative `rejection` | declared rejection observed | {observed} | PASS / FAIL / N/A | `{evidence-ref}` |
| negative `expected_state` | state remains at the declared rejection state | {observed} | PASS / FAIL / N/A | `{evidence-ref}` |
| negative `recovery` | declared recovery succeeds, or sourced N/A is recorded | {observed} | PASS / FAIL / N/A | `{evidence/ref or source_refs + reason}` |

## 8. Evidence Validity + Status-Code Distribution

| Field | Value |
|:---|:---|
| module package fingerprint | `{sha256}` |
| review round | {n} |
| app fingerprint | `{sha256/commit}` |
| supersedes | `{evidence-id}` / none |
| invalidated by | `{impact-id}` / none |

Status-code distribution (sanity check on coverage, e.g.
`2xx: 47, 3xx: 3, 4xx: 2 (expected), 5xx: 0`):

| Class | Count | Expected | Notes |
|:---|:---|:---|:---|
| 2xx | {n} | {expected} | {notes} |
| 3xx | {n} | {expected} | {notes} |
| 4xx | {n} | {expected} | {notes} |
| 5xx | {n} | 0 | {notes} |

Evidence inventory paths:

```text
docs/reports/e2e/evidence/{RUN_ID}.jsonl
docs/reports/e2e/screenshots/{RUN_ID}-{step}.png
```

## 9. BUG Drafts Surfaced + Idempotent Re-Execution

BUG drafts surfaced from this bundle (one block per finding that should
become a canonical BUG; shape matches `docs/reports/bugs/BUG-NNN.md` so
filing is a copy-paste):

### BUG draft: {short title}
- 严重度: P0 / P1 / P2
- 关联 flow: F-NNN
- 关联 step: 步骤 N
- 关联 prototype: {file}.html
- 现象: {what happened, with screenshot ref}
- 期望: {what flows.md said should happen}
- 复现: {exact `pnpm playwright test --grep` command}
- CDP evidence: {JSONL path + line number}
- 关联代码: {file:line if CDP stack trace identified it}
- 建议根因: {hypothesis, marked as such — do not assert without code review}

Idempotent re-execution instructions — anyone running these commands must see
the same pass/fail pattern (timestamps differ; counts and verdicts match):

```bash
# Environment
export WEB_BASE_URL=http://127.0.0.1:58080
docker compose up -d   # if backend services needed

# Seed (per-persona tokens + test data)
pnpm run seed:test

# Run
pnpm playwright test web/e2e/{module}/ --grep 'CASE-|PATH-'

# Evidence lands at
ls docs/reports/e2e/evidence/*.jsonl
ls docs/reports/e2e/screenshots/*.png
```

## Result

```text
pass / finding / req_change_required / release_blocked
```

Requested lifecycle event: `{event}`

Findings are symptoms. They enter S8 finding investigation before any repair
Builder is activated.
