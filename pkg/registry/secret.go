package registry

import "k8s.io/apimachinery/pkg/runtime/schema"

func init() {
	GetGlobalRegistry().Register(&ResourceDefinition{
		Name:    "secret",
		Aliases: []string{"secrets"},
		GroupVersionResource: schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "secrets",
		},
		Namespaced:    true,
		DefaultFields: []string{"name", "type", "age"},
		Fields: map[string]FieldDefinition{
			"name": {
				Name:        "name",
				JSONPath:    "{.metadata.name}",
				Description: "Secret name",
				Type:        "string",
			},
			"namespace": {
				Name:        "namespace",
				Aliases:     []string{"ns"},
				JSONPath:    "{.metadata.namespace}",
				Description: "Namespace",
				Type:        "string",
			},
			"type": {
				Name:        "type",
				JSONPath:    "{.type}",
				Description: "Secret type",
				Type:        "string",
			},
			"data-keys": {
				Name:        "data-keys",
				JSONPath:    "{.data}",
				Description: "Data keys",
				Type:        "map",
			},
			"age": {
				Name:        "age",
				JSONPath:    "{.metadata.creationTimestamp}",
				Description: "Age",
				Type:        "time",
			},
		},
	})
}
