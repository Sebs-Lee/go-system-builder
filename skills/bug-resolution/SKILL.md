---
name: bug-resolution
description: Use when S7 produces an investigation-ready Finding batch, or when an approved RepairContract is ready for S9 execution
category: methodology
version: 2.0.0
---
# S8 Investigation and Bug Projection

## Authority

Runtime legality is defined by `docs/control/loop-definition.json`; the S8/S9 stage contract is defined by `docs/control/agent-protocol.md`. This Skill explains how an Agent supplies the missing Case/Contract facts and does not override either authority.

S8 has one investigation chain:

```text
sealed S7 Finding / ObservationBatch
  -> reversible InvestigationCase
  -> HypothesisResults and CausalModel
  -> approved RepairContract
  -> S9 repair
```

- S7 `Finding` and its encounter/raw evidence are immutable facts. S8 consumes them; it does not rewrite them.
- `InvestigationCase` is the S8 authority. It owns grouping, split/merge revisions, hypotheses, evidence, causal reasoning, blast radius, detection gap, and route.
- `RepairContract` is the only repair authority. S9 executes its approved revision and hash.
- A canonical BUG is a human-readable compatibility projection created **after** `RepairContract` approval. It is never an S8 intake, grouping, approval, or repair prerequisite.

## Entry Conditions

- A sealed S7 `ObservationBatch` or an exact Finding set is available.
- Each Finding has an immutable identity/hash, baseline, expected/observed facts, and the recorded boundary or code-inspection scope needed for investigation.
- The Loop is in S8/`bug_resolution`.
- Existing Cases and approved Contracts have been checked for overlap and duplicates.

An incomplete but safe S7 batch enters S8 with explicit `capture_gaps`; do not silently turn a gap into a reproduction task. If the Finding set is not sealed or cannot be identified exactly, stop intake and repair the S7 handoff first.

## Required Inputs

| Input | Use |
|:---|:---|
| Finding exact set and hashes | Bind the Case without changing S7 facts |
| Encounter/raw evidence | Read `last_good -> wall_action -> first_bad`, timeline, state delta, side effects, and terminal state before considering reproduction |
| Baseline and traceability refs | Bound the affected implementation/specification surface |
| Existing Cases/Contracts | Detect duplicate, overlap, split, or follow-on work |
| S7 capture gaps and detection claims | Carry known uncertainty into the causal and repair contracts |
| Finder/checkpoint availability | Continue with the original investigator or an authorized replacement |

## Evidence Policy

1. **Do not reproduce the S7 symptom by default.** First consume the frozen encounter, raw evidence, state deltas, and code-inspection trail.
2. Request a `FindingSupplement` only when a discriminator is missing and a new observation can distinguish competing hypotheses. The supplement is append-only and must reference the Finding and the discriminator it answers.
3. If the discriminator cannot be obtained safely, record the gap and route `investigate_more` or `human_req_change`; do not manufacture certainty by repeating the same journey.
4. S8 owns root-cause investigation and detection-gap reasoning. S7 owns claim completeness and observation capture. S8 may use S7 gaps as facts, but does not rerun the S7 review protocol merely to make the intake look complete.

## Procedure

1. **Ingest and bind.** Create or resume one `InvestigationCase` from the exact Finding set, ObservationBatch hash, and baseline. Reject changed or ambiguous input. Do not create a BUG at this point.
2. **Group reversibly.** Group by a shared candidate mechanism and compatible repair boundary, never merely by file, symptom wording, or agent ownership. Record the grouping rationale. Split or merge by Case revision while preserving prior hashes and the exact Finding coverage.
3. **Form competing hypotheses.** Each hypothesis must state the invariant or authority it may violate, the evidence that would support/refute it, and a discriminator. Do not write a root cause as an untested conclusion.
4. **Dispatch independent questions.** Assign sub agents by evidence question, not one agent per Finding. Allocate a unique `assignment-*` id, register it on the Hypothesis, then run `runtime investigation dispatch` so the id is bound to the Investigator workgroup/Task/Agent lifecycle. An assignment includes Finding IDs, boundary refs, hypothesis, discriminator, read/command scope, expected evidence, stop condition, and forbidden product changes. On ordinary `plan_checkpoint`, the Investigator fills the Hook's exact professional HypothesisResult questionnaire and saves it; do not ask for a duplicate generic `PLAN_REPORT`. An explicit `plan_approval_required` assignment still waits for its original approval. Workers submit evidence and `HypothesisResult` only, using the same bound Assignment id.
   In packet mode, the Hook automatically prepares each eligible registered
   question and supplies its exact view and answer paths. The Investigator
   fills only the professional result fields and saves with Write/Edit/
   MultiEdit; the original Case consumer generates the receipt and updates
   Case facts. No professional result is prefilled. New hypothesis content
   receives machine IDs and still awaits Main dispatch. Exact retry compares
   saved bytes; changed answers conflict. Do not manually issue/consume or
   submit direct result flags on this normal path.
   Terminal `consume --agent` and legacy result flags cannot bypass an issued
   question. A received personal Assignment does not approve the Case route
   or RepairContract.
5. **Build the causal model.** The supported model must connect `trigger -> violated invariant -> faulty mechanism -> propagation -> symptoms`, and must state blast radius and detection gap. Every source Finding is explained, explicitly split, linked as duplicate, or routed as unresolved; unexplained Findings cannot be hidden in prose.
6. **Choose exactly one route per Case.** Use only:
   - `s9_repair` — implementation, contract, test, tooling, configuration, database, dependency, or environment repair is required;
   - `s2_spec_rework` — the specification/design is internally inconsistent or incomplete;
   - `human_req_change` — the requested behavior or requirement must change;
   - `s7_no_change` — expected behavior, test-only false alarm, or transient condition with no product/spec change;
   - `investigate_more` — evidence or discrimination is insufficient;
   - `duplicate` — follow an existing canonical Case/Contract.
7. **Create and approve the RepairContract.** For `s9_repair`, bind the exact Finding set, violated invariant, causal model, repair scope, forbidden scope, compatibility/migration/rollback expectations, symptom assertions, root-invariant assertion, detection-gap assertion, and stop/escalation conditions. Main performs the technical review; authority is either explicit human approval or a current bounded-repair delegation (docs/bounded-repair-autonomy.md). Architect review is conditional on a cross-module, source-of-truth, or contract-boundary change.
   When neither a pinned project repair policy nor a valid bounded grant applies, first register a `human_decision` evidence artifact scoped to `s8_contract_approval:<runtime_id>` and record the exact draft SHA-256 in that decision artifact. The approval handoff is then the CAS command `runtime investigation contract approve --case-id <case> --file <draft> --approved-by <actor> --approval-hash <sha256> --approval-evidence-id <evidence-id>`. It must succeed before S9 work begins; it writes the approved Contract hash, creates the next immutable Case revision, and advances to `bug_resolution.repair_readback`.
8. **Project and hand off.** Only after approval, generate the canonical BUG compatibility projection and the S9 work package idempotently from the Case and RepairContract. S9 reads the approved Contract hash and does not rediscover the root cause. A projection failure must not roll back the approved investigation; retry the projection from the same Case/Contract revision.
   The legacy unissued file-based S9 CLI route is a compatibility/diagnostic path, not the normal packet-enabled Builder procedure. It can retain the original Session/Plan/Assignment, execution barrier, result, Changeset, Impact, independent Targeted and Handoff commands. In the packet-enabled normal path, the Hook supplies the PlanReport and RepairResult exact answer paths; the Builder saves only those professional answers and the original consumers/adapter update lifecycle facts. Do not submit a separate generic PLAN_REPORT or generic completion_report, manually issue/consume a questionnaire, or run the execution-begin command after the graph has advanced automatically. Main still performs candidate integration, required checks, one independent Targeted result per approved Contract, and Handoff gates. If the questionnaire path is absent, report a Main/system preparation gap instead of guessing or falling back to the legacy CLI.

## Blocking and Recovery

Every blocked result must expose all four fields:

```text
BLOCKED: <short reason>
Missing: <field, evidence, or authorization>
Next: <one concrete command or action>
Verify: <status/readback command or artifact>
```

Use these recovery rules:

- **Unsealed or changed Finding set:** stop before Case creation; return to S7 to seal or re-export the exact batch, then verify the batch hash.
- **Missing discriminator:** append a `FindingSupplement` for the named discriminator; if it is unsafe or unavailable, route `investigate_more` instead of repeating the symptom.
- **Original finder unavailable:** resume from the saved Case/checkpoint with an authorized replacement investigator; do not discard the Case or require a fresh reproduction.
- **Case revision conflict:** reload the latest Case, preserve its source Finding exact set, and submit a new revision; never overwrite a newer revision.
- **Unexplained Finding or unsupported causal edge:** keep the Case open, add a targeted hypothesis assignment, and verify that every source ID is covered before Contract approval.
- **RepairContract incomplete:** revise the Case/Contract and run the completeness check again; do not create a BUG or dispatch S9.
- **Route requires specification or requirement change:** set the corresponding route, preserve evidence and causal work, and hand off to S2 or the human decision point.
- **S9 projection or dispatch failure after approval:** retry the idempotent projection using the approved Contract hash; do not create a second Case or alter the approved Contract.
- **S9 targeted failure:** read the failed TargetedReverification from `runtime repair status`. For `fail_same_cause`, `fail_new_cause`, `scope_changed`, or `stale`, use the exact `investigation route ... --reassessment-evidence <targeted-path>` action shown by `next_action`; it is the same Case's causal reassessment revision, not a new BUG or an S9 retry loop. For `blocked`, resolve the environment/authority blocker, run `runtime repair targeted resume --actor <actor> --reason <resolution>`, and submit a new independent targeted result first; this path does not create causal evidence.

## Outputs

- A versioned `InvestigationCase` with exact Finding coverage, grouping history, hypotheses, evidence, causal model, blast radius, detection gap, and route.
- For `s9_repair`, an approved `RepairContract` with revision/hash and S9 handoff.
- For other routes, the route rationale and the required next owner/action.
- A canonical BUG only as a post-approval compatibility projection.

## Stop Conditions

Stop and surface the named route when:

- the evidence cannot support a causal model;
- a requirement or specification decision is needed;
- a Finding remains unexplained;
- the proposed repair only hides a symptom or changes the source contract without authorization;
- a stale revision, integrity failure, or missing authority cannot be recovered through the rules above.

## Exit Conditions

- Every source Finding is covered by an InvestigationCase disposition without rewriting the immutable observation.
- A Case routed to `s9_repair` has an approved RepairContract with a causal model, bounded blast radius, detection-gap assertion, symptom assertions, and exact source Finding coverage.
- A Case routed elsewhere records one route, its reason, the next owner, and the evidence needed to resume.
- S9 receives the approved RepairContract hash; it does not receive an open-ended BUG report or rediscover the root cause.

## Non-Goals

- Do not create a rich BUG before the Case and RepairContract exist.
- Do not use a BUG as the source of truth for grouping, root cause, approval, or repair scope.
- Do not default to E2E reproduction when S7 encounter/raw evidence is available.
- Do not let S8 repair product code or let S9 reinterpret the causal model.
- Do not count targeted re-verification as a new clean S7 round.

## Repair authorization preflight

Before requesting approval, cross-check each repair unit's implementation and test files against prospective_scope and forbidden_scope. The approval path checks declared unit scopes using the same containment rule as S9; reviewers must additionally identify necessary but undeclared wiring, shared types, UI mounts and test files. Do not turn a scope mismatch into a component-local copy of a business rule.

Define actual registered entry-point checks for missing publication/wiring, compatibility and rollback evidence, and a negative control that fails when the connection is removed. A test asserting the production entry is absent documents a finding, not a successful repair. Existing S8 evidence is reusable only within its version, scope and role provenance; matching assignment IDs alone do not make a substituted platform role equivalent.

## Complete repair scope before approval

Trace every promised slot family through the real production entry and transaction before finalizing prospective_scope. A representative helper test cannot establish every-family reachability. Record necessary persistence/API/shared-type files in the existing unit scope and bind a red/green check to the real boundary. Separate a wiring-removal negative control from a clean historical baseline comparison; reverting one file while retaining other repair edits is not a pre-repair baseline. Matching failure names/counts alone does not establish matching causes. Explain inconsistent comparisons before claiming no regression. Do not split atomic save+clear operations into independent transactions to fit an underspecified scope.

## Optional bounded repair delegation

When a project repair policy was explicitly pinned at REQ bind, or an explicit current human grant exists, follow docs/bounded-repair-autonomy.md: register repair_contract_review using the Driver identity and pass --delegation-evidence-id to contract approve. Do not label this technical review human_decision or sign as the user. Use `binding-policy:<pinned-sha>` for the pinned-policy authority reference. If neither authority exists, the existing human approval path remains required. Budget, expiry, exact scope and semantic review remain mandatory; ordinary recoverable technical problems are handled by the Driver before requesting any missing authority.

## Issued S9 PlanReport adapter

For a registered repair Builder, the Hook supplies the PlanReport view and
answer paths. On ordinary `plan_checkpoint`, fill only the professional
`answer.plan`, `answer.red_checks`, and `answer.proposed_paths`, then save. The
original PlanReport consumer validates Session/Plan/Assignment facts and the
existing adapter projects that same answer onto the lifecycle checkpoint; do
not send an additional generic PLAN_REPORT. The receipt records that this
planning responsibility was accepted, not that implementation is complete.
After every required Assignment report is accepted, the original graph/listener
advances to implementation. Do not manually run `runtime repair execution begin`
on the normal path. Explicit `plan_approval_required` keeps its original
approval. An exact answer replay
returns the original fact; changed answers conflict. Missing answer paths are
a Main/system preparation gap, not a reason to fall back to manual issue/consume.

## Issued S9 RepairResult after implementation

When the original S9 graph reaches `fixing`, the Hook prepares the
post-execution RepairResult questionnaire and gives its exact view and
`answer_path`.
The Builder fills only the professional verdict, before-fix results, unit
results, detection-gap evidence, scope deviations, risks, migration and
rollback fields, then saves through the platform PostToolUse Hook. Do not
author machine changed-path or check facts, infer responsibility `done` from green checks,
or use terminal `stage-answer consume --agent` for an issued answer.

The original RepairResult consumer validates the professional answer; the
original candidate Writer checks actual changed paths and required checks and
projects that same answer onto the completion record. The existing
`work_complete=true` field projects only the Builder's current responsibility
to `done`; Main still performs the original
workspace integration, all required Main checks and final RepairResult
Writer. Keep the Worker source until that path completes. The pre-execution
PlanReport receipt remains `work_complete=false`; it does not replace the
post-execution answer. When no typed Result packet was issued, preserve the
existing `repair-result.json` and `s9_delivery_command` workflow.
