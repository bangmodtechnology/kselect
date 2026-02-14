package registry

import (
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ResourceDefinition struct {
	Name                 string
	Aliases              []string
	GroupVersionResource schema.GroupVersionResource
	Namespaced           bool     // true = namespaced, false = cluster-scoped (e.g. node)
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
// Supports dot-notation for map fields (e.g. "labels.app" stays as-is if "labels" is a known field).
// Returns the original name if no alias match is found.
func (d *ResourceDefinition) ResolveFieldAlias(alias string) string {
	// Check if it's already a canonical field name
	if _, ok := d.Fields[alias]; ok {
		return alias
	}

	// Handle dot-notation: resolve the base part alias (e.g. "lbl.app" → "labels.app")
	if dotIdx := strings.IndexByte(alias, '.'); dotIdx != -1 {
		base := alias[:dotIdx]
		subKey := alias[dotIdx+1:]
		resolved := d.resolveBaseAlias(base)
		return resolved + "." + subKey
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

// resolveBaseAlias resolves just the base field name from aliases.
func (d *ResourceDefinition) resolveBaseAlias(base string) string {
	if _, ok := d.Fields[base]; ok {
		return base
	}
	for name, field := range d.Fields {
		for _, a := range field.Aliases {
			if a == base {
				return name
			}
		}
	}
	return base
}

// IsMapField checks if a field name (possibly with dot-notation) references a map-type field.
// Returns the base field name and sub-key if it's a map sub-field access.
func (d *ResourceDefinition) IsMapSubField(fieldName string) (baseName, subKey string, ok bool) {
	dotIdx := strings.IndexByte(fieldName, '.')
	if dotIdx == -1 {
		return "", "", false
	}
	baseName = fieldName[:dotIdx]
	subKey = fieldName[dotIdx+1:]
	if fd, exists := d.Fields[baseName]; exists && fd.Type == "map" {
		return baseName, subKey, true
	}
	return "", "", false
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
