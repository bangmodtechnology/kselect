package registry

import "k8s.io/apimachinery/pkg/runtime/schema"

func init() {
	GetGlobalRegistry().Register(&ResourceDefinition{
		Name:    "hpa",
		Aliases: []string{"horizontalpodautoscaler", "horizontalpodautoscalers"},
		GroupVersionResource: schema.GroupVersionResource{
			Group:    "autoscaling",
			Version:  "v2",
			Resource: "horizontalpodautoscalers",
		},
		Namespaced:    true,
		DefaultFields: []string{"name", "reference", "min", "max", "replicas", "age"},
		Fields: map[string]FieldDefinition{
			"name": {
				Name:        "name",
				JSONPath:    "{.metadata.name}",
				Description: "HPA name",
				Type:        "string",
			},
			"namespace": {
				Name:        "namespace",
				Aliases:     []string{"ns"},
				JSONPath:    "{.metadata.namespace}",
				Description: "Namespace",
				Type:        "string",
			},
			"reference": {
				Name:        "reference",
				Aliases:     []string{"ref", "target"},
				JSONPath:    "{.spec.scaleTargetRef.name}",
				Description: "Scale target reference",
				Type:        "string",
			},
			"ref-kind": {
				Name:        "ref-kind",
				JSONPath:    "{.spec.scaleTargetRef.kind}",
				Description: "Scale target kind",
				Type:        "string",
			},
			"min": {
				Name:        "min",
				Aliases:     []string{"minreplicas"},
				JSONPath:    "{.spec.minReplicas}",
				Description: "Min replicas",
				Type:        "int",
			},
			"max": {
				Name:        "max",
				Aliases:     []string{"maxreplicas"},
				JSONPath:    "{.spec.maxReplicas}",
				Description: "Max replicas",
				Type:        "int",
			},
			"replicas": {
				Name:        "replicas",
				JSONPath:    "{.status.currentReplicas}",
				Description: "Current replicas",
				Type:        "int",
			},
			"desired": {
				Name:        "desired",
				JSONPath:    "{.status.desiredReplicas}",
				Description: "Desired replicas",
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
