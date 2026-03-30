package main

import (
	"bufio"
	"os"
	"strings"
)

type EnvVar struct {
	Name        string
	Default     string
	Description string
	Section     string
	Sensitive   bool
	Placeholder bool
}

type State struct {
	Vars         []EnvVar
	Values       map[string]string
	Skipped      map[string]bool
	TemplatePath string
	OutputPath   string
}

// IsSet returns true if the key has a non-empty value and is not skipped.
func (s *State) IsSet(name string) bool {
	return s.Values[name] != "" && !s.Skipped[name]
}

func (s *State) CountSet() int {
	n := 0
	for _, v := range s.Vars {
		if s.IsSet(v.Name) {
			n++
		}
	}
	return n
}

func (s *State) CountSkipped() int { return len(s.Skipped) }

func parseTemplate(path string) ([]EnvVar, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var vars []EnvVar
	var comments []string
	section := ""

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			if len(comments) > 0 {
				section = extractSection(comments)
				comments = nil
			}
			continue
		}

		if strings.HasPrefix(trimmed, "#") {
			comments = append(comments, trimmed)
			continue
		}

		key, val := parseKeyValue(trimmed)
		if key == "" {
			continue
		}

		desc := ""
		if len(comments) > 0 {
			if isDecorative(comments[0]) || isSectionHeader(comments[0], len(comments) > 1) {
				section = extractSection(comments[:1])
				if len(comments) > 1 {
					desc = joinComments(comments[1:])
				}
			} else {
				desc = joinComments(comments)
			}
		}
		comments = nil

		vars = append(vars, EnvVar{
			Name:        key,
			Default:     val,
			Description: desc,
			Section:     section,
			Sensitive:   isSensitive(key),
			Placeholder: isPlaceholder(val),
		})
	}

	return vars, scanner.Err()
}

func parseExisting(path string) (map[string]string, map[string]bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return make(map[string]string), make(map[string]bool), err
	}
	defer f.Close()

	values := make(map[string]string)
	skipped := make(map[string]bool)
	nextSkipped := false

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "# env-pilot:skipped" {
			nextSkipped = true
			continue
		}

		if strings.HasPrefix(trimmed, "#") || trimmed == "" {
			nextSkipped = false
			continue
		}

		key, val := parseKeyValue(trimmed)
		if key == "" {
			nextSkipped = false
			continue
		}

		values[key] = val
		if nextSkipped {
			skipped[key] = true
		}
		nextSkipped = false
	}

	return values, skipped, scanner.Err()
}

func parseKeyValue(line string) (string, string) {
	line = strings.TrimPrefix(line, "export ")
	idx := strings.IndexByte(line, '=')
	if idx < 1 {
		return "", ""
	}
	key := strings.TrimSpace(line[:idx])
	val := strings.TrimSpace(line[idx+1:])

	// Strip surrounding quotes — track whether value was quoted
	quoted := false
	if len(val) >= 2 {
		if (val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'') {
			val = val[1 : len(val)-1]
			quoted = true
		}
	}

	// Strip inline comments only for unquoted values
	if !quoted {
		if idx := strings.Index(val, " #"); idx >= 0 {
			val = strings.TrimSpace(val[:idx])
		}
	}

	return key, val
}

func isSensitive(name string) bool {
	upper := strings.ToUpper(name)
	for _, p := range []string{"_KEY", "_SECRET", "_TOKEN", "_PASSWORD", "CREDENTIAL", "_PRIVATE"} {
		if strings.Contains(upper, p) {
			return true
		}
	}
	return false
}

func isPlaceholder(val string) bool {
	if val == "" {
		return false
	}
	lower := strings.ToLower(val)
	for _, p := range []string{"your-", "your_", "change-me", "changeme", "xxx", "todo", "replace", "insert-", "enter-", "put-your"} {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

func isDecorative(comment string) bool {
	return strings.Contains(comment, "──") || strings.Contains(comment, "===") ||
		strings.Contains(comment, "---") || strings.Contains(comment, "###")
}

// isSectionHeader detects short, label-like comments that look like section names
// (e.g., "# Database") when followed by longer description comments.
func isSectionHeader(comment string, hasMore bool) bool {
	if !hasMore {
		return false
	}
	text := strings.TrimPrefix(comment, "#")
	text = strings.TrimSpace(text)
	// Short, no sentence-ending punctuation, no URLs — looks like a label
	return len(text) <= 30 && !strings.ContainsAny(text, ".:;,/()=")
}

func extractSection(comments []string) string {
	if len(comments) == 0 {
		return ""
	}
	s := comments[0]
	s = strings.TrimPrefix(s, "#")
	s = strings.TrimSpace(s)
	for _, dec := range []string{"──", "─", "===", "==", "---", "--"} {
		s = strings.ReplaceAll(s, dec, "")
	}
	return strings.TrimSpace(s)
}

func joinComments(comments []string) string {
	var lines []string
	for _, c := range comments {
		c = strings.TrimPrefix(c, "#")
		c = strings.TrimSpace(c)
		if c != "" {
			lines = append(lines, c)
		}
	}
	return strings.Join(lines, " ")
}

func detectTemplate() string {
	candidates := []string{".env.example", ".env.sample", ".env.template", ".env.dist", ".env.defaults"}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}
