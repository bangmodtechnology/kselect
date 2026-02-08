package registry

import "k8s.io/apimachinery/pkg/runtime/schema"

func init() {
	GetGlobalRegistry().Register(&ResourceDefinition{
		Name:    "job",
		Aliases: []string{"jobs"},
		GroupVersionResource: schema.GroupVersionResource{
			Group:    "batch",
			Version:  "v1",
			Resource: "jobs",
		},
		Namespaced:    true,
		DefaultFields: []string{"name", "completions", "succeeded", "failed", "age"},
		Fields: map[string]FieldDefinition{
			"name": {
				Name:        "name",
				JSONPath:    "{.metadata.name}",
				Description: "Job name",
				Type:        "string",
			},
			"namespace": {
				Name:        "namespace",
				Aliases:     []string{"ns"},
				JSONPath:    "{.metadata.namespace}",
				Description: "Namespace",
				Type:        "string",
			},
			"completions": {
				Name:        "completions",
				JSONPath:    "{.spec.completions}",
				Description: "Desired completions",
				Type:        "int",
			},
			"succeeded": {
				Name:        "succeeded",
				JSONPath:    "{.status.succeeded}",
				Description: "Succeeded count",
				Type:        "int",
			},
			"failed": {
				Name:        "failed",
				JSONPath:    "{.status.failed}",
				Description: "Failed count",
				Type:        "int",
			},
			"active": {
				Name:        "active",
				JSONPath:    "{.status.active}",
				Description: "Active count",
				Type:        "int",
			},
			"parallelism": {
				Name:        "parallelism",
				JSONPath:    "{.spec.parallelism}",
				Description: "Parallelism",
				Type:        "int",
			},
			"backofflimit": {
				Name:        "backofflimit",
				JSONPath:    "{.spec.backoffLimit}",
				Description: "Backoff limit",
				Type:        "int",
			},
			"image": {
				Name:        "image",
				JSONPath:    "{.spec.template.spec.containers[*].image}",
				Description: "Container images",
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
