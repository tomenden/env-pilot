package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	from := flag.String("from", "", "Path to template file (default: auto-detect)")
	out := flag.String("out", ".env", "Path to output file")
	review := flag.Bool("review", false, "Show review screen")
	status := flag.Bool("status", false, "Print key counts and exit (non-interactive)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: env-pilot [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Guided, interactive .env file setup from a template.\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	templatePath := *from
	if templatePath == "" {
		templatePath = detectTemplate()
		if templatePath == "" {
			fmt.Fprintln(os.Stderr, "No template file found. Looked for: .env.example, .env.sample, .env.template, .env.dist, .env.defaults")
			os.Exit(1)
		}
	}

	vars, err := parseTemplate(templatePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", templatePath, err)
		os.Exit(1)
	}

	values, skipped, err := parseExisting(*out)
	if err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", *out, err)
		os.Exit(1)
	}

	state := &State{
		Vars:          vars,
		Values:        values,
		Skipped:       skipped,
		TemplatePath: templatePath,
		OutputPath:   *out,
	}

	if *status {
		printStatus(state)
		return
	}

	if *review {
		printReview(state)
		return
	}

	m := newModel(state)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
