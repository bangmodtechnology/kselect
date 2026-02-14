package repl

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bangmodtechnology/kselect/pkg/describe"
	"github.com/bangmodtechnology/kselect/pkg/executor"
	"github.com/bangmodtechnology/kselect/pkg/output"
	"github.com/bangmodtechnology/kselect/pkg/parser"
	"github.com/bangmodtechnology/kselect/pkg/registry"
	"github.com/c-bata/go-prompt"
)

// REPL represents the interactive REPL session
type REPL struct {
	executor      *executor.Executor
	history       *History
	savedQueries  map[string]string
	outputFormat  output.Format
	namespace     string
	allNamespaces bool
	useColor      bool
}

// Config holds REPL configuration
type Config struct {
	OutputFormat  string
	Namespace     string
	AllNamespaces bool
	UseColor      bool
}

// New creates a new REPL instance
func New(exec *executor.Executor, config Config) (*REPL, error) {
	history, err := NewHistory()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize history: %w", err)
	}

	format := output.FormatTable
	if config.OutputFormat != "" {
		format = output.Format(config.OutputFormat)
	}

	return &REPL{
		executor:      exec,
		history:       history,
		savedQueries:  make(map[string]string),
		outputFormat:  format,
		namespace:     config.Namespace,
		allNamespaces: config.AllNamespaces,
		useColor:      config.UseColor,
	}, nil
}

// Start starts the interactive REPL
func (r *REPL) Start() {
	fmt.Println("kselect interactive mode")
	fmt.Println("Type your queries or special commands (\\help for help)")
	fmt.Println()

	p := prompt.New(
		r.executeLine,
		r.completer,
		prompt.OptionPrefix("kselect> "),
		prompt.OptionTitle("kselect"),
		prompt.OptionHistory(r.history.GetAll()),
		prompt.OptionPrefixTextColor(prompt.Green),
		prompt.OptionPreviewSuggestionTextColor(prompt.Blue),
		prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
		prompt.OptionSuggestionBGColor(prompt.DarkGray),
	)

	p.Run()
}

// executeLine executes a line of input
func (r *REPL) executeLine(line string) {
	line = strings.TrimSpace(line)

	// Empty line
	if line == "" {
		return
	}

	// Special commands
	if strings.HasPrefix(line, "\\") {
		r.handleCommand(line)
		return
	}

	// Add to history
	r.history.Add(line)

	// Parse and execute query
	query, err := parser.Parse(line)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing query: %v\n", err)
		return
	}

	// Apply namespace settings
	if r.allNamespaces {
		query.Namespace = "*"
	} else if r.namespace != "" {
		query.Namespace = r.namespace
	} else if query.Namespace == "" {
		query.Namespace = r.executor.CurrentNamespace
	}

	// Execute query
	spin := output.NewSpinner(fmt.Sprintf("Fetching %s...", query.Resource))
	spin.Start()
	start := time.Now()
	results, fields, err := r.executor.Execute(query)
	elapsed := time.Since(start)
	spin.Stop()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing query: %v\n", err)
		return
	}

	// Format output
	formatter := output.NewFormatter(r.outputFormat)
	formatter.SetElapsed(elapsed)
	if err := formatter.Print(results, fields); err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
		return
	}

	fmt.Println()
}

// completer provides auto-completion suggestions
func (r *REPL) completer(d prompt.Document) []prompt.Suggest {
	suggestions := []prompt.Suggest{}

	// Get text before cursor
	text := d.TextBeforeCursor()

	// Special commands
	if strings.HasPrefix(text, "\\") {
		commandSugs := []prompt.Suggest{
			{Text: "\\help", Description: "Show help message"},
			{Text: "\\history", Description: "Show query history"},
			{Text: "\\clear", Description: "Clear screen"},
			{Text: "\\save", Description: "Save current query"},
			{Text: "\\load", Description: "Load saved query"},
			{Text: "\\list", Description: "List saved queries"},
			{Text: "\\describe", Description: "Describe a resource"},
			{Text: "\\resources", Description: "List available resources"},
			{Text: "\\set", Description: "Set REPL options"},
			{Text: "\\show", Description: "Show current settings"},
			{Text: "\\exit", Description: "Exit REPL"},
			{Text: "\\quit", Description: "Exit REPL"},
		}
		return prompt.FilterHasPrefix(commandSugs, text, true)
	}

	// SQL keywords
	keywords := []prompt.Suggest{
		{Text: "FROM", Description: "Specify resource type"},
		{Text: "WHERE", Description: "Filter conditions"},
		{Text: "AND", Description: "Logical AND"},
		{Text: "OR", Description: "Logical OR"},
		{Text: "ORDER BY", Description: "Sort results"},
		{Text: "LIMIT", Description: "Limit number of results"},
		{Text: "OFFSET", Description: "Skip results"},
		{Text: "GROUP BY", Description: "Group results"},
		{Text: "HAVING", Description: "Filter grouped results"},
		{Text: "DISTINCT", Description: "Remove duplicates"},
		{Text: "INNER JOIN", Description: "Inner join"},
		{Text: "LEFT JOIN", Description: "Left join"},
		{Text: "RIGHT JOIN", Description: "Right join"},
		{Text: "ON", Description: "Join condition"},
		{Text: "DESCRIBE", Description: "Show resource schema"},
		{Text: "ASC", Description: "Ascending order"},
		{Text: "DESC", Description: "Descending order"},
		{Text: "LIKE", Description: "Pattern matching"},
		{Text: "NOT LIKE", Description: "Negative pattern matching"},
		{Text: "IN", Description: "Value in list"},
		{Text: "NOT IN", Description: "Value not in list"},
		{Text: "COUNT", Description: "Count aggregation"},
		{Text: "SUM", Description: "Sum aggregation"},
		{Text: "AVG", Description: "Average aggregation"},
		{Text: "MIN", Description: "Minimum aggregation"},
		{Text: "MAX", Description: "Maximum aggregation"},
	}
	suggestions = append(suggestions, keywords...)

	// Check if we're after FROM keyword
	upperText := strings.ToUpper(text)
	if strings.Contains(upperText, " FROM ") || strings.HasSuffix(upperText, "FROM") {
		// Resource suggestions
		reg := registry.GetGlobalRegistry()
		resources := reg.ListResources()
		for _, res := range resources {
			scope := "namespaced"
			if !res.Namespaced {
				scope = "cluster-scoped"
			}
			suggestions = append(suggestions, prompt.Suggest{
				Text:        res.Name,
				Description: fmt.Sprintf("%s resource", scope),
			})
			for _, alias := range res.Aliases {
				suggestions = append(suggestions, prompt.Suggest{
					Text:        alias,
					Description: fmt.Sprintf("alias for %s", res.Name),
				})
			}
		}
	}

	// Field suggestions (simplified - suggest fields from registry)
	if !strings.HasPrefix(text, "\\") && !strings.Contains(upperText, " FROM ") {
		// Try to suggest common fields
		commonFields := []prompt.Suggest{
			{Text: "name", Description: "Resource name"},
			{Text: "namespace", Description: "Namespace"},
			{Text: "status", Description: "Status"},
			{Text: "age", Description: "Age"},
			{Text: "labels", Description: "Labels"},
		}
		suggestions = append(suggestions, commonFields...)
	}

	return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
}

// handleCommand handles special REPL commands
func (r *REPL) handleCommand(cmd string) {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return
	}

	command := parts[0]
	args := parts[1:]

	switch command {
	case "\\help", "\\h", "\\?":
		r.showHelp()
	case "\\history":
		r.showHistory(args)
	case "\\clear", "\\c":
		fmt.Print("\033[H\033[2J")
	case "\\save":
		r.saveQuery(args)
	case "\\load":
		r.loadQuery(args)
	case "\\list":
		r.listSavedQueries()
	case "\\describe", "\\desc":
		r.describeResource(args)
	case "\\resources", "\\res":
		r.listResources()
	case "\\set":
		r.setSetting(args)
	case "\\show":
		r.showSettings()
	case "\\exit", "\\quit", "\\q":
		fmt.Println("Goodbye!")
		os.Exit(0)
	default:
		fmt.Printf("Unknown command: %s (type \\help for available commands)\n", command)
	}
}

func (r *REPL) showHelp() {
	fmt.Println("Special Commands:")
	fmt.Println("  \\help, \\h, \\?       Show this help message")
	fmt.Println("  \\history [n]         Show last n queries (default: 10)")
	fmt.Println("  \\clear, \\c           Clear screen")
	fmt.Println("  \\save <name>         Save last query with a name")
	fmt.Println("  \\load <name>         Load and execute a saved query")
	fmt.Println("  \\list                List all saved queries")
	fmt.Println("  \\describe <resource> Show resource schema")
	fmt.Println("  \\resources, \\res     List available resources")
	fmt.Println("  \\set <key> <value>   Set REPL option (format, namespace, color)")
	fmt.Println("  \\show                Show current settings")
	fmt.Println("  \\exit, \\quit, \\q     Exit REPL")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  name,status FROM pod WHERE namespace=default")
	fmt.Println("  name,replicas FROM deployment ORDER BY name")
	fmt.Println("  namespace, COUNT as total FROM pod -A GROUP BY namespace")
	fmt.Println()
}

func (r *REPL) showHistory(args []string) {
	count := 10
	if len(args) > 0 {
		fmt.Sscanf(args[0], "%d", &count)
	}

	entries := r.history.GetLast(count)
	if len(entries) == 0 {
		fmt.Println("No history")
		return
	}

	fmt.Printf("Last %d queries:\n", len(entries))
	for i, entry := range entries {
		fmt.Printf("  %d: %s\n", i+1, entry)
	}
	fmt.Println()
}

func (r *REPL) saveQuery(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: \\save <name>")
		return
	}

	name := args[0]
	lastQuery := r.history.GetLast(1)
	if len(lastQuery) == 0 {
		fmt.Println("No query to save")
		return
	}

	r.savedQueries[name] = lastQuery[0]
	fmt.Printf("Query saved as '%s'\n", name)
}

func (r *REPL) loadQuery(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: \\load <name>")
		return
	}

	name := args[0]
	query, exists := r.savedQueries[name]
	if !exists {
		fmt.Printf("No saved query named '%s'\n", name)
		return
	}

	fmt.Printf("Executing: %s\n", query)
	r.executeLine(query)
}

func (r *REPL) listSavedQueries() {
	if len(r.savedQueries) == 0 {
		fmt.Println("No saved queries")
		return
	}

	fmt.Println("Saved queries:")
	for name, query := range r.savedQueries {
		fmt.Printf("  %s: %s\n", name, query)
	}
	fmt.Println()
}

func (r *REPL) describeResource(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: \\describe <resource>")
		return
	}

	resourceName := args[0]
	reg := registry.GetGlobalRegistry()

	if err := describe.Resource(reg, resourceName); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func (r *REPL) listResources() {
	reg := registry.GetGlobalRegistry()
	describe.AllResources(reg)
}

func (r *REPL) setSetting(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: \\set <key> <value>")
		fmt.Println("Keys: format, namespace, color")
		return
	}

	key := args[0]
	value := args[1]

	switch key {
	case "format", "output":
		r.outputFormat = output.Format(value)
		fmt.Printf("Output format set to: %s\n", value)
	case "namespace", "ns":
		r.namespace = value
		r.allNamespaces = false
		fmt.Printf("Namespace set to: %s\n", value)
	case "all-namespaces", "all":
		if value == "true" || value == "1" || value == "yes" {
			r.allNamespaces = true
			fmt.Println("All namespaces enabled")
		} else {
			r.allNamespaces = false
			fmt.Println("All namespaces disabled")
		}
	case "color":
		if value == "true" || value == "1" || value == "yes" {
			r.useColor = true
			output.SetColorEnabled(true)
			fmt.Println("Color output enabled")
		} else {
			r.useColor = false
			output.SetColorEnabled(false)
			fmt.Println("Color output disabled")
		}
	default:
		fmt.Printf("Unknown setting: %s\n", key)
	}
}

func (r *REPL) showSettings() {
	fmt.Println("Current Settings:")
	fmt.Printf("  Output format:   %s\n", r.outputFormat)
	if r.allNamespaces {
		fmt.Printf("  Namespace:       * (all namespaces)\n")
	} else if r.namespace != "" {
		fmt.Printf("  Namespace:       %s\n", r.namespace)
	} else {
		fmt.Printf("  Namespace:       %s (current context)\n", r.executor.CurrentNamespace)
	}
	fmt.Printf("  Color output:    %v\n", r.useColor)
	fmt.Println()
}
