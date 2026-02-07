package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type ConditionOperator string

const (
	OpEqual        ConditionOperator = "="
	OpNotEqual     ConditionOperator = "!="
	OpGreaterThan  ConditionOperator = ">"
	OpLessThan     ConditionOperator = "<"
	OpGreaterEqual ConditionOperator = ">="
	OpLessEqual    ConditionOperator = "<="
	OpLike         ConditionOperator = "LIKE"
	OpNotLike      ConditionOperator = "NOT LIKE"
	OpIn           ConditionOperator = "IN"
	OpNotIn        ConditionOperator = "NOT IN"
)

type LogicalOperator string

const (
	LogicalAnd LogicalOperator = "AND"
	LogicalOr  LogicalOperator = "OR"
)

type Condition struct {
	Field    string
	Operator ConditionOperator
	Value    string
}

type ConditionGroup struct {
	Conditions      []Condition
	LogicalOperator LogicalOperator
	SubGroups       []*ConditionGroup
}

func ParseConditions(whereClause string) (*ConditionGroup, error) {
	if whereClause == "" {
		return &ConditionGroup{
			LogicalOperator: LogicalAnd,
		}, nil
	}

	// Split by OR (lower precedence)
	orParts := splitByLogicalOperator(whereClause, "OR")
	if len(orParts) > 1 {
		group := &ConditionGroup{
			LogicalOperator: LogicalOr,
		}
		for _, part := range orParts {
			subGroup, err := parseAndGroup(part)
			if err != nil {
				return nil, err
			}
			group.SubGroups = append(group.SubGroups, subGroup)
		}
		return group, nil
	}

	return parseAndGroup(whereClause)
}

func parseAndGroup(clause string) (*ConditionGroup, error) {
	group := &ConditionGroup{
		LogicalOperator: LogicalAnd,
	}

	andParts := splitByLogicalOperator(clause, "AND")
	for _, part := range andParts {
		cond, err := parseCondition(strings.TrimSpace(part))
		if err != nil {
			return nil, err
		}
		group.Conditions = append(group.Conditions, *cond)
	}

	return group, nil
}

func parseCondition(condStr string) (*Condition, error) {
	cond := &Condition{}

	operators := []string{
		" NOT LIKE ", " LIKE ",
		" NOT IN ", " IN ",
		">=", "<=", "!=", "=", ">", "<",
	}

	condUpper := strings.ToUpper(condStr)

	for _, op := range operators {
		opUpper := strings.ToUpper(op)
		idx := strings.Index(condUpper, opUpper)

		if idx != -1 {
			cond.Field = strings.TrimSpace(condStr[:idx])
			cond.Value = strings.TrimSpace(condStr[idx+len(op):])
			cond.Operator = ConditionOperator(strings.TrimSpace(opUpper))

			// Remove quotes
			cond.Value = strings.Trim(cond.Value, "'\"")

			return cond, nil
		}
	}

	return nil, fmt.Errorf("invalid condition: %s", condStr)
}

func splitByLogicalOperator(clause string, operator string) []string {
	var parts []string
	var current strings.Builder
	var inQuotes bool
	var quoteChar rune

	words := strings.Fields(clause)
	i := 0

	for i < len(words) {
		word := words[i]

		if len(word) > 0 {
			if word[0] == '\'' || word[0] == '"' {
				if !inQuotes {
					inQuotes = true
					quoteChar = rune(word[0])
				}
			}
			if len(word) > 0 && rune(word[len(word)-1]) == quoteChar {
				inQuotes = false
			}
		}

		if !inQuotes && strings.ToUpper(word) == operator {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		} else {
			if current.Len() > 0 {
				current.WriteString(" ")
			}
			current.WriteString(word)
		}

		i++
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	if len(parts) == 0 {
		return []string{clause}
	}

	return parts
}

func (c *Condition) Evaluate(value interface{}) bool {
	valStr := fmt.Sprintf("%v", value)

	switch c.Operator {
	case OpEqual:
		return valStr == c.Value
	case OpNotEqual:
		return valStr != c.Value
	case OpLike:
		pattern := strings.ReplaceAll(c.Value, "%", ".*")
		pattern = strings.ReplaceAll(pattern, "_", ".")
		matched, _ := regexp.MatchString("(?i)^"+pattern+"$", valStr)
		return matched
	case OpNotLike:
		pattern := strings.ReplaceAll(c.Value, "%", ".*")
		pattern = strings.ReplaceAll(pattern, "_", ".")
		matched, _ := regexp.MatchString("(?i)^"+pattern+"$", valStr)
		return !matched
	case OpIn:
		inValues := strings.Split(strings.Trim(c.Value, "()"), ",")
		for _, v := range inValues {
			if strings.TrimSpace(strings.Trim(v, "'\"")) == valStr {
				return true
			}
		}
		return false
	case OpNotIn:
		inValues := strings.Split(strings.Trim(c.Value, "()"), ",")
		for _, v := range inValues {
			if strings.TrimSpace(strings.Trim(v, "'\"")) == valStr {
				return false
			}
		}
		return true
	case OpGreaterThan:
		return compareValues(valStr, c.Value) > 0
	case OpLessThan:
		return compareValues(valStr, c.Value) < 0
	case OpGreaterEqual:
		return compareValues(valStr, c.Value) >= 0
	case OpLessEqual:
		return compareValues(valStr, c.Value) <= 0
	}

	return false
}

func compareValues(a, b string) int {
	// Try numeric comparison first
	aNum, aErr := strconv.ParseFloat(a, 64)
	bNum, bErr := strconv.ParseFloat(b, 64)
	if aErr == nil && bErr == nil {
		if aNum < bNum {
			return -1
		}
		if aNum > bNum {
			return 1
		}
		return 0
	}
	// Fall back to string comparison
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func (g *ConditionGroup) Evaluate(obj map[string]interface{}) bool {
	if g.LogicalOperator == LogicalAnd {
		for _, cond := range g.Conditions {
			if !cond.Evaluate(obj[cond.Field]) {
				return false
			}
		}
		for _, subGroup := range g.SubGroups {
			if !subGroup.Evaluate(obj) {
				return false
			}
		}
		return true
	}

	// OR
	for _, cond := range g.Conditions {
		if cond.Evaluate(obj[cond.Field]) {
			return true
		}
	}
	for _, subGroup := range g.SubGroups {
		if subGroup.Evaluate(obj) {
			return true
		}
	}
	return len(g.Conditions) == 0 && len(g.SubGroups) == 0
}
