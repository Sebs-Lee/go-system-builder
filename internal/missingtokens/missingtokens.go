package missingtokens

import (
	"fmt"
	"sort"
	"strings"
)

// missingTokenLegend maps GATE-BUILDER-BATCH-READY missing tokens to their
// meaning and the next executable action (L3-S6 §9.3 — error messages are
// process orchestration). The legend is shared by the Hook's not_ready
// recovery packet and `explain TR-006` so the agent reads one vocabulary
// wherever the tokens appear.
type missingTokenRule struct {
	prefix   string // token equality when exact, otherwise prefix match
	exact    bool
	meaning  string
	nextStep string
}

var builderBatchMissingTokenRules = []missingTokenRule{
	{
		prefix:   "batch:execution_batch_empty",
		exact:    true,
		meaning:  "no task documents are registered at the current generation — the TR-003 execution batch is missing",
		nextStep: "check whether TR-003 committed; run `loop-harness runtime reconcile` if the state is inconsistent",
	},
	{
		prefix:   "evidence:completion_report",
		exact:    true,
		meaning:  "no qualified completion envelope exists at all",
		nextStep: "run `runtime task-complete` for each unfinished TASK in the batch",
	},
	{
		prefix:   "evidence:completion_report:",
		meaning:  "that TASK has no Builder Result envelope",
		nextStep: "run `runtime task-complete` for the named TASK",
	},
	{
		prefix:   "checks:",
		meaning:  "the completion envelope records a check that is not `pass`",
		nextStep: "fix the failure and re-run `runtime task-complete`; the newer envelope supersedes the older one",
	},
	{
		prefix:   "scope_deviations:",
		meaning:  "the completion envelope declares an unapproved write-scope deviation",
		nextStep: "revise the assignment scope (new manifest row) or fix the implementation to stay inside `write_paths`",
	},
	{
		prefix:   "integration_checkpoint:",
		meaning:  "that TASK has no durable worktree integration checkpoint at `verified` or beyond",
		nextStep: "run `runtime task-integrate --assignment-id <id>` for that assignment (or re-trigger the Builder's SubagentStop)",
	},
}

// RenderMissingTokenLegend returns a compact legend block for the missing
// tokens of one gate evaluation. Empty when nothing matches (unknown tokens
// are listed verbatim with a generic fallback line so silence never hides a
// new token shape).
func RenderMissingTokenLegend(gateID string, missing []string) string {
	if len(missing) == 0 {
		return ""
	}
	var rules []missingTokenRule
	rules, hasRules := gateRules(gateID)
	if !hasRules {
		return ""
	}
	var b strings.Builder
	b.WriteString("MISSING TOKENS:\n")
	covered := make(map[string]bool)
	for _, token := range missing {
		rule, ok := matchMissingTokenRule(rules, token)
		if !ok {
			continue
		}
		if covered[rule.prefix] {
			continue
		}
		covered[rule.prefix] = true
		fmt.Fprintf(&b, "- %s — %s. Next: %s.\n", legendTokenLabel(rule, token), rule.meaning, rule.nextStep)
	}
	unknown := unknownTokens(rules, missing)
	if len(unknown) > 0 {
		sort.Strings(unknown)
		fmt.Fprintf(&b, "- %s — no legend entry; inspect the gate implementation or run `loop-harness explain %s`.\n",
			strings.Join(unknown, ", "), gateID)
	}
	return strings.TrimRight(b.String(), "\n")
}

func matchMissingTokenRule(rules []missingTokenRule, token string) (missingTokenRule, bool) {
	for _, rule := range rules {
		if rule.exact {
			if token == rule.prefix {
				return rule, true
			}
			continue
		}
		if strings.HasPrefix(token, rule.prefix) {
			return rule, true
		}
	}
	return missingTokenRule{}, false
}

// RenderGateTokenLegend prints the full token table for one gate — used by
// `explain <transition>` where no evaluation exists yet, so the agent can
// read the vocabulary before the first not_ready packet.
func RenderGateTokenLegend(gateID string) string {
	rules, ok := gateRules(gateID)
	if !ok {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "GATE MISSING-TOKEN LEGEND (%s):\n", gateID)
	for _, rule := range rules {
		fmt.Fprintf(&b, "- %s — %s. Next: %s.\n", legendTokenLabel(rule, rule.prefix), rule.meaning, rule.nextStep)
	}
	return strings.TrimRight(b.String(), "\n")
}

func gateRules(gateID string) ([]missingTokenRule, bool) {
	if gateID != "GATE-BUILDER-BATCH-READY" {
		return nil, false
	}
	return builderBatchMissingTokenRules, true
}

// legendTokenLabel keeps one line per rule family: the bare prefix for the
// per-TASK families (the specific TASK rides in the evaluation's token
// list) and the exact token otherwise.
func legendTokenLabel(rule missingTokenRule, token string) string {
	if rule.exact {
		return "`" + token + "`"
	}
	return "`" + rule.prefix + "<TASK>`"
}

func unknownTokens(rules []missingTokenRule, missing []string) []string {
	var unknown []string
	for _, token := range missing {
		if _, ok := matchMissingTokenRule(rules, token); !ok {
			unknown = append(unknown, token)
		}
	}
	return unknown
}
