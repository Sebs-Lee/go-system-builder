---
name: playwright-e2e
description: Use when Playwright browser tests, end-to-end flows, browser fixtures, or UI test evidence change
category: best-practice
version: 0.3.1
---
# Playwright E2E
## Authority
Technical practice only. S7 responsibility and evidence ownership remain in `docs/control/agent-protocol.md`; `e2e-browser-testing` supplies browser verification methods. Main dispatches under the current ReviewPlan, and the original S7 consumer and exit gate decide clean-round qualification.
## Applicability
Apply to browser automation, executable user flows, fixtures, selectors, or E2E evidence.
## Required Inputs
Read the module current-truth package (`scenario-model.json`, generated cases/coverage,
fixture contract, `stories.md`, `flows.md`, and current HTML), browser support target,
test data, and environment contract.
## Quality Criteria
Execute human navigation steps, use stable semantic selectors, isolate test data, assert
common oracle fields (`visible`, `terminal_state`, `persisted_effects`,
`forbidden_side_effects`) and, for negative cases, `rejection`, `expected_state`, and
`recovery`; retain traceable evidence and source-backed recovery N/A. Playwright or
Storybook snapshot equality proves implementation drift only (`tools/visual-qa/README.md`);
it does not prove the Design Thesis.
## Outputs
Repeatable browser tests or a scoped E2E review conclusion with browser evidence.
## N/A Criteria
N/A only when the affected behavior has no browser-visible user path.
## Stop Conditions
Stop on URL-jump-only coverage, missing required control selectors, or an unrepeatable environment.
## Non-Goals
Do not replace S7 responsibility assignment or clean-round decisions.

## Operating Procedure
1. Read the assigned CASE, branch polarity, oracle, PATH, controls, and fixture contract from the module package; define isolated persona/data fixtures and the initial browser state.
2. Implement each human action using role, label, or approved `data-test` locators. Use direct API/URL operations only to establish fixtures, never to skip the route under test.
3. Assert visible checkpoints and relevant request outcomes with auto-waited conditions, not sleeps. Add trace, screenshot, or video on failure for diagnosis.
4. Run the assigned independent verification, retain its actual browser evidence and reference it in the original S7 answer using the methods in `e2e-browser-testing`. Repeat only when needed to resolve a real observation; do not repeat unsafe P0 actions or create another report.

## Evidence Checklist
- Module, CASE ID, branch polarity, flow/PATH ID, personas, fixture setup/cleanup, and control-to-locator map.
- Assertions for each checkpoint, exception/permission branch, and terminal user-visible result.
- Repeatable command plus trace/screenshot references for any failure.

## Common Failure Modes
- `page.goto()` reaches an assertion while bypassing buttons and navigation logic.
- CSS/text/DOM-chain locators or arbitrary `waitForTimeout` hide real regressions.
- One shared account or record makes tests order-dependent or masks permission defects.

## Primary Sources
- [Playwright best practices](https://playwright.dev/docs/best-practices)
- [Playwright locators](https://playwright.dev/docs/locators)
