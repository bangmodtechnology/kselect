package registry

import "k8s.io/apimachinery/pkg/runtime/schema"

func init() {
	GetGlobalRegistry().Register(&ResourceDefinition{
		Name:    "resourcequota",
		Aliases: []string{"resourcequotas", "quota", "quotas"},
		GroupVersionResource: schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "resourcequotas",
		},
		Namespaced:    true,
		DefaultFields: []string{"name", "hard", "used", "age"},
		Fields: map[string]FieldDefinition{
			"name": {
				Name:        "name",
				JSONPath:    "{.metadata.name}",
				Description: "ResourceQuota name",
				Type:        "string",
			},
			"namespace": {
				Name:        "namespace",
				Aliases:     []string{"ns"},
				JSONPath:    "{.metadata.namespace}",
				Description: "Namespace",
				Type:        "string",
			},
			"hard": {
				Name:        "hard",
				JSONPath:    "{.status.hard}",
				Description: "Hard limits",
				Type:        "map",
			},
			"used": {
				Name:        "used",
				JSONPath:    "{.status.used}",
				Description: "Used resources",
				Type:        "map",
			},
			"scopes": {
				Name:        "scopes",
				JSONPath:    "{.spec.scopes}",
				Description: "Quota scopes",
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
