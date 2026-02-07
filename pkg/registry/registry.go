package registry

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ResourceDefinition struct {
	Name                 string
	Aliases              []string
	GroupVersionResource schema.GroupVersionResource
	DefaultFields        []string // fields shown when user omits field list
	Fields               map[string]FieldDefinition
}

type FieldDefinition struct {
	Name        string
	Aliases     []string // short names, e.g. "ns" for "namespace"
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

// ResolveFieldAlias resolves a field alias (e.g. "ns") to its canonical name (e.g. "namespace").
// Returns the original name if no alias match is found.
func (d *ResourceDefinition) ResolveFieldAlias(alias string) string {
	// Check if it's already a canonical field name
	if _, ok := d.Fields[alias]; ok {
		return alias
	}
	// Search aliases
	for name, field := range d.Fields {
		for _, a := range field.Aliases {
			if a == alias {
				return name
			}
		}
	}
	return alias
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
