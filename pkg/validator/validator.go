package validator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bangmodtechnology/kselect/pkg/parser"
	"github.com/bangmodtechnology/kselect/pkg/registry"
)

// ValidationError represents a query validation error
type ValidationError struct {
	Message     string
	Suggestions []string
}

func (e *ValidationError) Error() string {
	if len(e.Suggestions) == 0 {
		return e.Message
	}

	msg := e.Message + "\n"
	if len(e.Suggestions) == 1 {
		msg += fmt.Sprintf("Did you mean: %s?", e.Suggestions[0])
	} else {
		msg += "Did you mean one of these?\n"
		for _, sug := range e.Suggestions {
			msg += fmt.Sprintf("  - %s\n", sug)
		}
	}
	return msg
}

// Validator validates queries
type Validator struct {
	registry *registry.Registry
}

// New creates a new Validator
func New(reg *registry.Registry) *Validator {
	return &Validator{
		registry: reg,
	}
}

// Validate validates a parsed query
func (v *Validator) Validate(query *parser.Query) error {
	// Validate resource exists
	if err := v.validateResource(query.Resource); err != nil {
		return err
	}

	// Get resource definition
	resource, ok := v.registry.Get(query.Resource)
	if !ok {
		// Should not happen since we validated above, but check anyway
		return &ValidationError{Message: fmt.Sprintf("Resource '%s' not found", query.Resource)}
	}

	// Validate fields
	if err := v.validateFields(resource, query.Fields); err != nil {
		return err
	}

	// Validate WHERE conditions
	if query.Conditions != nil {
		if err := v.validateConditionGroup(resource, query.Conditions); err != nil {
			return err
		}
	}

	// Validate JOINs
	for _, join := range query.Joins {
		if err := v.validateJoin(&join); err != nil {
			return err
		}
	}

	// Validate ORDER BY (needs aggregates info for alias checking)
	if err := v.validateOrderBy(resource, query.OrderBy, query.Aggregates, query.Fields); err != nil {
		return err
	}

	// Validate GROUP BY
	if err := v.validateGroupBy(resource, query.GroupBy); err != nil {
		return err
	}

	// Validate aggregations
	if err := v.validateAggregates(resource, query.Aggregates); err != nil {
		return err
	}

	// Validate DISTINCT first (simpler check)
	if err := v.validateDistinct(query); err != nil {
		return err
	}

	// Validate aggregation consistency
	if err := v.validateAggregationConsistency(query); err != nil {
		return err
	}

	// Validate HAVING clause
	if query.Having != nil {
		if err := v.validateHaving(resource, query.Having, query.GroupBy, query.Aggregates); err != nil {
			return err
		}
	}

	return nil
}

// validateResource checks if a resource exists
func (v *Validator) validateResource(resourceName string) error {
	if resourceName == "" {
		return &ValidationError{Message: "Resource name is required"}
	}

	_, ok := v.registry.Get(resourceName)
	if !ok {
		// Find similar resource names
		suggestions := v.findSimilarResources(resourceName)
		return &ValidationError{
			Message:     fmt.Sprintf("Resource '%s' not found", resourceName),
			Suggestions: suggestions,
		}
	}

	return nil
}

// validateFields checks if fields exist in the resource
func (v *Validator) validateFields(resource *registry.ResourceDefinition, fields []string) error {
	// Empty fields or "*" is valid (will use default fields)
	if len(fields) == 0 || (len(fields) == 1 && fields[0] == "*") {
		return nil
	}

	for _, field := range fields {
		// Skip aggregate functions (COUNT, SUM, AVG, etc.)
		if isAggregateField(field) {
			continue
		}

		// Resolve alias to canonical name
		canonicalField := resource.ResolveFieldAlias(field)

		// Check if field exists (including dot-notation map sub-fields like labels.app)
		if _, ok := resource.Fields[canonicalField]; !ok {
			if _, _, ok := resource.IsMapSubField(canonicalField); !ok {
				// Find similar field names
				suggestions := v.findSimilarFields(resource, field)
				return &ValidationError{
					Message:     fmt.Sprintf("Field '%s' not found in resource '%s'", field, resource.Name),
					Suggestions: suggestions,
				}
			}
		}
	}

	return nil
}

// validateConditionGroup validates a condition group recursively
func (v *Validator) validateConditionGroup(resource *registry.ResourceDefinition, group *parser.ConditionGroup) error {
	// Validate conditions in this group
	for _, cond := range group.Conditions {
		// Skip subqueries
		if cond.SubQuery != nil {
			if err := v.validateResource(cond.SubQuery.Resource); err != nil {
				return fmt.Errorf("subquery validation failed: %w", err)
			}
			continue
		}

		// Resolve alias to canonical name
		canonicalField := resource.ResolveFieldAlias(cond.Field)

		// Check if field exists (including dot-notation map sub-fields)
		if _, ok := resource.Fields[canonicalField]; !ok {
			if _, _, ok := resource.IsMapSubField(canonicalField); !ok {
				suggestions := v.findSimilarFields(resource, cond.Field)
				return &ValidationError{
					Message:     fmt.Sprintf("Field '%s' in WHERE clause not found in resource '%s'", cond.Field, resource.Name),
					Suggestions: suggestions,
				}
			}
		}
	}

	// Recursively validate subgroups
	for _, subGroup := range group.SubGroups {
		if err := v.validateConditionGroup(resource, subGroup); err != nil {
			return err
		}
	}

	return nil
}

// validateJoin validates JOIN conditions
func (v *Validator) validateJoin(join *parser.JoinClause) error {
	// Validate joined resource exists
	if err := v.validateResource(join.Resource); err != nil {
		return fmt.Errorf("JOIN validation failed: %w", err)
	}

	// Note: Full validation of JOIN fields requires knowing the left resource,
	// which is more complex. For now, just check if the resource exists.

	return nil
}

// validateOrderBy validates ORDER BY fields
func (v *Validator) validateOrderBy(resource *registry.ResourceDefinition, orderBy []parser.OrderByField, aggregates []parser.AggregateFunc, fields []string) error {
	for _, ob := range orderBy {
		// Skip aggregate functions (e.g., COUNT, SUM.field)
		if isAggregateField(ob.Field) {
			continue
		}

		// Check if it's an aggregate alias
		isAggregateAlias := false
		for _, agg := range aggregates {
			if agg.Alias == ob.Field {
				isAggregateAlias = true
				break
			}
		}

		if isAggregateAlias {
			// Valid aggregate alias, allow it
			continue
		}

		// Check if it's a field alias in the SELECT list
		isSelectField := false
		for _, f := range fields {
			if f == ob.Field || resource.ResolveFieldAlias(f) == resource.ResolveFieldAlias(ob.Field) {
				isSelectField = true
				break
			}
		}

		if !isSelectField {
			// Resolve alias
			canonicalField := resource.ResolveFieldAlias(ob.Field)

			// Check if field exists in resource (including dot-notation map sub-fields)
			if _, ok := resource.Fields[canonicalField]; !ok {
				if _, _, ok := resource.IsMapSubField(canonicalField); !ok {
					suggestions := v.findSimilarFields(resource, ob.Field)
					return &ValidationError{
						Message:     fmt.Sprintf("Field '%s' in ORDER BY clause not found in resource '%s'", ob.Field, resource.Name),
						Suggestions: suggestions,
					}
				}
			}
		}
	}

	return nil
}

// validateGroupBy validates GROUP BY fields
func (v *Validator) validateGroupBy(resource *registry.ResourceDefinition, groupBy []string) error {
	for _, field := range groupBy {
		// Resolve alias
		canonicalField := resource.ResolveFieldAlias(field)

		// Check if field exists (including dot-notation map sub-fields)
		if _, ok := resource.Fields[canonicalField]; !ok {
			if _, _, ok := resource.IsMapSubField(canonicalField); !ok {
				suggestions := v.findSimilarFields(resource, field)
				return &ValidationError{
					Message:     fmt.Sprintf("Field '%s' in GROUP BY clause not found in resource '%s'", field, resource.Name),
					Suggestions: suggestions,
				}
			}
		}
	}

	return nil
}

// validateAggregates validates aggregate functions
func (v *Validator) validateAggregates(resource *registry.ResourceDefinition, aggregates []parser.AggregateFunc) error {
	for _, agg := range aggregates {
		// COUNT without field is valid (COUNT(*))
		if agg.Function == "COUNT" && (agg.Field == "" || agg.Field == "*") {
			continue
		}

		// Validate field exists
		if agg.Field != "" && agg.Field != "*" {
			canonicalField := resource.ResolveFieldAlias(agg.Field)
			if _, ok := resource.Fields[canonicalField]; !ok {
				suggestions := v.findSimilarFields(resource, agg.Field)
				return &ValidationError{
					Message:     fmt.Sprintf("Field '%s' in %s() aggregation not found in resource '%s'", agg.Field, agg.Function, resource.Name),
					Suggestions: suggestions,
				}
			}
		}
	}

	return nil
}

// findSimilarResources finds resource names similar to the given name
func (v *Validator) findSimilarResources(name string) []string {
	resources := v.registry.ListResources()
	var candidates []struct {
		name     string
		distance int
	}

	for _, res := range resources {
		// Check primary name
		dist := levenshteinDistance(strings.ToLower(name), strings.ToLower(res.Name))
		if dist <= 3 { // threshold
			candidates = append(candidates, struct {
				name     string
				distance int
			}{res.Name, dist})
		}

		// Check aliases
		for _, alias := range res.Aliases {
			dist := levenshteinDistance(strings.ToLower(name), strings.ToLower(alias))
			if dist <= 3 {
				candidates = append(candidates, struct {
					name     string
					distance int
				}{alias, dist})
			}
		}
	}

	// Sort by distance
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].distance < candidates[j].distance
	})

	// Return top 3 suggestions
	var suggestions []string
	limit := 3
	if len(candidates) < limit {
		limit = len(candidates)
	}
	for i := 0; i < limit; i++ {
		suggestions = append(suggestions, candidates[i].name)
	}

	return suggestions
}

// findSimilarFields finds field names similar to the given name
func (v *Validator) findSimilarFields(resource *registry.ResourceDefinition, name string) []string {
	var candidates []struct {
		name     string
		distance int
	}

	for fieldName, fieldDef := range resource.Fields {
		// Check primary name
		dist := levenshteinDistance(strings.ToLower(name), strings.ToLower(fieldName))
		if dist <= 3 {
			candidates = append(candidates, struct {
				name     string
				distance int
			}{fieldName, dist})
		}

		// Check aliases
		for _, alias := range fieldDef.Aliases {
			dist := levenshteinDistance(strings.ToLower(name), strings.ToLower(alias))
			if dist <= 3 {
				candidates = append(candidates, struct {
					name     string
					distance int
				}{alias, dist})
			}
		}
	}

	// Sort by distance
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].distance < candidates[j].distance
	})

	// Return top 3 suggestions
	var suggestions []string
	limit := 3
	if len(candidates) < limit {
		limit = len(candidates)
	}
	for i := 0; i < limit; i++ {
		suggestions = append(suggestions, candidates[i].name)
	}

	return suggestions
}

// validateHaving validates HAVING clause
func (v *Validator) validateHaving(resource *registry.ResourceDefinition, having *parser.ConditionGroup, groupBy []string, aggregates []parser.AggregateFunc) error {
	if having == nil {
		return nil
	}

	// HAVING requires GROUP BY
	if len(groupBy) == 0 && len(aggregates) == 0 {
		return &ValidationError{
			Message: "HAVING clause requires GROUP BY or aggregate functions",
		}
	}

	// Validate fields in HAVING clause
	return v.validateHavingGroup(resource, having, groupBy, aggregates)
}

// validateHavingGroup validates a HAVING condition group recursively
func (v *Validator) validateHavingGroup(resource *registry.ResourceDefinition, group *parser.ConditionGroup, groupBy []string, aggregates []parser.AggregateFunc) error {
	for _, cond := range group.Conditions {
		// HAVING can only reference GROUP BY fields or aggregate functions
		if !isAggregateField(cond.Field) {
			// Check if field is in GROUP BY
			canonicalField := resource.ResolveFieldAlias(cond.Field)
			found := false
			for _, gb := range groupBy {
				if resource.ResolveFieldAlias(gb) == canonicalField {
					found = true
					break
				}
			}

			if !found {
				return &ValidationError{
					Message: fmt.Sprintf("Field '%s' in HAVING clause must be in GROUP BY or be an aggregate function", cond.Field),
				}
			}

			// Also validate that field exists (including dot-notation map sub-fields)
			if _, ok := resource.Fields[canonicalField]; !ok {
				if _, _, ok := resource.IsMapSubField(canonicalField); !ok {
					suggestions := v.findSimilarFields(resource, cond.Field)
					return &ValidationError{
						Message:     fmt.Sprintf("Field '%s' in HAVING clause not found in resource '%s'", cond.Field, resource.Name),
						Suggestions: suggestions,
					}
				}
			}
		}
	}

	// Recursively validate subgroups
	for _, subGroup := range group.SubGroups {
		if err := v.validateHavingGroup(resource, subGroup, groupBy, aggregates); err != nil {
			return err
		}
	}

	return nil
}

// validateAggregationConsistency validates aggregation rules
func (v *Validator) validateAggregationConsistency(query *parser.Query) error {
	hasAggregates := len(query.Aggregates) > 0
	hasGroupBy := len(query.GroupBy) > 0

	// If there are aggregates but no GROUP BY, all fields must be aggregates
	if hasAggregates && !hasGroupBy {
		// Check if any non-aggregate fields exist
		for _, field := range query.Fields {
			if field == "*" {
				return &ValidationError{
					Message: "Cannot use '*' with aggregate functions without GROUP BY",
				}
			}
			if !isAggregateField(field) {
				return &ValidationError{
					Message: fmt.Sprintf("Field '%s' must appear in GROUP BY or be an aggregate function", field),
				}
			}
		}
	}

	// If there is GROUP BY, all non-aggregate fields must be in GROUP BY
	if hasGroupBy {
		reg := registry.GetGlobalRegistry()
		resource, ok := reg.Get(query.Resource)
		if !ok {
			return &ValidationError{Message: fmt.Sprintf("Resource '%s' not found", query.Resource)}
		}

		for _, field := range query.Fields {
			if field == "*" {
				return &ValidationError{
					Message: "Cannot use '*' with GROUP BY clause",
				}
			}

			if !isAggregateField(field) {
				// Field must be in GROUP BY
				canonicalField := resource.ResolveFieldAlias(field)
				found := false
				for _, gb := range query.GroupBy {
					if resource.ResolveFieldAlias(gb) == canonicalField {
						found = true
						break
					}
				}

				if !found {
					return &ValidationError{
						Message: fmt.Sprintf("Field '%s' must appear in GROUP BY or be an aggregate function", field),
					}
				}
			}
		}
	}

	// If there is HAVING, must have GROUP BY or aggregates
	if query.Having != nil && !hasGroupBy && !hasAggregates {
		return &ValidationError{
			Message: "HAVING clause requires GROUP BY or aggregate functions",
		}
	}

	return nil
}

// validateDistinct validates DISTINCT usage
func (v *Validator) validateDistinct(query *parser.Query) error {
	if !query.Distinct {
		return nil
	}

	// DISTINCT cannot be used with aggregates (unless all fields are aggregates)
	if len(query.Aggregates) > 0 {
		return &ValidationError{
			Message: "DISTINCT cannot be used with aggregate functions",
		}
	}

	// DISTINCT cannot be used with GROUP BY
	if len(query.GroupBy) > 0 {
		return &ValidationError{
			Message: "DISTINCT cannot be used with GROUP BY clause",
		}
	}

	// DISTINCT with * is allowed
	if len(query.Fields) == 1 && query.Fields[0] == "*" {
		return nil
	}

	return nil
}

// isAggregateField checks if a field is an aggregate function
func isAggregateField(field string) bool {
	// Check for patterns like COUNT, SUM.field, AVG.field, etc.
	upper := strings.ToUpper(field)
	aggregates := []string{"COUNT", "SUM", "AVG", "MIN", "MAX"}

	for _, agg := range aggregates {
		if upper == agg || strings.HasPrefix(upper, agg+".") {
			return true
		}
	}

	return false
}
