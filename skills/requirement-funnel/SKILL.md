---
name: requirement-funnel
description: Use when a requirement is being drafted or amended in S0 and the expected outcome needs to become a human-locked REQ baseline
category: methodology
version: 0.1.0
---
# Requirement Funnel

## Authority
Stage contract: `docs/agent-protocol.md` #s0 (primary_skill). Structure and field definitions: `docs/requirements/REQ-template.md` (§A→§B→§C funnel). This skill carries process only — how to think, converge, and hand up.

## Entry Conditions
- S0 starts (new REQ), or an amendment reopens a locked REQ (new generation) — both run the same funnel.

## Required Inputs
| Input | Path / field | Why |
|:--|:--|:--|
| Expected outcome | the human's natural-language statement (conversation, never recorded verbatim) | the funnel's raw material |
| Project facts | `docs/project-map.md` | grounding for §B constraints |
| Module truth packages | `docs/design/prototypes/<module>/` (when touched) | UI impact and regression scope |

## Procedure
1. **Mode**: the human states the expected outcome; you own the design end-to-end; the human only approves (拍板). You are not an interviewer — keep questions to yourself, hand proposals up.
2. **§A 理念**: filter the expected outcome — strip ambiguous colloquial wording, surface implicit premises — into a structured restatement (A1) and have the human confirm it says what they meant. Recording raw quotes records ambiguity. Then draft A2-A4, stakeholders, glossary. Hand up: complete §A proposal, ≤3 decision points.
3. **§B 方向与约束**: design 2-3 viable directions, prune to one recommendation with a one-line rejection reason per discarded direction; challenge every "must" constraint ("what does breaking it cost?" — no answer means preference, demote it); log assumptions and risks. Hand up: complete §B proposal, ≤3 decision points.
4. **§C 具体需求**: decompose into requirement points, each tracing back to a §A item (untraceable = scope creep, challenge it on the spot); flows with rejection paths; testable acceptance criteria; NFRs; migration inventory; UI impact table (three-valued). Hand up: complete §C proposal, ≤3 decision points.
5. **Per-layer loop**: diverge exhaustively → prune with the layer above → self-review → hand up. Never design the next layer before the current one is approved.
6. **Proposal discipline** (the only acceptable hand-up shape): a complete proposal — recommendation, rationale, rejected alternative with its cost. Open questions are forbidden: decide what you can, record the rationale in the matching field; only non-derivable value judgments (direction / scope boundary / priority / risk tolerance) become decision points, ≤3 per layer. More than 3 means the design has not converged — go back and prune, do not keep asking. Dumping questions on the human feeds L1's named failure mode of addictive dependence on "having a human re-check each time".
7. **Decision-point format**: "Recommend X (rationale); alternative Y (cost); the one thing you must decide: <one sentence>."
8. **Self-review** (light before each hand-up; full before proposing lock) — attack your own reasoning in three roles: implementer ("which acceptance criterion can't I build?"), user ("which sentence would I misread?"), maintainer ("where do §A4 exclusions and §B constraints collide?"). Findings: fix, or downgrade to a decision-point annotation. Review the reasoning chain (A→B→C traceable), not the format. The full-review conclusion is one paragraph recorded before §E.
9. **Escalate with a proposal** (never escalate a bare question): contradictions between approved layers, unresolvable intent, or value tradeoffs without a judgment basis — "I lean toward A because …; if your intent is B, then X/Y/Z must be redone."

## Outputs
- Per layer: a complete proposal with ≤3 decision points, recorded into the REQ §A/§B/§C sections.
- §D entries for open clarifications (each marked blocking / non-blocking).
- Before lock: the self-review conclusion paragraph (§E) accompanying the lock proposal.

## Exit Conditions
- Human confirms §A-§C match their intent and signs the §E lock record; `status: locked` (bind itself is human-only, outside this skill).

## Stop Conditions
- Intent cannot be understood even after restatement attempts → stop, escalate with candidate interpretations.
- Approved layers contradict each other → stop, escalate for arbitration.
- A value tradeoff has no judgment basis available to the agent → stop, hand the decision point up.

## Non-Goals
- No architecture design (§B ends at direction and constraints; S2 owns architecture).
- No value decisions on the human's behalf (priority, scope, direction sign-off).
- No bind execution or REQ mutation after lock (human-only; hook-enforced).
- No restating template field definitions — the template is self-describing.
