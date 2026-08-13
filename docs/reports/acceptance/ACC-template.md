# Acceptance Evidence: ACC-{id}

> Status: draft / passed / blocked / invalidated
> Runtime ref: `{runtime-id}@{revision}`
> Source REQ refs: REQ-{id} / none
> Accepted module current truth: `docs/design/prototypes/{module}/`
> Baseline generation: {n}
> Clean review round: {n}
> Clean-round evidence: `{review-evidence-ref}`
> PM / Architect: {name}

## 1. Fingerprinted Baseline

| Artifact | Path | Version | SHA-256 |
|:---|:---|:---|:---|
| REQ | `docs/requirements/REQ-{id}.md` | {version} | `{sha256}` |
| module current truth | `docs/design/prototypes/{module}/` (scenario four-pack + stories/flows/index/*.html) | current | `{sha256/N/A}` |
| contracts | `docs/contracts/CONTRACTS-{id}.md` | {version} | `{sha256}` |
| tasks | `docs/tasks/index.md` | {version} | `{sha256}` |

## 2. Clean Round

| Evidence | Workgroup manifest | Review round | Result | Validity |
|:---|:---|:---|:---|:---|
| Delivery verification | `{manifest/ref}` | {n} | PASS | current |
| QA | `{manifest/ref}` | {n} | PASS | current |
| E2E Browser | `{manifest/ref}` | {n} | PASS | current |
| open blocking BUGs | `{bug-index/ref}` | {n} | none | current |

All required dimensions must be PASS or evidence-backed N/A in the same round.

## 3. Requirement Acceptance

| REQ source_ref / Rule / CASE / Story / PATH / Spec | Expected | Evidence | Result |
|:---|:---|:---|:---|
| REQ-{id}/FR-{id} / BR-{id} / CASE-{id} / S-{id} / F-{id} / PATH-{id} / `web/e2e/{module}/*.spec.ts` | {behavior and oracle} | {REV/QA/E2E/test/sample} | pass / fail |

### Module scenario acceptance

| Gate | Expected | Evidence | Result |
|:---|:---|:---|:---|
| required allow branches | 100% | `scenario-coverage.json` | pass / fail |
| required reject branches | 100% | `scenario-coverage.json` | pass / fail |
| positive/negative ratio | meets `coverage_profile` | `scenario-coverage.json` | pass / fail |
| module regression | all current required CASE/PATH | E2E round {n} | pass / fail |

## 4. Delivery And Operations

| Item | Value | Evidence / Owner |
|:---|:---|:---|
| delivered scope | {modules/interfaces/config/data} | {TASK/commit} |
| deployment order | {steps} | {runbook/owner} |
| migration/data handling | {steps/N/A} | {script/evidence} |
| runtime verification | {health/critical path} | {command/result} |
| rollback | {method} | {owner} |
| operations handoff | {monitoring/alerts/manual controls} | {owner} |

## 5. Remaining Non-blocking Risks

| Risk | Severity | Owner | Tracking artifact |
|:---|:---|:---|:---|
| {risk} | P3 | {owner} | {REQ/BUG/TD} |

## 6. Decision

```text
passed / blocked
```

Acceptance does not authorize release. The next evidence is a release
architecture audit; final publication still requires human release approval.
