---
name: api-contracts
description: Use when APIs, events, shared schemas, integration payloads, or compatibility promises change
category: best-practice
version: 1.1.1
---
# API Contracts
## Authority
S3 uses the locked REQ and registered S2 design to define contracts. Locked contracts own implementation semantics. Stage routing lives in `docs/control/agent-protocol.md`; work dispatched through an Assignment uses its manifest and Agent message.
## Applicability
Apply to `api` risk and shared component boundaries.
## Required Inputs
For S3, read the frozen REQ/S2 inputs and supplied contract templates. For changes to an existing boundary, read its FE/BE/SYNC contracts, schemas, consumers, version policy and error model.
## Quality Criteria
Check field meaning, validation, compatibility, defaults, error codes, idempotency, pagination, ordering, retries, and side effects.
## Outputs
S3 contract design and its professional answer, or contract-aligned implementation and compatibility evidence for an implementation Assignment.
## N/A Criteria
N/A when no machine-consumed boundary changes.
## Stop Conditions
Stop on consumer disagreement, unspecified compatibility, or ambiguous side effects.
## Non-Goals
Do not silently extend locked contracts.

## Inlined Methodology

For implementation, read the exact native task brief, its TASK and linked
contract clauses, shared definitions and applicable design/rules. Repair work uses
the approved repair context. Form your professional plan, then implement and
self-check within that responsibility. Save the supplied professional outputs and
return their exact references for SubagentStop reception; Main owns integration.
An explicit domain approval remains required when the task names one. Existing
registered frozen executions use their original [recovery contract](../agent-dispatch/frozen-execution.md).

For S3 Main, follow the frozen REQ/S2 input view and
[S3 stage instructions](../../docs/control/agent-protocol.md#s3).
Use the exact answer, output and template paths supplied by the Hook.
Create the current REQ's index and applicable contracts; fill the professional
answer once and save. The Hook names the bound development branch and any
contract outputs read from Git: commit those completed professional files and
their Git-sourced model/example dependencies there before saving the answer.
S3 outputs can change during this entry; REQ/S2 inputs remain frozen.
The original consumer captures model dependencies and
runs checks and registration. S3 needs no Assignment, second envelope or
repeated manual checks.

Keep the current REQ's original index and complete applicable contract set.
Module behavior spans requirements; retain historical CASEs and real sources.
Another REQ's contract text cannot close a missing current CASE mapping.
Bare FR IDs refer to the current REQ; use the existing explicit `REQ/FR` form
when citing another requirement. Do not copy or relabel a source to satisfy the
check. Missing or ambiguous original sources belong to Main recovery. The normal
consumer uses the original delivery paths; no second scope file or manual check
is required. An unbound all-project diagnostic does not prove this batch complete.

PreToolUse observes graph progress and supplies the current questionnaire's
exact path, missing duties and next action. It does not authorize or deny tools.
Fill the professional answers and save at the supplied path; the original
consumer and stage gates determine reception and progression. Missing input,
conflicts or unspecified compatibility must be reported explicitly. Human
release decisions and independent reviews remain in their original stages.
Prompts reference the relevant contract clauses, role, inputs and expected
outputs; do not duplicate the full methodology or ask the author to copy
machine-known identity and evidence metadata.

## Shared source and reading order

Apply [shared-model contract rules in the factory](../../docs/rules/shared-model-contracts.md) (after installation: [project rule](../../docs/rules/shared-model-contracts.md)).
Converge the shared model and SYNC behavior before finalizing FE/BE contracts.
Use native definitions and exact references; do not copy field tables. Check
actual consumer/provider/Mock behavior, including structurally valid business
rejections. Read only the operation and model slices named by the current TASK
or the exact S3 Hook view.
