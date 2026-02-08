package registry

import "k8s.io/apimachinery/pkg/runtime/schema"

func init() {
	GetGlobalRegistry().Register(&ResourceDefinition{
		Name:    "node",
		Aliases: []string{"nodes", "no"},
		GroupVersionResource: schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "nodes",
		},
		DefaultFields: []string{"name", "status", "roles", "version", "internal-ip", "age"},
		Fields: map[string]FieldDefinition{
			"name": {
				Name:        "name",
				JSONPath:    "{.metadata.name}",
				Description: "Node name",
				Type:        "string",
			},
			"status": {
				Name:        "status",
				JSONPath:    "{.status.conditions[*].type}",
				Description: "Node status",
				Type:        "string",
			},
			"roles": {
				Name:        "roles",
				JSONPath:    "{.metadata.labels}",
				Description: "Node roles",
				Type:        "string",
			},
			"version": {
				Name:        "version",
				JSONPath:    "{.status.nodeInfo.kubeletVersion}",
				Description: "Kubelet version",
				Type:        "string",
			},
			"internal-ip": {
				Name:        "internal-ip",
				JSONPath:    "{.status.addresses[*].address}",
				Description: "Internal IP address",
				Type:        "string",
			},
			"external-ip": {
				Name:        "external-ip",
				JSONPath:    "{.status.addresses[*].address}",
				Description: "External IP address",
				Type:        "string",
			},
			"os": {
				Name:        "os",
				JSONPath:    "{.status.nodeInfo.osImage}",
				Description: "OS image",
				Type:        "string",
			},
			"kernel": {
				Name:        "kernel",
				JSONPath:    "{.status.nodeInfo.kernelVersion}",
				Description: "Kernel version",
				Type:        "string",
			},
			"container-runtime": {
				Name:        "container-runtime",
				JSONPath:    "{.status.nodeInfo.containerRuntimeVersion}",
				Description: "Container runtime version",
				Type:        "string",
			},
			"cpu": {
				Name:        "cpu",
				JSONPath:    "{.status.capacity.cpu}",
				Description: "CPU capacity",
				Type:        "string",
			},
			"memory": {
				Name:        "memory",
				JSONPath:    "{.status.capacity.memory}",
				Description: "Memory capacity",
				Type:        "string",
			},
			"pods": {
				Name:        "pods",
				JSONPath:    "{.status.capacity.pods}",
				Description: "Pod capacity",
				Type:        "string",
			},
			"arch": {
				Name:        "arch",
				JSONPath:    "{.status.nodeInfo.architecture}",
				Description: "Architecture",
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
