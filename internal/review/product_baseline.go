package review

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"sort"

	"github.com/entroforge/go-system-builder/internal/buildsurface"
)

// readCanonicalProductBaseline chooses the strict shared S6 projection before
// any legacy completion-report projection. A malformed canonical registry is
// an error from Read, never a reason to fall back to legacy evidence.
func readCanonicalProductBaseline(root string, state map[string]any) (*buildsurface.Snapshot, error) {
	return readCanonicalProductBaselineContext(context.Background(), root, state)
}

func readCanonicalProductBaselineContext(ctx context.Context, root string, state map[string]any) (*buildsurface.Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !buildsurface.HasCanonical(state) {
		return nil, nil
	}
	snapshot, err := buildsurface.ReadContext(ctx, root, state)
	if err != nil {
		return nil, fmt.Errorf("canonical S6 product baseline: %w", err)
	}
	return &snapshot, nil
}

// frozenSubjectsForProductBaseline projects the exact task sources and final
// product diff into the ReviewPlan's established frozen_subjects DTO. Git
// deletion preimage authority stays in ProductBaseline; Kind=deleted is only
// the review-facing declaration and never proves deletion by itself.
func frozenSubjectsForProductBaseline(snapshot *buildsurface.Snapshot) ([]FrozenSubject, error) {
	if snapshot == nil {
		return nil, fmt.Errorf("canonical product baseline is missing")
	}
	byPath := make(map[string]FrozenSubject, len(snapshot.Tasks)+len(snapshot.Paths))
	for _, task := range snapshot.Tasks {
		if task.Path == "" || task.SHA256 == "" {
			return nil, fmt.Errorf("canonical product baseline has an incomplete TASK source")
		}
		if previous, exists := byPath[task.Path]; exists && previous.SHA256 != task.SHA256 {
			return nil, fmt.Errorf("canonical product baseline has conflicting frozen bytes for %s", task.Path)
		}
		byPath[task.Path] = FrozenSubject{Path: task.Path, SHA256: task.SHA256, Kind: "task"}
	}
	for _, path := range snapshot.Paths {
		if path.Path == "" || path.SHA256 == "" {
			return nil, fmt.Errorf("canonical product baseline has an incomplete changed path")
		}
		if path.Status == "deleted" {
			// Unattributed deleted surface stays in review coverage through the
			// baseline snapshot and drift notes; it is not a frozen subject, so
			// its reappearance never blocks ReviewPlan registration. A TASK
			// source colliding with a deleted surface is still rejected below.
			continue
		}
		kind := "changed_surface"
		if path.Status != "added" && path.Status != "modified" {
			return nil, fmt.Errorf("canonical product baseline path %s has unknown status %q", path.Path, path.Status)
		}
		if previous, exists := byPath[path.Path]; exists {
			if previous.SHA256 != path.SHA256 {
				return nil, fmt.Errorf("TASK source and changed surface disagree on frozen bytes for %s", path.Path)
			}
			// Deleted surface never reaches here (skipped above); a TASK path
			// stays a TASK source in frozen_subjects. Its product status and
			// Git blob binding are separately covered by the snapshot.
			continue
		}
		byPath[path.Path] = FrozenSubject{Path: path.Path, SHA256: path.SHA256, Kind: kind}
	}
	out := make([]FrozenSubject, 0, len(byPath))
	for _, subject := range byPath {
		out = append(out, subject)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

func verifyProductBaselineDeclaration(root string, state map[string]any, plan *Plan) error {
	return verifyProductBaselineDeclarationContext(context.Background(), root, state, plan)
}

func verifyProductBaselineDeclarationContext(ctx context.Context, root string, state map[string]any, plan *Plan) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if plan == nil {
		return fmt.Errorf("product baseline verification requires a ReviewPlan")
	}
	canonical := buildsurface.HasCanonical(state)
	if !canonical {
		if plan.ProductBaseline != nil {
			return fmt.Errorf("ReviewPlan declares a canonical product baseline but Runtime has no canonical S6 Assignment registry")
		}
		return nil
	}
	if plan.ProductBaseline == nil {
		return fmt.Errorf("canonical S6 generation requires ReviewPlan.product_baseline")
	}
	if err := buildsurface.VerifyContext(ctx, root, state, *plan.ProductBaseline); err != nil {
		return fmt.Errorf("ReviewPlan product baseline is stale or invalid: %w", err)
	}
	want, err := frozenSubjectsForProductBaseline(plan.ProductBaseline)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(plan.FrozenSubjects, want) {
		return fmt.Errorf("ReviewPlan frozen_subjects do not exactly match canonical TASK sources and changed paths")
	}
	return ctx.Err()
}

// verifyCanonicalFrozenSubjectFiles complements the Git-side product baseline
// proof with the live worktree check. Deleted paths must remain absent; their
// preimage hash is verified from the ProductBaseline's pinned Git object.
func verifyCanonicalFrozenSubjectFiles(root string, plan *Plan, contexts ...context.Context) error {
	ctx := baselineContext(contexts)
	for _, subject := range plan.FrozenSubjects {
		if err := ctx.Err(); err != nil {
			return err
		}
		path, err := repositoryContainedPath(root, subject.Path)
		if err != nil {
			return fmt.Errorf("frozen subject %s is outside repository: %w", subject.Path, err)
		}
		if subject.Kind == "deleted" {
			// Deleted surface no longer enters frozen subjects from the
			// canonical baseline; a surviving deleted subject is a legacy
			// TASK-attributed deletion and must remain absent.
			if _, statErr := os.Lstat(path); statErr == nil {
				return fmt.Errorf("frozen deleted subject %s reappeared on disk; restore its absence before registration", subject.Path)
			} else if !os.IsNotExist(statErr) {
				return fmt.Errorf("frozen deleted subject %s cannot be checked: %w", subject.Path, statErr)
			}
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("frozen subject %s is unreadable: %w", subject.Path, err)
		}
		if actual := sha256Of(data); actual != subject.SHA256 {
			return fmt.Errorf("frozen subject %s drifted: plan pins %s but disk contains %s", subject.Path, subject.SHA256, actual)
		}
	}
	return nil
}

// baselineContext chooses the request budget for legacy-compatible helper APIs.
func baselineContext(contexts []context.Context) context.Context {
	if len(contexts) > 0 && contexts[0] != nil {
		return contexts[0]
	}
	return context.Background()
}
