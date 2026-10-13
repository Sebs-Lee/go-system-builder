package runtime

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"

	"github.com/entroforge/go-system-builder/internal/fileview"
	"github.com/entroforge/go-system-builder/internal/workpacket"
)

// ValidateWorkPacket checks current ownership and the complete frozen subject
// set. It is repeated under the original Writer lock before consumption.
func ValidateWorkPacket(root string, state map[string]any, packet workpacket.Packet, files fileview.Reader) error {
	if !PacketEnabled(state) || files == nil {
		return fmt.Errorf("packet consumer is not enabled with an authoritative file view")
	}
	lifecycle, _ := state["lifecycle"].(map[string]any)
	baseline, _ := state["baseline"].(map[string]any)
	generation, err := integerField(baseline, "generation")
	if err != nil {
		return err
	}
	entry, err := integerField(lifecycle, "phase_revision")
	if err != nil {
		return err
	}
	b := packet.Binding
	if state["runtime_id"] != b.RuntimeID || generation != b.Generation || lifecycle["state"] != "document_verification" || entry != b.StageEntry {
		return fmt.Errorf("packet runtime, generation or stage entry is stale")
	}
	expected, err := workpacket.BuildS5(b, packet.Inputs)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(packet, expected) {
		return fmt.Errorf("packet differs from registered template projection")
	}
	if err := validatePacketAssignment(root, state, b); err != nil {
		return err
	}
	documents, _ := state["documents"].([]any)
	refs := []workpacket.Reference{}
	for _, raw := range documents {
		d, _ := raw.(map[string]any)
		g, err := integerField(d, "generation")
		if err != nil {
			return err
		}
		if g != generation || (d["status"] != "registered" && d["status"] != "locked") {
			continue
		}
		ref := workpacket.Reference{ID: stringValue(d["id"]), Path: stringValue(d["path"]), Version: stringValue(d["version"]), SHA256: stringValue(d["sha256"])}
		data, err := files.ReadFile(ref.Path)
		if err != nil {
			return fmt.Errorf("read frozen packet input: %w", err)
		}
		if workpacket.Digest(data) != ref.SHA256 {
			return fmt.Errorf("frozen packet input drifted")
		}
		physical, err := safeEvidencePath(root, ref.Path)
		if err != nil {
			return fmt.Errorf("packet input containment: %w", err)
		}
		disk, err := os.ReadFile(filepath.Join(root, physical))
		if err != nil {
			return fmt.Errorf("read packet input checkout: %w", err)
		}
		if workpacket.Digest(disk) != ref.SHA256 {
			return fmt.Errorf("packet input checkout drifted")
		}
		refs = append(refs, ref)
	}
	sort.Slice(refs, func(i, j int) bool { return refs[i].Path < refs[j].Path })
	if !reflect.DeepEqual(refs, packet.Inputs) {
		return fmt.Errorf("packet does not cover the exact current document set")
	}
	return nil
}

func validatePacketAssignment(root string, state map[string]any, b workpacket.Binding) error {
	entities, _ := state["entities"].(map[string]any)
	agents, _ := entities["agents"].([]any)
	var agent map[string]any
	for _, raw := range agents {
		a, _ := raw.(map[string]any)
		if a["id"] == b.Owner {
			agent = a
			break
		}
	}
	if agent == nil || (agent["role"] != "document-verifier" && agent["role"] != "document_verifier") || agent["state"] == "stopped" || agent["state"] == "blocked" || agent["state"] == "queued" {
		return fmt.Errorf("packet owner is not a current document reviewer")
	}
	teams, _ := entities["teams"].([]any)
	var team map[string]any
	for _, raw := range teams {
		t, _ := raw.(map[string]any)
		if t["id"] == agent["team_id"] {
			team = t
			break
		}
	}
	if team == nil || team["kind"] != "document_verifier" || (team["status"] != "planned" && team["status"] != "active") || !containsString(team["agent_ids"], b.Owner) || !containsString(team["responsibility_ids"], b.Responsibility) {
		return fmt.Errorf("packet owner has no current registered document responsibility")
	}
	manifestRef, _ := team["manifest_ref"].(string)
	if agent["prompt_ref"] != manifestRef+"#"+b.AssignmentID {
		return fmt.Errorf("packet assignment does not match registered prompt")
	}
	manifestPath, err := safeEvidencePath(root, manifestRef)
	if err != nil {
		return fmt.Errorf("packet manifest path: %w", err)
	}
	manifestData, err := os.ReadFile(filepath.Join(root, manifestPath))
	if err != nil {
		return fmt.Errorf("read packet manifest: %w", err)
	}
	if workpacket.Digest(manifestData) != b.AssignmentSHA {
		return fmt.Errorf("packet assignment manifest drifted")
	}
	var manifest struct {
		Assignments []struct {
			ID             string `json:"assignment_id"`
			Owner          string `json:"agent_id"`
			Responsibility string `json:"responsibility_id"`
		} `json:"assignments"`
	}
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return fmt.Errorf("decode packet manifest: %w", err)
	}
	for _, a := range manifest.Assignments {
		if a.ID == b.AssignmentID && a.Owner == b.Owner && a.Responsibility == b.Responsibility {
			return nil
		}
	}
	return fmt.Errorf("packet responsibility is not assigned to this owner")
}
