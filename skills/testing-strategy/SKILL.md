---
name: testing-strategy
description: Use when executable behavior changes or unit, integration, regression, or failure-path evidence is reviewed
category: best-practice
version: 1.0.0
---
# Testing Strategy
## Authority
Tests prove contract behavior; they do not redefine it. Stage routing lives in `docs/control/agent-protocol.md`; team scope comes from the current manifest and runtime evidence; the reusable testing method is inlined below.
## Applicability
Apply to `behavior-change`, QA unit-test, and QA integration-test responsibilities.
## Required Inputs
Read acceptance clauses, Closing Contract, changed paths, failure modes, and integration boundaries.
## Quality Criteria
Map `source_refs → Rule → CASE → Story → PATH → Spec → Evidence`; cover every required
allow/reject branch, boundaries, errors, states, side effects, rollback, retries,
regressions, and meaningful assertions. Enforce the module positive/negative ratio and
verify tests fail for the intended defect.
## Outputs
Test plan, executable evidence, coverage gaps, or one scoped QA conclusion.
## N/A Criteria
N/A only with evidence that no executable behavior changed.
## Stop Conditions
Stop on untestable acceptance, hidden dependencies, or misleading green tests.
## Non-Goals
Do not equate line coverage with behavioral completeness.

## Review method

Derive tests from the assigned acceptance clauses and failure model, then inspect
whether their assertions can distinguish the intended defect from a healthy run.
Check boundary values, error propagation, state transitions, side effects and
recovery as well as the happy path. Verify relevant regression behavior without
reducing the current module obligation to only the triggering requirement. Explain
unexecuted work and any mocked boundary honestly; coverage counts do not establish
semantic correctness.

Record actual methods, observations, evidence and coverage gaps in the current
professional answer or its referenced existing artifacts. Main owns responsibility
grouping, dependencies and independence under the current ReviewPlan; this method
does not add actor-count, file-count, activation or repeat-report requirements.
