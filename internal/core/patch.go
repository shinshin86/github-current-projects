package core

import (
	"fmt"
	"strings"
)

// PatchResult holds the result of a README patch operation.
type PatchResult struct {
	Content string
	Patched bool
}

// PatchREADME replaces the marker section in the existing README content
// with the new section. If appendIfMissing is true and markers are not found,
// the section is appended at the end.
func PatchREADME(existing, newSection, marker string, appendIfMissing bool) (PatchResult, error) {
	beginMarker := fmt.Sprintf("<!-- BEGIN %s -->", marker)
	endMarker := fmt.Sprintf("<!-- END %s -->", marker)

	// Detect line ending style
	lineEnding := detectLineEnding(existing)

	beginIdx := strings.Index(existing, beginMarker)
	endIdx := strings.Index(existing, endMarker)

	if beginIdx == -1 && endIdx == -1 {
		// No markers found
		if !appendIfMissing {
			return PatchResult{}, fmt.Errorf("marker %q not found in README; use --append-if-missing to add it", marker)
		}
		// Append at the end
		separator := lineEnding
		if !strings.HasSuffix(existing, lineEnding) && existing != "" {
			separator = lineEnding + lineEnding
		} else if existing != "" {
			separator = lineEnding
		}
		adapted := adaptLineEnding(newSection, lineEnding)
		return PatchResult{
			Content: existing + separator + adapted,
			Patched: true,
		}, nil
	}

	if beginIdx == -1 || endIdx == -1 {
		return PatchResult{}, fmt.Errorf("only one of BEGIN/END markers for %q found; README markers are inconsistent", marker)
	}

	if beginIdx > endIdx {
		return PatchResult{}, fmt.Errorf("END marker appears before BEGIN marker for %q", marker)
	}

	// Find the end of the END marker line
	endOfEndMarker := endIdx + len(endMarker)
	if endOfEndMarker < len(existing) && existing[endOfEndMarker] == '\r' {
		endOfEndMarker++
	}
	if endOfEndMarker < len(existing) && existing[endOfEndMarker] == '\n' {
		endOfEndMarker++
	}

	before := existing[:beginIdx]
	after := existing[endOfEndMarker:]

	adapted := adaptLineEnding(newSection, lineEnding)

	return PatchResult{
		Content: before + adapted + after,
		Patched: true,
	}, nil
}

// detectLineEnding returns "\r\n" if the content uses CRLF, otherwise "\n".
func detectLineEnding(content string) string {
	if strings.Contains(content, "\r\n") {
		return "\r\n"
	}
	return "\n"
}

// adaptLineEnding converts LF newlines to the target line ending.
func adaptLineEnding(content, lineEnding string) string {
	if lineEnding == "\r\n" {
		// First normalize to LF, then convert to CRLF
		content = strings.ReplaceAll(content, "\r\n", "\n")
		content = strings.ReplaceAll(content, "\n", "\r\n")
	}
	return content
}
