package executor

import (
	"fmt"
	"strings"

	"github.com/bangmodtechnology/kselect/pkg/parser"
)

func (e *Executor) executeJoin(query *parser.Query) ([]map[string]interface{}, []string, error) {
	// Get primary resource
	primaryDef, ok := e.registry.Get(query.Resource)
	if !ok {
		return nil, nil, fmt.Errorf("unknown resource: %s", query.Resource)
	}

	primaryItems, err := e.fetchResources(primaryDef, query)
	if err != nil {
		return nil, nil, err
	}

	// Extract primary rows
	var primaryRows []map[string]interface{}
	for _, item := range primaryItems {
		row := make(map[string]interface{})
		for fieldName, fieldDef := range primaryDef.Fields {
			value := e.extractField(&item, fieldDef.JSONPath)
			// Store with alias prefix if alias is set
			prefix := query.Resource
			if query.ResourceAlias != "" {
				prefix = query.ResourceAlias
			}
			row[prefix+"."+fieldName] = value
			row[fieldName] = value
		}
		primaryRows = append(primaryRows, row)
	}

	results := primaryRows

	// Process each JOIN
	for _, join := range query.Joins {
		joinDef, ok := e.registry.Get(join.Resource)
		if !ok {
			return nil, nil, fmt.Errorf("unknown resource in JOIN: %s", join.Resource)
		}

		joinQuery := &parser.Query{
			Namespace: query.Namespace,
			Labels:    make(map[string]string),
		}
		joinItems, err := e.fetchResources(joinDef, joinQuery)
		if err != nil {
			return nil, nil, err
		}

		// Extract join rows
		var joinRows []map[string]interface{}
		for _, item := range joinItems {
			row := make(map[string]interface{})
			for fieldName, fieldDef := range joinDef.Fields {
				value := e.extractField(&item, fieldDef.JSONPath)
				prefix := join.Resource
				if join.Alias != "" {
					prefix = join.Alias
				}
				row[prefix+"."+fieldName] = value
				row[fieldName] = value
			}
			joinRows = append(joinRows, row)
		}

		results = performJoin(results, joinRows, join)
	}

	// Resolve output fields
	fields := resolveJoinFields(query)

	// Apply WHERE conditions
	if query.Conditions != nil {
		var filtered []map[string]interface{}
		for _, row := range results {
			if query.Conditions.Evaluate(row) {
				filtered = append(filtered, row)
			}
		}
		results = filtered
	}

	// Apply ORDER BY
	if len(query.OrderBy) > 0 {
		sortResults(results, query.OrderBy)
	}

	// Apply LIMIT/OFFSET
	results = applyLimitOffset(results, query.Limit, query.Offset)

	return results, fields, nil
}

func performJoin(left, right []map[string]interface{}, join parser.JoinClause) []map[string]interface{} {
	var results []map[string]interface{}

	switch join.Type {
	case parser.InnerJoin:
		for _, lRow := range left {
			for _, rRow := range right {
				if matchJoinCondition(lRow, rRow, join) {
					results = append(results, mergeRows(lRow, rRow))
				}
			}
		}

	case parser.LeftJoin:
		for _, lRow := range left {
			matched := false
			for _, rRow := range right {
				if matchJoinCondition(lRow, rRow, join) {
					results = append(results, mergeRows(lRow, rRow))
					matched = true
				}
			}
			if !matched {
				results = append(results, lRow)
			}
		}

	case parser.RightJoin:
		for _, rRow := range right {
			matched := false
			for _, lRow := range left {
				if matchJoinCondition(lRow, rRow, join) {
					results = append(results, mergeRows(lRow, rRow))
					matched = true
				}
			}
			if !matched {
				results = append(results, rRow)
			}
		}
	}

	return results
}

func matchJoinCondition(left, right map[string]interface{}, join parser.JoinClause) bool {
	leftVal := resolveFieldValue(left, join.LeftField)
	rightVal := resolveFieldValue(right, join.RightField)

	if leftVal == nil || rightVal == nil {
		return false
	}

	return fmt.Sprintf("%v", leftVal) == fmt.Sprintf("%v", rightVal)
}

func resolveFieldValue(row map[string]interface{}, field string) interface{} {
	// Try exact match first (including dotted prefix like "pod.name")
	if val, ok := row[field]; ok {
		return val
	}

	// Try nested map access for label/selector fields: "pod.label.app"
	parts := strings.SplitN(field, ".", 3)
	if len(parts) >= 2 {
		// Try "resource.field" pattern
		key := parts[0] + "." + parts[1]
		if val, ok := row[key]; ok {
			if len(parts) == 3 {
				// Access sub-field in a map value
				if m, ok := val.(map[string]interface{}); ok {
					return m[parts[2]]
				}
			}
			return val
		}
	}

	return nil
}

func mergeRows(left, right map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range left {
		merged[k] = v
	}
	for k, v := range right {
		merged[k] = v
	}
	return merged
}

func resolveJoinFields(query *parser.Query) []string {
	if len(query.Fields) == 1 && query.Fields[0] == "*" {
		return []string{"*"}
	}
	return query.Fields
}
