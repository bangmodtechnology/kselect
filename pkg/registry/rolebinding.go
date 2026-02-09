package registry

import "k8s.io/apimachinery/pkg/runtime/schema"

func init() {
	GetGlobalRegistry().Register(&ResourceDefinition{
		Name:    "rolebinding",
		Aliases: []string{"rolebindings"},
		GroupVersionResource: schema.GroupVersionResource{
			Group:    "rbac.authorization.k8s.io",
			Version:  "v1",
			Resource: "rolebindings",
		},
		Namespaced:    true,
		DefaultFields: []string{"name", "role-ref", "subjects", "age"},
		Fields: map[string]FieldDefinition{
			"name": {
				Name:        "name",
				JSONPath:    "{.metadata.name}",
				Description: "RoleBinding name",
				Type:        "string",
			},
			"namespace": {
				Name:        "namespace",
				Aliases:     []string{"ns"},
				JSONPath:    "{.metadata.namespace}",
				Description: "Namespace",
				Type:        "string",
			},
			"role-ref": {
				Name:        "role-ref",
				JSONPath:    "{.roleRef.name}",
				Description: "Referenced role name",
				Type:        "string",
			},
			"role-kind": {
				Name:        "role-kind",
				JSONPath:    "{.roleRef.kind}",
				Description: "Referenced role kind (Role/ClusterRole)",
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
