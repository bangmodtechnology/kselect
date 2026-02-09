package registry

import "k8s.io/apimachinery/pkg/runtime/schema"

func init() {
	GetGlobalRegistry().Register(&ResourceDefinition{
		Name:    "cronjob",
		Aliases: []string{"cronjobs", "cj"},
		GroupVersionResource: schema.GroupVersionResource{
			Group:    "batch",
			Version:  "v1",
			Resource: "cronjobs",
		},
		Namespaced:    true,
		DefaultFields: []string{"name", "schedule", "suspend", "active", "last-schedule", "age"},
		Fields: map[string]FieldDefinition{
			"name": {
				Name:        "name",
				JSONPath:    "{.metadata.name}",
				Description: "CronJob name",
				Type:        "string",
			},
			"namespace": {
				Name:        "namespace",
				Aliases:     []string{"ns"},
				JSONPath:    "{.metadata.namespace}",
				Description: "Namespace",
				Type:        "string",
			},
			"schedule": {
				Name:        "schedule",
				JSONPath:    "{.spec.schedule}",
				Description: "Cron schedule",
				Type:        "string",
			},
			"suspend": {
				Name:        "suspend",
				JSONPath:    "{.spec.suspend}",
				Description: "Suspended",
				Type:        "string",
			},
			"active": {
				Name:        "active",
				JSONPath:    "{.status.active}",
				Description: "Active jobs",
				Type:        "list",
			},
			"last-schedule": {
				Name:        "last-schedule",
				JSONPath:    "{.status.lastScheduleTime}",
				Description: "Last schedule time",
				Type:        "time",
			},
			"last-success": {
				Name:        "last-success",
				JSONPath:    "{.status.lastSuccessfulTime}",
				Description: "Last successful time",
				Type:        "time",
			},
			"concurrency": {
				Name:        "concurrency",
				JSONPath:    "{.spec.concurrencyPolicy}",
				Description: "Concurrency policy",
				Type:        "string",
			},
			"image": {
				Name:        "image",
				JSONPath:    "{.spec.jobTemplate.spec.template.spec.containers[*].image}",
				Description: "Container images",
				Type:        "list",
			},
			"cpu.req": {
				Name:        "cpu.req",
				JSONPath:    "{.spec.jobTemplate.spec.template.spec.containers[*].resources.requests.cpu}",
				Description: "CPU requests",
				Type:        "string",
			},
			"cpu.limit": {
				Name:        "cpu.limit",
				JSONPath:    "{.spec.jobTemplate.spec.template.spec.containers[*].resources.limits.cpu}",
				Description: "CPU limits",
				Type:        "string",
			},
			"mem.req": {
				Name:        "mem.req",
				JSONPath:    "{.spec.jobTemplate.spec.template.spec.containers[*].resources.requests.memory}",
				Description: "Memory requests",
				Type:        "string",
			},
			"mem.limit": {
				Name:        "mem.limit",
				JSONPath:    "{.spec.jobTemplate.spec.template.spec.containers[*].resources.limits.memory}",
				Description: "Memory limits",
				Type:        "string",
			},
			"cpu.req-m": {
				Name:        "cpu.req-m",
				JSONPath:    "{.spec.jobTemplate.spec.template.spec.containers[*].resources.requests.cpu}",
				Description: "CPU requests in millicores",
				Type:        "int",
			},
			"cpu.limit-m": {
				Name:        "cpu.limit-m",
				JSONPath:    "{.spec.jobTemplate.spec.template.spec.containers[*].resources.limits.cpu}",
				Description: "CPU limits in millicores",
				Type:        "int",
			},
			"mem.req-mi": {
				Name:        "mem.req-mi",
				JSONPath:    "{.spec.jobTemplate.spec.template.spec.containers[*].resources.requests.memory}",
				Description: "Memory requests in MiB",
				Type:        "int",
			},
			"mem.limit-mi": {
				Name:        "mem.limit-mi",
				JSONPath:    "{.spec.jobTemplate.spec.template.spec.containers[*].resources.limits.memory}",
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
