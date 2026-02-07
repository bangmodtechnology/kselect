package registry

import "testing"

func TestResolveFieldAlias(t *testing.T) {
	def := &ResourceDefinition{
		Name: "pod",
		Fields: map[string]FieldDefinition{
			"namespace": {
				Name:    "namespace",
				Aliases: []string{"ns"},
			},
			"name": {
				Name: "name",
			},
			"status": {
				Name:    "status",
				Aliases: []string{"st"},
			},
		},
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"ns", "namespace"},
		{"namespace", "namespace"},
		{"st", "status"},
		{"status", "status"},
		{"name", "name"},
		{"unknown", "unknown"}, // returns as-is if no match
	}

	for _, tt := range tests {
		result := def.ResolveFieldAlias(tt.input)
		if result != tt.expected {
			t.Errorf("ResolveFieldAlias(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
