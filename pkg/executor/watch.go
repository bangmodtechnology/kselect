package executor

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bangmodtechnology/kselect/pkg/output"
	"github.com/bangmodtechnology/kselect/pkg/parser"
)

func (e *Executor) ExecuteWatch(query *parser.Query, interval time.Duration, format output.Format) error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run immediately, then on interval
	if err := e.runAndPrint(query, format); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := e.runAndPrint(query, format); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
		case <-sigCh:
			fmt.Println("\nWatch stopped.")
			return nil
		}
	}
}

func (e *Executor) runAndPrint(query *parser.Query, format output.Format) error {
	results, fields, err := e.Execute(query)
	if err != nil {
		return err
	}

	// Clear screen
	fmt.Print("\033[2J\033[H")

	// Print header with timestamp
	fmt.Printf("Every %s | %s | %s\n",
		"2s",
		time.Now().Format("2006-01-02 15:04:05"),
		buildQuerySummary(query),
	)
	fmt.Println(strings.Repeat("-", 80))

	formatter := output.NewFormatter(format)
	return formatter.Print(results, fields)
}

func buildQuerySummary(query *parser.Query) string {
	var parts []string
	parts = append(parts, strings.Join(query.Fields, ","))
	parts = append(parts, "FROM", query.Resource)
	if query.Namespace != "" {
		parts = append(parts, fmt.Sprintf("(ns: %s)", query.Namespace))
	}
	return strings.Join(parts, " ")
}
