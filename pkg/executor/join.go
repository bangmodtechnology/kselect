package executor

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bangmodtechnology/kselect/pkg/parser"
	"github.com/bangmodtechnology/kselect/pkg/registry"
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

	// Resolve output fields (expand * using registry)
	fields := resolveJoinFields(query, e.registry)

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

// performJoin uses a hash join strategy for O(n+m) performance.
func performJoin(left, right []map[string]interface{}, join parser.JoinClause) []map[string]interface{} {
	conditions := join.Conditions
	// Backward compat: if Conditions is empty, fall back to single LeftField/RightField
	if len(conditions) == 0 && join.LeftField != "" {
		conditions = []parser.JoinCondition{{LeftField: join.LeftField, RightField: join.RightField}}
	}

	var results []map[string]interface{}

	switch join.Type {
	case parser.InnerJoin:
		// Build hash index on right rows keyed by right-side ON fields
		rightIndex := buildRightIndex(right, conditions)
		for _, lRow := range left {
			key := buildJoinKey(lRow, conditions, true)
			if key == "" {
				continue
			}
			for _, rRow := range rightIndex[key] {
				results = append(results, mergeRows(lRow, rRow))
			}
		}

	case parser.LeftJoin:
		rightIndex := buildRightIndex(right, conditions)
		for _, lRow := range left {
			key := buildJoinKey(lRow, conditions, true)
			matches := rightIndex[key]
			if key != "" && len(matches) > 0 {
				for _, rRow := range matches {
					results = append(results, mergeRows(lRow, rRow))
				}
			} else {
				results = append(results, copyRow(lRow))
			}
		}

	case parser.RightJoin:
		// Build hash index on left rows keyed by left-side ON fields
		leftIndex := buildLeftIndex(left, conditions)
		matchedRight := make(map[int]bool)
		for i, rRow := range right {
			key := buildJoinKey(rRow, conditions, false)
			matches := leftIndex[key]
			if key != "" && len(matches) > 0 {
				matchedRight[i] = true
				for _, lRow := range matches {
					results = append(results, mergeRows(lRow, rRow))
				}
			}
		}
		// Include unmatched right rows
		for i, rRow := range right {
			if !matchedRight[i] {
				results = append(results, copyRow(rRow))
			}
		}
	}

	return results
}

// buildRightIndex creates a hash map from right rows keyed by their ON field values.
func buildRightIndex(rows []map[string]interface{}, conditions []parser.JoinCondition) map[string][]map[string]interface{} {
	index := make(map[string][]map[string]interface{})
	for _, row := range rows {
		key := buildJoinKey(row, conditions, false)
		if key != "" {
			index[key] = append(index[key], row)
		}
	}
	return index
}

// buildLeftIndex creates a hash map from left rows keyed by their ON field values.
func buildLeftIndex(rows []map[string]interface{}, conditions []parser.JoinCondition) map[string][]map[string]interface{} {
	index := make(map[string][]map[string]interface{})
	for _, row := range rows {
		key := buildJoinKey(row, conditions, true)
		if key != "" {
			index[key] = append(index[key], row)
		}
	}
	return index
}

// buildJoinKey builds a composite key from the ON condition fields.
// isLeft=true uses LeftField, isLeft=false uses RightField.
// Uses \x00 as separator to prevent collisions between field values.
func buildJoinKey(row map[string]interface{}, conditions []parser.JoinCondition, isLeft bool) string {
	parts := make([]string, len(conditions))
	for i, cond := range conditions {
		field := cond.RightField
		if isLeft {
			field = cond.LeftField
		}
		val := resolveFieldValue(row, field)
		if val == nil {
			return ""
		}
		parts[i] = fmt.Sprintf("%v", val)
	}
	return strings.Join(parts, "\x00")
}

// matchJoinConditions checks all conditions (AND semantics).
func matchJoinConditions(left, right map[string]interface{}, conditions []parser.JoinCondition) bool {
	for _, cond := range conditions {
		leftVal := resolveFieldValue(left, cond.LeftField)
		rightVal := resolveFieldValue(right, cond.RightField)
		if leftVal == nil || rightVal == nil {
			return false
		}
		if fmt.Sprintf("%v", leftVal) != fmt.Sprintf("%v", rightVal) {
			return false
		}
	}
	return true
}

func copyRow(row map[string]interface{}) map[string]interface{} {
	cp := make(map[string]interface{}, len(row))
	for k, v := range row {
		cp[k] = v
	}
	return cp
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
	merged := make(map[string]interface{}, len(left)+len(right))
	for k, v := range left {
		merged[k] = v
	}
	for k, v := range right {
		merged[k] = v
	}
	return merged
}

// resolveJoinFields expands * into prefixed field names from all joined resources.
func resolveJoinFields(query *parser.Query, reg *registry.Registry) []string {
	if len(query.Fields) != 1 || query.Fields[0] != "*" {
		return query.Fields
	}

	var fields []string

	// Primary resource fields
	prefix := query.Resource
	if query.ResourceAlias != "" {
		prefix = query.ResourceAlias
	}
	if primaryDef, ok := reg.Get(query.Resource); ok {
		fields = append(fields, allFieldNames(primaryDef, prefix)...)
	}

	// Join resource fields
	for _, join := range query.Joins {
		jPrefix := join.Resource
		if join.Alias != "" {
			jPrefix = join.Alias
		}
		if joinDef, ok := reg.Get(join.Resource); ok {
			fields = append(fields, allFieldNames(joinDef, jPrefix)...)
		}
	}

	return fields
}

// allFieldNames returns "prefix.field" names for a resource definition.
// Uses DefaultFields if available, otherwise returns all fields sorted.
func allFieldNames(def *registry.ResourceDefinition, prefix string) []string {
	if len(def.DefaultFields) > 0 {
		result := make([]string, len(def.DefaultFields))
		for i, f := range def.DefaultFields {
			result[i] = prefix + "." + f
		}
		return result
	}

	names := make([]string, 0, len(def.Fields))
	for name := range def.Fields {
		names = append(names, name)
	}
	sort.Strings(names)

	result := make([]string, len(names))
	for i, name := range names {
		result[i] = prefix + "." + name
	}
	return result
}
