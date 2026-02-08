package registry

import "k8s.io/apimachinery/pkg/runtime/schema"

func init() {
	GetGlobalRegistry().Register(&ResourceDefinition{
		Name:    "replicaset",
		Aliases: []string{"replicasets", "rs"},
		GroupVersionResource: schema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: "replicasets",
		},
		Namespaced:    true,
		DefaultFields: []string{"name", "desired", "current", "ready", "age"},
		Fields: map[string]FieldDefinition{
			"name": {
				Name:        "name",
				JSONPath:    "{.metadata.name}",
				Description: "ReplicaSet name",
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
				JSONPath:    "{.spec.replicas}",
				Description: "Desired replicas",
				Type:        "int",
			},
			"current": {
				Name:        "current",
				JSONPath:    "{.status.replicas}",
				Description: "Current replicas",
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
			"image": {
				Name:        "image",
				JSONPath:    "{.spec.template.spec.containers[*].image}",
				Description: "Container images",
				Type:        "list",
			},
			"owner": {
				Name:        "owner",
				JSONPath:    "{.metadata.ownerReferences[*].name}",
				Description: "Owner (Deployment)",
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
