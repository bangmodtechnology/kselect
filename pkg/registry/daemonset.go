package registry

import "k8s.io/apimachinery/pkg/runtime/schema"

func init() {
	GetGlobalRegistry().Register(&ResourceDefinition{
		Name:    "daemonset",
		Aliases: []string{"daemonsets", "ds"},
		GroupVersionResource: schema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: "daemonsets",
		},
		Namespaced:    true,
		DefaultFields: []string{"name", "desired", "current", "ready", "available", "age"},
		Fields: map[string]FieldDefinition{
			"name": {
				Name:        "name",
				JSONPath:    "{.metadata.name}",
				Description: "DaemonSet name",
				Type:        "string",
			},
			"namespace": {
				Name:        "namespace",
				Aliases:     []string{"ns"},
				JSONPath:    "{.metadata.namespace}",
				Description: "Namespace",
				Type:        "string",
			},
			"desired": {
				Name:        "desired",
				JSONPath:    "{.status.desiredNumberScheduled}",
				Description: "Desired number scheduled",
				Type:        "int",
			},
			"current": {
				Name:        "current",
				JSONPath:    "{.status.currentNumberScheduled}",
				Description: "Current number scheduled",
				Type:        "int",
			},
			"ready": {
				Name:        "ready",
				JSONPath:    "{.status.numberReady}",
				Description: "Ready",
				Type:        "int",
			},
			"available": {
				Name:        "available",
				JSONPath:    "{.status.numberAvailable}",
				Description: "Available",
				Type:        "int",
			},
			"updated": {
				Name:        "updated",
				JSONPath:    "{.status.updatedNumberScheduled}",
				Description: "Updated",
				Type:        "int",
			},
			"misscheduled": {
				Name:        "misscheduled",
				JSONPath:    "{.status.numberMisscheduled}",
				Description: "Misscheduled",
				Type:        "int",
			},
			"image": {
				Name:        "image",
				JSONPath:    "{.spec.template.spec.containers[*].image}",
				Description: "Container images",
				Type:        "list",
			},
			"selector": {
				Name:        "selector",
				JSONPath:    "{.spec.selector.matchLabels}",
				Description: "Label selector",
				Type:        "map",
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
