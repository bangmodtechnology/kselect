package registry

import "k8s.io/apimachinery/pkg/runtime/schema"

func init() {
	GetGlobalRegistry().Register(&ResourceDefinition{
		Name:    "deployment",
		Aliases: []string{"deployments", "deploy"},
		GroupVersionResource: schema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: "deployments",
		},
		Namespaced:    true,
		DefaultFields: []string{"name", "replicas", "ready", "available", "age"},
		Fields: map[string]FieldDefinition{
			"name": {
				Name:        "name",
				JSONPath:    "{.metadata.name}",
				Description: "Deployment name",
				Type:        "string",
			},
			"namespace": {
				Name:        "namespace",
				Aliases:     []string{"ns"},
				JSONPath:    "{.metadata.namespace}",
				Description: "Namespace",
				Type:        "string",
			},
			"replicas": {
				Name:        "replicas",
				JSONPath:    "{.spec.replicas}",
				Description: "Desired replicas",
				Type:        "int",
			},
			"ready": {
				Name:        "ready",
				JSONPath:    "{.status.readyReplicas}",
				Description: "Ready replicas",
				Type:        "int",
			},
			"available": {
				Name:        "available",
				JSONPath:    "{.status.availableReplicas}",
				Description: "Available replicas",
				Type:        "int",
			},
			"updated": {
				Name:        "updated",
				JSONPath:    "{.status.updatedReplicas}",
				Description: "Updated replicas",
				Type:        "int",
			},
			"image": {
				Name:        "image",
				JSONPath:    "{.spec.template.spec.containers[*].image}",
				Description: "Container images",
				Type:        "list",
			},
			"strategy": {
				Name:        "strategy",
				JSONPath:    "{.spec.strategy.type}",
				Description: "Deployment strategy",
				Type:        "string",
			},
			"cpu.req": {
				Name:        "cpu.req",
				JSONPath:    "{.spec.template.spec.containers[*].resources.requests.cpu}",
				Description: "CPU requests",
				Type:        "string",
			},
			"cpu.limit": {
				Name:        "cpu.limit",
				JSONPath:    "{.spec.template.spec.containers[*].resources.limits.cpu}",
				Description: "CPU limits",
				Type:        "string",
			},
			"mem.req": {
				Name:        "mem.req",
				JSONPath:    "{.spec.template.spec.containers[*].resources.requests.memory}",
				Description: "Memory requests",
				Type:        "string",
			},
			"mem.limit": {
				Name:        "mem.limit",
				JSONPath:    "{.spec.template.spec.containers[*].resources.limits.memory}",
				Description: "Memory limits",
				Type:        "string",
			},
			"cpu.req-m": {
				Name:        "cpu.req-m",
				JSONPath:    "{.spec.template.spec.containers[*].resources.requests.cpu}",
				Description: "CPU requests in millicores",
				Type:        "int",
			},
			"cpu.limit-m": {
				Name:        "cpu.limit-m",
				JSONPath:    "{.spec.template.spec.containers[*].resources.limits.cpu}",
				Description: "CPU limits in millicores",
				Type:        "int",
			},
			"mem.req-mi": {
				Name:        "mem.req-mi",
				JSONPath:    "{.spec.template.spec.containers[*].resources.requests.memory}",
				Description: "Memory requests in MiB",
				Type:        "int",
			},
			"mem.limit-mi": {
				Name:        "mem.limit-mi",
				JSONPath:    "{.spec.template.spec.containers[*].resources.limits.memory}",
				Description: "Memory limits in MiB",
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
