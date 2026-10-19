---
name: specification-planning
description: Use when designing architecture, final UI design packages, contracts, and task decomposition in the planning phase
category: methodology
version: 2.2.2
---
# Specification Planning

## Authority
The locked REQ is the baseline. Design, contracts, and tasks must trace back to it. Runtime authority lives in `docs/control/loop-definition.json`; stage contracts live in `docs/control/agent-protocol.md`; Project Design Foundation lives in `docs/rules/design-foundation.md`; the method summary is inlined below.

## Entry Conditions
- The Loop is in `planning` (phases advance design → contracts → tasks via PTR-PLAN-01/02 and TR-002).
- The locked REQ is readable and its fingerprint matches the runtime baseline.

## Required Inputs
| Input | Path / field | Why |
|:---|:---|:--|
| Locked REQ | runtime `bound_req.path` | source of acceptance criteria and scope |
| Existing design | `docs/design/**` | reuse and conflict detection |
| Design Foundation | `docs/design/DESIGN.md` **§0 Next-agent card first**, then `design-language.md` and the Surface row if needed | inherit interaction/visual constraints when UI impact is `changed`; do not invent brand from a library hex; Foundation does not replace architecture |
| Derivation template | `docs/design/derivation/DERIVATION-template.md` | write `docs/design/derivation/REQ-<id>.md` before expanding HTML |
| Module current truth | UI path: `docs/design/prototypes/<module>/{index.html, stories.md, flows.md, scenario-model.json, cross-matrix.json, cases.json, scenario-coverage.json, fixture-contract.json, *.html}`; service-only path in the same directory: `scenario-model.json`, `fixture-contract.json`, `cross-matrix.json`, `stories.md`, `flows.md`, generated `cases.json` and `scenario-coverage.json` | current module behavior package; HTML/PATH/browser spec are required only for applicable browser cases |
| Rules | `docs/rules/*.md` | naming, security, api-design, state-machine, design-foundation constraints |
| Loop Definition | `docs/control/loop-definition.json` | planning exit transition and executable guards |

## Procedure — dual-track convergence (v2.0.0)

**Step 0 — route by UI impact before doing anything else**: read the bound
REQ's top `UI impact` field. `unknown` → stop, resolve it in the REQ's §D
first (the `ui_impact_resolved` guard blocks PTR-PLAN-01). `none` → skip
Foundation derivation and UI assets (page HTML, `index.html`, PATH, and browser
specs). If any AC targets an FR, still run the behavior steps and build/update
the module's 7-file service package: `scenario-model.json`,
`fixture-contract.json`, `cross-matrix.json`, `stories.md`, `flows.md`, and the
generated `cases.json` and `scenario-coverage.json`. Use an API caller/worker as
the story role, an F-NNN operation flow without PATH, and
`browser_required:false` on each service branch. Only when all ACs have an
endorsed NFR or §A4 target and no module behavior changed may the scenario
package work be skipped. Also update the same package when existing service
behavior truth changes; if the bound REQ has no FR to source that change,
resolve it through REQ amendment before locking the package. `changed` → run
the full UI and behavior flow below.
Do not treat `none` as “no design”: locked ARCHITECTURE remains required, and
`none` never removes an FR-targeted CASE.

**Step 0.5 — consume Project Design Foundation (`changed` only)**: read
`docs/design/DESIGN.md` **§0 Next-agent card** first, then the current `SUR-*` diff
named by the REQ, then `design-language.md` only if the card cannot answer
Must not. The card must already answer action role (no hex), value **and**
status-tag color, and which commitments are buttons vs sentences. Do not treat
a hex or library Primary from a prior construction note as the brand. Do not
open `design-language.md` on Core+thin when the card already answers Must not.
If Foundation is `local` (investment tier `local`) → write
`docs/design/derivation/REQ-<id>.md` with `Foundation: local` and module-local
decisions, without publishing a project Foundation. If status is `provisional`
/ `draft` / `in-review` / `superseded` or confirmation is PENDING, it is not a lock (only `published` with dates recorded is). If Foundation is missing,
draft, in-review, provisional, superseded, lacks §0, or cannot cover this surface, stop and load
`skills/design-foundation` — do not batch-generate pages. inherit must not
add a new information posture (KPI dashboard) without `extend`. Otherwise write
`docs/design/derivation/REQ-<id>.md` as a **referenced-ID consumption list**:
active `LAW/ANTI/INV` by stable ID + opening `GR-*`, `Must not` by source ID,
`Bindings` by `ROLE/PATTERN`/component path, `Proof` by `PROOF-*` with
resolvable path, one macro composition, one stress state; `inherit`/`extend`/
`exception` posture (extend → promotion candidate note; exception →
`docs/design/decisions/EX-*.md`). One macro composition and one stress HTML
state must exist before the remaining module pages. If any required ID, version
(`DESIGN.md@version`, `SUR-*@version`), or Proof path cannot be resolved,
stop at semantic fix — do not invent a skin. See
`docs/rules/design-foundation.md`. This step is a semantic duty, not a
PTR-PLAN-01 predicate. It does not replace the architecture track. The cold-start
handoff packet is only `§0 + SUR-* diff + current Derivation` (together
≤120 lines / ≤12 KB); later implementers read that packet, not the full thesis.

The S2 portion runs as two tracks with free ordering within one agent (not
subagent parallelism), converging twice. "Free ordering" means either track
may start first — inside Convergence 1 there is a hard order: **stories.md
must exist before rules/branches** (every branch's `story_refs` cites an
S-NNN; writing rules first means citing stories that do not exist yet).
Architecture constrains journeys; when a story challenges the architecture,
escalate that trade-off to the ADR human gate — never silently let one side
win.

**Track system (architecture → facts):**
1. Draft the architecture: components, data model, state machines, data
   flow. Record decisions as ADRs (decision/risk/consequence/alternative)
   under `docs/design/decisions/` — the single human gate of S2 is the
   ADR direction sign-off; the N/A list of the AC bridge joins the same
   sign-off package.
2. Land facts partitions from the architecture vocabulary.

**Track user (REQ §A → behavior-role stories):**
3. Write stories (S-NNN citing REQ-id) from the REQ's §A intent only. For a
   service-only package, express the API caller/worker's intent without
   inventing a screen or UI journey.
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
   records a no-branch reason — silence is not N/A. Machine floors (enforced
   by `scenario validate` / generate):
   - every declared fact and every story must appear in at least one cell
     (per-fact and per-story floors; the fact×story combinations themselves
     are your hunting judgment, **not** a cartesian-product requirement);
   - a branch cell must be a branch whose rule actually cites the cell's
     `REQ-<id>/FR-<id>` in `source_refs` (the matrix joins the model, it
     does not assert alongside it); `req_ref` may only reference the bound
     REQ;
   - a `no_branch_reason` is a rationale: name the why — at least 8
     characters including a letter (a bare "不需要"/"." is rejected;
     free-word escapes are not endorsed N/A).
5. Immediately run `.claude/bin/loop-harness scenario bridge --root .` — the
   AC source check (AC→FR→BR) needs no generated outputs; fix gaps now, not
   at close.

**Fixtures:**
6. After branches settle, write `fixture-contract.json` (synthetic data +
   cleanup) — data needs are branch-certain by now.

**Convergence 2 (operational flow):**
7. Write an F-NNN operation flow for every behavior path. Service-only
   branches use that flow without PATH, HTML, or Playwright and set
   `browser_required:false`. UI branches that require browser verification
   bind a PATH here and include the applicable prototype pages (4-field
   header); `browser_required:true` requires at least one PATH.

**Depth self-review, then close:**
8. Before closing, attack your own package in three roles — implementer:
   "which oracle can't I build or can't distinguish a wrong implementation
   by?"; e2e-tester: "which negative case can't I evidence across the
   seven oracle dimensions? (visible, terminal_state, persisted_effects,
   forbidden_side_effects — plus, for negative branches, rejection,
   expected_state, recovery)"; maintainer: "which rule will clash with
   module evolution?" Record the conclusion as one paragraph under a
   `## Depth Self-Review` heading in the ADR package. These are three views
   of one semantic review, not three reports. Reference generated counts and
   coverage instead of copying statistics into multiple documents; do not add
   direct-AC or count proofs. Profile positive/negative ratios are risk
   references, not quotas; retain every real required allow/reject branch.
9. When a module behavior package is required, close it with
   `.claude/bin/loop-harness scenario generate --module <module> --root .`
   then `scenario validate --module <module> --root .`; this validates the
   source package and generated outputs. Run `scenario bridge --root .` for
   the REQ-side AC check. The bridge requires every AC to reach a CASE or
   carry an endorsed N/A: a declared NFR id or an explicit §A4 negative-space
   pointer — §A4 is the REQ's "明确不做" table; free text is rejected as a
   silent removal from the verification denominator. If no module behavior
   package is needed because all ACs have endorsed N/A targets, run the
   bridge and omit only the package commands. Close the architecture side by
   flipping `ARCHITECTURE-<id>.md`'s top `状态` to `locked`, then close S2
   through the packet-enabled or packet-off legacy path in the
   Planning Gate Reception section below. The original design judgment and
   PTR-PLAN-01 gate remain authoritative.

**S3/S4 steps:**
10. For S3, converge the shared data model and SYNC protocol first, then derive
    the applicable FE/BE responsibilities from the four templates under
    `docs/dev/contracts/` (CONTRACTS / BE / FE / SYNC). Each contract traces to
    the bound REQ and applicable current-truth inputs. The CONTRACTS index
    coverage matrix is for this REQ's clause universe — one `{id} §{n}` cell
    per clause, matching the clause number declared by the target contract.
    S3 freezes only the locked REQ, registered S2 design, and conditional
    sources they reference; it does not use its own not-yet-authored CONTRACTS
    index as an issuance prerequisite. The Hook supplies the exact answer path
    and known or candidate output paths/templates. Create one index and the
    complete applicable contract set for the bound REQ. S3-created shared
    models/examples are outputs: saving the supplied answer runs the original
    shared-model checker, which discovers their dependencies and records them
    in the existing evidence subjects. Do not change an already frozen S2
    model. Set finalized contracts to `locked`, fill the existing professional
    answer. The original consumer reads the declared source view: commit Git-sourced professional outputs to
    the exact development branch named by the Hook before saving that answer.
    S3 outputs can change during the current entry; REQ/S2 inputs remain frozen.
    Save the answer once. The gate evaluates the full current set; PTR-PLAN-02 performs the original
    registration. Do not hand-create evidence, copy machine metadata, or run
    the same checks as ordinary steps. Run `contracts check` or `scenario
    bridge` only to diagnose a specific reported failure. See [S3](../../docs/control/agent-protocol.md#s3)
    for the current packet workflow; packet-off legacy handling is below.
11. Decompose into TASKs: each TASK binds one primary contract, declares
    its Delivered Clauses (§3) and Module Impact (§3.1), has a Closing
    Contract (§7 four assert lines), and obeys single-responsibility.
    Never hand-copy fingerprints or versions — runtime documents[] owns
    them; §2 keeps read order only.
    **拆分纪律（builder 视角）**：用一句话给出交付目标，按独立验收、责任和真实变更耦合拆分；只有实际依赖共享实现产物的下游任务才声明 foundation TASK 依赖。每个任务的 §2 清单引用必要条款切片，并标明 required/conditional/optional 与条件。结合共享读集、同文件写入、依赖和执行中的 compact/返工观察判断粒度；字节数、文件数或一次 compact 不能自动决定拆分。调整已批准批次时走原规划/审查/失效路径，并保留合同覆盖与可执行依赖。
    When indexing existing engineering choices, use the empty
    `docs/rules/project-conventions-template.json` to author the optional
    `docs/rules/project-conventions.json`. Reference the exact locked Rule or
    originally registered locked architecture/decision, its source section,
    hash/version, explicit path scope and exceptions. The default text must
    exist in that source section; do not invent a new business decision in an
    index. Examples are references to actual implementation/tests. New decisions
    still need the original architecture/change-control path before indexing.
12. For S4, complete the TASK batch and current REQ index at the exact paths and
    templates supplied by the Hook. Mark each TASK document `complete` or
    `cancelled`; these document values do not change a work responsibility's
    status. Commit declared Git-sourced outputs to the Hook-named branch, fill
    the existing S4 answer, and save it once. The original consumer runs
    `tasks check`, projects planning evidence, and evaluates the existing gate
    and `TR-002`; do not run those checks, create an evidence envelope, or
    recover Runtime state as ordinary authoring steps. Existing `tasks check`
    and `tasks check --json` output remains available for a specific diagnosis
    or professional review of granularity and source-backed advisories. Read
    concrete observations against the actual responsibility; byte counts are
    not token measurements or split thresholds. Unresolved assertions still
    require a professional oracle, and the analysis never executes them.
    `behavior_examples` only projects CASE IDs explicitly named in that TASK's
    Manifest Clauses against its current `cases.json`. It validates the original
    scenario source package and preserves inputs, oracle, fixture and source
    hashes; Mode and unknown conditions remain visible. Resolve `unresolved`
    questions through the original contract/design responsibility. A source
    example is not a product test result or approval of a new requirement.
    Inspect `task_conventions` with its source mode: use only proved
    `declared_default` entries and their exceptions; resolve candidate, stale,
    conflicting or migrating entries through the original owner. S4 authoring
    has no automatic document-registration proof. Absent/no matching registry
    permits ordinary authorized engineering choices, and never changes the
    original TASK coverage, gates or permission.
13. If document verification returns `document_fix_required` (`TR-004`),
    repair the affected documents. Re-open an architecture or UI decision
    only when verification evidence shows that the decision itself is
    invalid; otherwise keep rework bounded to the flagged contract or TASK.

## Planning Gate Reception（S2/S3/S4）

For S2, complete the actual applicable architecture/UI/service/ADR work. The
original responsibility-start/binding completion supplies the exact current
view/answer paths, full duties and conditional material pointers. Fill only the
existing `architecture_path` and `conclusion`; empty conclusion is draft. Normal
save records the original professional report and runs its original outlet.
Already recorded but unfinished S2 work continues on the same save without
re-registration; completed old-stage replay is read only. Ordinary Pre/Stop
only remind from current facts, without preparation or source reproof. Missing
actual material is reported precisely; pass alone cannot replace the design.
Main/Architect labels, packet equality and source hashes do not qualify authors.
Do not copy metadata, create envelopes, issue/consume or manually recover.

S3/S4 retain their existing questionnaires and professional duties. Read the
supplied exact view/answer paths and concrete missing items, then save the
professional answer through their original consumer; do not guess directories
or duplicate an envelope. The compatibility envelope example below is not an
S2 normal authoring obligation or a producer-name whitelist for S2.

For S4, the Hook also supplies the current REQ index path, TASK candidate paths
and templates, frozen source inputs, and the bound Git branch. Main decides the
task breakdown and professional contents, commits the declared outputs to that
branch, fills the existing answer once, and lets the original consumer perform
the checks and `TR-002` registration. Do not manually repeat `tasks check` or
recover a received answer; follow the precise missing items and next action in
the Hook feedback. `complete/cancelled` apply to TASK documents, not the shared
work-responsibility states.

The manual envelope path below is only for packet-off legacy Runtime flows
that have no current typed S2/S3/S4 Main packet. Do not use it to bypass or
substitute for an issued packet.

### Packet-off legacy evidence envelope

Each legacy planning gate (design/contracts/tasks) requires its original
JSON evidence envelope in addition to disk facts. Registering a Markdown report
does not provide `conclusion` and fails with `evidence:<id>:schema`. Use the
matching kind and responsibility in the template:

```json
{
  "schema_version": "1.0.0",
  "evidence_id": "planning-{design|contracts|tasks}-pass",
  "kind": "planning_design | planning_contract | planning_task",
  "runtime_id": "从 .claude/loop-state.json 顶部复制",
  "baseline_generation": "{当前 baseline generation——数字，如 1}",
  "producer_agent_id": "你的 agent id",
  "producer_responsibility": "Architect（S2）/ Contract Planner（S3）/ Task Planner（S4）——gate 按此词白名单，逐字匹配",
  "subject_refs": [],
  "conclusion": "pass",
  "created_at": "ISO 时间戳"
}
```

此 legacy 模板可将 `subject_refs` 留空，且不写 `review_round`。将文件保存为
`docs/reports/planning/{id}.json`；仅在 packet mode 关闭且不存在当前 typed packet
时注册：

```text
.claude/bin/loop-harness runtime evidence add --id planning-design-pass   --kind planning_design --path docs/reports/planning/planning-design-pass.json   --produced-by <你的 agent id> --responsibility Architect
```

For S3/S4 substitute `--kind planning_contract --responsibility "Contract Planner"`
or `--kind planning_task --responsibility "Task Planner"`, with IDs
`planning-contracts-pass` or `planning-tasks-pass`. Reusing an ID is rejected,
including an invalid prior entry; use an `-r2` suffix for a later legacy attempt.
The missing tokens remain `evidence:planning_design_record`,
`evidence:planning_contract_record`, and `evidence:planning_task_record`.
`evidence:<id>:schema` means the envelope fields and registration do not match.

## Outputs
- Architecture and ADR records (with the depth self-review paragraph and
  the endorsed-N/A list) under `docs/design/`.
- Design Derivation Note at `docs/design/derivation/REQ-<id>.md` when UI
  impact is `changed`.
- Locked current module package when needed: a nine-file UI package when
  UI changes, or a seven-file service behavior package when an FR-targeted
  AC has no UI impact. The service package has no HTML, PATH, or browser spec;
  its CASEs remain in the S2→S7 denominator. Module packages are current truth,
  not per-REQ copies.
- FE/BE/SYNC contracts with REQ and design traceability.
- TASK batch with Closing Contracts and single-responsibility assignments.
- Current planning evidence required by `TR-002`.

## Exit Conditions
- The planning checkpoint is committed and the Loop transitioned to `document_verification`.
- Next (NOT your job): dispatch two document-verifier subagents per `docs/control/agent-protocol.md #s5` — planning does not continue into S5; the activation envelopes should name any Triggered Deep-Dives (see the document-verification SKILL) whose conditions the REQ/contracts meet.

## Stop Conditions
Stop immediately and surface to the human if any of:
- The REQ is ambiguous about a core acceptance criterion.
- A design decision conflicts with an immutable rule and cannot be resolved without REQ clarification.
- A story challenges the architecture in a way that changes the ADR direction (escalate to the human gate).
- An AC cannot reach a CASE and has no endorseable N/A (category + pointer) — surface the criterion, do not free-text it away.
- A behavior-role story, operation flow, and applicable prototype contradict each other in a way that changes product semantics.
- A contract cannot trace to a REQ clause.

## Non-Goals
- Do not implement code (that is the Builder Agent's job).
- Do not verify contracts (that is `document-verification`).
- Do not create the Agent Team (that is `team-planning`).

## Inlined Methodology

Planning is an executable Loop phase machine (design → contracts →
tasks via PTR-PLAN-01/02), not a document-production state machine.
Architecture, optional UI design, contracts, and TASKs are work products
within those phases; they do not each require a runtime transition.
`TR-002 planning_ready` is the only planning exit and evaluates the
current planning package (registered contracts + the complete TASK batch). `TR-004
document_fix_required` returns failed documents to planning for
evidence-bounded rework. Runtime 1.3 stores the shared five-state responsibility
status on its Agent and TASK work-state fields. Legacy Runtime 1.1/1.2 TASK
lifecycle values (`candidate -> reviewed -> locked -> in_progress -> review ->
done`) remain only in their original compatibility format and are not rewritten
in place. TASK document `Status: complete` means only that the planning artifact
is finished. Agents follow the current Runtime/Hook projection and never author
a generic status. Contracts and TASKs must trace back to the locked REQ. Invalid
or guard-failing events do not
change state, execute no side effect, and record a rejected event.
Idempotency uses CAS revision checks; one committed transition per runtime
revision.

## Executable acceptance checks before locking TASKs

Before locking a check command, verify its selected test inventory, browser
support policy, runtime ownership, required dependencies, and known baseline
failures. A command named cross-browser is not evidence that the current
requirement is selected by those browser projects. Reuse unchanged baseline
receipts only when their commit/configuration and declared scope apply; do not
run the entire suite at every planning step. Unknown baseline health remains
unknown, never PASS. Browser coverage changes use the existing specification
change route, not a new approval workflow.

Manual debugging and integration must use the same project runner/environment
profile for comparable evidence. Do not replace a required custom reporter,
filter away failed tests, or hide the command exit code behind `tail`/`tee`.
Resource controls may limit concurrency without reducing test coverage.

## Production-path ownership before S4 close

For each user-visible or externally consumed capability, trace the real caller through the applicable UI mount, API wrapper, registered route, authorization, service and response projection. Reuse the existing TASK/contract traceability sections: name the owning TASK for each required connection, including changes to shared types. A service implementation alone does not establish a production consumer. For backend-only work start at its actual external caller; do not invent a UI obligation.

Cross-check the files needed to implement each connection against its planned write scope. Missing ownership or contradictory scope is unfinished planning; correct it before locking/dispatch, not by copying business rules into the allowed component. Retain one canonical authority for business decisions; client-side validation requires explicit compatibility and differential checks.

For critical capabilities name at least one check crossing the actual registered boundary and one counterexample (such as reverted wiring) that would make it fail. Counts, test names, a mock returning the desired response, and a test proving an entry is absent are not delivery evidence. Record coverage gaps explicitly; do not defer a TASK's own delivery obligations to S7. This is a semantic review obligation, not a claim that the current machine gate proves end-to-end correctness.


For new batches apply [shared-model contracts in the factory](../../docs/rules/shared-model-contracts.md) (after installation: [project rule](../../docs/rules/shared-model-contracts.md)): explicit policy, native source references, validated examples, committed design closure and role-specific TASK links. Do not create a registry or copy schemas per REQ. Independent generation may run in parallel; only actual shared implementation artifacts require foundation TASK dependencies.

## Overall dispatch plan (waves-v1)

Follow [factory dispatch rules](../../docs/rules/dispatch-plan.md), or after installation
[project dispatch rules](../../docs/rules/dispatch-plan.md). S4 delivers the REQ's
index as an ordered wave checklist; S5 reviews and freezes it with TASKs. S6 reads
that plan and `s6 status --capacity <actual total slots>` before dispatch, then
recomputes after integration or capacity release. Start every compatible ready
TASK within actual capacity; do not wait for unrelated tasks in an earlier wave.
Only verified integration into the root development branch releases consumers.
Never write runtime progress back into frozen plans or TASKs.
