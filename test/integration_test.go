package test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/bangmodtechnology/kselect/pkg/output"
	"github.com/bangmodtechnology/kselect/pkg/parser"
	"github.com/bangmodtechnology/kselect/pkg/registry"
	"github.com/bangmodtechnology/kselect/pkg/validator"
)

// TestParserValidatorIntegration tests the integration between parser and validator
func TestParserValidatorIntegration(t *testing.T) {
	reg := registry.GetGlobalRegistry()
	v := validator.New(reg)

	tests := []struct {
		name        string
		queryStr    string
		shouldParse bool
		shouldValidate bool
		errContains string
	}{
		{
			name:        "valid basic query",
			queryStr:    "name,namespace,status FROM pod",
			shouldParse: true,
			shouldValidate: true,
		},
		{
			name:        "valid query with WHERE",
			queryStr:    "name,status FROM pod WHERE namespace=default",
			shouldParse: true,
			shouldValidate: true,
		},
		{
			name:        "valid aggregation with GROUP BY",
			queryStr:    "namespace, COUNT as count FROM pod GROUP BY namespace",
			shouldParse: true,
			shouldValidate: true,
		},
		{
			name:        "valid ORDER BY",
			queryStr:    "name,restarts FROM pod ORDER BY restarts DESC",
			shouldParse: true,
			shouldValidate: true,
		},
		{
			name:        "invalid resource",
			queryStr:    "name FROM invalid_resource",
			shouldParse: true,
			shouldValidate: false,
			errContains: "Resource 'invalid_resource' not found",
		},
		{
			name:        "invalid field",
			queryStr:    "invalid_field FROM pod",
			shouldParse: true,
			shouldValidate: false,
			errContains: "Field 'invalid_field' not found",
		},
		{
			name:        "invalid aggregation - missing GROUP BY",
			queryStr:    "name, COUNT as count FROM pod",
			shouldParse: true,
			shouldValidate: false,
			errContains: "must appear in GROUP BY",
		},
		{
			name:        "invalid - wildcard with GROUP BY",
			queryStr:    "* FROM pod GROUP BY namespace",
			shouldParse: true,
			shouldValidate: false,
			errContains: "Cannot use '*' with GROUP BY",
		},
		{
			name:        "invalid - DISTINCT with GROUP BY",
			queryStr:    "DISTINCT namespace FROM pod GROUP BY namespace",
			shouldParse: true,
			shouldValidate: false,
			errContains: "DISTINCT cannot be used with GROUP BY",
		},
		{
			name:        "valid subquery",
			queryStr:    "name FROM pod WHERE name IN kselect name FROM deployment",
			shouldParse: true,
			shouldValidate: true,
		},
		{
			name:        "invalid subquery resource",
			queryStr:    "name FROM pod WHERE name IN kselect name FROM invalid",
			shouldParse: true,
			shouldValidate: false,
			errContains: "subquery validation failed",
		},
		{
			name:        "valid LIMIT",
			queryStr:    "name FROM pod LIMIT 10",
			shouldParse: true,
			shouldValidate: true,
		},
		{
			name:        "valid DISTINCT",
			queryStr:    "DISTINCT name,namespace FROM pod",
			shouldParse: true,
			shouldValidate: true,
		},
		{
			name:        "valid with alias",
			queryStr:    "name,ns FROM pod",
			shouldParse: true,
			shouldValidate: true,
		},
		{
			name:        "invalid WHERE field",
			queryStr:    "name FROM pod WHERE invalid_field=value",
			shouldParse: true,
			shouldValidate: false,
			errContains: "WHERE clause not found",
		},
		{
			name:        "invalid ORDER BY field",
			queryStr:    "name FROM pod ORDER BY invalid_field",
			shouldParse: true,
			shouldValidate: false,
			errContains: "ORDER BY clause not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test parsing
			query, err := parser.Parse(tt.queryStr)

			if tt.shouldParse && err != nil {
				t.Fatalf("Parse failed unexpectedly: %v", err)
			}

			if !tt.shouldParse && err == nil {
				t.Fatal("Parse should have failed but didn't")
			}

			if !tt.shouldParse {
				return // Skip validation if parse failed
			}

			// Test validation
			err = v.Validate(query)

			if tt.shouldValidate && err != nil {
				t.Errorf("Validation failed unexpectedly: %v", err)
			}

			if !tt.shouldValidate && err == nil {
				t.Error("Validation should have failed but didn't")
			}

			if !tt.shouldValidate && err != nil && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Error should contain '%s', got: %v", tt.errContains, err)
				}
			}
		})
	}
}

// TestFormatterIntegration tests output formatter with various data
func TestFormatterIntegration(t *testing.T) {
	tests := []struct {
		name     string
		format   output.Format
		results  []map[string]interface{}
		fields   []string
		contains []string
	}{
		{
			name:   "table format",
			format: output.FormatTable,
			results: []map[string]interface{}{
				{"name": "pod-1", "status": "Running"},
				{"name": "pod-2", "status": "Pending"},
			},
			fields:   []string{"name", "status"},
			contains: []string{"NAME", "STATUS", "pod-1", "Running", "pod-2", "Pending"},
		},
		{
			name:   "json format",
			format: output.FormatJSON,
			results: []map[string]interface{}{
				{"name": "pod-1", "namespace": "default"},
			},
			fields:   []string{"name", "namespace"},
			contains: []string{`"name"`, `"pod-1"`, `"namespace"`, `"default"`},
		},
		{
			name:   "csv format",
			format: output.FormatCSV,
			results: []map[string]interface{}{
				{"name": "pod-1", "status": "Running"},
				{"name": "pod-2", "status": "Pending"},
			},
			fields:   []string{"name", "status"},
			contains: []string{"name,status", "pod-1,Running", "pod-2,Pending"},
		},
		{
			name:     "empty results",
			format:   output.FormatTable,
			results:  []map[string]interface{}{},
			fields:   []string{"name", "status"},
			contains: []string{"NAME", "STATUS"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := output.NewFormatter(tt.format)

			// Redirect stdout to buffer (formatter prints to stdout)
			// For testing, we'll use a custom approach
			// Since formatter.Print writes to stdout, we need to capture it
			// For now, just test that it doesn't error
			err := formatter.Print(tt.results, tt.fields)
			if err != nil {
				t.Errorf("Formatter.Print() error = %v", err)
			}

			// Note: To properly test output, we'd need to capture stdout
			// or modify the formatter to accept an io.Writer
			// For now, we're just testing that it doesn't crash
			_ = buf
		})
	}
}

// TestComplexQueryScenarios tests complex real-world query scenarios
func TestComplexQueryScenarios(t *testing.T) {
	reg := registry.GetGlobalRegistry()
	v := validator.New(reg)

	scenarios := []struct {
		name        string
		queryStr    string
		description string
		shouldWork  bool
	}{
		{
			name:        "aggregation with multiple GROUP BY fields",
			queryStr:    "namespace,status,COUNT as count FROM pod GROUP BY namespace,status",
			description: "Count pods by namespace and status",
			shouldWork:  true,
		},
		{
			name:        "SUM with GROUP BY",
			queryStr:    "namespace,SUM.restarts as total_restarts FROM pod GROUP BY namespace",
			description: "Sum of restarts per namespace",
			shouldWork:  true,
		},
		{
			name:        "complex WHERE with multiple conditions",
			queryStr:    "name,status FROM pod WHERE namespace=default AND status=Running",
			description: "Filter by namespace and status",
			shouldWork:  true,
		},
		{
			name:        "ORDER BY with LIMIT",
			queryStr:    "name,restarts FROM pod ORDER BY restarts DESC LIMIT 5",
			description: "Top 5 pods by restart count",
			shouldWork:  true,
		},
		{
			name:        "subquery with IN operator",
			queryStr:    "name,namespace FROM pod WHERE name IN kselect name FROM deployment",
			description: "Pods matching deployment names",
			shouldWork:  true,
		},
		{
			name:        "DISTINCT with multiple fields",
			queryStr:    "DISTINCT namespace,node FROM pod",
			description: "Unique namespace-node combinations",
			shouldWork:  true,
		},
		{
			name:        "wildcard without aggregation",
			queryStr:    "* FROM pod WHERE namespace=default",
			description: "All fields from pods in default namespace",
			shouldWork:  true,
		},
		{
			name:        "default fields (empty field list)",
			queryStr:    "FROM pod WHERE namespace=default",
			description: "Use default fields for pod resource",
			shouldWork:  true,
		},
		{
			name:        "field alias usage",
			queryStr:    "name,ns FROM pod WHERE ns=default",
			description: "Use 'ns' alias for 'namespace'",
			shouldWork:  true,
		},
		{
			name:        "comparison operators",
			queryStr:    "name,restarts FROM pod WHERE restarts GT 5",
			description: "Pods with more than 5 restarts",
			shouldWork:  true,
		},
		{
			name:        "LIKE pattern matching",
			queryStr:    "name FROM pod WHERE name LIKE 'nginx-%'",
			description: "Pods with names starting with 'nginx-'",
			shouldWork:  true,
		},
		{
			name:        "AVG aggregation",
			queryStr:    "namespace,AVG.restarts as avg_restarts FROM pod GROUP BY namespace",
			description: "Average restarts per namespace",
			shouldWork:  true,
		},
		{
			name:        "MIN/MAX aggregation",
			queryStr:    "namespace,MIN.restarts as min,MAX.restarts as max FROM pod GROUP BY namespace",
			description: "Min and max restarts per namespace",
			shouldWork:  true,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Logf("Description: %s", scenario.description)
			t.Logf("Query: %s", scenario.queryStr)

			// Parse
			query, err := parser.Parse(scenario.queryStr)
			if err != nil {
				if scenario.shouldWork {
					t.Fatalf("Parse failed: %v", err)
				}
				return
			}

			// Validate
			err = v.Validate(query)
			if scenario.shouldWork && err != nil {
				t.Errorf("Validation failed: %v", err)
			}
			if !scenario.shouldWork && err == nil {
				t.Error("Expected validation to fail but it passed")
			}
		})
	}
}

// TestQueryTransformationFlow tests the full flow of query transformation
func TestQueryTransformationFlow(t *testing.T) {
	reg := registry.GetGlobalRegistry()
	v := validator.New(reg)

	queryStr := "namespace,COUNT as pod_count FROM pod GROUP BY namespace ORDER BY pod_count DESC LIMIT 10"

	// Step 1: Parse
	query, err := parser.Parse(queryStr)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify parse results
	if query.Resource != "pod" && query.Resource != "pods" {
		t.Errorf("Expected resource 'pod', got '%s'", query.Resource)
	}
	if len(query.GroupBy) != 1 || query.GroupBy[0] != "namespace" {
		t.Errorf("Expected GROUP BY [namespace], got %v", query.GroupBy)
	}
	if query.Limit != 10 {
		t.Errorf("Expected LIMIT 10, got %d", query.Limit)
	}

	// Step 2: Validate
	err = v.Validate(query)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	// Step 3: Verify query structure
	if len(query.Fields) != 1 || query.Fields[0] != "namespace" {
		t.Errorf("Expected fields [namespace], got %v", query.Fields)
	}
	if len(query.Aggregates) != 1 {
		t.Errorf("Expected 1 aggregate, got %d", len(query.Aggregates))
	}
	if query.Aggregates[0].Function != "COUNT" {
		t.Errorf("Expected COUNT aggregate, got %s", query.Aggregates[0].Function)
	}
}

// TestFuzzyMatchingSuggestions tests the fuzzy matching suggestion system
func TestFuzzyMatchingSuggestions(t *testing.T) {
	reg := registry.GetGlobalRegistry()
	v := validator.New(reg)

	tests := []struct {
		name           string
		queryStr       string
		expectSuggestion string
	}{
		{
			name:           "typo in resource name",
			queryStr:       "name FROM pode",
			expectSuggestion: "pod",
		},
		{
			name:           "typo in field name",
			queryStr:       "nam FROM pod",
			expectSuggestion: "name",
		},
		{
			name:           "similar resource name",
			queryStr:       "name FROM deployment",
			expectSuggestion: "", // Should work, no suggestion needed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := parser.Parse(tt.queryStr)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			err = v.Validate(query)

			if tt.expectSuggestion == "" {
				// Query should be valid
				if err != nil {
					t.Errorf("Expected query to be valid, got error: %v", err)
				}
			} else {
				// Query should fail with suggestions
				if err == nil {
					t.Error("Expected validation error with suggestions")
					return
				}

				valErr, ok := err.(*validator.ValidationError)
				if !ok {
					t.Errorf("Expected ValidationError, got %T", err)
					return
				}

				if len(valErr.Suggestions) == 0 {
					t.Error("Expected suggestions but got none")
					return
				}

				found := false
				for _, sug := range valErr.Suggestions {
					if sug == tt.expectSuggestion {
						found = true
						break
					}
				}

				if !found {
					t.Errorf("Expected suggestion '%s' in %v", tt.expectSuggestion, valErr.Suggestions)
				}
			}
		})
	}
}

// TestEdgeCases tests edge cases and corner scenarios
func TestEdgeCases(t *testing.T) {
	reg := registry.GetGlobalRegistry()
	v := validator.New(reg)

	tests := []struct {
		name        string
		queryStr    string
		shouldWork  bool
		description string
	}{
		{
			name:        "empty resource name",
			queryStr:    "name FROM",
			shouldWork:  false,
			description: "FROM without resource should fail",
		},
		{
			name:        "multiple aggregates",
			queryStr:    "namespace,COUNT as count,SUM.restarts as sum,AVG.restarts as avg FROM pod GROUP BY namespace",
			shouldWork:  true,
			description: "Multiple aggregate functions in one query",
		},
		{
			name:        "complex WHERE with AND",
			queryStr:    "name FROM pod WHERE status=Running AND namespace=default",
			shouldWork:  true,
			description: "Complex WHERE with AND operator",
		},
		{
			name:        "ORDER BY aggregate alias",
			queryStr:    "namespace,COUNT as total FROM pod GROUP BY namespace ORDER BY total DESC",
			shouldWork:  true,
			description: "ORDER BY should work with aggregate aliases",
		},
		{
			name:        "LIMIT without ORDER BY",
			queryStr:    "name FROM pod LIMIT 10",
			shouldWork:  true,
			description: "LIMIT without ORDER BY is valid",
		},
		{
			name:        "OFFSET with LIMIT",
			queryStr:    "name FROM pod LIMIT 10 OFFSET 5",
			shouldWork:  true,
			description: "OFFSET with LIMIT",
		},
		{
			name:        "multiple GROUP BY fields",
			queryStr:    "namespace,status,node,COUNT as count FROM pod GROUP BY namespace,status,node",
			shouldWork:  true,
			description: "GROUP BY with multiple fields",
		},
		{
			name:        "empty field list defaults to resource defaults",
			queryStr:    "FROM pod",
			shouldWork:  true,
			description: "Empty field list should use default fields",
		},
		{
			name:        "wildcard with WHERE",
			queryStr:    "* FROM pod WHERE namespace=kube-system",
			shouldWork:  true,
			description: "Wildcard should work with WHERE clause",
		},
		{
			name:        "case insensitive keywords",
			queryStr:    "name from pod where namespace=default",
			shouldWork:  true,
			description: "Keywords should be case insensitive",
		},
		{
			name:        "resource alias",
			queryStr:    "name FROM pods",
			shouldWork:  true,
			description: "Resource aliases should work (pods -> pod)",
		},
		{
			name:        "field alias in WHERE",
			queryStr:    "name FROM pod WHERE ns=default",
			shouldWork:  true,
			description: "Field alias 'ns' for 'namespace' should work",
		},
		{
			name:        "NOT IN subquery",
			queryStr:    "name FROM pod WHERE name NOT IN kselect name FROM deployment",
			shouldWork:  true,
			description: "NOT IN with subquery",
		},
		{
			name:        "aggregate without alias but with GROUP BY",
			queryStr:    "namespace,COUNT FROM pod GROUP BY namespace",
			shouldWork:  true,
			description: "Aggregate without explicit alias (uses default)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Description: %s", tt.description)

			query, err := parser.Parse(tt.queryStr)
			if err != nil {
				if tt.shouldWork {
					t.Fatalf("Parse failed: %v", err)
				} else {
					t.Logf("Parse failed as expected: %v", err)
					return
				}
			}

			err = v.Validate(query)
			if tt.shouldWork && err != nil {
				t.Errorf("Validation should pass but failed: %v", err)
			}
			if !tt.shouldWork && err == nil {
				t.Error("Validation should fail but passed")
			}
		})
	}
}

// TestMultiResourceQueries tests queries involving multiple resources (JOINs, subqueries)
func TestMultiResourceQueries(t *testing.T) {
	reg := registry.GetGlobalRegistry()
	v := validator.New(reg)

	tests := []struct {
		name        string
		queryStr    string
		shouldWork  bool
		description string
	}{
		{
			name:        "simple subquery",
			queryStr:    "name FROM pod WHERE namespace IN kselect name FROM namespace",
			shouldWork:  true,
			description: "Subquery returning namespace names",
		},
		{
			name:        "subquery with invalid resource",
			queryStr:    "name FROM pod WHERE name IN kselect name FROM nonexistent",
			shouldWork:  false,
			description: "Subquery with non-existent resource should fail",
		},
		{
			name:        "nested subquery concept",
			queryStr:    "name FROM pod WHERE name IN kselect name FROM deployment",
			shouldWork:  true,
			description: "Pods matching deployment names",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Description: %s", tt.description)

			query, err := parser.Parse(tt.queryStr)
			if err != nil {
				if !tt.shouldWork {
					t.Logf("Parse failed as expected: %v", err)
					return
				}
				t.Fatalf("Parse failed: %v", err)
			}

			err = v.Validate(query)
			if tt.shouldWork && err != nil {
				t.Errorf("Validation should pass but failed: %v", err)
			}
			if !tt.shouldWork && err == nil {
				t.Error("Validation should fail but passed")
			}
		})
	}
}

// TestValidationErrorQuality tests the quality of validation error messages
func TestValidationErrorQuality(t *testing.T) {
	reg := registry.GetGlobalRegistry()
	v := validator.New(reg)

	tests := []struct {
		name        string
		queryStr    string
		errContains []string
	}{
		{
			name:        "resource not found with suggestions",
			queryStr:    "name FROM pode",
			errContains: []string{"Resource 'pode' not found", "Did you mean"},
		},
		{
			name:        "field not found with suggestions",
			queryStr:    "nam FROM pod",
			errContains: []string{"Field 'nam' not found", "Did you mean"},
		},
		{
			name:        "aggregation consistency error",
			queryStr:    "name,COUNT as count FROM pod",
			errContains: []string{"must appear in GROUP BY", "aggregate function"},
		},
		{
			name:        "DISTINCT with GROUP BY error",
			queryStr:    "DISTINCT namespace FROM pod GROUP BY namespace",
			errContains: []string{"DISTINCT cannot be used with GROUP BY"},
		},
		{
			name:        "wildcard with aggregates error",
			queryStr:    "*,COUNT as count FROM pod",
			errContains: []string{"Cannot use '*' with aggregate"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := parser.Parse(tt.queryStr)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			err = v.Validate(query)
			if err == nil {
				t.Fatal("Expected validation error but got none")
			}

			errMsg := err.Error()
			for _, substr := range tt.errContains {
				if !strings.Contains(errMsg, substr) {
					t.Errorf("Error message should contain '%s', got: %s", substr, errMsg)
				}
			}
		})
	}
}
