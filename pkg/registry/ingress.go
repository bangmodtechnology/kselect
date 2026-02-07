package registry

import "k8s.io/apimachinery/pkg/runtime/schema"

func init() {
	GetGlobalRegistry().Register(&ResourceDefinition{
		Name:    "ingress",
		Aliases: []string{"ingresses", "ing"},
		GroupVersionResource: schema.GroupVersionResource{
			Group:    "networking.k8s.io",
			Version:  "v1",
			Resource: "ingresses",
		},
		Fields: map[string]FieldDefinition{
			"name": {
				Name:        "name",
				JSONPath:    "{.metadata.name}",
				Description: "Ingress name",
				Type:        "string",
			},
			"namespace": {
				Name:        "namespace",
				JSONPath:    "{.metadata.namespace}",
				Description: "Namespace",
				Type:        "string",
			},
			"class": {
				Name:        "class",
				JSONPath:    "{.spec.ingressClassName}",
				Description: "Ingress class",
				Type:        "string",
			},
			"host": {
				Name:        "host",
				JSONPath:    "{.spec.rules[*].host}",
				Description: "Hostnames",
				Type:        "list",
			},
			"address": {
				Name:        "address",
				JSONPath:    "{.status.loadBalancer.ingress[*].ip}",
				Description: "Load balancer address",
				Type:        "list",
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
