# Rule Index

---
rule_id: RULES-INDEX
category: Process
status: locked
owner: Project Manager / Architect
scope: all agents
updated: YYYY-MM-DD
---

## Read Order

1. Root `AGENTS.md`
2. This index and every Required Rule
3. Rules referenced by the current REQ, specification, assignment, BUG, or release audit
4. The primary Methodology Skill and selected Best Practices for the current action

Rules are stable policies. Runtime gates belong to the Loop Definition and
Hooks; procedures belong to Skills; role behavior belongs to Agent Definitions.

## Required Rules

| Rule | File | Trigger |
|:---|:---|:---|
| Communication | `communication.md` | collaboration, evidence, escalation |
| Naming | `naming.md` | identifiers, documents, branches, APIs |
| Change Control | `change-control.md` | baseline change |
| Git Branch Release | `git-branch-release.md` | branch, merge, release |
| Security | `security.md` | code, config, data, auth, secrets |

## Conditional Rules

| Rule | File | Trigger |
|:---|:---|:---|
| API Design | `api-design.md` | API, event, schema, error code |
| State Machine | `state-machine.md` | lifecycle or transition |
| Error Handling | `error-handling.md` | error, retry, failure path |
| Bugfix Review | `bugfix-review.md` | BUG investigation or closure |
| Release Architecture Audit | `release-architecture-audit.md` | release preparation |
| UI Design Package | `ui-prototype.md` | visible frontend change |
| Project Design Foundation | `design-foundation.md` | user-visible UI, brand language, first UI REQ, new product surface |

## Immutable Policy

- Chat does not change a baseline.
- Locked REQ changes require human approval through change control.
- Agent permissions are the intersection of definition, manifest, activation, runtime, and Hook policy.
- Release architecture audit is engineering evidence; human release approval is a separate final boundary.

- [Shared models and contract reading](shared-model-contracts.md): common data/protocol authority, explicit model policy, progressive TASK links and real consumer validation.

## Overall dispatch plan

S4 → S5 → S6 follow [dispatch plan rules](dispatch-plan.md). Read the approved wave checklist and live S6 board; schedule compatible tasks that are ready for dispatch and integrate results promptly. `ready` is scheduling eligibility, not a responsibility status.

## Optional project conventions

Start from [the empty conventions template](project-conventions-template.json)
when Architect indexes existing engineering decisions in
`docs/rules/project-conventions.json`. The embedded `project-conventions.schema.json`
defines source references, path scopes, examples, exceptions and declared status.
The installer ships the empty template and creates no approved project defaults.

`tasks check --json` returns `task_conventions` selected for each TASK's declared
write scope. A `declared_default` must match the original locked Rules index or
the original registered locked document, its hash/version, a real source section
and current example hashes. Registry status alone proves no approval. An absent
registry creates candidates only for API/error rules explicitly cited by the TASK.
Candidate/conflict/stale/migration rows belong to the original Architect or rule
owner; they never change business values, permission, coverage or stage gates.
Consume resolved `entries[].status`, not the registry's nested `entry.status`
claim. The projection proves source consistency; reviewers still judge semantic
applicability and whether an implementation example follows the rule.
