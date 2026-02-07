package registry

import "k8s.io/apimachinery/pkg/runtime/schema"

func init() {
	GetGlobalRegistry().Register(&ResourceDefinition{
		Name:    "service",
		Aliases: []string{"services", "svc"},
		GroupVersionResource: schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "services",
		},
		Fields: map[string]FieldDefinition{
			"name": {
				Name:        "name",
				JSONPath:    "{.metadata.name}",
				Description: "Service name",
				Type:        "string",
			},
			"namespace": {
				Name:        "namespace",
				JSONPath:    "{.metadata.namespace}",
				Description: "Namespace",
				Type:        "string",
			},
			"type": {
				Name:        "type",
				JSONPath:    "{.spec.type}",
				Description: "Service type",
				Type:        "string",
			},
			"cluster-ip": {
				Name:        "cluster-ip",
				JSONPath:    "{.spec.clusterIP}",
				Description: "Cluster IP",
				Type:        "string",
			},
			"external-ip": {
				Name:        "external-ip",
				JSONPath:    "{.status.loadBalancer.ingress[*].ip}",
				Description: "External IP",
				Type:        "list",
			},
			"port": {
				Name:        "port",
				JSONPath:    "{.spec.ports[*].port}",
				Description: "Service ports",
				Type:        "list",
			},
			"targetport": {
				Name:        "targetport",
				JSONPath:    "{.spec.ports[*].targetPort}",
				Description: "Target ports",
				Type:        "list",
			},
			"selector": {
				Name:        "selector",
				JSONPath:    "{.spec.selector}",
				Description: "Label selector",
				Type:        "map",
			},
			"age": {
				Name:        "age",
				JSONPath:    "{.metadata.creationTimestamp}",
				Description: "Age",
				Type:        "time",
			},
		},
	})
}
