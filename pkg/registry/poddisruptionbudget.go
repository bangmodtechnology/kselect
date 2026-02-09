package registry

import "k8s.io/apimachinery/pkg/runtime/schema"

func init() {
	GetGlobalRegistry().Register(&ResourceDefinition{
		Name:    "poddisruptionbudget",
		Aliases: []string{"poddisruptionbudgets", "pdb", "pdbs"},
		GroupVersionResource: schema.GroupVersionResource{
			Group:    "policy",
			Version:  "v1",
			Resource: "poddisruptionbudgets",
		},
		Namespaced:    true,
		DefaultFields: []string{"name", "min-available", "max-unavailable", "current-healthy", "desired-healthy", "age"},
		Fields: map[string]FieldDefinition{
			"name": {
				Name:        "name",
				JSONPath:    "{.metadata.name}",
				Description: "PodDisruptionBudget name",
				Type:        "string",
			},
			"namespace": {
				Name:        "namespace",
				Aliases:     []string{"ns"},
				JSONPath:    "{.metadata.namespace}",
				Description: "Namespace",
				Type:        "string",
			},
			"min-available": {
				Name:        "min-available",
				JSONPath:    "{.spec.minAvailable}",
				Description: "Minimum available pods",
				Type:        "string",
			},
			"max-unavailable": {
				Name:        "max-unavailable",
				JSONPath:    "{.spec.maxUnavailable}",
				Description: "Maximum unavailable pods",
				Type:        "string",
			},
			"selector": {
				Name:        "selector",
				JSONPath:    "{.spec.selector.matchLabels}",
				Description: "Pod selector",
				Type:        "map",
			},
			"current-healthy": {
				Name:        "current-healthy",
				JSONPath:    "{.status.currentHealthy}",
				Description: "Current healthy pods",
				Type:        "int",
			},
			"desired-healthy": {
				Name:        "desired-healthy",
				JSONPath:    "{.status.desiredHealthy}",
				Description: "Desired healthy pods",
				Type:        "int",
			},
			"disruptions-allowed": {
				Name:        "disruptions-allowed",
				JSONPath:    "{.status.disruptionsAllowed}",
				Description: "Disruptions allowed",
				Type:        "int",
			},
			"expected-pods": {
				Name:        "expected-pods",
				JSONPath:    "{.status.expectedPods}",
				Description: "Expected pods",
				Type:        "int",
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
