package cli

import (
	"fmt"
	"path/filepath"

	"github.com/entroforge/go-system-builder/internal/submission"
	"github.com/entroforge/go-system-builder/internal/workpacket"
)

// The existing bounded display cache records attempts, never reception facts.
// Packet identity isolates owners, rounds and revisions without reading drafts.
func nativeS7DiagnosticKind(packet string) string { return "S7 native Stop " + packet }

func nativeS7ReturnDiagnostic(root string, state map[string]any, p workpacket.ReviewResultPacket) string {
	references := fmt.Sprintf("S7 owner %s; packet %s; view %s; answer %s. ", p.Binding.OwnerAgentID, p.ID, filepath.Join(root, p.ViewPath), filepath.Join(root, p.AnswerPath))
	if diagnostic := submission.PlanningDiagnostic(root, state, nativeS7DiagnosticKind(p.ID)); diagnostic != "" {
		return references + diagnostic + ". This is the last attempted reception, not a fresh validation or an accepted result. Main restores the named source/environment gap or resumes this original Reviewer for the specific professional omissions. "
	}
	return references + "No original Stop diagnostic is recorded; check the actual native execution/return before deciding recovery. "
}
