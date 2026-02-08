package parser

import (
	"fmt"
	"regexp"
	"strings"
)

type JoinType string

const (
	InnerJoin JoinType = "INNER"
	LeftJoin  JoinType = "LEFT"
	RightJoin JoinType = "RIGHT"
)

type JoinClause struct {
	Type      JoinType
	Resource  string
	Alias     string
	LeftField  string
	RightField string
}

type AggregateFunc struct {
	Function string // COUNT, SUM, AVG, MIN, MAX
	Field    string // field name or *
	Alias    string // AS alias
}

type OrderByField struct {
	Field      string
	Descending bool
}

type Query struct {
	Fields        []string
	Aggregates    []AggregateFunc
	Resource      string
	ResourceAlias string
	Namespace     string
	Labels        map[string]string
	FieldSelector string
	Conditions    *ConditionGroup
	Joins         []JoinClause
	GroupBy       []string
	Having        *ConditionGroup
	OrderBy       []OrderByField
	Limit         int
	Offset        int
	Distinct      bool
	UseDefault    bool // true when user omits field list
}

func Parse(input string) (*Query, error) {
	query := &Query{
		Labels: make(map[string]string),
	}

	input = strings.TrimSpace(input)

	// Strip optional SELECT keyword
	selectRe := regexp.MustCompile(`(?i)^SELECT\s+`)
	input = selectRe.ReplaceAllString(input, "")

	// Find FROM keyword to split fields and the rest
	fromIdx := findKeywordIndex(input, "FROM")
	if fromIdx == -1 {
		return nil, fmt.Errorf("invalid syntax: missing FROM keyword")
	}

	fieldStr := strings.TrimSpace(input[:fromIdx])
	rest := strings.TrimSpace(input[fromIdx+4:]) // skip "FROM"

	// No fields specified → show all fields
	if fieldStr == "" {
		query.Fields = []string{"*"}
	}

	// Parse DISTINCT
	if strings.HasPrefix(strings.ToUpper(fieldStr), "DISTINCT ") {
		query.Distinct = true
		fieldStr = strings.TrimSpace(fieldStr[9:])
	}

	// Parse fields and aggregates
	if err := parseFieldsAndAggregates(query, fieldStr); err != nil {
		return nil, err
	}

	// Parse resource (with optional alias) and remaining clauses
	if err := parseFromAndClauses(query, rest); err != nil {
		return nil, err
	}

	// Default namespace
	if query.Namespace == "" {
		query.Namespace = "default"
	}

	return query, nil
}

func parseFieldsAndAggregates(query *Query, fieldStr string) error {
	if fieldStr == "*" {
		query.Fields = []string{"*"}
		return nil
	}

	parts := splitTopLevel(fieldStr, ',')
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Check for aggregate function: COUNT(*), SUM(field), etc.
		aggRe := regexp.MustCompile(`(?i)^(COUNT|SUM|AVG|MIN|MAX)\s*\(\s*([^)]*)\s*\)(?:\s+AS\s+(\w+))?$`)
		if matches := aggRe.FindStringSubmatch(part); len(matches) > 0 {
			agg := AggregateFunc{
				Function: strings.ToUpper(matches[1]),
				Field:    strings.TrimSpace(matches[2]),
				Alias:    matches[3],
			}
			if agg.Alias == "" {
				if agg.Field == "*" {
					agg.Alias = strings.ToLower(agg.Function)
				} else {
					agg.Alias = strings.ToLower(agg.Function) + "_" + agg.Field
				}
			}
			query.Aggregates = append(query.Aggregates, agg)
		} else {
			query.Fields = append(query.Fields, part)
		}
	}

	return nil
}

func parseFromAndClauses(query *Query, rest string) error {
	// Extract resource name (and optional alias) before any keyword
	tokens := strings.Fields(rest)
	if len(tokens) == 0 {
		return fmt.Errorf("missing resource name after FROM")
	}

	query.Resource = strings.ToLower(tokens[0])
	consumed := 1

	// Check for alias (next token that isn't a keyword)
	if consumed < len(tokens) && !isKeyword(tokens[consumed]) {
		query.ResourceAlias = tokens[consumed]
		consumed++
	}

	// Rejoin remaining tokens
	remaining := strings.TrimSpace(strings.Join(tokens[consumed:], " "))

	// Parse JOIN clauses
	remaining, err := parseJoins(query, remaining)
	if err != nil {
		return err
	}

	// Parse WHERE clause
	remaining, err = parseWhere(query, remaining)
	if err != nil {
		return err
	}

	// Parse GROUP BY clause
	remaining, err = parseGroupBy(query, remaining)
	if err != nil {
		return err
	}

	// Parse HAVING clause
	remaining, err = parseHaving(query, remaining)
	if err != nil {
		return err
	}

	// Parse ORDER BY
	remaining, err = parseOrderByClause(query, remaining)
	if err != nil {
		return err
	}

	// Parse LIMIT / OFFSET
	parseLimitOffset(query, remaining)

	return nil
}

func parseJoins(query *Query, input string) (string, error) {
	joinRe := regexp.MustCompile(`(?i)(INNER|LEFT|RIGHT)\s+JOIN\s+(\w+)(?:\s+(\w+))?\s+ON\s+(\S+)\s*=\s*(\S+)`)
	remaining := input

	for {
		matches := joinRe.FindStringSubmatchIndex(remaining)
		if matches == nil {
			break
		}

		fullMatch := remaining[matches[0]:matches[1]]
		joinType := strings.ToUpper(remaining[matches[2]:matches[3]])
		resource := remaining[matches[4]:matches[5]]
		alias := ""
		if matches[6] != -1 {
			alias = remaining[matches[6]:matches[7]]
		}
		leftField := remaining[matches[8]:matches[9]]
		rightField := remaining[matches[10]:matches[11]]

		join := JoinClause{
			Type:       JoinType(joinType),
			Resource:   strings.ToLower(resource),
			Alias:      alias,
			LeftField:  leftField,
			RightField: rightField,
		}
		query.Joins = append(query.Joins, join)
		remaining = strings.Replace(remaining, fullMatch, "", 1)
		remaining = strings.TrimSpace(remaining)
	}

	return remaining, nil
}

func parseWhere(query *Query, input string) (string, error) {
	whereRe := regexp.MustCompile(`(?i)\bWHERE\s+(.+?)(?:\s+GROUP\s+BY\b|\s+ORDER\s+BY\b|\s+LIMIT\b|\s+HAVING\b|$)`)
	matches := whereRe.FindStringSubmatch(input)
	if len(matches) > 1 {
		whereClause := strings.TrimSpace(matches[1])
		conditions, err := ParseConditions(whereClause)
		if err != nil {
			return "", fmt.Errorf("error parsing WHERE clause: %w", err)
		}
		query.Conditions = conditions
		extractNamespace(query, conditions)

		// Remove WHERE clause from input
		whereIdx := findKeywordIndex(input, "WHERE")
		if whereIdx != -1 {
			// Find the end of WHERE clause (next keyword or end)
			endRe := regexp.MustCompile(`(?i)\s+(GROUP\s+BY|ORDER\s+BY|LIMIT|HAVING)\b`)
			endMatch := endRe.FindStringIndex(input[whereIdx:])
			if endMatch != nil {
				input = input[:whereIdx] + input[whereIdx+endMatch[0]:]
			} else {
				input = input[:whereIdx]
			}
		}
	}
	return strings.TrimSpace(input), nil
}

func parseGroupBy(query *Query, input string) (string, error) {
	groupRe := regexp.MustCompile(`(?i)\bGROUP\s+BY\s+(.+?)(?:\s+HAVING\b|\s+ORDER\s+BY\b|\s+LIMIT\b|$)`)
	matches := groupRe.FindStringSubmatch(input)
	if len(matches) > 1 {
		fields := strings.Split(matches[1], ",")
		for _, f := range fields {
			f = strings.TrimSpace(f)
			if f != "" {
				query.GroupBy = append(query.GroupBy, f)
			}
		}

		// Remove GROUP BY clause
		gbIdx := findKeywordIndex(input, "GROUP")
		if gbIdx != -1 {
			endRe := regexp.MustCompile(`(?i)\s+(HAVING|ORDER\s+BY|LIMIT)\b`)
			endMatch := endRe.FindStringIndex(input[gbIdx:])
			if endMatch != nil {
				input = input[:gbIdx] + input[gbIdx+endMatch[0]:]
			} else {
				input = input[:gbIdx]
			}
		}
	}
	return strings.TrimSpace(input), nil
}

func parseHaving(query *Query, input string) (string, error) {
	havingRe := regexp.MustCompile(`(?i)\bHAVING\s+(.+?)(?:\s+ORDER\s+BY\b|\s+LIMIT\b|$)`)
	matches := havingRe.FindStringSubmatch(input)
	if len(matches) > 1 {
		havingClause := strings.TrimSpace(matches[1])
		conditions, err := ParseConditions(havingClause)
		if err != nil {
			return "", fmt.Errorf("error parsing HAVING clause: %w", err)
		}
		query.Having = conditions

		// Remove HAVING clause
		hIdx := findKeywordIndex(input, "HAVING")
		if hIdx != -1 {
			endRe := regexp.MustCompile(`(?i)\s+(ORDER\s+BY|LIMIT)\b`)
			endMatch := endRe.FindStringIndex(input[hIdx:])
			if endMatch != nil {
				input = input[:hIdx] + input[hIdx+endMatch[0]:]
			} else {
				input = input[:hIdx]
			}
		}
	}
	return strings.TrimSpace(input), nil
}

func parseOrderByClause(query *Query, input string) (string, error) {
	orderRe := regexp.MustCompile(`(?i)\bORDER\s+BY\s+(.+?)(?:\s+LIMIT\b|$)`)
	matches := orderRe.FindStringSubmatch(input)
	if len(matches) > 1 {
		query.OrderBy = parseOrderBy(strings.TrimSpace(matches[1]))

		// Remove ORDER BY clause
		obIdx := findKeywordIndex(input, "ORDER")
		if obIdx != -1 {
			endRe := regexp.MustCompile(`(?i)\s+LIMIT\b`)
			endMatch := endRe.FindStringIndex(input[obIdx:])
			if endMatch != nil {
				input = input[:obIdx] + input[obIdx+endMatch[0]:]
			} else {
				input = input[:obIdx]
			}
		}
	}
	return strings.TrimSpace(input), nil
}

func parseLimitOffset(query *Query, input string) {
	limitRe := regexp.MustCompile(`(?i)\bLIMIT\s+(\d+)(?:\s+OFFSET\s+(\d+))?`)
	matches := limitRe.FindStringSubmatch(input)
	if len(matches) > 1 {
		fmt.Sscanf(matches[1], "%d", &query.Limit)
		if len(matches) > 2 && matches[2] != "" {
			fmt.Sscanf(matches[2], "%d", &query.Offset)
		}
	}
}

func parseOrderBy(orderBy string) []OrderByField {
	var fields []OrderByField
	parts := strings.Split(orderBy, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		tokens := strings.Fields(part)

		if len(tokens) == 0 {
			continue
		}

		field := OrderByField{
			Field:      tokens[0],
			Descending: false,
		}

		if len(tokens) > 1 && strings.ToUpper(tokens[1]) == "DESC" {
			field.Descending = true
		}

		fields = append(fields, field)
	}

	return fields
}

func extractNamespace(query *Query, conditions *ConditionGroup) {
	for _, cond := range conditions.Conditions {
		if (cond.Field == "namespace" || cond.Field == "ns") && cond.Operator == OpEqual {
			query.Namespace = cond.Value
			return
		}
	}

	for _, subGroup := range conditions.SubGroups {
		extractNamespace(query, subGroup)
	}
}

func findKeywordIndex(input string, keyword string) int {
	upper := strings.ToUpper(input)
	kw := strings.ToUpper(keyword)

	idx := 0
	for {
		pos := strings.Index(upper[idx:], kw)
		if pos == -1 {
			return -1
		}
		absPos := idx + pos
		// Ensure it's a word boundary
		before := absPos == 0 || input[absPos-1] == ' ' || input[absPos-1] == '\t'
		after := absPos+len(kw) >= len(input) || input[absPos+len(kw)] == ' ' || input[absPos+len(kw)] == '\t'
		if before && after {
			return absPos
		}
		idx = absPos + len(kw)
	}
}

func isKeyword(token string) bool {
	keywords := []string{
		"WHERE", "ORDER", "BY", "LIMIT", "OFFSET",
		"GROUP", "HAVING", "INNER", "LEFT", "RIGHT",
		"JOIN", "ON", "AND", "OR", "AS",
	}
	upper := strings.ToUpper(token)
	for _, kw := range keywords {
		if upper == kw {
			return true
		}
	}
	return false
}

// splitTopLevel splits by delimiter but respects parentheses nesting
func splitTopLevel(s string, delim byte) []string {
	var parts []string
	var current strings.Builder
	depth := 0

	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '(' {
			depth++
		} else if ch == ')' {
			depth--
		}
		if ch == delim && depth == 0 {
			parts = append(parts, current.String())
			current.Reset()
		} else {
			current.WriteByte(ch)
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}
