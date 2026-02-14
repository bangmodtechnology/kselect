package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/bangmodtechnology/kselect/pkg/executor"
	"github.com/bangmodtechnology/kselect/pkg/parser"
)

// Status color styles (mirrors pkg/output/color.go logic)
var (
	styleGreen  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	styleYellow = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	styleRed    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("6")).
			Padding(0, 1)

	filterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("3")).
			Padding(0, 1)

	footerStyle = lipgloss.NewStyle().
			Faint(true).
			Padding(0, 1)
)

var greenStatuses = map[string]bool{
	"running": true, "succeeded": true, "active": true,
	"bound": true, "ready": true, "complete": true, "available": true,
}

var yellowStatuses = map[string]bool{
	"pending": true, "terminating": true, "unknown": true,
	"creating": true, "init": true, "waiting": true, "containercreating": true,
}

var redStatuses = map[string]bool{
	"failed": true, "error": true, "crashloopbackoff": true,
	"imagepullbackoff": true, "errimagepull": true, "oomkilled": true,
	"evicted": true, "backoff": true,
}

type model struct {
	table           table.Model
	query           *parser.Query
	exec            *executor.Executor
	allResults      []map[string]interface{}
	filteredResults []map[string]interface{}
	fields          []string
	filterInput     string
	filtering       bool
	sortCol         int
	sortDesc        bool
	elapsed         time.Duration
	err             error
	width           int
	height          int
}

type refreshMsg struct {
	results []map[string]interface{}
	fields  []string
	elapsed time.Duration
	err     error
}

// Run starts the TUI with pre-fetched results.
func Run(exec *executor.Executor, query *parser.Query, results []map[string]interface{}, fields []string, elapsed time.Duration) error {
	m := newModel(exec, query, results, fields, elapsed)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func newModel(exec *executor.Executor, query *parser.Query, results []map[string]interface{}, fields []string, elapsed time.Duration) model {
	m := model{
		query:      query,
		exec:       exec,
		allResults: results,
		fields:     fields,
		elapsed:    elapsed,
		sortCol:    -1,
		width:      80,
		height:     24,
	}
	m.filteredResults = results
	m.table = m.buildTable()
	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.filtering {
			return m.handleFilterKey(msg)
		}
		return m.handleNormalKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table = m.buildTable()
		return m, nil

	case refreshMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.allResults = msg.results
		m.fields = msg.fields
		m.elapsed = msg.elapsed
		m.err = nil
		m.applyFilter()
		m.applySort()
		m.table = m.buildTable()
		return m, nil
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) handleNormalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "/":
		m.filtering = true
		m.filterInput = ""
		return m, nil
	case "s":
		m.cycleSort(false)
		m.table = m.buildTable()
		return m, nil
	case "S":
		m.cycleSort(true)
		m.table = m.buildTable()
		return m, nil
	case "r":
		return m, m.refreshCmd()
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *model) handleFilterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.filtering = false
		m.filterInput = ""
		m.filteredResults = m.allResults
		m.applySort()
		m.table = m.buildTable()
		return m, nil
	case "enter":
		m.filtering = false
		return m, nil
	case "backspace":
		if len(m.filterInput) > 0 {
			m.filterInput = m.filterInput[:len(m.filterInput)-1]
		}
		m.applyFilter()
		m.applySort()
		m.table = m.buildTable()
		return m, nil
	default:
		if len(msg.String()) == 1 || msg.String() == " " {
			m.filterInput += msg.String()
			m.applyFilter()
			m.applySort()
			m.table = m.buildTable()
		}
		return m, nil
	}
}

func (m *model) applyFilter() {
	if m.filterInput == "" {
		m.filteredResults = m.allResults
		return
	}

	filter := strings.ToLower(m.filterInput)
	var filtered []map[string]interface{}
	for _, row := range m.allResults {
		for _, field := range m.fields {
			val := fmt.Sprintf("%v", row[field])
			if strings.Contains(strings.ToLower(val), filter) {
				filtered = append(filtered, row)
				break
			}
		}
	}
	m.filteredResults = filtered
}

func (m *model) cycleSort(desc bool) {
	if desc {
		if m.sortCol < 0 {
			m.sortCol = 0
		}
		m.sortDesc = true
	} else {
		if m.sortCol < 0 {
			m.sortCol = 0
			m.sortDesc = false
		} else if !m.sortDesc {
			m.sortDesc = true
		} else {
			m.sortCol++
			m.sortDesc = false
			if m.sortCol >= len(m.fields) {
				m.sortCol = 0
			}
		}
	}
	m.applySort()
}

func (m *model) applySort() {
	if m.sortCol < 0 || m.sortCol >= len(m.fields) {
		return
	}
	field := m.fields[m.sortCol]
	desc := m.sortDesc
	sort.SliceStable(m.filteredResults, func(i, j int) bool {
		vi := fmt.Sprintf("%v", m.filteredResults[i][field])
		vj := fmt.Sprintf("%v", m.filteredResults[j][field])
		if desc {
			return vi > vj
		}
		return vi < vj
	})
}

func (m model) refreshCmd() tea.Cmd {
	return func() tea.Msg {
		start := time.Now()
		results, fields, err := m.exec.Execute(m.query)
		elapsed := time.Since(start)
		return refreshMsg{results: results, fields: fields, elapsed: elapsed, err: err}
	}
}

func (m model) buildTable() table.Model {
	// Calculate available height for table (header bar + filter bar + footer = 3 lines)
	tableHeight := m.height - 4
	if tableHeight < 3 {
		tableHeight = 3
	}

	// Build columns with proportional widths
	cols := m.buildColumns()

	// Build rows
	rows := make([]table.Row, len(m.filteredResults))
	for i, result := range m.filteredResults {
		row := make(table.Row, len(m.fields))
		for j, field := range m.fields {
			val := formatValue(result[field])
			row[j] = val
		}
		rows[i] = row
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(tableHeight),
	)

	// Style the table
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return t
}

func (m model) buildColumns() []table.Column {
	numCols := len(m.fields)
	if numCols == 0 {
		return nil
	}

	// Reserve 2 chars for padding per column
	available := m.width - (numCols * 2)
	if available < numCols {
		available = numCols
	}
	colWidth := available / numCols
	if colWidth < 6 {
		colWidth = 6
	}

	cols := make([]table.Column, numCols)
	for i, field := range m.fields {
		title := strings.ToUpper(field)
		if i == m.sortCol {
			if m.sortDesc {
				title += " v"
			} else {
				title += " ^"
			}
		}
		cols[i] = table.Column{Title: title, Width: colWidth}
	}
	return cols
}

func (m model) View() string {
	var b strings.Builder

	// Header
	ns := m.query.Namespace
	if ns == "" || ns == "*" {
		ns = "all"
	}
	sortInfo := ""
	if m.sortCol >= 0 && m.sortCol < len(m.fields) {
		dir := "^"
		if m.sortDesc {
			dir = "v"
		}
		sortInfo = fmt.Sprintf(" | Sort: %s %s", m.fields[m.sortCol], dir)
	}
	header := fmt.Sprintf("kselect | %s (%s) | %d items%s | (%.2fs)",
		m.query.Resource, ns, len(m.filteredResults), sortInfo, m.elapsed.Seconds())
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")

	// Filter bar
	if m.filtering {
		b.WriteString(filterStyle.Render(fmt.Sprintf("Filter: %s_", m.filterInput)))
	} else if m.filterInput != "" {
		b.WriteString(filterStyle.Render(fmt.Sprintf("Filter: %s", m.filterInput)))
	} else {
		b.WriteString("")
	}
	b.WriteString("\n")

	// Error display
	if m.err != nil {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Padding(0, 1)
		b.WriteString(errStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n")
	}

	// Table with colored cells
	b.WriteString(m.renderTableWithColor())

	// Footer
	footer := " ^/v/jk Navigate  / Filter  s Sort  r Refresh  q Quit"
	b.WriteString(footerStyle.Render(footer))

	return b.String()
}

// renderTableWithColor renders the table view and applies status coloring.
func (m model) renderTableWithColor() string {
	// Use the default table view
	view := m.table.View()

	// Apply color to known status values in the rendered output
	for status := range greenStatuses {
		titled := capitalize(status)
		if strings.Contains(view, titled) {
			view = strings.ReplaceAll(view, titled, styleGreen.Render(titled))
		}
	}
	for status := range yellowStatuses {
		titled := capitalize(status)
		if strings.Contains(view, titled) {
			view = strings.ReplaceAll(view, titled, styleYellow.Render(titled))
		}
	}
	for status := range redStatuses {
		titled := capitalize(status)
		if strings.Contains(view, titled) {
			view = strings.ReplaceAll(view, titled, styleRed.Render(titled))
		}
	}

	return view
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func formatValue(val interface{}) string {
	if val == nil {
		return "<none>"
	}
	switch v := val.(type) {
	case []interface{}:
		strs := make([]string, len(v))
		for i, item := range v {
			strs[i] = fmt.Sprintf("%v", item)
		}
		return strings.Join(strs, ",")
	case map[string]interface{}:
		var pairs []string
		for k, val := range v {
			pairs = append(pairs, fmt.Sprintf("%s=%v", k, val))
		}
		return strings.Join(pairs, ",")
	default:
		return fmt.Sprintf("%v", v)
	}
}
