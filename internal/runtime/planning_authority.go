package runtime

import (
	"fmt"
	"reflect"

	"github.com/entroforge/go-system-builder/internal/fileview"
	"github.com/entroforge/go-system-builder/internal/workpacket"
)

// CurrentDesignPacket projects an actual Main responsibility, not a workgroup.
// Issuance and consumption must repeat ValidateDesignPacket inside the Writer.
func CurrentDesignPacket(state map[string]any) (workpacket.DesignPacket, error) {
	if !PacketEnabled(state) {
		return workpacket.DesignPacket{}, fmt.Errorf("explicit packet mode is required")
	}
	life, _ := state["lifecycle"].(map[string]any)
	if life["state"] != "planning" || life["phase"] != "design" {
		return workpacket.DesignPacket{}, fmt.Errorf("S2 requires the original planning.design cursor")
	}
	bound, _ := state["bound_req"].(map[string]any)
	if bound == nil || bound["status"] != "locked" {
		return workpacket.DesignPacket{}, fmt.Errorf("S2 requires the authorized locked REQ")
	}
	baseline, _ := state["baseline"].(map[string]any)
	generation, err := integerField(baseline, "generation")
	if err != nil {
		return workpacket.DesignPacket{}, err
	}
	entry, err := integerField(life, "phase_revision")
	if err != nil {
		return workpacket.DesignPacket{}, err
	}
	return workpacket.BuildDesign(workpacket.PlanningBinding{RuntimeID: stringValue(state["runtime_id"]), Generation: generation, StageEntry: entry, Stage: "S2", Owner: "Main", Requirement: workpacket.Reference{ID: stringValue(bound["id"]), Path: stringValue(bound["path"]), Version: stringValue(bound["version"]), SHA256: stringValue(bound["sha256"])}})
}

func ValidateDesignPacket(root string, state map[string]any, p workpacket.DesignPacket, files fileview.Reader) error {
	if files == nil {
		return fmt.Errorf("design packet requires original authoritative file view")
	}
	if err := fileview.ValidateAuthority(root, state); err != nil {
		return err
	}
	expected, err := CurrentDesignPacket(state)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(expected, p) {
		return fmt.Errorf("design packet differs from actual current Main stage and template")
	}
	view, err := files.ReadFile(p.Binding.Requirement.Path)
	if err != nil {
		return err
	}
	disk, err := workpacket.ReadFile(root, p.Binding.Requirement.Path, 2<<20)
	if err != nil {
		return err
	}
	if workpacket.Digest(view) != p.Binding.Requirement.SHA256 || workpacket.Digest(disk) != p.Binding.Requirement.SHA256 {
		return fmt.Errorf("design packet frozen REQ source drifted")
	}
	if verified, ok := files.(interface{ Verify() error }); ok {
		if err := verified.Verify(); err != nil {
			return fmt.Errorf("design packet source moved: %w", err)
		}
	}
	return nil
}
