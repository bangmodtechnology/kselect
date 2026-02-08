package registry

import "k8s.io/apimachinery/pkg/runtime/schema"

func init() {
	GetGlobalRegistry().Register(&ResourceDefinition{
		Name:    "pod",
		Aliases: []string{"pods", "po"},
		GroupVersionResource: schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "pods",
		},
		Namespaced:    true,
		DefaultFields: []string{"name", "status", "ip", "node", "restarts", "age"},
		Fields: map[string]FieldDefinition{
			"name": {
				Name:        "name",
				JSONPath:    "{.metadata.name}",
				Description: "Pod name",
				Type:        "string",
			},
			"namespace": {
				Name:        "namespace",
				Aliases:     []string{"ns"},
				JSONPath:    "{.metadata.namespace}",
				Description: "Namespace",
				Type:        "string",
			},
			"status": {
				Name:        "status",
				JSONPath:    "{.status.phase}",
				Description: "Pod status",
				Type:        "string",
			},
			"ip": {
				Name:        "ip",
				JSONPath:    "{.status.podIP}",
				Description: "Pod IP address",
				Type:        "string",
			},
			"node": {
				Name:        "node",
				JSONPath:    "{.spec.nodeName}",
				Description: "Node name",
				Type:        "string",
			},
			"cpu.req": {
				Name:        "cpu.req",
				JSONPath:    "{.spec.containers[*].resources.requests.cpu}",
				Description: "CPU requests",
				Type:        "string",
			},
			"cpu.limit": {
				Name:        "cpu.limit",
				JSONPath:    "{.spec.containers[*].resources.limits.cpu}",
				Description: "CPU limits",
				Type:        "string",
			},
			"mem.req": {
				Name:        "mem.req",
				JSONPath:    "{.spec.containers[*].resources.requests.memory}",
				Description: "Memory requests",
				Type:        "string",
			},
			"mem.limit": {
				Name:        "mem.limit",
				JSONPath:    "{.spec.containers[*].resources.limits.memory}",
				Description: "Memory limits",
				Type:        "string",
			},
			"image": {
				Name:        "image",
				JSONPath:    "{.spec.containers[*].image}",
				Description: "Container images",
				Type:        "list",
			},
			"restarts": {
				Name:        "restarts",
				JSONPath:    "{.status.containerStatuses[*].restartCount}",
				Description: "Restart count",
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
