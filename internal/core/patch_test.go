package core

import (
	"strings"
	"testing"
)

func TestPatchREADMENormal(t *testing.T) {
	existing := `# Profile

<!-- BEGIN CURRENT PROJECTS -->
## Current Projects

- [old](https://github.com/u/old) (Go) - Old project
<!-- END CURRENT PROJECTS -->

## About
`
	newSection := `<!-- BEGIN CURRENT PROJECTS -->
## Current Projects

- [new](https://github.com/u/new) (Rust) - New project
<!-- END CURRENT PROJECTS -->
`

	result, err := PatchREADME(existing, newSection, "CURRENT PROJECTS", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Patched {
		t.Error("expected Patched=true")
	}
	if !strings.Contains(result.Content, "[new]") {
		t.Error("new section not found")
	}
	if strings.Contains(result.Content, "[old]") {
		t.Error("old section should have been replaced")
	}
	if !strings.Contains(result.Content, "# Profile") {
		t.Error("content before marker should be preserved")
	}
	if !strings.Contains(result.Content, "## About") {
		t.Error("content after marker should be preserved")
	}
}

func TestPatchREADMENoMarkers(t *testing.T) {
	existing := "# Profile\n\n## About\n"
	newSection := "<!-- BEGIN CURRENT PROJECTS -->\n## Current Projects\n<!-- END CURRENT PROJECTS -->\n"

	_, err := PatchREADME(existing, newSection, "CURRENT PROJECTS", false)
	if err == nil {
		t.Fatal("expected error when markers not found")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestPatchREADMEAppendIfMissing(t *testing.T) {
	existing := "# Profile\n\n## About\n"
	newSection := "<!-- BEGIN CURRENT PROJECTS -->\n## Current Projects\n<!-- END CURRENT PROJECTS -->\n"

	result, err := PatchREADME(existing, newSection, "CURRENT PROJECTS", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Patched {
		t.Error("expected Patched=true")
	}
	if !strings.Contains(result.Content, "<!-- BEGIN CURRENT PROJECTS -->") {
		t.Error("appended section not found")
	}
	if !strings.Contains(result.Content, "# Profile") {
		t.Error("original content should be preserved")
	}
}

func TestPatchREADMEOnlyBeginMarker(t *testing.T) {
	existing := "# Profile\n<!-- BEGIN CURRENT PROJECTS -->\nstuff\n"
	newSection := "new\n"

	_, err := PatchREADME(existing, newSection, "CURRENT PROJECTS", false)
	if err == nil {
		t.Fatal("expected error for inconsistent markers")
	}
	if !strings.Contains(err.Error(), "inconsistent") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestPatchREADMEOnlyEndMarker(t *testing.T) {
	existing := "# Profile\n<!-- END CURRENT PROJECTS -->\nstuff\n"
	newSection := "new\n"

	_, err := PatchREADME(existing, newSection, "CURRENT PROJECTS", false)
	if err == nil {
		t.Fatal("expected error for inconsistent markers")
	}
}

func TestPatchREADMEReversedMarkers(t *testing.T) {
	existing := "# Profile\n<!-- END CURRENT PROJECTS -->\nstuff\n<!-- BEGIN CURRENT PROJECTS -->\n"
	newSection := "new\n"

	_, err := PatchREADME(existing, newSection, "CURRENT PROJECTS", false)
	if err == nil {
		t.Fatal("expected error for reversed markers")
	}
	if !strings.Contains(err.Error(), "before BEGIN") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestPatchREADMECRLF(t *testing.T) {
	existing := "# Profile\r\n\r\n<!-- BEGIN CURRENT PROJECTS -->\r\n## Old\r\n<!-- END CURRENT PROJECTS -->\r\n\r\n## About\r\n"
	newSection := "<!-- BEGIN CURRENT PROJECTS -->\n## Current Projects\n\n- [repo](url)\n<!-- END CURRENT PROJECTS -->\n"

	result, err := PatchREADME(existing, newSection, "CURRENT PROJECTS", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The new section should be adapted to CRLF
	if strings.Contains(result.Content, "\n") && !strings.Contains(result.Content, "\r\n") {
		t.Error("expected CRLF line endings in patched content")
	}

	// Verify no bare LF exists (all \n should be preceded by \r)
	for i, c := range result.Content {
		if c == '\n' && (i == 0 || result.Content[i-1] != '\r') {
			t.Errorf("found bare LF at position %d", i)
			break
		}
	}
}

func TestPatchREADMEPreservesLF(t *testing.T) {
	existing := "# Profile\n\n<!-- BEGIN CURRENT PROJECTS -->\n## Old\n<!-- END CURRENT PROJECTS -->\n\n## About\n"
	newSection := "<!-- BEGIN CURRENT PROJECTS -->\n## Current Projects\n<!-- END CURRENT PROJECTS -->\n"

	result, err := PatchREADME(existing, newSection, "CURRENT PROJECTS", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(result.Content, "\r\n") {
		t.Error("LF content should not gain CRLF")
	}
}
