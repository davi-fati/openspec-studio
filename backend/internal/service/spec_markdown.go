package service

import (
	"strings"

	"github.com/davidlima/openspec-studio/backend/internal/domain"
)

// parsedSpecMarkdown is spec.md broken into its Purpose and Requirement
// blocks. Each requirement's body (description + scenarios) is kept as raw
// markdown rather than a deeper AST, so edits round-trip byte-for-byte
// through save/reload without a full markdown parser.
type parsedSpecMarkdown struct {
	purpose      string
	requirements []domain.SpecRequirement
}

func parseSpecMarkdown(content string) parsedSpecMarkdown {
	lines := strings.Split(content, "\n")
	var purpose strings.Builder
	var requirements []domain.SpecRequirement
	section := ""
	var curName string
	var curBody strings.Builder

	flush := func() {
		if curName != "" {
			requirements = append(requirements, domain.SpecRequirement{
				Name: curName,
				Body: strings.TrimRight(curBody.String(), "\n"),
			})
		}
		curName = ""
		curBody.Reset()
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "## Purpose"):
			flush()
			section = "purpose"
			continue
		case strings.HasPrefix(trimmed, "## Requirements"), strings.HasPrefix(trimmed, "## ADDED Requirements"):
			flush()
			section = "requirements"
			continue
		case strings.HasPrefix(trimmed, "## "):
			flush()
			section = ""
			continue
		}

		switch section {
		case "purpose":
			purpose.WriteString(line)
			purpose.WriteString("\n")
		case "requirements":
			if strings.HasPrefix(trimmed, "### Requirement:") {
				flush()
				curName = strings.TrimSpace(strings.TrimPrefix(trimmed, "### Requirement:"))
				continue
			}
			if curName != "" {
				curBody.WriteString(line)
				curBody.WriteString("\n")
			}
		}
	}
	flush()

	return parsedSpecMarkdown{
		purpose:      strings.TrimSpace(purpose.String()),
		requirements: requirements,
	}
}

func renderSpecMarkdown(purpose string, requirements []domain.SpecRequirement) string {
	var b strings.Builder
	b.WriteString("## Purpose\n\n")
	b.WriteString(strings.TrimSpace(purpose))
	b.WriteString("\n\n## Requirements\n")
	for _, r := range requirements {
		b.WriteString("\n### Requirement: ")
		b.WriteString(r.Name)
		b.WriteString("\n")
		b.WriteString(strings.Trim(r.Body, "\n"))
		b.WriteString("\n")
	}
	return b.String()
}
