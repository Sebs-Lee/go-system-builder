package agentcontext_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/entroforge/go-system-builder/internal/agentcontext"
)

func removeSkillBlock(data []byte) string {
	source := string(data)
	start := strings.Index(source, "skills:") + len("skills:")
	tail := strings.Index(source[start:], "\n---")
	return source[:start] + source[start+tail:]
}
func TestFirstBuilderCoreRolePreservesEveryAuthorityFieldAndBody(t *testing.T) {
	original, err := os.ReadFile("../../agents/backend-builder.md")
	if err != nil {
		t.Fatal(err)
	}
	projected, err := agentcontext.RenderCoreRole("backend-builder", original)
	if err != nil {
		t.Fatal(err)
	}
	if removeSkillBlock(original) != removeSkillBlock(projected) {
		t.Fatal("core generation changed non-Skill fields or responsibility body")
	}
	frontmatter := strings.SplitN(string(projected), "---", 3)[1]
	for _, name := range []string{"agent-dispatch", "backend-engineering", "testing-strategy", "code-quality"} {
		if !strings.Contains(frontmatter, "  - "+name+"\n") {
			t.Fatalf("missing core %s", name)
		}
	}
	for _, name := range []string{"gorm", "gin", "s3-object-storage", "jwt-authentication"} {
		if strings.Contains(frontmatter, "  - "+name+"\n") {
			t.Fatalf("undeclared stack was still preloaded: %s", name)
		}
	}
	if len(projected) >= len(original) || bytes.Equal(projected, original) {
		t.Fatal("generation did not remove broad preload")
	}
	// The generation is not an activation producer and cannot mutate its source.
	reread, _ := os.ReadFile("../../agents/backend-builder.md")
	if !bytes.Equal(original, reread) {
		t.Fatal("generator modified the installed/source role")
	}
}
func TestCoreRoleRejectsAmbiguousFrontmatterAndKeepsCRLFAuthorityBytes(t *testing.T) {
	source := []byte("---\r\nname: backend-builder\r\ntools: Read, Write\r\npermissionMode: default\r\nskills:\r\n  - gorm\r\nmaxTurns: 9\r\n---\r\nExact body\r\n")
	projected, err := agentcontext.RenderCoreRole("backend-builder", source)
	if err != nil || !bytes.HasSuffix(projected, []byte("maxTurns: 9\r\n---\r\nExact body\r\n")) || !bytes.Contains(projected, []byte("tools: Read, Write\r\npermissionMode: default\r\n")) {
		t.Fatalf("CRLF authority changed: %v %s", err, projected)
	}
	for _, bad := range []string{"---\nname: backend-builder\nskills: []\n---\nbody", "---\nname: backend-builder\nskills:\n  - a\nskills:\n  - b\n---\nbody", "---\nname: foreign-role\nskills:\n  - a\n---\nbody"} {
		if _, err := agentcontext.RenderCoreRole("backend-builder", []byte(bad)); err == nil {
			t.Fatal("ambiguous/different role was generated")
		}
	}
	if _, err := agentcontext.RenderCoreRole("frontend-builder", source); err == nil {
		t.Fatal("unsupported rollout role became enabled")
	}
}
