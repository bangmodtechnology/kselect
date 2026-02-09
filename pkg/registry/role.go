package registry

import "k8s.io/apimachinery/pkg/runtime/schema"

func init() {
	GetGlobalRegistry().Register(&ResourceDefinition{
		Name:    "role",
		Aliases: []string{"roles"},
		GroupVersionResource: schema.GroupVersionResource{
			Group:    "rbac.authorization.k8s.io",
			Version:  "v1",
			Resource: "roles",
		},
		Namespaced:    true,
		DefaultFields: []string{"name", "age"},
		Fields: map[string]FieldDefinition{
			"name": {
				Name:        "name",
				JSONPath:    "{.metadata.name}",
				Description: "Role name",
				Type:        "string",
			},
			"namespace": {
				Name:        "namespace",
				Aliases:     []string{"ns"},
				JSONPath:    "{.metadata.namespace}",
				Description: "Namespace",
				Type:        "string",
			},
			"rules": {
				Name:        "rules",
				JSONPath:    "{.rules}",
				Description: "Policy rules",
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
