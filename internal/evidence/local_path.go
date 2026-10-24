package evidence

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
)

// S8EvidencePath identifies a captured local source without creating another
// evidence index. The original path identity and content both remain bound.
func S8EvidencePath(packetID, sourcePath, digest string) string {
	key := sha256.Sum256([]byte(sourcePath))
	return ".claude/evidence/s8/" + packetID + "/" + hex.EncodeToString(key[:]) + "/" + digest
}

// LocalPathReference parses the existing path: spelling for new observations.
// Missing hashes are authored inputs, never accepted immutable references.
func LocalPathReference(ref string) (string, string, error) {
	rel, digest, err := ParsePathReference(ref)
	if err != nil {
		return "", "", err
	}
	clean := filepath.ToSlash(filepath.Clean(rel))
	if !strings.HasPrefix(ref, "path:") || rel == "" || filepath.IsAbs(rel) || strings.Contains(rel, "\\") || clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", "", fmt.Errorf("path evidence must name a repository-contained file")
	}
	return clean, digest, nil
}

// MatchesS8PathBinding verifies the structural projection of original author
// input into the immutable Case. Physical bytes are checked at reception, not
// by a read-only historical receipt lookup.
func MatchesS8PathBinding(packetID, authored, bound string) bool {
	rel, want, err := LocalPathReference(authored)
	if err != nil {
		return false
	}
	captured, digest, err := LocalPathReference(bound)
	if err != nil || digest == "" || (want != "" && want != digest) {
		return false
	}
	return captured == S8EvidencePath(packetID, rel, digest)
}
