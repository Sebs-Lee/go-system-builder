---
name: bug-resolution
description: Use when investigating a sealed S7 Finding batch or an original S11 defect decision in S8, or executing an approved RepairContract in S9
category: methodology
version: 2.0.0
---
# S8 Investigation and Bug Projection


## Native S9 dispatch

Main passes the exact prepared S9 brief to Claude Code native Agent with worktree isolation. The Builder saves its professional PlanReport and real prechecks before implementation, continues repair/self-checks without activation or per-tool receipts, commits products and returns its professional Result. The professional PlanReport and Result are private outputs in the actual native worktree. The Main-root question is read-only input; an isolated Worker never writes Main. At Stop the adapter imports the actual worktree Plan payload into the existing questionnaire and invokes the original consumer, without Main authoring the answer. SubagentStop receives the original professional artifacts; Main integrates and runs actual necessary checks separately. An independent Targeted verifier receives its exact original answer path and returns it for Stop reception. No Team, workspace prepare/launch, generic readback, platform ID or session reservation is required. Existing installed workspace responsibilities keep their frozen recovery contract. Keep actual professional independence, Contract assertions and the fresh S7 exit. S8 retains its own investigation and approval responsibilities.

Main reads its own division/recovery questions, the exact native dispatch brief and accepted professional artifacts. Worker private Plan/Result questions remain that Worker's authoring responsibility; an accepted answer is not another fill task. After candidate reception Main performs the original integration and actual checks immediately, then reads the original Final/Impact. Do not wait for Final before integration or dispatch the completed Builder to fill another Result. A recorded blocker retains Main's original recovery question.


## Native S5 / S7 / S8 professional work

Main carries the exact original view and answer references into Claude Code native Agent. Read its actual frozen subjects, perform only your independent professional responsibility and save the original answer. Return that exact reference; SubagentStop calls the original consumer. Ordinary tools do not activate, deliver or require a receipt. No Team, opaque platform ID, workspace prepare/launch, generic readback/PLAN or begin is required for new work. Main retains professional division, ReviewPlan waves, Case routing and stage closure; real approval modes and actual independence remain. Frozen registered workspaces keep their original sources and recovery; never fall back to Main for a lost frozen execution. S5 checks the independent document responsibility; S7 answers its Claims/lens; S8 investigates its Hypothesis without product writes.

## Authority

Runtime legality is defined by `docs/control/loop-definition.json`; the S8/S9 stage contract is defined by `docs/control/agent-protocol.md`. This Skill explains how an Agent supplies the missing Case/Contract facts and does not override either authority.

S8 has one investigation chain:

```text
sealed S7 Finding / ObservationBatch OR original S11 defect decision
  -> reversible InvestigationCase
  -> HypothesisResults and CausalModel
  -> approved RepairContract
  -> S9 repair
```

- S7 `Finding` and its encounter/raw evidence are immutable facts. S8 consumes them; it does not rewrite them.
- Preserve registered evidence references across the handoff: `<id>` and `runtime:<id>` select the same indexed artifact with its original validity/generation/round/SHA checks. Do not register a prefixed alias or rewrite an accepted Finding to make the reference resolve. Runtime evidence does not substitute for an execution anchor where the original gate requires actual execution.
- `InvestigationCase` is the S8 authority. It owns grouping, split/merge revisions, hypotheses, evidence, causal reasoning, blast radius, detection gap, and route.
- `RepairContract` is the only repair authority. S9 executes its approved revision and hash.
- A canonical BUG is a human-readable compatibility projection created **after** `RepairContract` approval. It is never an S8 intake, grouping, approval, or repair prerequisite.

## Entry Conditions

- A sealed S7 `ObservationBatch` or the original human defect decision consumed by TR-027 is available.
- Each Finding has an immutable identity/hash, baseline, expected/observed facts, and the recorded boundary or code-inspection scope needed for investigation.
- The Loop is in S8/`bug_resolution`.
- Existing Cases and approved Contracts have been checked for overlap and duplicates.

An incomplete but safe S7 batch enters S8 with explicit `capture_gaps`; do not silently turn a gap into a reproduction task. If the Finding set is not sealed or cannot be identified exactly, stop intake and repair the S7 handoff first.

A human-source Case uses the original decision ID/path and the existing frozen
ReviewPlan product baseline. The system preserves that source through the original
Case revision history even when professional boundary descriptions evolve. Do not
re-register the decision, invent Findings or run S7 just to obtain IDs. With no
Findings, Finding explanation arrays and hypothesis source_finding_ids remain [].
The issued question supplies the exact report path and evidence metadata. Use the
original decision ID/path for source_boundary_refs; it identifies the reported
problem and does not prove that a discriminator executed. Investigate and record
actual observations normally. Missing or changed original source/baseline is a
Main/system recovery responsibility. A release rejection is not repair approval:
the original final-contract approval remains necessary before S9.

## Required Inputs

| Input | Use |
|:---|:---|
| Finding exact set and hashes | Bind the Case without changing S7 facts |
| Original S11 defect decision, for human-source Cases | Preserve the actual report without manufacturing a Finding or Batch |
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

The normal authoring contract follows L3-S8's 2026-10-17 minimum workflow.
Main sees the current professional questions in the original order: intake/grouping,
competing hypotheses and real dispatch, causal model/route, then Contract content
and genuine approval. Known IDs, hashes, revisions and source bindings come from
the system. Main and Investigator save their professional answers once; neither
manually issues/consumes, splits the same content into three route JSON files,
or assembles duplicate registration commands/reports. The listener supplies the
exact current questionnaire paths; fill the professional content and save for
the original consumers. A missing questionnaire or unresolved source is a
specific Main/system preparation gap to recover through the original flow.
Implementation and acceptance evidence are tracked in the S8 mechanism review;
final candidate, full S8 and platform acceptance depend on their recorded scope.

1. **Ingest and bind.** Main reads the sealed Batch and immutable Findings, or the original S11 defect decision, and supplies the intake/grouping rationale. The system binds or resumes the original `InvestigationCase` using that source and its real frozen baseline. Batch sources retain the exact Finding set and ObservationBatch hash; human sources retain the original decision with an empty Finding set. Reject changed or ambiguous input. Do not create a BUG at this point.
2. **Group reversibly.** Group by a shared candidate mechanism and compatible repair boundary, never merely by file, symptom wording, or agent ownership. Record the grouping rationale. Split or merge by Case revision while preserving prior hashes and the exact Finding coverage.
3. **Form competing hypotheses.** Each hypothesis must state the invariant or authority it may violate, the evidence that would support/refute it, and a discriminator. Do not write a root cause as an untested conclusion.
4. **Dispatch independent questions.** Main assigns investigators by evidence question, not one agent per Finding, and performs the real platform spawn/resume. The system reuses original Hypothesis registration and `runtime investigation dispatch` consumers to generate/bind the unique Assignment to the Investigator workgroup/Task/Agent lifecycle. An assignment includes Finding IDs, boundary refs, hypothesis, discriminator, read/command scope, expected evidence, stop condition, and forbidden product changes. These bindings are not questionnaire author fields. On ordinary `plan_checkpoint`, the Investigator fills the Hook's exact professional HypothesisResult questionnaire and saves it; do not ask for a duplicate generic `PLAN_REPORT`. An explicit `plan_approval_required` assignment still waits for its original approval. Workers fill only the existing professional `answer`; Hook listening does not perform or pretend to complete the platform spawn.
   In packet mode, the Hook automatically prepares each eligible registered
   question and supplies its exact view and answer paths. The Investigator
   fills only the professional result fields, saves that exact file, and returns
   the exact answer reference. Native SubagentStop invokes the original Case
   consumer, which generates the receipt and updates Case facts. No professional result is prefilled. New hypothesis content
   receives machine IDs and still awaits Main dispatch. For a lost acceptance
   response, Main reads the original receipt; native Agent return already uses
   that read-only path. It works without private answer/view files and never
   asks the Investigator to recreate accepted work. An actual resubmission
   still reads its saved bytes and compares the digest; changed answers
   conflict, and arbitrary Save replay after file deletion is not promised.
   Missing canonical Case/event proof remains a Main recovery issue; an
   unreceived packet is not completion. Do not manually issue/consume or
   submit direct result flags on this normal path.
   Terminal `consume --agent` and legacy result flags cannot bypass an issued
   question. A received personal Assignment does not approve the Case route
   or RepairContract.
5. **Build the causal model.** The supported model must connect `trigger -> violated invariant -> faulty mechanism -> propagation -> symptoms`, and must state blast radius and detection gap. Every source Finding is explained, explicitly split, linked as duplicate, or routed as unresolved; unexplained Findings cannot be hidden in prose.
6. **Choose exactly one route per Case.** Use only:
   - `s9_repair` — implementation, contract, test, tooling, configuration, database, dependency, or environment repair is required;
   - `s2_spec_rework` — the specification/design is internally inconsistent or incomplete;
   - `human_req_change` — the requested behavior or requirement must change;
   - `s7_no_change` — expected behavior, a false alarm, or a transient condition requiring no product/spec/test/config artifact change; use the existing route_reason to explain actual observations, why nothing needs changing, and what could disprove that judgment;
   - `investigate_more` — evidence or discrimination is insufficient;
   - `duplicate` — follow an existing canonical Case/Contract.
   Current unanswered questions block repair. Keep an earlier inconclusive result truthful; later accepted discriminators may resolve its evidence gap, and Main explains that reasoning in the original route_reason. Add a distinct question for new work; do not overwrite an accepted question/result or manufacture a new evidence reference when the actual source is unchanged. Only supported/refuted results for the exact current Hypothesis/Assignment count as that evidence. A new question or edited route reason alone does not unlock repeated investigate_more.
7. **Create and approve the RepairContract.** For `s9_repair`, bind the exact Finding set, violated invariant, causal model, repair scope, forbidden scope, compatibility/migration/rollback expectations, symptom assertions, root-invariant assertion, detection-gap assertion, and stop/escalation conditions. Main performs the technical review; authority is either explicit human approval or a current bounded-repair delegation (docs/bounded-repair-autonomy.md). Architect review is conditional on a cross-module, source-of-truth, or contract-boundary change.
   Before the real review, the system projects current Case revision, exact Finding set and other known machine fields into the final draft and supplies its actual path. Approval binds those final real bytes; do not auto-edit the approved draft afterward. When neither a pinned project repair policy nor a valid bounded grant applies, Main obtains a genuine human decision and the system records the original `human_decision` evidence scoped to `s8_contract_approval:<runtime_id>` with the final draft hash. Saving professional content does not approve it. Only after valid original authority exists may the system call the original `runtime investigation contract approve` CAS transaction; it must succeed before S9 begins and retains the approved Contract hash and immutable Case revision.
8. **Project and hand off.** Main supplies the current route's professional judgment and handles genuine human decisions. When the original route conditions and authority are met, the system calls the original consume/handoff; Main does not manually consume it. Only after approval, generate the canonical BUG compatibility projection and the S9 work package idempotently from the Case and RepairContract. S9 reads the approved Contract hash and does not rediscover the root cause. A projection failure must not roll back the approved investigation; retry the projection from the same Case/Contract revision.
   The legacy unissued file-based S9 CLI route is a compatibility/diagnostic path, not the normal packet-enabled Builder procedure. It can retain the original Session/Plan/Assignment, execution barrier, result, Changeset, Impact, independent Targeted and Handoff commands. In the packet-enabled normal path, the Hook supplies the PlanReport and RepairResult exact answer paths; the Builder saves only those professional answers and the original consumers/adapter update lifecycle facts. Do not submit a separate generic PLAN_REPORT or generic completion_report, manually issue/consume a questionnaire, or run the execution-begin command after the graph has advanced automatically. Main still performs candidate integration, required checks, one independent Targeted result per approved Contract, and Handoff gates. If the questionnaire path is absent, report a Main/system preparation gap instead of guessing or falling back to the legacy CLI.

## Blocking and Recovery

For an issued S8 answer, a blocked discriminator uses `result="inconclusive"`.
Use existing `answer.method` for actual attempts and unexecuted steps, and
`answer.observed` for observed facts, the concrete blocker, missing material and
Main's next action. Keep genuine evidence/boundary references; do not invent
execution, add blocker fields or deliver a second report. If mandatory source
material is unavailable, identify a Main/system preparation gap. Main retains
route decisions and the original consumers retain their completeness checks.

The following is the descriptive legacy/recovery message format, not additional
fields or a second deliverable for the normal issued S8 questionnaire:

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
- **Case revision conflict:** Main/system reloads the original current Case and preserves its source Finding exact set through the existing CAS recovery; authors do not calculate a new revision or overwrite a newer one.
- **Unexplained Finding or unsupported causal edge:** keep the Case open, add a targeted hypothesis assignment, and verify that every source ID is covered before Contract approval.
- **RepairContract incomplete:** revise the Case/Contract and run the completeness check again; do not create a BUG or dispatch S9.
- **Route requires specification or requirement change:** set the corresponding route, preserve evidence and causal work, and hand off to S2 or the human decision point.
- **S9 projection or dispatch failure after approval:** retry the idempotent projection using the approved Contract hash; do not create a second Case or alter the approved Contract.
- **S9 targeted failure:** preserve the failed original artifact and follow the precise current questionnaire/recovery feedback. Non-blocked failures return the same Case for original causal reassessment, not a new BUG or an S9 retry loop. A blocked verifier records actual evidence; after the blocker is resolved, the original Resume/Create/Commit consumers handle the new independent result. The explicit status/route/targeted-resume CLI commands remain legacy/recovery references, not another normal author submission. Missing normal paths belong to Main/system preparation; do not guess or fabricate PASS.

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

S9 Main ChangeImpact and independent Targeted forms keep their original flat
request DTO shape. Fill the supplied professional fields at the top level;
do not invent a `title/questions/answer` wrapper or a replacement form. A
rejected, unreceived authoring draft is corrected at its exact supplied path.
Fresh IDs and fresh files apply to new immutable registered artifacts, not
ordinary draft corrections. The existing consumers retain all checks.

## Original S8 human approval artifact

After a genuine human decision, the existing JSON approval consumer expects
these exact fields (substitute the current original references):

```json
{
  "decision": "approve_contract",
  "decision_id": "<fresh-evidence-id>",
  "runtime_id": "<current-runtime-id>",
  "case_id": "<current-case-id>",
  "contract_id": "<current-contract-id>",
  "approval_hash": "<sha256-of-the-final-presented-contract-draft>"
}
```

Register the completed decision with `runtime evidence add --kind human_decision
--id <same-fresh-evidence-id> --path <fresh-decision-file> --produced-by
<actual-approver> --scope-ref s8_contract_approval:<runtime-id> --root <Main>`.
Then the original `runtime investigation contract approve` transaction uses
that same evidence ID, actual approver and exact draft hash. A professional
save or Agent return grants no approval. Explicitly authorized disposable
simulations must label the decision as synthetic; they are not real human
consent for a product release.

A registered artifact is fingerprinted. Never correct it in place or register
different decisions against the same mutable file. A rejected decision remains
history; correct input uses a fresh decision ID **and a fresh file path**. This
is the existing approval format and source lifetime, not a dispatch handshake.

## Complete repair scope before approval

Trace every promised slot family through the real production entry and transaction before finalizing prospective_scope. A representative helper test cannot establish every-family reachability. Record necessary persistence/API/shared-type files in the existing unit scope and bind a red/green check to the real boundary. Separate a wiring-removal negative control from a clean historical baseline comparison; reverting one file while retaining other repair edits is not a pre-repair baseline. Matching failure names/counts alone does not establish matching causes. Explain inconsistent comparisons before claiming no regression. Do not split atomic save+clear operations into independent transactions to fit an underspecified scope.

## Optional bounded repair delegation

When a project repair policy was explicitly pinned at REQ bind, or an explicit current human grant exists, follow docs/bounded-repair-autonomy.md: register repair_contract_review using the Driver identity and pass --delegation-evidence-id to contract approve. Do not label this technical review human_decision or sign as the user. Use `binding-policy:<pinned-sha>` for the pinned-policy authority reference. If neither authority exists, the existing human approval path remains required. Budget, expiry, exact scope and semantic review remain mandatory; ordinary recoverable technical problems are handled by the Driver before requesting any missing authority.

## Issued S9 PlanReport adapter

Main reads the approved Contract and RepairPlan, selects actual specialized
Builders and independent verification work, and performs real native spawn/resume.
The system supplies exact personal view/answer paths and the current-step brief;
Main carries these into the native task without authoring or saving Worker answers.
Read the Main-root question as input. The native brief supplies private
PlanReport and Result paths relative to the actual native worktree. The Builder
personally saves `plan`, real `red_checks` and `proposed_paths` there before
implementation, then continues without a save receipt or another startup step.
Each precheck is `{name,command,result,evidence_refs}` with the actual attempted
command and honest `fail/pass/blocked` result. Cite real logs or existing
references; a blocked attempt or passing precheck does not require invented red
evidence. These commands record past work, not Main's later check selection.

Return the exact brief and absolute Result path. SubagentStop imports the private
PlanReport into the original questionnaire and invokes the existing consumers.
Main owns division, dependencies and necessary integration; Claude Code owns
worktree creation and execution. Do not write Main, wait for peer plans, manually
issue/consume, copy machine bindings or add a generic activation report. Missing
paths/environment go to Main as concrete blockers. Existing frozen executions
use their original [recovery contract](../agent-dispatch/frozen-execution.md).

## Issued S9 RepairResult after implementation

At the supplied Worker root, implement and self-check, or complete assigned
no-change tests, then personally save the exact RepairResult answer. Fill the
professional verdict, actual assertion observations, unit results, evidence,
scope deviations, risks and applicable migration/rollback fields. Known machine
and Git/check facts come from the system; green checks do not replace judgment.
No-change testing requires no new commit. Existing Contract or Git references may
support actual migration/rollback needs; nonempty “No migration” text does not
force two references or another report.

PlanReport and RepairResult record the two different professional steps before
and after work. Return the saved Result reference for native Stop reception,
without a second generic completion prerequisite. Actual changes still require real Git delivery, Main
integration and necessary Main checks; retain the Worker until Main receives it.
Main uses the supplied `runtime workspace integrate --root <Main-root> --assignment <execution-assignment-id> --agent <owner-agent-id> --check-command '<actual necessary Main check command>'`; root/assignment/owner come from the current execution, while Main selects actual necessary checks (repeat `--check-command` for multiple checks). This typed S9 path records real integration and Final; do not substitute generic `runtime task-integrate`.
No-change results still receive real required checks. If every accepted Final has no product change, Main keeps the supplied empty artifact set and change_types=[], explains the original Case/Contract disposition in the existing Impact decisions, and still arranges all necessary independent assertion observations and a fresh full S7 round. If the repair premise is no longer valid, return to the original S8 disposition instead. Do not invent a diff, commit or change type to advance. Save through any supported
editor, script or extension tool; do not send a separate completion report or
copy identities, hashes, revisions or check receipts into professional answers.

## S9 impact, independent verification and handoff

Actual Contract approval prepares the original Session/Plan and Main division
drafts. Explicit session startup followed by plan compile also prepares those
existing division drafts. Ordinary tools only read current responsibilities and
paths; they do not perform this preparation. After real Main integration, the
system prepares ChangeImpact only when the current repair facts reach that step;
otherwise it reports the actual remaining assignment work.

Main saves the exact ChangeImpact questionnaire with professional change types,
per-change decisions/rationale and evidence dispositions; system-known diff and
bindings come from accepted Final batch facts, merged with the existing last-wins
rule. No physical Session rescan, exact-set, freshness, registered-evidence or hash
proof qualifies this judgment. Actual file/output paths are valid evidence. Use the
professional field shapes/enums in that questionnaire,
including `decision` in `invalidate/supersede/retain/reverify`.
Successful Impact saving prepares the existing Targeted questionnaire for the
actually assigned independent verifier. If no verifier is selected yet, use Main's
current independent verifier division path; saving that selection at the
targeted step prepares that person's answer.
Main actually spawns/notifies the independent verifier. The questionnaire supplies
all necessary Contract assertion text and system slots; the verifier performs the
real observations and saves
`continuity_reason`, `assertion_results` (`assertion_id,result,evidence_refs`),
`scope_compliance`, `result` and the applicable `failure_class` at the exact path.
Assertion results are `pass/fail/blocked`; scope is `pass/fail`; overall result is
`pass/fail/blocked/scope_changed`. Use actual evidence, not the Builder's conclusion;
the verifier must not have implemented or merged the work being verified. Every
required assertion needs an actual result; specific missing observations prevent
advancement. Extra observations and ID spelling/count/aliases, owner pools, issue
chains or evidence registration are not report qualifications. Honest fail/blocked
reports can be recorded without pretending that unfinished checks passed.

For blockers, save actual attempts/gaps and use the accurate current resolution
question. Do not fabricate PASS or manually issue a successor answer. The system
records the existing Handoff directly after the final successful Targeted save
and reports its actual artifact and fresh S7 seed paths; there is no new Handoff
form or next ordinary tool preparation step. Main then continues fresh S7 responsibility.
Hook listening adds no tool/identity/scope ACL; actual checks, independence,
actual pending user decisions under their existing authorization and the original
Writer’s atomicity/CAS/idempotency remain. Candidate, Main integration and Final
are distinct real steps; removing duplicate proof does not remove integration.
