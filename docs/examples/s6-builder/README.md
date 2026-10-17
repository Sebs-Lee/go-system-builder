# S6 Builder workgroup manifest

The ordinary S6 path prepares a non-authoritative manifest from the current
eligible TASK and registered dispatch inputs. Main chooses the TASK, Builder
role, and actual platform Agent and Team IDs; the command projects Runtime,
REQ, generation, locked document fingerprints, TASK scope, satisfied
dependencies, and check commands:

```sh
loop-harness s6 prepare --root <authority-root> --task <TASK> \
  --role backend-builder --agent-id <actual-platform-agent-id> \
  --platform-team-id <actual-platform-team-id>
```

If the locked TASK has no concrete Worker-owned Closing Contract predicate,
`s6 prepare` prints that missing professional decision and writes the draft
without inventing one. Add the concrete predicate to the draft before
registration. The output gives the exact draft path and the original
registration command. The draft itself does not register an Assignment or
record a check result.

[`workgroup-manifest.json`](workgroup-manifest.json) remains a schema and
compatibility reference for explicit native or legacy setup. Do not copy it
for the ordinary bound-workspace flow or manually replace Runtime metadata,
hashes, scope, or checks.

The original registration consumer is:

```sh
loop-harness runtime register-workgroup \
  --manifest <exact-manifest-path-returned-by-prepare> \
  --task-id <TASK> \
  --task docs/dev/tasks/<TASK>.md
```

For the ordinary bound-workspace route, prepare and launch each registered
Assignment using the returned assignment and owner IDs:

```sh
loop-harness runtime workspace prepare \
  --root <authority-root> --assignment <id> --agent <owner>
loop-harness runtime workspace launch \
  --root <authority-root> --assignment <id> --agent <owner>
```

`workspace launch` starts the real Builder in that prepared workspace. Main
tracks that original process/session and its receipt, including when the host
returns a background task handle. Continue the same execution with the original
`workspace launch --resume` when needed. One Builder carries this execution;
an additional native Agent in the same worktree would duplicate its work.

The Hook supplies the active PLAN/Result view and exact answer paths. Save the
professional answers there; the original adapters project the checkpoint and
completion facts. The prepared canonical contract puts TASK checks on Main's
required-check list; before commit the Builder still runs those TASK commands and
states their commands and observed outputs in the existing Result summary, or
explains any unrun check. Main runs the checks again during original integration. No PASS is
recorded by `s6 prepare`. This route does not need `team launch`, a separate
readback template, or a manually created worktree. See the protocol for the
explicitly selected native/legacy teammate route and approval-required
assignments.

The reference example's zero hashes are placeholders, and its `validation`
object is not proof that a real manifest passed. The original validator checks
the actual current files and registration facts.

For a code correction after failed integration, use the Hook's existing rework
instruction. `s6 prepare` also accepts `--revise-assignment <id>` and
`--revision-intent-key <new-key>` with the new Builder identities. It preserves
the original contract and returns the exact revision draft path and original
registration command. `runtime workspace pending --root <authority-root>` is
the read-only diagnostic for the current retry/rework actions.
