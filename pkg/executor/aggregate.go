package executor

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/bangmodtechnology/kselect/pkg/parser"
)

func applyAggregation(results []map[string]interface{}, query *parser.Query, fields []string) ([]map[string]interface{}, []string) {
	if len(query.GroupBy) == 0 && len(query.Aggregates) > 0 {
		// No GROUP BY: aggregate over all results
		row := computeAggregates(results, query.Aggregates)
		outFields := aggregateOutputFields(query)
		return []map[string]interface{}{row}, outFields
	}

	if len(query.GroupBy) > 0 {
		return applyGroupBy(results, query)
	}

	return results, fields
}

func applyGroupBy(results []map[string]interface{}, query *parser.Query) ([]map[string]interface{}, []string) {
	// Group rows by GROUP BY fields
	groups := make(map[string][]map[string]interface{})
	var groupOrder []string

	for _, row := range results {
		key := groupKey(row, query.GroupBy)
		if _, exists := groups[key]; !exists {
			groupOrder = append(groupOrder, key)
		}
		groups[key] = append(groups[key], row)
	}

	// Compute aggregates per group
	var output []map[string]interface{}
	for _, key := range groupOrder {
		groupRows := groups[key]
		row := make(map[string]interface{})

		// Add GROUP BY field values from first row in group
		for _, field := range query.GroupBy {
			row[field] = groupRows[0][field]
		}

		// Compute aggregates
		aggRow := computeAggregates(groupRows, query.Aggregates)
		for k, v := range aggRow {
			row[k] = v
		}

		// Also add regular selected fields from first row
		for _, f := range query.Fields {
			if _, exists := row[f]; !exists {
				row[f] = groupRows[0][f]
			}
		}

		output = append(output, row)
	}

	// Apply HAVING
	if query.Having != nil {
		var filtered []map[string]interface{}
		for _, row := range output {
			if query.Having.Evaluate(row) {
				filtered = append(filtered, row)
			}
		}
		output = filtered
	}

	outFields := aggregateOutputFields(query)
	return output, outFields
}

func groupKey(row map[string]interface{}, groupBy []string) string {
	var parts []string
	for _, field := range groupBy {
		parts = append(parts, fmt.Sprintf("%v", row[field]))
	}
	return strings.Join(parts, "|")
}

func computeAggregates(rows []map[string]interface{}, aggregates []parser.AggregateFunc) map[string]interface{} {
	result := make(map[string]interface{})

	for _, agg := range aggregates {
		switch agg.Function {
		case "COUNT":
			if agg.Field == "*" {
				result[agg.Alias] = len(rows)
			} else {
				count := 0
				for _, row := range rows {
					if row[agg.Field] != nil {
						count++
					}
				}
				result[agg.Alias] = count
			}

		case "SUM":
			sum := 0.0
			for _, row := range rows {
				sum += toFloat(row[agg.Field])
			}
			result[agg.Alias] = sum

		case "AVG":
			sum := 0.0
			count := 0
			for _, row := range rows {
				if row[agg.Field] != nil {
					sum += toFloat(row[agg.Field])
					count++
				}
			}
			if count > 0 {
				result[agg.Alias] = math.Round(sum/float64(count)*100) / 100
			} else {
				result[agg.Alias] = 0.0
			}

		case "MIN":
			var minVal *float64
			for _, row := range rows {
				if row[agg.Field] != nil {
					v := toFloat(row[agg.Field])
					if minVal == nil || v < *minVal {
						minVal = &v
					}
				}
			}
			if minVal != nil {
				result[agg.Alias] = *minVal
			} else {
				result[agg.Alias] = nil
			}

		case "MAX":
			var maxVal *float64
			for _, row := range rows {
				if row[agg.Field] != nil {
					v := toFloat(row[agg.Field])
					if maxVal == nil || v > *maxVal {
						maxVal = &v
					}
				}
			}
			if maxVal != nil {
				result[agg.Alias] = *maxVal
			} else {
				result[agg.Alias] = nil
			}
		}
	}

	return result
}

func aggregateOutputFields(query *parser.Query) []string {
	var fields []string
	fields = append(fields, query.GroupBy...)
	fields = append(fields, query.Fields...)
	for _, agg := range query.Aggregates {
		fields = append(fields, agg.Alias)
	}
	// Deduplicate
	seen := make(map[string]bool)
	var unique []string
	for _, f := range fields {
		if !seen[f] && f != "" {
			seen[f] = true
			unique = append(unique, f)
		}
	}
	return unique
}

func toFloat(val interface{}) float64 {
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return f
	default:
		f, _ := strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
		return f
	}
}
