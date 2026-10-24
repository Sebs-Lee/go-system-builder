---
name: code-quality
description: Use when code is authored or reviewed for maintainability, reuse, abstraction, hardcoding, readability, or complexity
category: best-practice
version: 1.0.0
---
# Code Quality
## Authority
Quality guidance only. Stage routing lives in `docs/control/agent-protocol.md`; team scope comes from the current manifest and runtime evidence; the reusable review method is inlined below.
## Applicability
Apply to code changes, QA module-code, and reuse/abstraction responsibilities.
## Required Inputs
Read changed code, local conventions, tests, public contracts, and dependency boundaries.
## Quality Criteria
Check correctness clarity, naming, cohesion, dependency direction, duplication, useful reuse, abstraction timing, hardcoding, magic values, oversized units, dead/debug code, and error handling.
## Outputs
Maintainable implementation or one auditable quality conclusion.
## N/A Criteria
N/A only when no executable or configuration code is in scope.
## Stop Conditions
Stop when improvement requires unrelated refactoring or contract change.
## Non-Goals
Do not optimize for fewer lines or speculative abstraction.

## Review method

Trace the assigned Claim from its actual contract to the implementation and caller.
Inspect normal and failure paths, state changes, ownership, error propagation and
boundary conversions. Explain a concrete correctness or maintenance consequence
for each finding; name the entry point and the first broken invariant. Distinguish
useful shared behavior from coincidental similarity before proposing abstraction.
For review work, preserve evidence in the assigned answer instead of editing the
product. A repair or unrelated refactor requires Main's original route.

Use the current ReviewPlan's Claims, scope and independence constraints. Main owns
specialist selection and compatible grouping under team-planning; this method does
not create another actor-count, file-count or dispatch gate.
