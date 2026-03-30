package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	clrCyan   = lipgloss.AdaptiveColor{Light: "6", Dark: "14"}
	clrGreen  = lipgloss.AdaptiveColor{Light: "2", Dark: "10"}
	clrYellow = lipgloss.AdaptiveColor{Light: "3", Dark: "11"}
	clrDim    = lipgloss.AdaptiveColor{Light: "245", Dark: "240"}
	clrBorder = lipgloss.AdaptiveColor{Light: "250", Dark: "238"}
)

var (
	stDim      = lipgloss.NewStyle().Foreground(clrDim)
	stBold     = lipgloss.NewStyle().Bold(true)
	stGreen    = lipgloss.NewStyle().Foreground(clrGreen)
	stYellow   = lipgloss.NewStyle().Foreground(clrYellow)
	stCyan     = lipgloss.NewStyle().Foreground(clrCyan)
	stSection  = lipgloss.NewStyle().Bold(true).Foreground(clrCyan)
	stBorder   = lipgloss.NewStyle().Foreground(clrBorder)
	stBadge    = lipgloss.NewStyle().Foreground(clrYellow)
	stInputBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(clrBorder).
			Padding(0, 1)
)

type inputMode int

const (
	modeNavigate inputMode = iota
	modeEdit
	modeConfirm
)

type model struct {
	state    *State
	mode     inputMode
	cursor   int
	input    textinput.Model
	width    int
	height   int
	saveErr  error

	confirmValue  string
	confirmReveal bool
	quitting      bool
}

func newModel(s *State) model {
	ti := textinput.New()
	ti.Prompt = "  ▸ "
	ti.PromptStyle = stCyan
	ti.Focus()

	return model{
		state:  s,
		input:  ti,
		width:  80,
		height: 24,
	}
}

func (m model) Init() tea.Cmd { return textinput.Blink }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
	}

	switch m.mode {
	case modeConfirm:
		return m.updateConfirm(msg)
	case modeEdit:
		return m.updateEdit(msg)
	default:
		return m.updateNavigate(msg)
	}
}

func (m model) View() string { return m.viewMain() }

// --- State mutation helpers ---

func (m *model) setValue(name, val string) {
	m.state.Values[name] = val
	delete(m.state.Skipped, name)
	m.save()
}

func (m *model) skipValue(name string) {
	m.state.Skipped[name] = true
	delete(m.state.Values, name)
	m.save()
}

func (m *model) resetValue(name string) {
	delete(m.state.Values, name)
	delete(m.state.Skipped, name)
	m.save()
}

func (m *model) advance() {
	if m.cursor < len(m.state.Vars)-1 {
		m.cursor++
	}
}

func (m *model) save() {
	m.saveErr = writeEnv(m.state)
}

// --- Update handlers ---

func (m model) updateNavigate(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.state.Vars)-1 {
				m.cursor++
			}
		case "tab", "enter":
			m.mode = modeEdit
			m.setupInput()
			return m, textinput.Blink
		case "s":
			m.skipValue(m.state.Vars[m.cursor].Name)
			m.advance()
		case "r":
			m.resetValue(m.state.Vars[m.cursor].Name)
		case "w":
			m.save()
			m.quitting = true
			return m, tea.Quit
		case "q":
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) updateEdit(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "enter":
			return m.submitValue()
		case "esc":
			m.mode = modeNavigate
			return m, nil
		case "up":
			if m.input.Value() == "" {
				m.mode = modeNavigate
				if m.cursor > 0 {
					m.cursor--
				}
				return m, nil
			}
		case "down":
			if m.input.Value() == "" {
				m.mode = modeNavigate
				m.advance()
				return m, nil
			}
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "enter":
			m.setValue(m.state.Vars[m.cursor].Name, m.confirmValue)
			m.mode = modeNavigate
			m.advance()
			return m, nil
		case "r":
			m.confirmReveal = !m.confirmReveal
		case "x":
			m.mode = modeEdit
			m.input.Reset()
			return m, textinput.Blink
		case "esc":
			m.mode = modeNavigate
		}
	}
	return m, nil
}

func (m model) submitValue() (tea.Model, tea.Cmd) {
	v := m.state.Vars[m.cursor]
	val := strings.TrimSpace(m.input.Value())

	if val == "" {
		if v.Default != "" && !v.Placeholder {
			val = v.Default
		} else {
			m.skipValue(v.Name)
			m.mode = modeNavigate
			m.advance()
			return m, nil
		}
	}

	if v.Sensitive {
		m.mode = modeConfirm
		m.confirmValue = val
		m.confirmReveal = false
		return m, nil
	}

	m.setValue(v.Name, val)
	m.mode = modeNavigate
	m.advance()
	return m, nil
}

func (m *model) setupInput() {
	v := m.state.Vars[m.cursor]
	m.input.Reset()
	m.input.Placeholder = ""
	if v.Default != "" && !v.Placeholder {
		m.input.Placeholder = v.Default
	}
	if v.Sensitive {
		m.input.EchoMode = textinput.EchoPassword
		m.input.EchoCharacter = '•'
	} else {
		m.input.EchoMode = textinput.EchoNormal
	}
}

// --- View ---

func (m model) viewMain() string {
	lw := m.leftWidth()
	rw := m.width - lw - 1
	if rw < 20 {
		rw = 20
	}
	contentH := m.height - 3
	if contentH < 5 {
		contentH = 5
	}

	header := m.renderHeader()
	leftLines := m.renderKeyListLines(lw, contentH)
	rightLines := strings.Split(m.renderDetail(rw-4), "\n")

	for len(leftLines) < contentH {
		leftLines = append(leftLines, "")
	}
	for len(rightLines) < contentH {
		rightLines = append(rightLines, "")
	}

	div := stBorder.Render("│")
	var contentLines []string
	for i := 0; i < contentH; i++ {
		left := leftLines[i]
		right := ""
		if i < len(rightLines) {
			right = rightLines[i]
		}
		contentLines = append(contentLines, padToWidth(left, lw)+div+"  "+right)
	}

	return header + "\n" + strings.Join(contentLines, "\n") + "\n" + renderBottomBar()
}

func (m model) leftWidth() int {
	if m.width < 40 {
		return 30
	}
	w := m.width * 37 / 100
	if w < 30 {
		w = 30
	}
	if w > 50 {
		w = 50
	}
	return w
}

func (m model) renderHeader() string {
	set := m.state.CountSet()
	skipped := m.state.CountSkipped()
	remaining := len(m.state.Vars) - set - skipped

	left := stDim.Render(m.state.TemplatePath + " → " + m.state.OutputPath)
	right := stGreen.Render(fmt.Sprintf("✓ %d set", set)) + "  " +
		stYellow.Render(fmt.Sprintf("⏭ %d skipped", skipped)) + "  " +
		stDim.Render(fmt.Sprintf("○ %d remaining", remaining))

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right) - 4
	if gap < 2 {
		gap = 2
	}
	return "  " + left + strings.Repeat(" ", gap) + right
}

func (m model) renderKeyListLines(lw, maxH int) []string {
	var lines []string
	curSection := ""
	cursorLine := 0

	for i, v := range m.state.Vars {
		if v.Section != curSection {
			curSection = v.Section
			if curSection != "" {
				if len(lines) > 0 {
					lines = append(lines, "")
				}
				lines = append(lines, stSection.Render(curSection))
			}
		}

		if i == m.cursor {
			cursorLine = len(lines)
		}

		icon := m.varIcon(v)
		nameMax := lw*45/100 - 1
		if nameMax < 5 {
			nameMax = 5
		}
		name := truncStr(v.Name, nameMax)

		valMax := lw - nameMax - 6
		if valMax < 5 {
			valMax = 5
		}
		valPreview := m.keyPreview(v, valMax)

		if i == m.cursor {
			lines = append(lines, fmt.Sprintf(" %s %s %-*s  %s", stCyan.Render("▸"), icon, nameMax, name, valPreview))
		} else {
			lines = append(lines, fmt.Sprintf("   %s %-*s  %s", icon, nameMax, name, valPreview))
		}
	}

	// Scroll to keep cursor visible (pure computation, no mutation)
	scrollOff := 0
	if maxH > 0 && len(lines) > maxH {
		if cursorLine >= maxH {
			scrollOff = cursorLine - maxH + 1
		}
		lines = lines[scrollOff:]
		if len(lines) > maxH {
			lines = lines[:maxH]
		}
	}

	return lines
}

func (m model) varIcon(v EnvVar) string {
	if m.state.IsSet(v.Name) {
		return stGreen.Render("✓")
	}
	if m.state.Skipped[v.Name] {
		return stYellow.Render("⏭")
	}
	return stDim.Render("○")
}

func (m model) keyPreview(v EnvVar, maxW int) string {
	if m.state.IsSet(v.Name) {
		val := m.state.Values[v.Name]
		if v.Sensitive {
			val = maskValue(val)
		}
		return stDim.Render(truncStr(val, maxW))
	}
	if m.state.Skipped[v.Name] {
		return stDim.Render(truncStr("(skipped)", maxW))
	}
	return stDim.Render(truncStr("(not set)", maxW))
}

func (m model) renderDetail(maxW int) string {
	v := m.state.Vars[m.cursor]
	var b strings.Builder

	b.WriteString(stBold.Render(v.Name))

	if v.Sensitive {
		b.WriteString("\n\n" + stBadge.Render("🔒 sensitive"))
	}

	if v.Description != "" {
		b.WriteString("\n\n" + stDim.Render(v.Description))
	}

	if v.Default != "" {
		label := "Default"
		if v.Placeholder {
			label = "Placeholder"
		}
		b.WriteString("\n\n" + stDim.Render(label+": ") + stGreen.Render(v.Default))
	}

	if m.state.IsSet(v.Name) {
		display := m.state.Values[v.Name]
		if v.Sensitive {
			display = maskValue(display)
		}
		b.WriteString("\n" + stDim.Render("Current: ") + stCyan.Render(display))
	} else if m.state.Skipped[v.Name] {
		b.WriteString("\n\n" + stDim.Render("Status: ") + stYellow.Render("skipped"))
	} else {
		b.WriteString("\n\n" + stDim.Render("Status: ") + stYellow.Render("not set"))
	}

	switch m.mode {
	case modeEdit:
		b.WriteString("\n")
		if v.Sensitive {
			b.WriteString("\n" + stDim.Render("Enter value:"))
			b.WriteString("\n" + stInputBox.Width(maxW-4).Render(m.input.View()))
			b.WriteString("\n\n" + stCyan.Render("[↵]") + stDim.Render(" submit  ") + stCyan.Render("[esc]") + stDim.Render(" cancel"))
		} else {
			hint := " submit"
			if m.input.Value() == "" {
				if v.Default != "" && !v.Placeholder {
					hint = " accept default"
				} else {
					hint = " skip"
				}
			}
			b.WriteString("\n" + m.input.View())
			b.WriteString("\n\n" + stCyan.Render("[↵]") + stDim.Render(hint+"  ") + stCyan.Render("[esc]") + stDim.Render(" cancel"))
		}

	case modeConfirm:
		b.WriteString("\n")
		if m.confirmReveal {
			b.WriteString("\n  ▸ " + m.confirmValue)
		} else {
			b.WriteString("\n  ▸ " + stDim.Render(maskValue(m.confirmValue)))
		}
		b.WriteString("\n\n" + stCyan.Render("[r]") + stDim.Render("eveal to verify?"))
		b.WriteString("\n\n" + stCyan.Render("[↵]") + stDim.Render(" accept  ") + stCyan.Render("[x]") + stDim.Render(" redo  ") + stCyan.Render("[esc]") + stDim.Render(" cancel"))

	default:
		b.WriteString("\n\n" + stDim.Render("Press ") + stCyan.Render("[tab]") + stDim.Render(" or ") + stCyan.Render("[enter]") + stDim.Render(" to edit"))
	}

	return b.String()
}

func renderBottomBar() string {
	items := []struct{ key, label string }{
		{"↑↓", " navigate"},
		{"⇥", " edit"},
		{"[s]", "kip"},
		{"[r]", "eset"},
		{"[w]", "rite & exit"},
		{"[q]", "uit"},
	}
	var parts []string
	for _, item := range items {
		parts = append(parts, stCyan.Render(item.key)+stDim.Render(item.label))
	}
	return "  " + strings.Join(parts, "   ")
}

// --- Helpers ---

func padToWidth(s string, w int) string {
	visible := lipgloss.Width(s)
	if visible >= w {
		return s
	}
	return s + strings.Repeat(" ", w-visible)
}

func truncStr(s string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max <= 1 {
		return "…"
	}
	return string(runes[:max-1]) + "…"
}

func maskValue(val string) string {
	if len(val) <= 7 {
		return strings.Repeat("•", len(val))
	}
	return strings.Repeat("•", len(val)-7) + val[len(val)-7:]
}

// --- Non-interactive ---

func printStatus(s *State) {
	total := len(s.Vars)
	set := s.CountSet()
	skipped := s.CountSkipped()
	fmt.Printf("%d total, %d set, %d skipped, %d remaining\n", total, set, skipped, total-set-skipped)
}

func printReview(s *State) {
	curSection := ""
	for _, v := range s.Vars {
		if v.Section != curSection {
			curSection = v.Section
			if curSection != "" {
				fmt.Println(stSection.Render(curSection))
			}
		}

		var icon, displayVal string
		if s.IsSet(v.Name) {
			icon = stGreen.Render("✓")
			displayVal = s.Values[v.Name]
			if v.Sensitive {
				displayVal = stDim.Render(maskValue(displayVal))
			}
		} else if s.Skipped[v.Name] {
			icon = stYellow.Render("⏭")
			displayVal = stDim.Render("(skipped)")
		} else {
			icon = stDim.Render("○")
			displayVal = stDim.Render("(not set)")
		}

		padding := 40 - len(v.Name)
		if padding < 2 {
			padding = 2
		}
		fmt.Printf("%s %-20s%s%s\n", icon, v.Name, strings.Repeat(" ", padding), displayVal)
	}
}
