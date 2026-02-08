package registry

import "k8s.io/apimachinery/pkg/runtime/schema"

func init() {
	GetGlobalRegistry().Register(&ResourceDefinition{
		Name:    "event",
		Aliases: []string{"events", "ev"},
		GroupVersionResource: schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "events",
		},
		Namespaced:    true,
		DefaultFields: []string{"type", "reason", "object", "message", "age"},
		Fields: map[string]FieldDefinition{
			"name": {
				Name:        "name",
				JSONPath:    "{.metadata.name}",
				Description: "Event name",
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
				Description: "Event type (Normal/Warning)",
				Type:        "string",
			},
			"reason": {
				Name:        "reason",
				JSONPath:    "{.reason}",
				Description: "Event reason",
				Type:        "string",
			},
			"object": {
				Name:        "object",
				JSONPath:    "{.involvedObject.name}",
				Description: "Involved object name",
				Type:        "string",
			},
			"object-kind": {
				Name:        "object-kind",
				JSONPath:    "{.involvedObject.kind}",
				Description: "Involved object kind",
				Type:        "string",
			},
			"message": {
				Name:        "message",
				JSONPath:    "{.message}",
				Description: "Event message",
				Type:        "string",
			},
			"count": {
				Name:        "count",
				JSONPath:    "{.count}",
				Description: "Event count",
				Type:        "int",
			},
			"source": {
				Name:        "source",
				JSONPath:    "{.source.component}",
				Description: "Source component",
				Type:        "string",
			},
			"first-seen": {
				Name:        "first-seen",
				JSONPath:    "{.firstTimestamp}",
				Description: "First seen",
				Type:        "time",
			},
			"last-seen": {
				Name:        "last-seen",
				JSONPath:    "{.lastTimestamp}",
				Description: "Last seen",
				Type:        "time",
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
