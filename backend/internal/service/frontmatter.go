package service

import (
	"fmt"
	"strings"

	"github.com/goccy/go-yaml"
)

// splitFrontmatter separates a "---\n<yaml>\n---\n<body>" file into its YAML
// frontmatter and markdown body. Used for review/debt-item files under
// openspec/context/, keeping the same "plain file, no database" convention
// specs and changes already use.
func splitFrontmatter(content string) (front string, body string) {
	const delim = "---"
	if !strings.HasPrefix(content, delim) {
		return "", content
	}
	rest := content[len(delim):]
	idx := strings.Index(rest, "\n"+delim)
	if idx == -1 {
		return "", content
	}
	front = strings.TrimSpace(rest[:idx])
	body = strings.TrimPrefix(rest[idx+len("\n"+delim):], "\n")
	return front, strings.TrimSpace(body) + "\n"
}

func renderFrontmatter(meta any, body string) (string, error) {
	yamlBytes, err := yaml.Marshal(meta)
	if err != nil {
		return "", fmt.Errorf("marshal frontmatter: %w", err)
	}
	return fmt.Sprintf("---\n%s---\n\n%s", string(yamlBytes), strings.TrimSpace(body)+"\n"), nil
}

func parseFrontmatter(content string, out any) (body string, err error) {
	front, body := splitFrontmatter(content)
	if front == "" {
		return body, nil
	}
	if err := yaml.Unmarshal([]byte(front), out); err != nil {
		return "", fmt.Errorf("parse frontmatter: %w", err)
	}
	return body, nil
}
