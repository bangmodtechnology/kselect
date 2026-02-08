package registry

import "k8s.io/apimachinery/pkg/runtime/schema"

func init() {
	GetGlobalRegistry().Register(&ResourceDefinition{
		Name:    "gateway",
		Aliases: []string{"gateways", "gw"},
		GroupVersionResource: schema.GroupVersionResource{
			Group:    "gateway.networking.k8s.io",
			Version:  "v1",
			Resource: "gateways",
		},
		Namespaced:    true,
		DefaultFields: []string{"name", "class", "addresses", "programmed", "age"},
		Fields: map[string]FieldDefinition{
			"name": {
				Name:        "name",
				JSONPath:    "{.metadata.name}",
				Description: "Gateway name",
				Type:        "string",
			},
			"namespace": {
				Name:        "namespace",
				Aliases:     []string{"ns"},
				JSONPath:    "{.metadata.namespace}",
				Description: "Namespace",
				Type:        "string",
			},
			"class": {
				Name:        "class",
				JSONPath:    "{.spec.gatewayClassName}",
				Description: "Gateway class",
				Type:        "string",
			},
			"addresses": {
				Name:        "addresses",
				JSONPath:    "{.status.addresses[*].value}",
				Description: "Gateway addresses",
				Type:        "list",
			},
			"listeners": {
				Name:        "listeners",
				JSONPath:    "{.spec.listeners[*].name}",
				Description: "Listener names",
				Type:        "list",
			},
			"programmed": {
				Name:        "programmed",
				JSONPath:    "{.status.conditions[*].status}",
				Description: "Programmed status",
				Type:        "string",
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
