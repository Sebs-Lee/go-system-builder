---
name: team-planning
description: Use when planning independent engineering responsibilities for native Agents while preserving professional coverage and frozen recovery
category: methodology
version: 2.0.0
---
# Professional Work Planning

Main divides engineering work and uses Claude Code native Agent. A manifest/Assignment records engineering responsibility; it does not create a Claude Team or reserve an actor/session. `agent_id` is the engineering owner label. `platform_team_id`, where the existing schema retains it, is a logical grouping label. Do not look up opaque platform IDs or run Team/workspace launch to populate these fields.

## Authority

Follow `docs/control/agent-protocol.md`, the original schemas and `docs/control/loop-definition.json`. Original domain Writers own responsibility and stage facts; a manifest is not a platform scheduler.

## Entry Conditions

Professional scope and actual source inputs are known. Main can identify required coverage and genuine independence constraints.

## Required Inputs

Current stage, frozen specifications/implementation, original responsibilities, professional risks, dependencies and exact original preparation outputs.

## Procedure — Plan actual work

1. Read current professional inputs and choose concrete responsibilities, reviewed scope and outputs.
2. Enumerate required coverage and real risk-triggered reviews. Explain genuine N/A; do not silently omit work or create empty obligations.
3. Separate incompatible permissions, different conclusion dimensions and independent verification. An implementer must not sign independent verification of the same output.
4. Select frontend-builder, backend-builder or test-builder by actual write ownership; use document-verifier, delivery-verifier, QA, e2e-tester and investigator for their professional duties. Keep E2E/browser work in its relevant role.
5. Preserve source freshness, real dependencies and resource conflicts. Main starts dependent work when actual inputs are ready; a framework queue is not the native Agent scheduler.
6. Prepare original engineering inputs and private answer paths, then carry exact references into native Agent. Claude Code owns context, messaging and resume. Registration is not a successful platform spawn.
7. Receive actual artifacts through Start/Stop and original consumers. Main handles integration, checks, conflicting conclusions and stage closure separately.

## Stage paths

- S6 (`s6-native-v1`): `s6 prepare --root <Main> --task <TASK> --role <role>` supplies the exact brief. Builder uses native worktree, professional plan, actual implementation/self-checks and Result. Main uses original task-integrate. Source/report revisions use the original revision Writer.
- S5: register this entry's two independent dimensions, then carry exact issued view/answer paths into separate native contexts. New read-only reviews need no workspace prepare.
- S7: ReviewPlan owns Claims, lenses, waves and dependencies; original manifest draft/registration prepares the engineering responsibility and exact question. Reviewer returns the original professional answer for Stop reception. A post-S9 seed is an unfinished Planner input, not a dispatchable professional decision. Resolve PLANNER-REFINE fields against the current scope and submit the first professional Plan through original `runtime review-plan --file <professional-plan.json>` while preserving the seed id/round/generation/baseline. Only the exact unstarted seed is eligible; later discovery revisions retain their normal budget and workspace constraints. If E2E is genuinely inapplicable, set e2e_coverage_state=not_applicable, keep the explicit N/A Claim with source_refs/na_rationale/na_checklist_id, set verification_artifact_workspace null; do not add the Runtime-only verification_artifact_digest to the Plan, and remove only its E2E Assignment. Keep all applicable delivery/QA Claims. The conservative cold_start seed hint must not manufacture a browser surface or mandatory browser task.
- S8: Main's Case/Hypothesis dispatch supplies the exact Investigator question. Investigator performs its discriminator without product writes and returns the original answer. Main owns routing/Contract approval.
- S9: professional division prepares each exact native brief. Builder saves PlanReport before implementation, continues in native worktree, self-checks and commits actual product changes when present, then returns Result. An allowed no-change result does not manufacture a commit. Main integrates/checks separately. Targeted is a separate independent native task; original handoff leads to fresh S7.

## S5 manifest assembly

Use the original team-manifest schema and actual accepted S4 input set. Both `DV-SPEC-CONSISTENCY` and `DV-TASK-EXECUTABILITY` cover the full current batch, with distinct engineering owners and one `independence` separation edge. Use entry-specific IDs/paths and retain history. Main chooses the professional split; use current original preparation/source facts instead of copied example hashes.

Set `workgroup_kind=document_verifier`, `review_round=null`, and current `validation.validated_at`. S5 stage entry is not an S7 verification round: use `review-e<phase_revision>` only in entry-specific IDs/paths, not in the manifest `review_round` field. The registration anchor is the first S4 dispatch-plan TASK; use its actual entity ID/path, never a REQ ID as `--task-id`. For `docs/dev/tasks/TASK-001-01.md`, use its canonical entity ID `TASK-001-01`.

Required `skill_refs`: `document-verification, specification-planning, api-contracts` for specification consistency; `document-verification, specification-planning, testing-strategy` for task executability. Do not add S6 Assignment Record/Closing Contract fields to S5. These engineering grouping/owner fields require no actual Claude Team or platform ID lookup.

### S5 original manifest shape

The installed shape below is the original professional registration DTO, not a platform Team or execution definition. Use it directly; do not extract schemas from the binary or copy frozen launcher instructions. Replace every placeholder, carry the actual current Runtime/generation/stage-entry and **all** current registered document refs, and choose entry-specific IDs and professional scope. `baseline_generation: 1` is only an example. Keep the two original responsibility profiles and separation edge. The original registration rechecks source fingerprints, required dimensions and independence; the example's validation fields do not establish a PASS. `read_paths` covers the actual complete current batch; no product writes or S6 fields belong to these S5 roles.

```json
{
  "schema_version": "1.0.0",
  "manifest_id": "<entry-specific manifest id>",
  "version": "v1.0.0",
  "runtime_id": "<current runtime id>",
  "req_id": "<bound REQ id>",
  "baseline_generation": 1,
  "review_round": null,
  "platform_team_id": "<logical engineering group label>",
  "workgroup_id": "<entry-specific workgroup id>",
  "workgroup_kind": "document_verifier",
  "status": "planned",
  "documents": [
    {
      "id": "<each current registered document id>",
      "path": "<actual path>",
      "version": "<actual version>",
      "sha256": "<actual registered SHA256>"
    }
  ],
  "risk_tags": [],
  "responsibility_dispositions": [
    {
      "responsibility_id": "DV-SPEC-CONSISTENCY",
      "disposition": "assigned",
      "trigger": "S5 mandatory independent dimension.",
      "assignment_ids": [
        "<spec assignment id>"
      ],
      "na_rationale": null,
      "evidence_ref": null
    },
    {
      "responsibility_id": "DV-TASK-EXECUTABILITY",
      "disposition": "assigned",
      "trigger": "S5 mandatory independent dimension.",
      "assignment_ids": [
        "<task assignment id>"
      ],
      "na_rationale": null,
      "evidence_ref": null
    }
  ],
  "assignments": [
    {
      "assignment_id": "<spec assignment id>",
      "responsibility_id": "DV-SPEC-CONSISTENCY",
      "role_family": "document-verifier",
      "scope": [
        "Review current specification-chain consistency."
      ],
      "agent_id": "<distinct spec engineering owner>",
      "agent_definition_ref": ".claude/agents/document-verifier.md",
      "skill_refs": [
        "document-verification",
        "specification-planning",
        "api-contracts"
      ],
      "read_paths": [
        "<actual current batch input paths>"
      ],
      "write_paths": [],
      "output_paths": [],
      "depends_on": [],
      "reuse_decision": "create",
      "grouping_rationale": "Different professional conclusions require independent contexts.",
      "status": "planned",
      "dispatch_mode": "plan_checkpoint",
      "done_when": [
        "Record the assigned professional conclusion and findings in its original private answer."
      ]
    },
    {
      "assignment_id": "<task assignment id>",
      "responsibility_id": "DV-TASK-EXECUTABILITY",
      "role_family": "document-verifier",
      "scope": [
        "Review executability of the complete current TASK batch."
      ],
      "agent_id": "<distinct task engineering owner>",
      "agent_definition_ref": ".claude/agents/document-verifier.md",
      "skill_refs": [
        "document-verification",
        "specification-planning",
        "testing-strategy"
      ],
      "read_paths": [
        "<actual current batch input paths>"
      ],
      "write_paths": [],
      "output_paths": [],
      "depends_on": [],
      "reuse_decision": "create",
      "grouping_rationale": "Different professional conclusions require independent contexts.",
      "status": "planned",
      "dispatch_mode": "plan_checkpoint",
      "done_when": [
        "Record the assigned professional conclusion and findings in its original private answer."
      ]
    }
  ],
  "separation_edges": [
    {
      "left_assignment_id": "<spec assignment id>",
      "right_assignment_id": "<task assignment id>",
      "reason": "independence"
    }
  ],
  "planned_agent_count": 2,
  "max_parallel_agents": 2,
  "quantity_rationale": "Two independent document dimensions require two separate contexts.",
  "validation": {
    "result": "pass",
    "missing_responsibilities": [],
    "unresolved_conflicts": [],
    "warnings": [],
    "validated_at": "<current RFC3339 time>"
  }
}
```

Main registers the completed file through original `runtime register-workgroup --root <Main> --manifest <file> --task-id <first S4 dispatch-plan TASK id>`. The original Main questionnaire preparation supplies exact views/answers; then use native Agent. Registration alone creates no native context.

## Native boundary and frozen compatibility

Follow [agent-dispatch](../agent-dispatch/SKILL.md) for exact references, professional artifacts and Start/Stop. No generic readback, begin, activation or ordinary-save receipt precedes new native execution. Keep real professional approval, independence and the stage's actual required coverage.

Only an existing registered pre-migration execution uses [frozen-execution.md](frozen-execution.md). Preserve its sources/recovery and avoid duplicate launch; do not apply that procedure to new native responsibilities.

## Outputs

Original professional responsibility plan/manifest, concrete coverage and exact native task references. Do not create a parallel platform execution registry.

## Exit Conditions

Required work and independence are represented truthfully, original registration/preparation has supplied exact paths, and Main has actually dispatched the intended native responsibilities. Registration alone is not platform execution.

## Stop Conditions

Report missing authoritative inputs, incompatible responsibility ownership, lost independence or original approval still pending. Keep concrete unfinished work visible.

## Non-Goals

No Team creation, opaque actor lookup, process/session launcher or generic activation workflow. Do not self-author independent professional results or advance stages manually.
