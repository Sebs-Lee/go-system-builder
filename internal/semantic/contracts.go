package semantic

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	contractTokenPatterns = []*regexp.Regexp{
		regexp.MustCompile(`CASE-[A-Z0-9]+(?:-[A-Z0-9]+)*`),
		regexp.MustCompile(`\bS-[0-9]{3}\b`),
		regexp.MustCompile(`\bF-[0-9]{3}\b`),
		regexp.MustCompile(`\bPATH-[A-Z0-9]+(?:-[A-Z0-9]+)*`),
		regexp.MustCompile(`\bFR-[A-Z0-9]+(?:-[A-Z0-9]+)*`),
	}
	contractClauseCellPattern = regexp.MustCompile(`\b(FE|BE|SYNC)-[A-Z0-9]+(?:-[A-Z0-9]+)*\s*§\d+`)
)

type ContractCheckResult struct {
	Contracts    int      `json:"contracts"`
	TokenRefs    int      `json:"token_refs"`
	Clauses      int      `json:"clauses"`
	Fingerprints int      `json:"fingerprints"`
	Problems     []string `json:"problems,omitempty"`
}

// ContractsCheck is S3's mechanical close (L3-S3 v4.0.1). Division of labor
// with the S2 AC bridge: the bridge owns REQ-side AC↔CASE; this owns
// contract-side token existence. Non-goal: free-text cell semantics.
func ContractsCheck(root string) (ContractCheckResult, error) {
	result := ContractCheckResult{Problems: []string{}}
	dir := filepath.Join(root, "docs", "contracts")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil // no contracts directory — nothing to reconcile
		}
		return result, fmt.Errorf("read docs/contracts: %w", err)
	}

	universe := map[string]bool{}
	modules, _ := os.ReadDir(filepath.Join(root, "docs", "design", "prototypes"))
	for _, module := range modules {
		if !module.IsDir() || module.Name() == "template" || module.Name() == "templates" {
			continue
		}
		mpath := filepath.Join(root, "docs", "design", "prototypes", module.Name())
		for _, file := range []string{"cases.json", "stories.md", "flows.md"} {
			data, err := os.ReadFile(filepath.Join(mpath, file))
			if err != nil {
				continue
			}
			for _, pattern := range contractTokenPatterns {
				for _, token := range pattern.FindAllString(string(data), -1) {
					universe[token] = true
				}
			}
		}
	}

	reqFRs := map[string]bool{}
	reqFiles, _ := filepath.Glob(filepath.Join(root, "docs", "requirements", "REQ-*.md"))
	for _, reqFile := range reqFiles {
		data, err := os.ReadFile(reqFile)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "| FR-") {
				cells := strings.Split(strings.Trim(trimmed, "|"), "|")
				if len(cells) > 0 {
					reqFRs[strings.TrimSpace(cells[0])] = true
				}
			}
		}
	}

	contractIDs := map[string]string{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") || strings.Contains(entry.Name(), "template") {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".md")
		contractIDs[id] = filepath.Join(dir, entry.Name())
	}

	for id, path := range contractIDs {
		result.Contracts++
		data, err := os.ReadFile(path)
		if err != nil {
			result.Problems = append(result.Problems, fmt.Sprintf("%s: unreadable: %v", id, err))
			continue
		}
		content := string(data)
		for _, pattern := range contractTokenPatterns {
			for _, token := range pattern.FindAllString(content, -1) {
				result.TokenRefs++
				if strings.HasPrefix(token, "FR-") {
					if !reqFRs[token] {
						result.Problems = append(result.Problems, fmt.Sprintf("%s: token %s does not exist in any REQ's FR table", id, token))
					}
					continue
				}
				if !universe[token] {
					result.Problems = append(result.Problems, fmt.Sprintf("%s: token %s does not exist in any module package (cases/stories/flows)", id, token))
				}
			}
		}
		for _, cell := range contractClauseCellPattern.FindAllString(content, -1) {
			result.Clauses++
			contractID := strings.TrimSpace(strings.Fields(cell)[0])
			if _, ok := contractIDs[contractID]; !ok {
				result.Problems = append(result.Problems, fmt.Sprintf("%s: clause cell %q points at unknown contract", id, cell))
			}
		}
		for _, line := range strings.Split(content, "\n") {
			trimmed := strings.TrimSpace(line)
			if !strings.HasPrefix(trimmed, "|") {
				continue
			}
			cells := strings.Split(strings.Trim(trimmed, "|"), "|")
			for i, cell := range cells {
				cell = strings.TrimSpace(cell)
				if isHex64(cell) && strings.Contains(strings.ToLower(strings.Join(cells[:i], " ")), "fingerprint") {
					result.Fingerprints++
					if resolved, ok := resolveContractFingerprint(root, trimmed); ok && resolved != cell {
						result.Problems = append(result.Problems, fmt.Sprintf("%s: fingerprint column does not match disk (recorded %s… actual %s…)", id, cell[:12], resolved[:12]))
					}
				}
			}
		}
	}
	return result, nil
}

func isHex64(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			return false
		}
	}
	return true
}

func resolveContractFingerprint(root, row string) (string, bool) {
	cells := strings.Split(strings.Trim(strings.TrimSpace(row), "|"), "|")
	var fileRef string
	for _, cell := range cells {
		cell = strings.TrimSpace(cell)
		for _, name := range []string{"scenario-model.json", "fixture-contract.json", "cross-matrix.json", "cases.json", "scenario-coverage.json", "index.html", "stories.md", "flows.md"} {
			if strings.HasSuffix(cell, name) {
				fileRef = cell
				break
			}
		}
		if fileRef != "" {
			break
		}
	}
	if fileRef == "" {
		return "", false
	}
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(fileRef)))
	if err != nil {
		return "", false
	}
	return fmt.Sprintf("%x", sha256.Sum256(data)), true
}
