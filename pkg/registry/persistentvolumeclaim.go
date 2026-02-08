package registry

import "k8s.io/apimachinery/pkg/runtime/schema"

func init() {
	GetGlobalRegistry().Register(&ResourceDefinition{
		Name:    "persistentvolumeclaim",
		Aliases: []string{"persistentvolumeclaims", "pvc"},
		GroupVersionResource: schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "persistentvolumeclaims",
		},
		Namespaced:    true,
		DefaultFields: []string{"name", "status", "volume", "capacity", "access-modes", "storageclass", "age"},
		Fields: map[string]FieldDefinition{
			"name": {
				Name:        "name",
				JSONPath:    "{.metadata.name}",
				Description: "PVC name",
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
				Description: "PVC status",
				Type:        "string",
			},
			"volume": {
				Name:        "volume",
				JSONPath:    "{.spec.volumeName}",
				Description: "Bound PV name",
				Type:        "string",
			},
			"capacity": {
				Name:        "capacity",
				JSONPath:    "{.status.capacity.storage}",
				Description: "Actual capacity",
				Type:        "string",
			},
			"request": {
				Name:        "request",
				JSONPath:    "{.spec.resources.requests.storage}",
				Description: "Requested storage",
				Type:        "string",
			},
			"access-modes": {
				Name:        "access-modes",
				JSONPath:    "{.status.accessModes}",
				Description: "Access modes",
				Type:        "list",
			},
			"storageclass": {
				Name:        "storageclass",
				Aliases:     []string{"sc"},
				JSONPath:    "{.spec.storageClassName}",
				Description: "Storage class",
				Type:        "string",
			},
			"volume-mode": {
				Name:        "volume-mode",
				JSONPath:    "{.spec.volumeMode}",
				Description: "Volume mode",
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
