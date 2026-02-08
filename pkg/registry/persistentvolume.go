package registry

import "k8s.io/apimachinery/pkg/runtime/schema"

func init() {
	GetGlobalRegistry().Register(&ResourceDefinition{
		Name:    "persistentvolume",
		Aliases: []string{"persistentvolumes", "pv"},
		GroupVersionResource: schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "persistentvolumes",
		},
		DefaultFields: []string{"name", "capacity", "access-modes", "reclaim-policy", "status", "claim", "storageclass", "age"},
		Fields: map[string]FieldDefinition{
			"name": {
				Name:        "name",
				JSONPath:    "{.metadata.name}",
				Description: "PV name",
				Type:        "string",
			},
			"capacity": {
				Name:        "capacity",
				JSONPath:    "{.spec.capacity.storage}",
				Description: "Storage capacity",
				Type:        "string",
			},
			"access-modes": {
				Name:        "access-modes",
				JSONPath:    "{.spec.accessModes}",
				Description: "Access modes",
				Type:        "list",
			},
			"reclaim-policy": {
				Name:        "reclaim-policy",
				JSONPath:    "{.spec.persistentVolumeReclaimPolicy}",
				Description: "Reclaim policy",
				Type:        "string",
			},
			"status": {
				Name:        "status",
				JSONPath:    "{.status.phase}",
				Description: "PV status",
				Type:        "string",
			},
			"claim": {
				Name:        "claim",
				JSONPath:    "{.spec.claimRef.name}",
				Description: "Bound PVC name",
				Type:        "string",
			},
			"claim-ns": {
				Name:        "claim-ns",
				JSONPath:    "{.spec.claimRef.namespace}",
				Description: "Bound PVC namespace",
				Type:        "string",
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
