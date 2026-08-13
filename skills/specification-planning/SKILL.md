---
name: specification-planning
description: Use when designing architecture, final UI design packages, contracts, and task decomposition in the planning phase
category: methodology
version: 1.1.0
---
# Specification Planning

## Authority
The locked REQ is the baseline. Design, contracts, and tasks must trace back to it. Runtime authority lives in `docs/loop-definition.json`; stage contracts live in `docs/agent-protocol.md`; the method summary is inlined below.

## Entry Conditions
- The Loop is in `planning` (any phase: initialize, design, ui_prototype, contract_drafting, task_drafting, rework).
- The locked REQ is readable and its fingerprint matches the runtime baseline.

## Required Inputs
| Input | Path / field | Why |
|:---|:---|:---|
| Locked REQ | runtime `bound_req.path` | source of acceptance criteria and scope |
| Existing design | `docs/design/**` | reuse and conflict detection |
| Module current truth | `docs/design/prototypes/<module>/{index.html, stories.md, flows.md, scenario-model.json, cases.json, scenario-coverage.json, fixture-contract.json, *.html}` | current module package and prototype gate input |
| Rules | `docs/rules/*.md` | naming, security, api-design, state-machine constraints |
| Loop Definition | `docs/loop-definition.json` | planning exit transition and executable guards |

## Procedure
1. Read the locked REQ end-to-end; extract acceptance criteria, scope boundaries, and non-goals.
2. Draft the architecture: components, data model, state machines, data flow. Record decisions as ADRs under `docs/design/decisions/`.
3. Determine UI impact. If the REQ changes a user-visible surface, complete the UI design package before locking affected contracts; otherwise record that UI impact is `none` and continue.
4. For UI-impacting or behavior-changing work, read the affected module's complete current package before editing. Reconcile the REQ into `index.html` + `stories.md` + `flows.md` + the four scenario JSON files + page HTML files. REQ is a `source_refs` input, never a package owner. The current implementation IS factual context; do not create a per-REQ, per-round, or versioned copy.
5. Derive and validate `Rule → CASE → Story → PATH → Spec → Evidence` before contracts: require explicit allow/reject branches, 100% required branch coverage, module profile ratio, synthetic fixture cleanup, and full-module regression.
6. Draft contracts in order: FE-contract → BE-contract → SYNC-contract. Each must link to the REQ source ref and the module current-truth package.
7. Decompose into TASKs: each TASK binds one contract, has a Closing Contract (forbidden paths + required evidence), and obeys single-responsibility. Verify no TASK spans two contracts and that scenario-bearing TASKs name the module regression sweep.
8. Check the actual `TR-002 planning_ready` contract: at least one locked contract and one complete TASK must exist, with the scenario and planning evidence current. Request `TR-002` only when that contract is satisfied.
9. If document verification returns `document_fix_required` (`TR-004`), repair the affected documents. Re-open an architecture or UI decision only when verification evidence shows that the decision itself is invalid; otherwise keep rework bounded to the flagged contract or TASK.

## Outputs
- Architecture and ADR records under `docs/design/`.
- Locked current module UI/scenario package (when UI impact or behavior is changed) with fingerprints for the scenario four-pack, HTML prototype, stories, flows, and module spec.
- FE/BE/SYNC contracts with REQ and design traceability.
- TASK batch with Closing Contracts and single-responsibility assignments.
- Current planning evidence required by `TR-002`.

## Exit Conditions
- The planning checkpoint is committed and the Loop transitioned to `document_verification`.

## Stop Conditions
Stop immediately and surface to the human if any of:
- The REQ is ambiguous about a core acceptance criterion.
- A design decision conflicts with an immutable rule and cannot be resolved without REQ clarification.
- UI design package review is rejected by the human.
- User story, user flow, and prototype contradict each other in a way that changes product semantics.
- A contract cannot trace to a REQ clause.

## Non-Goals
- Do not implement code (that is the Builder Agent's job).
- Do not verify contracts (that is `document-verification`).
- Do not create the Agent Team (that is `team-planning`).

## Inlined Methodology

Planning is a single executable Loop phase (`planning.design`), not a document-production state machine. Architecture, optional UI design, contracts, and TASKs are work products within that phase; they do not each require a runtime transition. `TR-002 planning_ready` is the only planning exit and evaluates the current planning package. `TR-004 document_fix_required` returns failed documents to planning for evidence-bounded rework. The TASK lifecycle remains `candidate -> reviewed -> locked -> in_progress -> review -> done`; contracts and TASKs must trace back to the locked REQ. Invalid or guard-failing events do not change state, execute no side effect, and record a rejected event. Idempotency uses CAS revision checks; one committed transition per runtime revision.
