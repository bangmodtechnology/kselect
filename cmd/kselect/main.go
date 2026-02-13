package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bangmodtechnology/kselect/pkg/completion"
	"github.com/bangmodtechnology/kselect/pkg/describe"
	"github.com/bangmodtechnology/kselect/pkg/executor"
	"github.com/bangmodtechnology/kselect/pkg/output"
	"github.com/bangmodtechnology/kselect/pkg/parser"
	"github.com/bangmodtechnology/kselect/pkg/registry"
	"github.com/bangmodtechnology/kselect/pkg/repl"
	"github.com/bangmodtechnology/kselect/pkg/validator"

	// Import resource definitions (auto-register via init())
	_ "github.com/bangmodtechnology/kselect/pkg/registry"
)

var (
	Version  = "dev"
	Codename = ""
)

func main() {
	// Extract flags from anywhere in args (Go flag stops at first non-flag arg)
	rawArgs := os.Args[1:]

	// Define all flags first
	outputFormat := flag.String("o", "table", "Output format: table, json, yaml, csv, wide")
	namespace := flag.String("n", "", "Namespace (like kubectl -n)")
	allNamespaces := flag.Bool("A", false, "All namespaces (like kubectl -A)")
	noColor := flag.Bool("no-color", false, "Disable color output")
	showVersion := flag.Bool("version", false, "Show version")
	listResources := flag.Bool("list", false, "List available resources and fields")
	pluginDir := flag.String("plugins", "", "Directory containing plugin YAML files")
	watch := flag.Bool("watch", false, "Watch mode: continuously refresh results")
	interval := flag.Duration("interval", 2*time.Second, "Watch refresh interval")
	interactive := flag.Bool("interactive", false, "Interactive REPL mode")
	dryRun := flag.Bool("dry-run", false, "Validate query without executing")
	describeResource := flag.String("describe", "", "Describe a resource schema")

	// Build flag maps from definitions (auto-syncs with defined flags)
	knownValueFlags, knownBoolFlags := buildFlagMaps()

	// Split args using generated maps
	flagArgs, queryArgs := splitArgs(rawArgs, knownValueFlags, knownBoolFlags)

	// Reset os.Args so flag.Parse() sees our extracted flags
	os.Args = append([]string{os.Args[0]}, flagArgs...)

	flag.Parse()

	// Include any remaining non-flag args from flag.Parse
	queryArgs = append(queryArgs, flag.Args()...)

	// Color: auto-detect TTY, respect --no-color flag
	format := output.Format(*outputFormat)
	useColor := !*noColor && output.DetectColor() && (format == output.FormatTable || format == output.FormatWide)
	output.SetColorEnabled(useColor)

	if *showVersion {
		if Codename != "" {
			fmt.Printf("kselect version %s (%s)\n", Version, Codename)
		} else {
			fmt.Printf("kselect version %s\n", Version)
		}
		return
	}

	// Load plugins
	if *pluginDir != "" {
		if err := registry.LoadPlugins(*pluginDir); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load plugins: %v\n", err)
		}
	}

	if *listResources {
		printResources()
		return
	}

	// Describe resource
	if *describeResource != "" {
		reg := registry.GetGlobalRegistry()
		if err := describe.Resource(reg, *describeResource); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Interactive mode
	if *interactive {
		exec, err := executor.NewExecutor()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to Kubernetes: %v\n", err)
			os.Exit(1)
		}

		config := repl.Config{
			OutputFormat:  *outputFormat,
			Namespace:     *namespace,
			AllNamespaces: *allNamespaces,
			UseColor:      useColor,
		}

		r, err := repl.New(exec, config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing REPL: %v\n", err)
			os.Exit(1)
		}

		r.Start()
		return
	}

	// Subcommand: completion
	if len(queryArgs) > 0 && queryArgs[0] == "completion" {
		if len(queryArgs) < 2 {
			fmt.Println("Usage: kselect completion [bash|zsh]")
			os.Exit(0)
		}
		switch queryArgs[1] {
		case "bash":
			completion.GenerateBash(os.Stdout)
		case "zsh":
			completion.GenerateZsh(os.Stdout)
		default:
			fmt.Fprintf(os.Stderr, "Unknown shell: %s (use bash or zsh)\n", queryArgs[1])
			os.Exit(1)
		}
		return
	}

	if len(queryArgs) == 0 {
		printHelp()
		os.Exit(0)
	}

	// Check for DESCRIBE command
	if len(queryArgs) >= 2 && strings.ToUpper(queryArgs[0]) == "DESCRIBE" {
		reg := registry.GetGlobalRegistry()
		resourceName := queryArgs[1]
		if err := describe.Resource(reg, resourceName); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	queryStr := strings.Join(queryArgs, " ")

	// Parse query
	query, err := parser.Parse(queryStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing query: %v\n", err)
		os.Exit(1)
	}

	// Dry-run mode: validate query without execution
	if *dryRun {
		reg := registry.GetGlobalRegistry()
		v := validator.New(reg)
		if err := v.Validate(query); err != nil {
			fmt.Fprintf(os.Stderr, "Validation error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Query validation passed ✓")
		fmt.Printf("Resource: %s\n", query.Resource)
		if len(query.Fields) > 0 && query.Fields[0] != "*" {
			fmt.Printf("Fields: %s\n", strings.Join(query.Fields, ", "))
		} else {
			resource, _ := reg.Get(query.Resource)
			if len(resource.DefaultFields) > 0 {
				fmt.Printf("Fields: %s (default)\n", strings.Join(resource.DefaultFields, ", "))
			}
		}
		return
	}

	// Create executor
	exec, err := executor.NewExecutor()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to Kubernetes: %v\n", err)
		os.Exit(1)
	}

	// Namespace priority: -A > -n flag > WHERE namespace > current kube context
	if *allNamespaces {
		query.Namespace = "*"
	} else if *namespace != "" {
		query.Namespace = *namespace
	} else if query.Namespace == "" {
		query.Namespace = exec.CurrentNamespace
	}

	// Watch mode
	if *watch {
		if err := exec.ExecuteWatch(query, *interval, format); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Single execution
	results, fields, err := exec.Execute(query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing query: %v\n", err)
		os.Exit(1)
	}

	// Format output
	formatter := output.NewFormatter(format)
	if err := formatter.Print(results, fields); err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
		os.Exit(1)
	}
}

// buildFlagMaps inspects all defined flags and creates lookup maps.
// This ensures flag maps stay in sync with actual flag definitions.
func buildFlagMaps() (valueFlags, boolFlags map[string]bool) {
	valueFlags = make(map[string]bool)
	boolFlags = make(map[string]bool)

	// Define the boolFlag interface (from Go's flag package)
	type boolFlag interface {
		IsBoolFlag() bool
	}

	flag.VisitAll(func(f *flag.Flag) {
		shortName := "-" + f.Name
		longName := "--" + f.Name

		// Check if it's a bool flag
		if bf, ok := f.Value.(boolFlag); ok && bf.IsBoolFlag() {
			boolFlags[shortName] = true
			boolFlags[longName] = true
		} else {
			valueFlags[shortName] = true
			valueFlags[longName] = true
		}
	})

	return
}

// splitArgs separates flag arguments from query arguments.
// This allows flags like --watch and -o json to appear anywhere in the command.
func splitArgs(args []string, knownValueFlags, knownBoolFlags map[string]bool) (flagArgs, queryArgs []string) {
	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Check for value flags with = syntax (e.g. --interval=5s)
		if eqIdx := strings.Index(arg, "="); eqIdx > 0 {
			prefix := arg[:eqIdx]
			if knownValueFlags[prefix] {
				flagArgs = append(flagArgs, arg)
				continue
			}
		}

		if knownBoolFlags[arg] {
			flagArgs = append(flagArgs, arg)
		} else if knownValueFlags[arg] {
			flagArgs = append(flagArgs, arg)
			if i+1 < len(args) {
				i++
				flagArgs = append(flagArgs, args[i])
			}
		} else {
			queryArgs = append(queryArgs, arg)
		}
	}
	return
}

func printResources() {
	reg := registry.GetGlobalRegistry()
	resources := reg.ListResources()

	fmt.Println("Available Resources:")
	fmt.Println()

	for _, res := range resources {
		fmt.Printf("  %s", res.Name)
		if len(res.Aliases) > 0 {
			fmt.Printf(" (%s)", strings.Join(res.Aliases, ", "))
		}
		fmt.Println()

		if len(res.DefaultFields) > 0 {
			fmt.Printf("  Default: %s\n", strings.Join(res.DefaultFields, ", "))
		}

		fmt.Println("  Fields:")
		for fieldName, fieldDef := range res.Fields {
			fmt.Printf("    - %-15s %s [%s]\n", fieldName, fieldDef.Description, fieldDef.Type)
		}
		fmt.Println()
	}
}

func printHelp() {
	fmt.Println("kselect - SQL-like query for Kubernetes resources")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  kselect [flags] [fields] FROM resource [WHERE conditions] [ORDER BY field] [LIMIT n]")
	fmt.Println()
	fmt.Println("  Omit fields to use default columns for each resource.")
	fmt.Println("  Example: kselect FROM pod WHERE namespace=default")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -n namespace    Namespace (like kubectl -n, default: current context)")
	fmt.Println("  -A              All namespaces (like kubectl -A)")
	fmt.Println("  -o string       Output format: table, json, yaml, csv, wide (default: table)")
	fmt.Println("  -interactive    Interactive REPL mode")
	fmt.Println("  -dry-run        Validate query without executing")
	fmt.Println("  -describe res   Describe resource schema (e.g., -describe pod)")
	fmt.Println("  -list           List available resources and fields")
	fmt.Println("  -plugins dir    Directory containing plugin YAML files")
	fmt.Println("  -watch          Watch mode: continuously refresh results")
	fmt.Println("  -interval dur   Watch refresh interval (default: 2s)")
	fmt.Println("  -no-color       Disable color output (auto-detects TTY)")
	fmt.Println("  -version        Show version")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Basic query (uses current context namespace)")
	fmt.Println("  kselect name,status,ip FROM pod")
	fmt.Println()
	fmt.Println("  # Specify namespace with -n flag")
	fmt.Println("  kselect name,status FROM pod -n production")
	fmt.Println()
	fmt.Println("  # All namespaces")
	fmt.Println("  kselect name,namespace,status FROM pod -A")
	fmt.Println()
	fmt.Println("  # With conditions (namespace in WHERE still works)")
	fmt.Println("  kselect name,status FROM pod WHERE status=Running -n default")
	fmt.Println()
	fmt.Println("  # Pattern matching")
	fmt.Println("  kselect name,image FROM pod WHERE name LIKE 'nginx-%'")
	fmt.Println()
	fmt.Println("  # Comparison (shell-safe: GT, GE, LT, LE, NE, EQ avoid shell redirection)")
	fmt.Println("  kselect name,restarts FROM pod WHERE restarts GT 10 ORDER BY restarts DESC")
	fmt.Println("  kselect name,restarts FROM pod WHERE restarts LE 5")
	fmt.Println()
	fmt.Println("  # Sorting and limiting")
	fmt.Println("  kselect name,restarts FROM pod ORDER BY restarts DESC LIMIT 5")
	fmt.Println()
	fmt.Println("  # Output as JSON")
	fmt.Println("  kselect name,status FROM pod -n default -o json")
	fmt.Println()
	fmt.Println("  # Export to CSV")
	fmt.Println("  kselect name,cpu.req,mem.req FROM pod -o csv > pods.csv")
	fmt.Println()
	fmt.Println("  # Aggregation (shell-safe: no parens needed)")
	fmt.Println("  kselect namespace, COUNT as pod_count FROM pod -A GROUP BY namespace")
	fmt.Println("  kselect namespace, SUM.restarts as total FROM pod GROUP BY namespace")
	fmt.Println()
	fmt.Println("  # Watch mode")
	fmt.Println("  kselect name,status FROM pod -n default --watch")
	fmt.Println()
	fmt.Println("  # Subquery (shell-safe: no parens needed)")
	fmt.Println("  kselect name,status FROM pod WHERE name IN kselect name FROM deployment")
	fmt.Println("  kselect name FROM pod WHERE name NOT IN kselect name FROM service")
	fmt.Println()
	fmt.Println("  # Join")
	fmt.Println("  kselect pod.name,svc.name FROM pod INNER JOIN service svc ON pod.label.app = svc.selector.app")
	fmt.Println()
	fmt.Println("  # Shell completion")
	fmt.Println("  source <(kselect completion bash)   # bash")
	fmt.Println("  source <(kselect completion zsh)    # zsh")
	fmt.Println()
	fmt.Println("  # Describe resource")
	fmt.Println("  kselect --describe pod")
	fmt.Println("  kselect DESCRIBE deployment")
}
