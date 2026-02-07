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

func TestParseEmptyCondition(t *testing.T) {
	group, err := ParseConditions("")
	if err != nil {
		t.Fatalf("ParseConditions failed: %v", err)
	}
	if group.LogicalOperator != LogicalAnd {
		t.Errorf("Expected AND operator for empty conditions")
	}
}
