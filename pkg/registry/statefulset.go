package registry

import "k8s.io/apimachinery/pkg/runtime/schema"

func init() {
	GetGlobalRegistry().Register(&ResourceDefinition{
		Name:    "statefulset",
		Aliases: []string{"statefulsets", "sts"},
		GroupVersionResource: schema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: "statefulsets",
		},
		Namespaced:    true,
		DefaultFields: []string{"name", "replicas", "ready", "age"},
		Fields: map[string]FieldDefinition{
			"name": {
				Name:        "name",
				JSONPath:    "{.metadata.name}",
				Description: "StatefulSet name",
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
			"current": {
				Name:        "current",
				JSONPath:    "{.status.currentReplicas}",
				Description: "Current replicas",
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
			"servicename": {
				Name:        "servicename",
				JSONPath:    "{.spec.serviceName}",
				Description: "Headless service name",
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
