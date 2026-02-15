package executor

import (
	"reflect"
	"testing"

	"github.com/bangmodtechnology/kselect/pkg/parser"
	"github.com/bangmodtechnology/kselect/pkg/registry"
)

func newTestRegistry() *registry.Registry {
	reg := registry.NewRegistry()
	reg.Register(&registry.ResourceDefinition{
		Name:          "pod",
		DefaultFields: []string{"name", "status", "namespace"},
		Fields: map[string]registry.FieldDefinition{
			"name":      {Name: "name"},
			"status":    {Name: "status"},
			"namespace": {Name: "namespace"},
			"ip":        {Name: "ip"},
		},
	})
	reg.Register(&registry.ResourceDefinition{
		Name:          "service",
		DefaultFields: []string{"name", "type", "cluster-ip"},
		Fields: map[string]registry.FieldDefinition{
			"name":       {Name: "name"},
			"type":       {Name: "type"},
			"cluster-ip": {Name: "cluster-ip"},
		},
	})
	reg.Register(&registry.ResourceDefinition{
		Name: "configmap",
		// No DefaultFields — should fall back to sorted field names
		Fields: map[string]registry.FieldDefinition{
			"name":      {Name: "name"},
			"namespace": {Name: "namespace"},
			"data-keys": {Name: "data-keys"},
		},
	})
	return reg
}

func TestResolveJoinFieldsStar(t *testing.T) {
	reg := newTestRegistry()

	query := &parser.Query{
		Fields:   []string{"*"},
		Resource: "pod",
		Joins: []parser.JoinClause{
			{
				Type:     parser.InnerJoin,
				Resource: "service",
				Alias:    "svc",
			},
		},
	}

	fields := resolveJoinFields(query, reg)

	// Primary: pod.name, pod.status, pod.namespace (from DefaultFields)
	// Join: svc.name, svc.type, svc.cluster-ip (from DefaultFields, with alias prefix)
	expected := []string{
		"pod.name", "pod.status", "pod.namespace",
		"svc.name", "svc.type", "svc.cluster-ip",
	}

	if !reflect.DeepEqual(fields, expected) {
		t.Errorf("Expected fields %v, got %v", expected, fields)
	}
}

func TestResolveJoinFieldsStarNoDefaults(t *testing.T) {
	reg := newTestRegistry()

	query := &parser.Query{
		Fields:   []string{"*"},
		Resource: "pod",
		Joins: []parser.JoinClause{
			{
				Type:     parser.InnerJoin,
				Resource: "configmap",
				Alias:    "cm",
			},
		},
	}

	fields := resolveJoinFields(query, reg)

	// configmap has no DefaultFields → sorted field names
	// pod: DefaultFields → pod.name, pod.status, pod.namespace
	// configmap: sorted → cm.data-keys, cm.name, cm.namespace
	expected := []string{
		"pod.name", "pod.status", "pod.namespace",
		"cm.data-keys", "cm.name", "cm.namespace",
	}

	if !reflect.DeepEqual(fields, expected) {
		t.Errorf("Expected fields %v, got %v", expected, fields)
	}
}

func TestResolveJoinFieldsExplicit(t *testing.T) {
	reg := newTestRegistry()

	query := &parser.Query{
		Fields:   []string{"pod.name", "svc.type"},
		Resource: "pod",
		Joins: []parser.JoinClause{
			{
				Type:     parser.InnerJoin,
				Resource: "service",
				Alias:    "svc",
			},
		},
	}

	fields := resolveJoinFields(query, reg)

	// Explicit fields should be returned as-is
	expected := []string{"pod.name", "svc.type"}
	if !reflect.DeepEqual(fields, expected) {
		t.Errorf("Expected fields %v, got %v", expected, fields)
	}
}

func TestPerformHashJoinInner(t *testing.T) {
	left := []map[string]interface{}{
		{"name": "pod-a", "namespace": "default"},
		{"name": "pod-b", "namespace": "default"},
		{"name": "pod-c", "namespace": "kube-system"},
	}
	right := []map[string]interface{}{
		{"name": "svc-a", "selector": "pod-a"},
		{"name": "svc-b", "selector": "pod-x"}, // no match
	}

	join := parser.JoinClause{
		Type: parser.InnerJoin,
		Conditions: []parser.JoinCondition{
			{LeftField: "name", RightField: "selector"},
		},
	}

	results := performJoin(left, right, join)
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0]["name"] != "svc-a" { // right overwrites left "name"
		t.Errorf("Expected merged name 'svc-a', got '%v'", results[0]["name"])
	}
}

func TestPerformHashJoinMultipleConditions(t *testing.T) {
	left := []map[string]interface{}{
		{"name": "pod-a", "namespace": "default"},
		{"name": "pod-a", "namespace": "production"},
	}
	right := []map[string]interface{}{
		{"svc-name": "svc-a", "svc-ns": "default"},
	}

	join := parser.JoinClause{
		Type: parser.InnerJoin,
		Conditions: []parser.JoinCondition{
			{LeftField: "name", RightField: "svc-name"},
			{LeftField: "namespace", RightField: "svc-ns"},
		},
	}

	// Only name=pod-a AND namespace=default should match (not production)
	// But right has svc-name=svc-a not pod-a, so no match... let me fix the test data
	right = []map[string]interface{}{
		{"svc-name": "pod-a", "svc-ns": "default"},
	}

	results := performJoin(left, right, join)
	if len(results) != 1 {
		t.Fatalf("Expected 1 result (only default namespace matches), got %d", len(results))
	}
	if results[0]["namespace"] != "default" {
		t.Errorf("Expected namespace 'default', got '%v'", results[0]["namespace"])
	}
}

func TestPerformHashJoinLeft(t *testing.T) {
	left := []map[string]interface{}{
		{"name": "pod-a"},
		{"name": "pod-b"},
	}
	right := []map[string]interface{}{
		{"selector": "pod-a", "svc": "svc-a"},
	}

	join := parser.JoinClause{
		Type: parser.LeftJoin,
		Conditions: []parser.JoinCondition{
			{LeftField: "name", RightField: "selector"},
		},
	}

	results := performJoin(left, right, join)
	if len(results) != 2 {
		t.Fatalf("Expected 2 results (left join keeps all left rows), got %d", len(results))
	}
	// pod-a should have svc field, pod-b should not
	if results[0]["svc"] != "svc-a" {
		t.Errorf("Expected svc 'svc-a' for pod-a match, got '%v'", results[0]["svc"])
	}
	if results[1]["svc"] != nil {
		t.Errorf("Expected no svc for pod-b (unmatched), got '%v'", results[1]["svc"])
	}
}

func TestPerformHashJoinRight(t *testing.T) {
	left := []map[string]interface{}{
		{"name": "pod-a"},
	}
	right := []map[string]interface{}{
		{"selector": "pod-a", "svc": "svc-a"},
		{"selector": "pod-x", "svc": "svc-x"},
	}

	join := parser.JoinClause{
		Type: parser.RightJoin,
		Conditions: []parser.JoinCondition{
			{LeftField: "name", RightField: "selector"},
		},
	}

	results := performJoin(left, right, join)
	if len(results) != 2 {
		t.Fatalf("Expected 2 results (right join keeps all right rows), got %d", len(results))
	}
}

func TestBuildJoinKeyComposite(t *testing.T) {
	row := map[string]interface{}{
		"name": "pod-a",
		"ns":   "default",
	}
	conditions := []parser.JoinCondition{
		{LeftField: "name", RightField: "svc-name"},
		{LeftField: "ns", RightField: "svc-ns"},
	}

	key := buildJoinKey(row, conditions, true)
	if key != "pod-a\x00default" {
		t.Errorf("Expected composite key 'pod-a\\x00default', got '%s'", key)
	}

	// Nil value should return empty key
	row2 := map[string]interface{}{"name": "pod-a"}
	key2 := buildJoinKey(row2, conditions, true)
	if key2 != "" {
		t.Errorf("Expected empty key for nil field, got '%s'", key2)
	}
}
