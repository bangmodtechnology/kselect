package executor

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bangmodtechnology/kselect/pkg/parser"
	"github.com/bangmodtechnology/kselect/pkg/registry"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Executor struct {
	dynamicClient dynamic.Interface
	registry      *registry.Registry
}

func NewExecutor() (*Executor, error) {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	return &Executor{
		dynamicClient: dynamicClient,
		registry:      registry.GetGlobalRegistry(),
	}, nil
}

func (e *Executor) Execute(query *parser.Query) ([]map[string]interface{}, []string, error) {
	// Handle JOIN queries
	if len(query.Joins) > 0 {
		return e.executeJoin(query)
	}

	resDef, ok := e.registry.Get(query.Resource)
	if !ok {
		return nil, nil, fmt.Errorf("unknown resource: %s (use --list to see available resources)", query.Resource)
	}

	// Resolve field aliases in query (e.g. "ns" → "namespace")
	resolveQueryAliases(query, resDef)

	// Fetch resources from K8s
	items, err := e.fetchResources(resDef, query)
	if err != nil {
		return nil, nil, err
	}

	// Resolve fields (expand * to all fields)
	fields := e.resolveFields(query, resDef)

	// Extract field values and apply WHERE filter
	var results []map[string]interface{}
	for _, item := range items {
		row := e.extractRow(&item, resDef, fields)

		// Always include namespace for filtering even if not in selected fields
		if _, has := row["namespace"]; !has {
			if nsDef, ok := resDef.Fields["namespace"]; ok {
				row["namespace"] = e.extractField(&item, nsDef.JSONPath)
			}
		}

		// Apply WHERE conditions
		if query.Conditions != nil && !query.Conditions.Evaluate(row) {
			continue
		}

		results = append(results, row)
	}

	// Apply aggregations if present
	if len(query.Aggregates) > 0 || len(query.GroupBy) > 0 {
		results, fields = applyAggregation(results, query, fields)
	}

	// Apply DISTINCT
	if query.Distinct {
		results = applyDistinct(results, fields)
	}

	// Apply ORDER BY
	if len(query.OrderBy) > 0 {
		sortResults(results, query.OrderBy)
	}

	// Apply LIMIT and OFFSET
	results = applyLimitOffset(results, query.Limit, query.Offset)

	return results, fields, nil
}

func (e *Executor) fetchResources(resDef *registry.ResourceDefinition, query *parser.Query) ([]unstructured.Unstructured, error) {
	listOptions := metav1.ListOptions{}
	if query.FieldSelector != "" {
		listOptions.FieldSelector = query.FieldSelector
	}
	if len(query.Labels) > 0 {
		var labels []string
		for k, v := range query.Labels {
			if v == "" {
				labels = append(labels, k)
			} else {
				labels = append(labels, fmt.Sprintf("%s=%s", k, v))
			}
		}
		listOptions.LabelSelector = strings.Join(labels, ",")
	}

	gvr := resDef.GroupVersionResource
	var list *unstructured.UnstructuredList
	var err error

	if query.Namespace == "*" || query.Namespace == "" {
		list, err = e.dynamicClient.Resource(gvr).Namespace("").List(context.TODO(), listOptions)
	} else {
		list, err = e.dynamicClient.Resource(gvr).Namespace(query.Namespace).List(context.TODO(), listOptions)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list %s: %w", resDef.Name, err)
	}

	return list.Items, nil
}

func (e *Executor) resolveFields(query *parser.Query, resDef *registry.ResourceDefinition) []string {
	// * or empty → use DefaultFields, fallback to all fields
	if len(query.Fields) == 0 || (len(query.Fields) == 1 && query.Fields[0] == "*") {
		if len(resDef.DefaultFields) > 0 {
			return resDef.DefaultFields
		}
		var fields []string
		for name := range resDef.Fields {
			fields = append(fields, name)
		}
		sort.Strings(fields)
		return fields
	}

	return query.Fields
}

func (e *Executor) extractRow(item *unstructured.Unstructured, resDef *registry.ResourceDefinition, fields []string) map[string]interface{} {
	row := make(map[string]interface{})
	// Extract all known fields so WHERE conditions can reference any field
	for fieldName, fieldDef := range resDef.Fields {
		row[fieldName] = e.extractField(item, fieldDef.JSONPath)
	}
	return row
}

func (e *Executor) extractField(obj *unstructured.Unstructured, jsonPath string) interface{} {
	path := strings.TrimPrefix(jsonPath, "{.")
	path = strings.TrimSuffix(path, "}")

	parts := strings.Split(path, ".")
	var current interface{} = obj.Object

	for _, part := range parts {
		if strings.Contains(part, "[*]") {
			part = strings.TrimSuffix(part, "[*]")
			m, ok := current.(map[string]interface{})
			if !ok {
				return nil
			}
			arr, ok := m[part]
			if !ok {
				return nil
			}
			// Return array values
			if slice, ok := arr.([]interface{}); ok {
				return slice
			}
			return arr
		}

		m, ok := current.(map[string]interface{})
		if !ok {
			return nil
		}

		current, ok = m[part]
		if !ok {
			return nil
		}
	}

	return current
}

func sortResults(results []map[string]interface{}, orderBy []parser.OrderByField) {
	sort.SliceStable(results, func(i, j int) bool {
		for _, order := range orderBy {
			vi := fmt.Sprintf("%v", results[i][order.Field])
			vj := fmt.Sprintf("%v", results[j][order.Field])

			// Try numeric comparison
			ni, errI := strconv.ParseFloat(vi, 64)
			nj, errJ := strconv.ParseFloat(vj, 64)
			if errI == nil && errJ == nil {
				if ni != nj {
					if order.Descending {
						return ni > nj
					}
					return ni < nj
				}
				continue
			}

			if vi != vj {
				if order.Descending {
					return vi > vj
				}
				return vi < vj
			}
		}
		return false
	})
}

func applyLimitOffset(results []map[string]interface{}, limit, offset int) []map[string]interface{} {
	if limit <= 0 && offset <= 0 {
		return results
	}

	start := offset
	if start > len(results) {
		return nil
	}

	results = results[start:]

	if limit > 0 && limit < len(results) {
		results = results[:limit]
	}

	return results
}

// resolveQueryAliases resolves field aliases (e.g. "ns" → "namespace") throughout the query.
func resolveQueryAliases(query *parser.Query, resDef *registry.ResourceDefinition) {
	// Resolve aliases in selected fields
	for i, f := range query.Fields {
		query.Fields[i] = resDef.ResolveFieldAlias(f)
	}

	// Resolve aliases in WHERE conditions
	if query.Conditions != nil {
		resolveConditionAliases(query.Conditions, resDef)
	}

	// Resolve aliases in ORDER BY
	for i, ob := range query.OrderBy {
		query.OrderBy[i].Field = resDef.ResolveFieldAlias(ob.Field)
	}

	// Resolve aliases in GROUP BY
	for i, gb := range query.GroupBy {
		query.GroupBy[i] = resDef.ResolveFieldAlias(gb)
	}

	// Resolve aliases in HAVING
	if query.Having != nil {
		resolveConditionAliases(query.Having, resDef)
	}
}

func resolveConditionAliases(group *parser.ConditionGroup, resDef *registry.ResourceDefinition) {
	for i, cond := range group.Conditions {
		group.Conditions[i].Field = resDef.ResolveFieldAlias(cond.Field)
	}
	for _, sub := range group.SubGroups {
		resolveConditionAliases(sub, resDef)
	}
}

func applyDistinct(results []map[string]interface{}, fields []string) []map[string]interface{} {
	seen := make(map[string]bool)
	var unique []map[string]interface{}

	for _, row := range results {
		var parts []string
		for _, f := range fields {
			parts = append(parts, fmt.Sprintf("%v", row[f]))
		}
		key := strings.Join(parts, "|")
		if !seen[key] {
			seen[key] = true
			unique = append(unique, row)
		}
	}

	return unique
}

func FormatAge(timestamp interface{}) string {
	if timestamp == nil {
		return "<none>"
	}

	ts, ok := timestamp.(string)
	if !ok {
		return fmt.Sprintf("%v", timestamp)
	}

	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return ts
	}

	duration := time.Since(t)
	switch {
	case duration.Hours() >= 24*365:
		return fmt.Sprintf("%dy", int(duration.Hours()/(24*365)))
	case duration.Hours() >= 24:
		return fmt.Sprintf("%dd", int(duration.Hours()/24))
	case duration.Hours() >= 1:
		return fmt.Sprintf("%dh", int(duration.Hours()))
	case duration.Minutes() >= 1:
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	default:
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	}
}
