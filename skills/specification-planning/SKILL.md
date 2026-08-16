---
name: specification-planning
description: Use when designing architecture, final UI design packages, contracts, and task decomposition in the planning phase
category: methodology
version: 2.0.0
---
# Specification Planning

## Authority
The locked REQ is the baseline. Design, contracts, and tasks must trace back to it. Runtime authority lives in `docs/loop-definition.json`; stage contracts live in `docs/agent-protocol.md`; the method summary is inlined below.

## Entry Conditions
- The Loop is in `planning` (any phase: initialize, design, ui_prototype, contract_drafting, task_drafting, rework).
- The locked REQ is readable and its fingerprint matches the runtime baseline.

## Required Inputs
| Input | Path / field | Why |
|:---|:---|:--|
| Locked REQ | runtime `bound_req.path` | source of acceptance criteria and scope |
| Existing design | `docs/design/**` | reuse and conflict detection |
| Module current truth | `docs/design/prototypes/<module>/{index.html, stories.md, flows.md, scenario-model.json, cross-matrix.json, cases.json, scenario-coverage.json, fixture-contract.json, *.html}` | current module package and prototype gate input |
| Rules | `docs/rules/*.md` | naming, security, api-design, state-machine constraints |
| Loop Definition | `docs/loop-definition.json` | planning exit transition and executable guards |

## Procedure — dual-track convergence (v2.0.0)

The S2 portion runs as two tracks with free ordering within one agent (not
subagent parallelism), converging twice. Architecture constrains journeys;
when a story challenges the architecture, escalate that trade-off to the
ADR human gate — never silently let one side win.

**Track system (architecture → facts):**
1. Draft the architecture: components, data model, state machines, data
   flow. Record decisions as ADRs (decision/risk/consequence/alternative)
   under `docs/design/decisions/` — the single human gate of S2 is the
   ADR direction sign-off; the N/A list of the AC bridge joins the same
   sign-off package.
2. Land facts partitions from the architecture vocabulary.

**Track user (REQ §A → stories):**
3. Write stories (S-NNN citing REQ-id) from the REQ's §A intent only.
   When evolving an existing module, read its complete current package
   first — new stories must not duplicate or contradict existing S-NNN.
   Stories seed behavior completeness; they do not depend on the
   architecture track.

**Convergence 1 (behavioral completeness):**
4. Draft rules/branches (oracle written with the branch — a branch's
   polarity forces its outcome thinking) by crossing facts × REQ demand
   points (FR) × stories, rejection paths included. `source_refs` cite
   `REQ-<id>/FR-<id>` so the AC bridge can resolve. The hunt's carrier is
   `cross-matrix.json`: every meaningful cell names its covering branch or
   records a no-branch reason — silence is not N/A.
5. Immediately run `loop-harness scenario bridge --root .` — the AC source
   check (AC→FR→BR) needs no generated outputs; fix gaps now, not at close.

**Fixtures:**
6. After branches settle, write `fixture-contract.json` (synthetic data +
   cleanup) — data needs are branch-certain by now.

**Convergence 2 (walkable):**
7. Write flows (F-NNN/PATH-*) and prototype pages (4-field header) as the
   convergence of stories' journeys and branches' behavior;
   `browser_required` branches bind their PATH here.

**Depth self-review, then close:**
8. Before closing, attack your own package in three roles — implementer:
   "which oracle can't I build or can't distinguish a wrong implementation
   by?"; e2e-tester: "which negative case can't I evidence across the
   seven dimensions?"; maintainer: "which rule will clash with module
   evolution?" Record the conclusion as one paragraph in the ADR package.
9. Close: `loop-harness scenario generate` then `validate` — ratio gates,
   reference existence, byte-frozen outputs, cross-matrix references, and
   the full AC bridge (every AC reaches a CASE or carries an endorsed
   N/A: a declared NFR id or an explicit §A4 pointer; free text is
   rejected as a silent removal from the verification denominator).
10. If `ui_impact` is `none`, skip tracks/convergences 3–9 and go straight
    to contracts (step 11); if `unknown`, stop — resolve it in the REQ's
    §D first (`ui_impact_resolved` guards PTR-PLAN-01).

**S3/S4 steps (unchanged):**
11. Draft contracts in order: FE-contract → BE-contract → SYNC-contract.
    Each must link to the REQ source ref and the module current-truth
    package.
12. Decompose into TASKs: each TASK binds one contract, has a Closing
    Contract (forbidden paths + required evidence), and obeys
    single-responsibility. Scenario-bearing TASKs name the module
    regression sweep.
13. Check the actual `TR-002 planning_ready` contract: at least one locked
    contract and one complete TASK must exist, with the scenario and
    planning evidence current. Request `TR-002` only when that contract
    is satisfied.
14. If document verification returns `document_fix_required` (`TR-004`),
    repair the affected documents. Re-open an architecture or UI decision
    only when verification evidence shows that the decision itself is
    invalid; otherwise keep rework bounded to the flagged contract or TASK.

## Outputs
- Architecture and ADR records (with the depth self-review paragraph and
  the endorsed-N/A list) under `docs/design/`.
- Locked current module UI/scenario package (when UI impact or behavior is
  changed) with fingerprints for the scenario package (now eight files incl.
  cross-matrix.json), HTML prototype, stories, flows, and module spec.
- FE/BE/SYNC contracts with REQ and design traceability.
- TASK batch with Closing Contracts and single-responsibility assignments.
- Current planning evidence required by `TR-002`.

## Exit Conditions
- The planning checkpoint is committed and the Loop transitioned to `document_verification`.

## Stop Conditions
Stop immediately and surface to the human if any of:
- The REQ is ambiguous about a core acceptance criterion.
- A design decision conflicts with an immutable rule and cannot be resolved without REQ clarification.
- A story challenges the architecture in a way that changes the ADR direction (escalate to the human gate).
- An AC cannot reach a CASE and has no endorseable N/A (category + pointer) — surface the criterion, do not free-text it away.
- User story, user flow, and prototype contradict each other in a way that changes product semantics.
- A contract cannot trace to a REQ clause.

## Non-Goals
- Do not implement code (that is the Builder Agent's job).
- Do not verify contracts (that is `document-verification`).
- Do not create the Agent Team (that is `team-planning`).

## Inlined Methodology

Planning is a single executable Loop phase (`planning.design`), not a
document-production state machine. Architecture, optional UI design,
contracts, and TASKs are work products within that phase; they do not each
require a runtime transition. `TR-002 planning_ready` is the only planning
exit and evaluates the current planning package. `TR-004
document_fix_required` returns failed documents to planning for
evidence-bounded rework. The TASK lifecycle remains `candidate ->
reviewed -> locked -> in_progress -> review -> done`; contracts and TASKs
must trace back to the locked REQ. Invalid or guard-failing events do not
change state, execute no side effect, and record a rejected event.
Idempotency uses CAS revision checks; one committed transition per runtime
revision.
