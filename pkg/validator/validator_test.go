package validator

import (
	"strings"
	"testing"

	"github.com/bangmodtechnology/kselect/pkg/parser"
	"github.com/bangmodtechnology/kselect/pkg/registry"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func createTestRegistry() *registry.Registry {
	reg := registry.NewRegistry()

	// Register test pod resource
	reg.Register(&registry.ResourceDefinition{
		Name:    "pod",
		Aliases: []string{"pods", "po"},
		GroupVersionResource: schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "pods",
		},
		Namespaced:    true,
		DefaultFields: []string{"name", "status", "ip"},
		Fields: map[string]registry.FieldDefinition{
			"name": {
				Name:        "name",
				Aliases:     []string{},
				JSONPath:    "{.metadata.name}",
				Description: "Pod name",
				Type:        "string",
			},
			"namespace": {
				Name:        "namespace",
				Aliases:     []string{"ns"},
				JSONPath:    "{.metadata.namespace}",
				Description: "Namespace",
				Type:        "string",
			},
			"status": {
				Name:        "status",
				JSONPath:    "{.status.phase}",
				Description: "Pod status",
				Type:        "string",
			},
			"ip": {
				Name:        "ip",
				JSONPath:    "{.status.podIP}",
				Description: "Pod IP",
				Type:        "string",
			},
			"restarts": {
				Name:        "restarts",
				JSONPath:    "{.status.containerStatuses[0].restartCount}",
				Description: "Restart count",
				Type:        "number",
			},
		},
	})

	// Register test deployment resource
	reg.Register(&registry.ResourceDefinition{
		Name:    "deployment",
		Aliases: []string{"deployments", "deploy"},
		GroupVersionResource: schema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: "deployments",
		},
		Namespaced:    true,
		DefaultFields: []string{"name", "replicas", "ready"},
		Fields: map[string]registry.FieldDefinition{
			"name": {
				Name:        "name",
				JSONPath:    "{.metadata.name}",
				Description: "Deployment name",
				Type:        "string",
			},
			"replicas": {
				Name:        "replicas",
				JSONPath:    "{.spec.replicas}",
				Description: "Desired replicas",
				Type:        "number",
			},
			"ready": {
				Name:        "ready",
				JSONPath:    "{.status.readyReplicas}",
				Description: "Ready replicas",
				Type:        "number",
			},
		},
	})

	return reg
}

func TestValidateResource(t *testing.T) {
	reg := createTestRegistry()
	v := New(reg)

	tests := []struct {
		name        string
		resource    string
		shouldError bool
		hasSuggestion bool
	}{
		{"valid resource", "pod", false, false},
		{"valid alias", "po", false, false},
		{"invalid resource", "invalid", true, false},
		{"typo in resource", "pode", true, true}, // should suggest "pod"
		{"empty resource", "", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.validateResource(tt.resource)

			if tt.shouldError && err == nil {
				t.Errorf("Expected error for resource '%s'", tt.resource)
			}

			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error for resource '%s': %v", tt.resource, err)
			}

			if tt.hasSuggestion {
				valErr, ok := err.(*ValidationError)
				if !ok {
					t.Errorf("Expected ValidationError, got %T", err)
				} else if len(valErr.Suggestions) == 0 {
					t.Errorf("Expected suggestions for typo '%s'", tt.resource)
				}
			}
		})
	}
}

func TestValidateFields(t *testing.T) {
	reg := createTestRegistry()
	v := New(reg)
	resource, _ := reg.Get("pod")

	tests := []struct {
		name        string
		fields      []string
		shouldError bool
		hasSuggestion bool
	}{
		{"valid fields", []string{"name", "status"}, false, false},
		{"valid with alias", []string{"name", "ns"}, false, false},
		{"wildcard", []string{"*"}, false, false},
		{"empty fields", []string{}, false, false},
		{"invalid field", []string{"invalid"}, true, false},
		{"typo in field", []string{"nam"}, true, true}, // should suggest "name"
		{"aggregate function", []string{"COUNT"}, false, false},
		{"aggregate with field", []string{"SUM.restarts"}, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.validateFields(resource, tt.fields)

			if tt.shouldError && err == nil {
				t.Errorf("Expected error for fields %v", tt.fields)
			}

			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error for fields %v: %v", tt.fields, err)
			}

			if tt.hasSuggestion {
				valErr, ok := err.(*ValidationError)
				if !ok {
					t.Errorf("Expected ValidationError, got %T", err)
				} else if len(valErr.Suggestions) == 0 {
					t.Errorf("Expected suggestions for typo in fields %v", tt.fields)
				}
			}
		})
	}
}

func TestValidateQuery(t *testing.T) {
	reg := createTestRegistry()
	v := New(reg)

	tests := []struct {
		name        string
		queryStr    string
		shouldError bool
	}{
		{"valid basic query", "name,status FROM pod", false},
		{"valid with WHERE", "name FROM pod WHERE namespace=default", false},
		{"valid with alias", "name,ns FROM pod", false},
		{"invalid resource", "name FROM invalid", true},
		{"invalid field", "invalid FROM pod", true},
		{"invalid WHERE field", "name FROM pod WHERE invalid=value", true},
		{"valid ORDER BY", "name FROM pod ORDER BY status", false},
		{"invalid ORDER BY", "name FROM pod ORDER BY invalid", true},
		{"valid aggregation", "namespace, COUNT FROM pod GROUP BY namespace", false},
		{"valid SUM", "namespace, SUM.restarts FROM pod GROUP BY namespace", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := parser.Parse(tt.queryStr)
			if err != nil {
				t.Fatalf("Failed to parse query: %v", err)
			}

			err = v.Validate(query)

			if tt.shouldError && err == nil {
				t.Errorf("Expected error for query '%s'", tt.queryStr)
			}

			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error for query '%s': %v", tt.queryStr, err)
			}
		})
	}
}

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		s1       string
		s2       string
		expected int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "a", 1},
		{"a", "a", 0},
		{"ab", "ab", 0},
		{"pod", "pode", 1},
		{"pod", "pot", 1},
		{"pod", "deployment", 8},
		{"namespace", "namespac", 1},
		{"status", "statuz", 1},
	}

	for _, tt := range tests {
		t.Run(tt.s1+"_"+tt.s2, func(t *testing.T) {
			dist := levenshteinDistance(tt.s1, tt.s2)
			if dist != tt.expected {
				t.Errorf("levenshteinDistance(%q, %q) = %d, expected %d", tt.s1, tt.s2, dist, tt.expected)
			}
		})
	}
}

func TestValidationErrorMessage(t *testing.T) {
	tests := []struct {
		name        string
		err         *ValidationError
		contains    []string
	}{
		{
			name: "no suggestions",
			err: &ValidationError{
				Message:     "Field not found",
				Suggestions: []string{},
			},
			contains: []string{"Field not found"},
		},
		{
			name: "one suggestion",
			err: &ValidationError{
				Message:     "Resource 'pode' not found",
				Suggestions: []string{"pod"},
			},
			contains: []string{"Resource 'pode' not found", "Did you mean: pod?"},
		},
		{
			name: "multiple suggestions",
			err: &ValidationError{
				Message:     "Field 'nam' not found",
				Suggestions: []string{"name", "namespace"},
			},
			contains: []string{"Field 'nam' not found", "Did you mean one of these?", "name", "namespace"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.err.Error()
			for _, substr := range tt.contains {
				if !strings.Contains(msg, substr) {
					t.Errorf("Error message should contain %q, got: %s", substr, msg)
				}
			}
		})
	}
}

func TestValidateAggregationConsistency(t *testing.T) {
	reg := createTestRegistry()
	v := New(reg)

	tests := []struct {
		name        string
		queryStr    string
		shouldError bool
		errContains string
	}{
		{
			name:        "aggregate with GROUP BY - valid",
			queryStr:    "namespace, COUNT as count FROM pod GROUP BY namespace",
			shouldError: false,
		},
		{
			name:        "aggregate without GROUP BY but all fields are aggregates - valid",
			queryStr:    "COUNT as count FROM pod",
			shouldError: false,
		},
		{
			name:        "non-aggregate field with aggregate but no GROUP BY - invalid",
			queryStr:    "name, COUNT as count FROM pod",
			shouldError: true,
			errContains: "must appear in GROUP BY",
		},
		{
			name:        "wildcard with aggregate - invalid",
			queryStr:    "*, COUNT as count FROM pod",
			shouldError: true,
			errContains: "Cannot use '*' with aggregate functions",
		},
		{
			name:        "wildcard with GROUP BY - invalid",
			queryStr:    "* FROM pod GROUP BY namespace",
			shouldError: true,
			errContains: "Cannot use '*' with GROUP BY",
		},
		{
			name:        "non-grouped field in SELECT with GROUP BY - invalid",
			queryStr:    "name, namespace FROM pod GROUP BY namespace",
			shouldError: true,
			errContains: "must appear in GROUP BY",
		},
		{
			name:        "SUM with GROUP BY - valid",
			queryStr:    "namespace, SUM.restarts as total FROM pod GROUP BY namespace",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := parser.Parse(tt.queryStr)
			if err != nil {
				t.Fatalf("Failed to parse query: %v", err)
			}

			err = v.Validate(query)

			if tt.shouldError && err == nil {
				t.Errorf("Expected error for query '%s'", tt.queryStr)
			}

			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error for query '%s': %v", tt.queryStr, err)
			}

			if tt.shouldError && err != nil && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Error should contain '%s', got: %v", tt.errContains, err)
				}
			}
		})
	}
}

func TestValidateDistinct(t *testing.T) {
	reg := createTestRegistry()
	v := New(reg)

	tests := []struct {
		name        string
		query       *parser.Query
		shouldError bool
		errContains string
	}{
		{
			name: "DISTINCT with normal fields - valid",
			query: &parser.Query{
				Fields:   []string{"name", "namespace"},
				Resource: "pod",
				Distinct: true,
			},
			shouldError: false,
		},
		{
			name: "DISTINCT with wildcard - valid",
			query: &parser.Query{
				Fields:   []string{"*"},
				Resource: "pod",
				Distinct: true,
			},
			shouldError: false,
		},
		{
			name: "DISTINCT with aggregates - invalid",
			query: &parser.Query{
				Fields:     []string{"namespace", "COUNT"},
				Resource:   "pod",
				Distinct:   true,
				Aggregates: []parser.AggregateFunc{{Function: "COUNT", Field: "*", Alias: "count"}},
			},
			shouldError: true,
			errContains: "DISTINCT cannot be used with aggregate functions",
		},
		{
			name: "DISTINCT with GROUP BY - invalid",
			query: &parser.Query{
				Fields:   []string{"namespace"},
				Resource: "pod",
				Distinct: true,
				GroupBy:  []string{"namespace"},
			},
			shouldError: true,
			errContains: "DISTINCT cannot be used with GROUP BY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.query)

			if tt.shouldError && err == nil {
				t.Errorf("Expected error")
			}

			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if tt.shouldError && err != nil && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Error should contain '%s', got: %v", tt.errContains, err)
				}
			}
		})
	}
}
