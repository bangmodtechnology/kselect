package registry

import "k8s.io/apimachinery/pkg/runtime/schema"

func init() {
	GetGlobalRegistry().Register(&ResourceDefinition{
		Name:    "networkpolicy",
		Aliases: []string{"networkpolicies", "netpol"},
		GroupVersionResource: schema.GroupVersionResource{
			Group:    "networking.k8s.io",
			Version:  "v1",
			Resource: "networkpolicies",
		},
		Namespaced:    true,
		DefaultFields: []string{"name", "pod-selector", "policy-types", "age"},
		Fields: map[string]FieldDefinition{
			"name": {
				Name:        "name",
				JSONPath:    "{.metadata.name}",
				Description: "NetworkPolicy name",
				Type:        "string",
			},
			"namespace": {
				Name:        "namespace",
				Aliases:     []string{"ns"},
				JSONPath:    "{.metadata.namespace}",
				Description: "Namespace",
				Type:        "string",
			},
			"pod-selector": {
				Name:        "pod-selector",
				JSONPath:    "{.spec.podSelector.matchLabels}",
				Description: "Pod selector labels",
				Type:        "map",
			},
			"policy-types": {
				Name:        "policy-types",
				JSONPath:    "{.spec.policyTypes}",
				Description: "Policy types (Ingress/Egress)",
				Type:        "list",
			},
			"ingress-rules": {
				Name:        "ingress-rules",
				JSONPath:    "{.spec.ingress}",
				Description: "Number of ingress rules",
				Type:        "list",
			},
			"egress-rules": {
				Name:        "egress-rules",
				JSONPath:    "{.spec.egress}",
				Description: "Number of egress rules",
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
