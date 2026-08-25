package review

import (
	"fmt"
	"sort"
	"strings"
)

func validateE2EAssetDeclarations(plan *Plan) error {
	if len(plan.E2EAssets) == 0 {
		return s7GateError(
			"S7_E2E_ASSET_FINGERPRINT",
			"regression_available requires at least one fingerprinted reusable E2E asset",
			[]string{"no CASE/PATH asset with a path and sha256 is declared"},
			[]string{"add e2e_assets entries for the existing spec/fixture/selector/environment assets, or change the plan to cold_start"},
			"runtime review-plan --file plan.json --expected-revision <N>",
		)
	}
	seen := map[string]bool{}
	for _, asset := range plan.E2EAssets {
		if strings.TrimSpace(asset.AssetID) == "" || strings.TrimSpace(asset.CaseRef) == "" || strings.TrimSpace(asset.Path) == "" || len(asset.SHA256) != 64 {
			return s7GateError(
				"S7_E2E_ASSET_FINGERPRINT",
				fmt.Sprintf("E2E asset %q has an incomplete fingerprint declaration", asset.AssetID),
				[]string{"asset_id, case_ref, repository path and 64-character sha256 are required"},
				[]string{"complete the e2e_assets entry from the current CASE/PATH files"},
				"runtime review-plan --file plan.json --expected-revision <N>",
			)
		}
		if seen[asset.AssetID] {
			return s7GateError(
				"S7_E2E_ASSET_FINGERPRINT",
				fmt.Sprintf("E2E asset %q is declared more than once", asset.AssetID),
				[]string{"duplicate asset_id"},
				[]string{"keep one immutable entry per reusable E2E asset"},
				"runtime review-plan --file plan.json --expected-revision <N>",
			)
		}
		seen[asset.AssetID] = true
	}
	return nil
}

func sortE2EAssets(assets []E2EAsset) []E2EAsset {
	result := append([]E2EAsset(nil), assets...)
	sort.Slice(result, func(i, j int) bool { return result[i].AssetID < result[j].AssetID })
	return result
}
