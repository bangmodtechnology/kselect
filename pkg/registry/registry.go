package registry

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ResourceDefinition struct {
	Name                 string
	Aliases              []string
	GroupVersionResource schema.GroupVersionResource
	Fields               map[string]FieldDefinition
}

type FieldDefinition struct {
	Name        string
	JSONPath    string
	Description string
	Type        string // string, int, list, map, time
}

type Registry struct {
	resources map[string]*ResourceDefinition
}

var globalRegistry = NewRegistry()

func NewRegistry() *Registry {
	return &Registry{
		resources: make(map[string]*ResourceDefinition),
	}
}

func (r *Registry) Register(def *ResourceDefinition) {
	r.resources[def.Name] = def
	for _, alias := range def.Aliases {
		r.resources[alias] = def
	}
}

func (r *Registry) Get(name string) (*ResourceDefinition, bool) {
	def, ok := r.resources[name]
	return def, ok
}

func (r *Registry) ListResources() []*ResourceDefinition {
	seen := make(map[string]bool)
	var resources []*ResourceDefinition

	for _, def := range r.resources {
		if !seen[def.Name] {
			resources = append(resources, def)
			seen[def.Name] = true
		}
	}
	return resources
}

func GetGlobalRegistry() *Registry {
	return globalRegistry
}
