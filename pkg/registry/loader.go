package registry

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type pluginDefinition struct {
	Name          string                    `yaml:"name"`
	Aliases       []string                  `yaml:"aliases"`
	Group         string                    `yaml:"group"`
	Version       string                    `yaml:"version"`
	Resource      string                    `yaml:"resource"`
	Namespaced    *bool                     `yaml:"namespaced"` // default true
	DefaultFields []string                  `yaml:"default_fields"`
	Fields        map[string]pluginFieldDef `yaml:"fields"`
}

type pluginFieldDef struct {
	JSONPath    string   `yaml:"jsonpath"`
	Type        string   `yaml:"type"`
	Description string   `yaml:"description"`
	Aliases     []string `yaml:"aliases"`
}

func LoadPlugins(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return fmt.Errorf("failed to glob plugin dir: %w", err)
	}

	ymlFiles, err := filepath.Glob(filepath.Join(dir, "*.yml"))
	if err != nil {
		return fmt.Errorf("failed to glob plugin dir: %w", err)
	}
	files = append(files, ymlFiles...)

	for _, file := range files {
		if err := loadPlugin(file); err != nil {
			return fmt.Errorf("failed to load plugin %s: %w", file, err)
		}
	}
	return nil
}

func loadPlugin(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var plugin pluginDefinition
	if err := yaml.Unmarshal(data, &plugin); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	if plugin.Name == "" || plugin.Resource == "" {
		return fmt.Errorf("plugin must have name and resource fields")
	}

	fields := make(map[string]FieldDefinition)
	for name, f := range plugin.Fields {
		fields[name] = FieldDefinition{
			Name:        name,
			Aliases:     f.Aliases,
			JSONPath:    f.JSONPath,
			Description: f.Description,
			Type:        f.Type,
		}
	}

	namespaced := true
	if plugin.Namespaced != nil {
		namespaced = *plugin.Namespaced
	}

	def := &ResourceDefinition{
		Name:    plugin.Name,
		Aliases: plugin.Aliases,
		GroupVersionResource: schema.GroupVersionResource{
			Group:    plugin.Group,
			Version:  plugin.Version,
			Resource: plugin.Resource,
		},
		Namespaced:    namespaced,
		DefaultFields: plugin.DefaultFields,
		Fields:        fields,
	}

	GetGlobalRegistry().Register(def)
	return nil
}
