# S6 native Builder delivery

Main selects an eligible reviewed TASK and role:

```sh
loop-harness s6 prepare --root <Main> --task <TASK> --role backend-builder
```

Carry the returned exact task brief into a native Claude Code Agent call. Use native worktree isolation for writes, with worktree.baseRef=head on the committed bound development branch. No Team ID, platform Agent ID or separate workspace launcher is needed. Start provides engineering guidance. The Builder reads actual sources, saves its professional plan, implements, runs TASK self-checks, commits and saves its professional Result. Return the exact brief/result reference. Stop receives the actual artifacts through the original domain consumer; Main explicitly integrates and runs required Main checks afterwards.

The plan and Result drafts contain professional fields only and live under the actual Worker worktree’s private `.claude/assignments/native/` directory. Commit TASK products; keep these drafts out of Git. Save does not mean activation or acceptance; there is no manual begin/task-complete step. Actual missing deliverables can be corrected; real blockers can be returned without waiting for Main integration.

Use the exact nested structures supplied in the brief. Each Result `assertion_evidence` item has only `assertion`, `oracle` (verbatim from the plan), and nonempty `evidence_refs`. References must be exact existing worktree-relative file paths within the TASK write scope or listed Read paths, such as `src/pkg/input.go`; omit line numbers, descriptions and command strings. Put command results and observations in `summary` or closing rationale. Stop reports specific professional field errors through the original S6 Schema and retains the original evidence scope checks.

The JSON workgroup-manifest example in this directory is an old explicit s6-v1.2 compatibility fixture. Existing frozen workspace assignments retain their original prepare/launch/resume and recovery commands. Do not copy that manifest or relaunch a running old assignment for the native path. Actual identity, checks and evidence are never inferred from the example's placeholders.
