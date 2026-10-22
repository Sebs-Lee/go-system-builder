# Contracts

This directory stores contract templates and contract rules.

An active S3 batch consists of its original current-REQ index and declared
contracts. Other REQ contracts remain readable, but their text cannot fill a
missing CASE citation in this batch. The module retains its complete current
behavior; only CASE obligations owned by this exact REQ must close here.
Unqualified FR IDs belong to this REQ; qualified `REQ/FR` references resolve
their actual owner, without borrowing another requirement's same-numbered ID.
The checker uses the original bound requirement path and original delivery set.
`contracts check` without an active REQ is an all-project authoring diagnostic,
and does not certify any one S3 batch.

Clause references must resolve to the target contract's existing `本合同条款`
column in a Markdown pipe table with a header and separator. Bare `§n` and
the target's own qualified ID are declarations; prose, fenced examples and
other-contract references are not. This checks numbered structure; S3 authors
and S5 reviewers still judge promises and coverage. No separate clause registry
or author proof is required.

## Lock Authority

Authors record actual contract scope, clauses, applicability and necessary validation sources. Runtime and the original consumer supply machine fingerprints and registration associations; authors do not assemble or copy fingerprint, receipt or revision metadata. Legal stage transitions remain in `docs/control/loop-definition.json`; this README does not maintain a second gate list.

## Locked Contract Means

The contract is the Builder's source of truth.

Any change to scope, API, fields, errors, state, side effects, or acceptance must go through `docs/rules/change-control.md`.

## Quality Gates

Locked contract quality gates, such as lint, typecheck, build, tests, or contract tests, must be explicit.

Builder command evidence belongs in `TASK-*`; delivery correctness belongs in `REV-*`; coverage and test quality belong in `QA-*`; real-browser user-flow evidence belongs in `E2E-*`; `ACC-*` links the final evidence.

## Templates

| Type | Template |
|:---|:---|
| Overview | `CONTRACTS-template.md` |
| Frontend | `FE-contract-template.md` |
| Backend | `BE-contract-template.md` |
| Sync/API | `SYNC-contract-template.md` |


## Shared model authority

Follow [shared-model contracts](../../rules/shared-model-contracts.md). Data shape has one native authoring source; SYNC owns shared behavior, FE/BE derive responsibilities. Structural negative examples fail schema validation; business negatives remain valid data and are rejected by the business rule.
