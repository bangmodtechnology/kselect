package registry

import "k8s.io/apimachinery/pkg/runtime/schema"

func init() {
	GetGlobalRegistry().Register(&ResourceDefinition{
		Name:    "clusterrolebinding",
		Aliases: []string{"clusterrolebindings"},
		GroupVersionResource: schema.GroupVersionResource{
			Group:    "rbac.authorization.k8s.io",
			Version:  "v1",
			Resource: "clusterrolebindings",
		},
		Namespaced:    false, // cluster-scoped
		DefaultFields: []string{"name", "role-ref", "subjects", "age"},
		Fields: map[string]FieldDefinition{
			"name": {
				Name:        "name",
				JSONPath:    "{.metadata.name}",
				Description: "ClusterRoleBinding name",
				Type:        "string",
			},
			"role-ref": {
				Name:        "role-ref",
				JSONPath:    "{.roleRef.name}",
				Description: "Referenced ClusterRole name",
				Type:        "string",
			},
			"role-kind": {
				Name:        "role-kind",
				JSONPath:    "{.roleRef.kind}",
				Description: "Referenced role kind",
				Type:        "string",
			},
			"subjects": {
				Name:        "subjects",
				JSONPath:    "{.subjects[*].name}",
				Description: "Subject names",
				Type:        "list",
			},
			"subject-kinds": {
				Name:        "subject-kinds",
				JSONPath:    "{.subjects[*].kind}",
				Description: "Subject kinds",
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
