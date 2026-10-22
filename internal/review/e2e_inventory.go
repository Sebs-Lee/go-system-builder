package review

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/entroforge/go-system-builder/internal/pathscope"
	"github.com/entroforge/go-system-builder/internal/workpacket"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// e2eScenario is the small CASE projection the S7 Planner needs. The full
// cases.json remains the authority; this projection only carries the fields
// needed to name an Assignment and build a concrete oracle prompt.
type e2eScenario struct {
	CasesPath       string
	Module          string
	ID              string
	Title           string
	Polarity        string
	FlowRefs        []string
	Oracle          map[string]any
	Required        bool
	BrowserRequired bool
}

type e2eSourceCandidate struct {
	Path          string
	SHA256        string
	PossibleCases []string
}

type e2eInventory struct {
	Cases      []e2eScenario
	Candidates []e2eSourceCandidate
}

type e2eCasesDocument struct {
	Module string            `json:"module"`
	Cases  []e2eScenarioJSON `json:"cases"`
}

type e2eScenarioJSON struct {
	ID              string         `json:"id"`
	Title           string         `json:"title"`
	Polarity        string         `json:"polarity"`
	FlowRefs        []string       `json:"flow_refs"`
	Oracle          map[string]any `json:"oracle"`
	Required        bool           `json:"required"`
	BrowserRequired bool           `json:"browser_required"`
}

var e2eSpecSuffixes = []string{".spec.ts", ".spec.tsx", ".spec.js", ".spec.jsx"}

// Discovery offers source hints, never an executable-asset declaration.
// The full cases path keeps module-local CASE/PATH identifiers distinct.
func discoverE2EInventory(root string, state map[string]any) (e2eInventory, []string) {
	var inventory e2eInventory
	var diagnostics []string
	moduleFilter := boundE2EModules(root, state)
	prototypes := filepath.Join(root, "docs", "design", "prototypes")
	entries, err := os.ReadDir(prototypes)
	if os.IsNotExist(err) {
		return inventory, nil
	}
	if err != nil {
		return inventory, []string{fmt.Sprintf("read E2E prototype root: %v", err)}
	}
	for _, entry := range entries {
		if !entry.IsDir() || (len(moduleFilter) > 0 && !moduleFilter[entry.Name()]) {
			continue
		}
		path := filepath.Join(prototypes, entry.Name(), "cases.json")
		data, readErr := workpacket.ReadFile(root, "docs/design/prototypes/"+entry.Name()+"/cases.json", 2<<20)
		if os.IsNotExist(readErr) {
			continue
		}
		if readErr != nil {
			diagnostics = append(diagnostics, fmt.Sprintf("read %s: %v", filepath.ToSlash(filepath.Join("docs", "design", "prototypes", entry.Name(), "cases.json")), readErr))
			continue
		}
		var document e2eCasesDocument
		if err := json.Unmarshal(data, &document); err != nil {
			diagnostics = append(diagnostics, fmt.Sprintf("decode %s: %v", filepath.ToSlash(filepath.Join("docs", "design", "prototypes", entry.Name(), "cases.json")), err))
			continue
		}
		seen := map[string]bool{}
		for _, item := range document.Cases {
			if seen[item.ID] {
				diagnostics = append(diagnostics, fmt.Sprintf("duplicate CASE %s in %s; Main must correct the scenario source", item.ID, filepath.ToSlash(path)))
				continue
			}
			seen[item.ID] = true
			if !item.Required || !item.BrowserRequired || strings.TrimSpace(item.ID) == "" {
				continue
			}
			inventory.Cases = append(inventory.Cases, e2eScenario{
				CasesPath: "docs/design/prototypes/" + entry.Name() + "/cases.json", Module: entry.Name(), ID: item.ID, Title: item.Title, Polarity: item.Polarity,
				FlowRefs: append([]string(nil), item.FlowRefs...), Oracle: item.Oracle,
				Required: item.Required, BrowserRequired: item.BrowserRequired,
			})
		}
	}
	if len(inventory.Cases) == 0 {
		return inventory, diagnostics
	}

	caseIDs := make(map[string][]string, len(inventory.Cases))
	for _, item := range inventory.Cases {
		caseIDs[item.ID] = append(caseIDs[item.ID], scenarioRef(item))
	}
	scope := pathscope.New(root)
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			diagnostics = append(diagnostics, fmt.Sprintf("walk E2E assets: %v", walkErr))
			return nil
		}
		if scope.Excludes(path) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", ".claude", "node_modules", "vendor", "dist", "build":
				return filepath.SkipDir
			}
			return nil
		}
		if !entry.Type().IsRegular() || !hasSuffix(path, e2eSpecSuffixes) {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		data, readErr := workpacket.ReadFile(root, rel, 2<<20)
		if readErr != nil {
			diagnostics = append(diagnostics, fmt.Sprintf("read E2E candidate %s: %v", rel, readErr))
			return nil
		}
		digest := sha256.Sum256(data)
		candidate := e2eSourceCandidate{Path: rel, SHA256: hex.EncodeToString(digest[:])}
		for caseID, refs := range caseIDs {
			if mentionsCaseID(string(data), caseID) {
				candidate.PossibleCases = append(candidate.PossibleCases, refs...)
			}
		}
		if len(candidate.PossibleCases) > 0 {
			sort.Strings(candidate.PossibleCases)
			inventory.Candidates = append(inventory.Candidates, candidate)
		}
		return nil
	})
	sort.Slice(inventory.Cases, func(i, j int) bool { return scenarioRef(inventory.Cases[i]) < scenarioRef(inventory.Cases[j]) })
	sort.Slice(inventory.Candidates, func(i, j int) bool { return inventory.Candidates[i].Path < inventory.Candidates[j].Path })
	return inventory, diagnostics
}

func hasSuffix(path string, suffixes []string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(strings.ToLower(path), suffix) {
			return true
		}
	}
	return false
}

var e2eModuleRefPattern = regexp.MustCompile(`docs/design/prototypes/([a-z0-9]+(?:-[a-z0-9]+)*)/`)

func boundE2EModules(root string, state map[string]any) map[string]bool {
	bound, _ := state["bound_req"].(map[string]any)
	path, _ := bound["path"].(string)
	if strings.TrimSpace(path) == "" {
		return nil
	}
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(filepath.Clean(path))))
	if err != nil {
		return nil
	}
	modules := map[string]bool{}
	for _, match := range e2eModuleRefPattern.FindAllStringSubmatch(string(data), -1) {
		if len(match) == 2 {
			modules[match[1]] = true
		}
	}
	return modules
}

func scenarioRef(s e2eScenario) string { return s.CasesPath + "#" + s.ID }

// A comment is still just a candidate, but CASE-0010 is not CASE-001.
func mentionsCaseID(text, id string) bool {
	if id == "" {
		return false
	}
	word := func(b byte) bool {
		return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9' || b == '_' || b == '-'
	}
	for offset := 0; offset < len(text); {
		i := strings.Index(text[offset:], id)
		if i < 0 {
			return false
		}
		i += offset
		end := i + len(id)
		if (i == 0 || !word(text[i-1])) && (end == len(text) || !word(text[end])) {
			return true
		}
		offset = end
	}
	return false
}
