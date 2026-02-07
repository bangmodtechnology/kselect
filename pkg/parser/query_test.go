package parser

import (
	"testing"
)

func TestParseSimpleQuery(t *testing.T) {
	query, err := Parse("name,status FROM pod WHERE namespace=default")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(query.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(query.Fields))
	}
	if query.Fields[0] != "name" {
		t.Errorf("Expected field 'name', got '%s'", query.Fields[0])
	}
	if query.Fields[1] != "status" {
		t.Errorf("Expected field 'status', got '%s'", query.Fields[1])
	}
	if query.Resource != "pod" {
		t.Errorf("Expected resource 'pod', got '%s'", query.Resource)
	}
	if query.Namespace != "default" {
		t.Errorf("Expected namespace 'default', got '%s'", query.Namespace)
	}
}

func TestParseWithOptionalSelect(t *testing.T) {
	// With SELECT keyword should also work
	query, err := Parse("SELECT name,status FROM pod WHERE namespace=default")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(query.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(query.Fields))
	}
	if query.Resource != "pod" {
		t.Errorf("Expected resource 'pod', got '%s'", query.Resource)
	}
}

func TestParseSelectStar(t *testing.T) {
	query, err := Parse("* FROM deployment WHERE namespace=production")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(query.Fields) != 1 || query.Fields[0] != "*" {
		t.Errorf("Expected fields ['*'], got %v", query.Fields)
	}
	if query.Namespace != "production" {
		t.Errorf("Expected namespace 'production', got '%s'", query.Namespace)
	}
}

func TestParseOrderBy(t *testing.T) {
	query, err := Parse("name,restarts FROM pod ORDER BY restarts DESC")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(query.OrderBy) != 1 {
		t.Fatalf("Expected 1 ORDER BY field, got %d", len(query.OrderBy))
	}
	if query.OrderBy[0].Field != "restarts" {
		t.Errorf("Expected ORDER BY field 'restarts', got '%s'", query.OrderBy[0].Field)
	}
	if !query.OrderBy[0].Descending {
		t.Error("Expected DESC order")
	}
}

func TestParseLimit(t *testing.T) {
	query, err := Parse("name FROM pod LIMIT 10")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if query.Limit != 10 {
		t.Errorf("Expected LIMIT 10, got %d", query.Limit)
	}
}

func TestParseLimitOffset(t *testing.T) {
	query, err := Parse("name FROM pod LIMIT 10 OFFSET 20")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if query.Limit != 10 {
		t.Errorf("Expected LIMIT 10, got %d", query.Limit)
	}
	if query.Offset != 20 {
		t.Errorf("Expected OFFSET 20, got %d", query.Offset)
	}
}

func TestParseDistinct(t *testing.T) {
	query, err := Parse("DISTINCT status FROM pod")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if !query.Distinct {
		t.Error("Expected DISTINCT to be true")
	}
	if len(query.Fields) != 1 || query.Fields[0] != "status" {
		t.Errorf("Expected field 'status', got %v", query.Fields)
	}
}

func TestParseGroupBy(t *testing.T) {
	query, err := Parse("namespace, COUNT(*) as pod_count FROM pod GROUP BY namespace")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(query.GroupBy) != 1 || query.GroupBy[0] != "namespace" {
		t.Errorf("Expected GROUP BY [namespace], got %v", query.GroupBy)
	}
	if len(query.Aggregates) != 1 {
		t.Fatalf("Expected 1 aggregate, got %d", len(query.Aggregates))
	}
	if query.Aggregates[0].Function != "COUNT" {
		t.Errorf("Expected COUNT function, got '%s'", query.Aggregates[0].Function)
	}
	if query.Aggregates[0].Alias != "pod_count" {
		t.Errorf("Expected alias 'pod_count', got '%s'", query.Aggregates[0].Alias)
	}
}

func TestParseComplexQuery(t *testing.T) {
	query, err := Parse("name,status FROM pod WHERE namespace=default AND status=Running ORDER BY name LIMIT 5")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if query.Resource != "pod" {
		t.Errorf("Expected resource 'pod', got '%s'", query.Resource)
	}
	if query.Conditions == nil {
		t.Fatal("Expected conditions to be parsed")
	}
	if len(query.OrderBy) != 1 {
		t.Errorf("Expected 1 ORDER BY field, got %d", len(query.OrderBy))
	}
	if query.Limit != 5 {
		t.Errorf("Expected LIMIT 5, got %d", query.Limit)
	}
}

func TestParseDefaultNamespace(t *testing.T) {
	query, err := Parse("name FROM pod")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if query.Namespace != "default" {
		t.Errorf("Expected default namespace 'default', got '%s'", query.Namespace)
	}
}

func TestParseNoFrom(t *testing.T) {
	_, err := Parse("name,status pod")
	if err == nil {
		t.Error("Expected error for missing FROM keyword")
	}
}

func TestParseNoFieldsUseDefault(t *testing.T) {
	query, err := Parse("FROM pod WHERE namespace=default")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if !query.UseDefault {
		t.Error("Expected UseDefault to be true")
	}
	if query.Resource != "pod" {
		t.Errorf("Expected resource 'pod', got '%s'", query.Resource)
	}
	if query.Namespace != "default" {
		t.Errorf("Expected namespace 'default', got '%s'", query.Namespace)
	}
}

func TestParseNoFieldsNoWhere(t *testing.T) {
	query, err := Parse("FROM deployment")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if !query.UseDefault {
		t.Error("Expected UseDefault to be true")
	}
	if query.Resource != "deployment" {
		t.Errorf("Expected resource 'deployment', got '%s'", query.Resource)
	}
}

func TestParseNamespaceAliasNs(t *testing.T) {
	query, err := Parse("name FROM pod WHERE ns=production")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if query.Namespace != "production" {
		t.Errorf("Expected namespace 'production' from alias 'ns', got '%s'", query.Namespace)
	}
	if query.Conditions == nil {
		t.Fatal("Expected conditions to be parsed")
	}
	if len(query.Conditions.Conditions) != 1 {
		t.Fatalf("Expected 1 condition, got %d", len(query.Conditions.Conditions))
	}
	if query.Conditions.Conditions[0].Field != "ns" {
		t.Errorf("Expected condition field 'ns', got '%s'", query.Conditions.Conditions[0].Field)
	}
}

func TestParseMultipleOrderBy(t *testing.T) {
	query, err := Parse("namespace,name FROM pod ORDER BY namespace ASC, name DESC")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(query.OrderBy) != 2 {
		t.Fatalf("Expected 2 ORDER BY fields, got %d", len(query.OrderBy))
	}
	if query.OrderBy[0].Field != "namespace" || query.OrderBy[0].Descending {
		t.Errorf("Expected 'namespace ASC', got field=%s desc=%v", query.OrderBy[0].Field, query.OrderBy[0].Descending)
	}
	if query.OrderBy[1].Field != "name" || !query.OrderBy[1].Descending {
		t.Errorf("Expected 'name DESC', got field=%s desc=%v", query.OrderBy[1].Field, query.OrderBy[1].Descending)
	}
}
