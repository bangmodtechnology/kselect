package registry

import "k8s.io/apimachinery/pkg/runtime/schema"

func init() {
	GetGlobalRegistry().Register(&ResourceDefinition{
		Name:    "clusterrole",
		Aliases: []string{"clusterroles"},
		GroupVersionResource: schema.GroupVersionResource{
			Group:    "rbac.authorization.k8s.io",
			Version:  "v1",
			Resource: "clusterroles",
		},
		Namespaced:    false, // cluster-scoped
		DefaultFields: []string{"name", "aggregation-rule", "age"},
		Fields: map[string]FieldDefinition{
			"name": {
				Name:        "name",
				JSONPath:    "{.metadata.name}",
				Description: "ClusterRole name",
				Type:        "string",
			},
			"rules": {
				Name:        "rules",
				JSONPath:    "{.rules}",
				Description: "Policy rules",
				Type:        "list",
			},
			"aggregation-rule": {
				Name:        "aggregation-rule",
				JSONPath:    "{.aggregationRule.clusterRoleSelectors}",
				Description: "Aggregation rule selectors",
				Type:        "list",
			},
			"age": {
				Name:        "age",
				JSONPath:    "{.metadata.creationTimestamp}",
				Description: "Age",
				Type:        "time",
			},
			"labels": {
				Name:        "labels",
				JSONPath:    "{.metadata.labels}",
				Description: "Labels",
				Type:        "map",
			},
		},
	})
}
