package main

import (
	"fmt"
	"os"
	"strings"
)

// writeEnv writes the .env file, preserving template key order.
func writeEnv(s *State) error {
	var b strings.Builder

	for _, v := range s.Vars {
		if s.Skipped[v.Name] {
			b.WriteString("# envsetup:skipped\n")
		}

		val := s.Values[v.Name]

		// Quote values that contain spaces, #, or special chars
		if needsQuoting(val) {
			val = fmt.Sprintf("%q", val)
		}

		fmt.Fprintf(&b, "%s=%s\n", v.Name, val)
	}

	return os.WriteFile(s.OutputPath, []byte(b.String()), 0644)
}

func needsQuoting(val string) bool {
	if val == "" {
		return false
	}
	return strings.ContainsAny(val, " \t#\"'\\$`") || strings.Contains(val, "\n")
}
