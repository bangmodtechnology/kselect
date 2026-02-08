package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"gopkg.in/yaml.v3"
)

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
	FormatCSV   Format = "csv"
	FormatWide  Format = "wide"
)

type Formatter struct {
	format Format
	writer io.Writer
}

func NewFormatter(format Format) *Formatter {
	return &Formatter{
		format: format,
		writer: os.Stdout,
	}
}

func (f *Formatter) Print(results []map[string]interface{}, fields []string) error {
	switch f.format {
	case FormatJSON:
		return f.printJSON(results)
	case FormatYAML:
		return f.printYAML(results)
	case FormatCSV:
		return f.printCSV(results, fields)
	case FormatWide:
		return f.printWide(results, fields)
	case FormatTable:
		fallthrough
	default:
		return f.printTable(results, fields)
	}
}

func (f *Formatter) printJSON(results []map[string]interface{}) error {
	encoder := json.NewEncoder(f.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(results)
}

func (f *Formatter) printYAML(results []map[string]interface{}) error {
	encoder := yaml.NewEncoder(f.writer)
	defer encoder.Close()
	return encoder.Encode(results)
}

func (f *Formatter) printCSV(results []map[string]interface{}, fields []string) error {
	writer := csv.NewWriter(f.writer)
	defer writer.Flush()

	if err := writer.Write(fields); err != nil {
		return err
	}

	for _, row := range results {
		values := make([]string, len(fields))
		for i, field := range fields {
			values[i] = FormatValue(row[field])
		}
		if err := writer.Write(values); err != nil {
			return err
		}
	}

	return nil
}

func (f *Formatter) printTable(results []map[string]interface{}, fields []string) error {
	if len(results) == 0 {
		fmt.Fprintln(f.writer, "No resources found.")
		return nil
	}

	w := tabwriter.NewWriter(f.writer, 0, 0, 3, ' ', 0)

	// Header
	headers := make([]string, len(fields))
	for i, f := range fields {
		headers[i] = strings.ToUpper(f)
	}
	fmt.Fprintln(w, strings.Join(headers, "\t"))

	// Rows
	for _, row := range results {
		values := make([]string, len(fields))
		for i, field := range fields {
			val := formatFieldValue(row[field], field)
			if colorEnabled {
				val = colorize(val, field)
				values[i] = truncateColored(val, 50)
			} else {
				values[i] = truncate(val, 50)
			}
		}
		fmt.Fprintln(w, strings.Join(values, "\t"))
	}

	w.Flush()
	fmt.Fprintf(f.writer, "\n%d resource(s) found.\n", len(results))
	return nil
}

func (f *Formatter) printWide(results []map[string]interface{}, fields []string) error {
	if len(results) == 0 {
		fmt.Fprintln(f.writer, "No resources found.")
		return nil
	}

	w := tabwriter.NewWriter(f.writer, 0, 0, 2, ' ', 0)

	headers := make([]string, len(fields))
	for i, f := range fields {
		headers[i] = strings.ToUpper(f)
	}
	fmt.Fprintln(w, strings.Join(headers, "\t"))

	for _, row := range results {
		values := make([]string, len(fields))
		for i, field := range fields {
			val := formatFieldValue(row[field], field)
			if colorEnabled {
				val = colorize(val, field)
			}
			values[i] = val
		}
		fmt.Fprintln(w, strings.Join(values, "\t"))
	}

	w.Flush()
	fmt.Fprintf(f.writer, "\n%d resource(s) found.\n", len(results))
	return nil
}

func formatFieldValue(val interface{}, fieldName string) string {
	if val == nil {
		return "<none>"
	}

	// Format age/time fields
	if fieldName == "age" {
		return formatAge(val)
	}

	return FormatValue(val)
}

func FormatValue(val interface{}) string {
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
		for k, v := range v {
			pairs = append(pairs, fmt.Sprintf("%s=%v", k, v))
		}
		return strings.Join(pairs, ",")
	default:
		return fmt.Sprintf("%v", v)
	}
}

func formatAge(val interface{}) string {
	if val == nil {
		return "<none>"
	}

	ts, ok := val.(string)
	if !ok {
		return fmt.Sprintf("%v", val)
	}

	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return ts
	}

	duration := time.Since(t)
	switch {
	case duration.Hours() >= 24*365:
		return fmt.Sprintf("%dy", int(duration.Hours()/(24*365)))
	case duration.Hours() >= 24:
		return fmt.Sprintf("%dd", int(duration.Hours()/24))
	case duration.Hours() >= 1:
		return fmt.Sprintf("%dh", int(duration.Hours()))
	case duration.Minutes() >= 1:
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	default:
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
