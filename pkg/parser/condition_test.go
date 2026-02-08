package parser

import (
	"testing"
)

func TestParseSimpleCondition(t *testing.T) {
	group, err := ParseConditions("namespace=default")
	if err != nil {
		t.Fatalf("ParseConditions failed: %v", err)
	}
	if len(group.Conditions) != 1 {
		t.Fatalf("Expected 1 condition, got %d", len(group.Conditions))
	}
	if group.Conditions[0].Field != "namespace" {
		t.Errorf("Expected field 'namespace', got '%s'", group.Conditions[0].Field)
	}
	if group.Conditions[0].Operator != OpEqual {
		t.Errorf("Expected operator '=', got '%s'", group.Conditions[0].Operator)
	}
	if group.Conditions[0].Value != "default" {
		t.Errorf("Expected value 'default', got '%s'", group.Conditions[0].Value)
	}
}

func TestParseAndConditions(t *testing.T) {
	group, err := ParseConditions("namespace=default AND status=Running")
	if err != nil {
		t.Fatalf("ParseConditions failed: %v", err)
	}
	if group.LogicalOperator != LogicalAnd {
		t.Errorf("Expected AND operator, got '%s'", group.LogicalOperator)
	}
	if len(group.Conditions) != 2 {
		t.Fatalf("Expected 2 conditions, got %d", len(group.Conditions))
	}
}

func TestParseOrConditions(t *testing.T) {
	group, err := ParseConditions("status=Running OR status=Pending")
	if err != nil {
		t.Fatalf("ParseConditions failed: %v", err)
	}
	if group.LogicalOperator != LogicalOr {
		t.Errorf("Expected OR operator, got '%s'", group.LogicalOperator)
	}
	if len(group.SubGroups) != 2 {
		t.Fatalf("Expected 2 subgroups, got %d", len(group.SubGroups))
	}
}

func TestParseNotEqualCondition(t *testing.T) {
	group, err := ParseConditions("status!=Failed")
	if err != nil {
		t.Fatalf("ParseConditions failed: %v", err)
	}
	if group.Conditions[0].Operator != OpNotEqual {
		t.Errorf("Expected operator '!=', got '%s'", group.Conditions[0].Operator)
	}
}

func TestParseLikeCondition(t *testing.T) {
	group, err := ParseConditions("name LIKE nginx-%")
	if err != nil {
		t.Fatalf("ParseConditions failed: %v", err)
	}
	if group.Conditions[0].Operator != OpLike {
		t.Errorf("Expected operator 'LIKE', got '%s'", group.Conditions[0].Operator)
	}
	if group.Conditions[0].Value != "nginx-%" {
		t.Errorf("Expected value 'nginx-%%', got '%s'", group.Conditions[0].Value)
	}
}

func TestParseComparisonConditions(t *testing.T) {
	tests := []struct {
		input    string
		operator ConditionOperator
	}{
		{"restarts > 5", OpGreaterThan},
		{"restarts < 10", OpLessThan},
		{"restarts >= 5", OpGreaterEqual},
		{"restarts <= 10", OpLessEqual},
	}

	for _, tt := range tests {
		group, err := ParseConditions(tt.input)
		if err != nil {
			t.Fatalf("ParseConditions failed for '%s': %v", tt.input, err)
		}
		if group.Conditions[0].Operator != tt.operator {
			t.Errorf("For '%s': expected operator '%s', got '%s'", tt.input, tt.operator, group.Conditions[0].Operator)
		}
	}
}

func TestEvaluateEqual(t *testing.T) {
	cond := Condition{Field: "status", Operator: OpEqual, Value: "Running"}
	if !cond.Evaluate("Running") {
		t.Error("Expected Running == Running to be true")
	}
	if cond.Evaluate("Pending") {
		t.Error("Expected Pending == Running to be false")
	}
}

func TestEvaluateLike(t *testing.T) {
	cond := Condition{Field: "name", Operator: OpLike, Value: "nginx-%"}
	if !cond.Evaluate("nginx-abc") {
		t.Error("Expected nginx-abc to match nginx-%")
	}
	if cond.Evaluate("apache-abc") {
		t.Error("Expected apache-abc to not match nginx-%")
	}
}

func TestEvaluateIn(t *testing.T) {
	cond := Condition{Field: "status", Operator: OpIn, Value: "(Running,Pending,Failed)"}
	if !cond.Evaluate("Running") {
		t.Error("Expected Running to be IN list")
	}
	if cond.Evaluate("Unknown") {
		t.Error("Expected Unknown to not be IN list")
	}
}

func TestEvaluateGreaterThan(t *testing.T) {
	cond := Condition{Field: "restarts", Operator: OpGreaterThan, Value: "5"}
	if !cond.Evaluate("10") {
		t.Error("Expected 10 > 5 to be true")
	}
	if cond.Evaluate("3") {
		t.Error("Expected 3 > 5 to be false")
	}
}

func TestEvaluateGroupAnd(t *testing.T) {
	group := &ConditionGroup{
		LogicalOperator: LogicalAnd,
		Conditions: []Condition{
			{Field: "status", Operator: OpEqual, Value: "Running"},
			{Field: "namespace", Operator: OpEqual, Value: "default"},
		},
	}

	obj := map[string]interface{}{
		"status":    "Running",
		"namespace": "default",
	}
	if !group.Evaluate(obj) {
		t.Error("Expected AND group to match")
	}

	obj["status"] = "Pending"
	if group.Evaluate(obj) {
		t.Error("Expected AND group to not match when status is Pending")
	}
}

func TestEvaluateGroupOr(t *testing.T) {
	group := &ConditionGroup{
		LogicalOperator: LogicalOr,
		SubGroups: []*ConditionGroup{
			{
				LogicalOperator: LogicalAnd,
				Conditions: []Condition{
					{Field: "status", Operator: OpEqual, Value: "Running"},
				},
			},
			{
				LogicalOperator: LogicalAnd,
				Conditions: []Condition{
					{Field: "status", Operator: OpEqual, Value: "Pending"},
				},
			},
		},
	}

	obj := map[string]interface{}{"status": "Running"}
	if !group.Evaluate(obj) {
		t.Error("Expected OR group to match Running")
	}

	obj["status"] = "Pending"
	if !group.Evaluate(obj) {
		t.Error("Expected OR group to match Pending")
	}

	obj["status"] = "Failed"
	if group.Evaluate(obj) {
		t.Error("Expected OR group to not match Failed")
	}
}

func TestParseShellSafeOperators(t *testing.T) {
	tests := []struct {
		input    string
		field    string
		operator ConditionOperator
		value    string
	}{
		{"restarts GT 10", "restarts", OpGreaterThan, "10"},
		{"restarts LT 5", "restarts", OpLessThan, "5"},
		{"restarts GE 10", "restarts", OpGreaterEqual, "10"},
		{"restarts LE 5", "restarts", OpLessEqual, "5"},
		{"status NE Running", "status", OpNotEqual, "Running"},
		{"namespace EQ default", "namespace", OpEqual, "default"},
		// case-insensitive
		{"restarts gt 10", "restarts", OpGreaterThan, "10"},
		{"restarts Gt 10", "restarts", OpGreaterThan, "10"},
	}

	for _, tt := range tests {
		group, err := ParseConditions(tt.input)
		if err != nil {
			t.Fatalf("ParseConditions failed for '%s': %v", tt.input, err)
		}
		if len(group.Conditions) != 1 {
			t.Fatalf("For '%s': expected 1 condition, got %d", tt.input, len(group.Conditions))
		}
		c := group.Conditions[0]
		if c.Field != tt.field {
			t.Errorf("For '%s': expected field '%s', got '%s'", tt.input, tt.field, c.Field)
		}
		if c.Operator != tt.operator {
			t.Errorf("For '%s': expected operator '%s', got '%s'", tt.input, tt.operator, c.Operator)
		}
		if c.Value != tt.value {
			t.Errorf("For '%s': expected value '%s', got '%s'", tt.input, tt.value, c.Value)
		}
	}
}

func TestParseShellSafeWithAndOr(t *testing.T) {
	group, err := ParseConditions("restarts GT 10 AND status NE Failed")
	if err != nil {
		t.Fatalf("ParseConditions failed: %v", err)
	}
	if len(group.Conditions) != 2 {
		t.Fatalf("Expected 2 conditions, got %d", len(group.Conditions))
	}
	if group.Conditions[0].Operator != OpGreaterThan {
		t.Errorf("Expected GT operator, got '%s'", group.Conditions[0].Operator)
	}
	if group.Conditions[1].Operator != OpNotEqual {
		t.Errorf("Expected NE operator, got '%s'", group.Conditions[1].Operator)
	}
}

func TestParseSubQuery(t *testing.T) {
	group, err := ParseConditions("name IN (name FROM pod WHERE status=Running)")
	if err != nil {
		t.Fatalf("ParseConditions failed: %v", err)
	}
	if len(group.Conditions) != 1 {
		t.Fatalf("Expected 1 condition, got %d", len(group.Conditions))
	}
	c := group.Conditions[0]
	if c.Field != "name" {
		t.Errorf("Expected field 'name', got '%s'", c.Field)
	}
	if c.Operator != OpIn {
		t.Errorf("Expected operator IN, got '%s'", c.Operator)
	}
	if c.SubQuery == nil {
		t.Fatal("Expected SubQuery to be set")
	}
	if c.SubQuery.Resource != "pod" {
		t.Errorf("Expected subquery resource 'pod', got '%s'", c.SubQuery.Resource)
	}
	if len(c.SubQuery.Fields) != 1 || c.SubQuery.Fields[0] != "name" {
		t.Errorf("Expected subquery fields [name], got %v", c.SubQuery.Fields)
	}
}

func TestParseSubQueryNotIn(t *testing.T) {
	group, err := ParseConditions("name NOT IN (name FROM deployment)")
	if err != nil {
		t.Fatalf("ParseConditions failed: %v", err)
	}
	c := group.Conditions[0]
	if c.Operator != OpNotIn {
		t.Errorf("Expected NOT IN, got '%s'", c.Operator)
	}
	if c.SubQuery == nil {
		t.Fatal("Expected SubQuery to be set")
	}
	if c.SubQuery.Resource != "deployment" {
		t.Errorf("Expected subquery resource 'deployment', got '%s'", c.SubQuery.Resource)
	}
}

func TestParseStaticInStillWorks(t *testing.T) {
	group, err := ParseConditions("status IN (Running,Pending)")
	if err != nil {
		t.Fatalf("ParseConditions failed: %v", err)
	}
	c := group.Conditions[0]
	if c.SubQuery != nil {
		t.Error("Expected SubQuery to be nil for static IN list")
	}
	if c.Operator != OpIn {
		t.Errorf("Expected IN, got '%s'", c.Operator)
	}
}

func TestSubQueryEvaluateWithValues(t *testing.T) {
	cond := Condition{
		Field:          "name",
		Operator:       OpIn,
		SubQuery:       &Query{Resource: "pod"}, // non-nil signals subquery
		SubQueryValues: []string{"nginx", "redis", "postgres"},
	}

	if !cond.Evaluate("nginx") {
		t.Error("Expected 'nginx' to be IN subquery values")
	}
	if cond.Evaluate("apache") {
		t.Error("Expected 'apache' to NOT be IN subquery values")
	}

	// NOT IN
	condNot := Condition{
		Field:          "name",
		Operator:       OpNotIn,
		SubQuery:       &Query{Resource: "pod"},
		SubQueryValues: []string{"nginx", "redis"},
	}
	if !condNot.Evaluate("apache") {
		t.Error("Expected 'apache' to pass NOT IN")
	}
	if condNot.Evaluate("nginx") {
		t.Error("Expected 'nginx' to fail NOT IN")
	}
}

func TestParseEmptyCondition(t *testing.T) {
	group, err := ParseConditions("")
	if err != nil {
		t.Fatalf("ParseConditions failed: %v", err)
	}
	if group.LogicalOperator != LogicalAnd {
		t.Errorf("Expected AND operator for empty conditions")
	}
}
