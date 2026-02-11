package describe

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bangmodtechnology/kselect/pkg/registry"
)

// Resource prints detailed information about a resource
func Resource(reg *registry.Registry, resourceName string) error {
	resource, ok := reg.Get(resourceName)
	if !ok {
		return fmt.Errorf("resource '%s' not found", resourceName)
	}

	// Header
	scope := "namespaced"
	if !resource.Namespaced {
		scope = "cluster-scoped"
	}

	fmt.Printf("Resource: %s\n", resource.Name)
	fmt.Printf("Scope:    %s\n", scope)

	if len(resource.Aliases) > 0 {
		fmt.Printf("Aliases:  %s\n", strings.Join(resource.Aliases, ", "))
	}

	fmt.Printf("Group:    %s\n", resource.GroupVersionResource.Group)
	fmt.Printf("Version:  %s\n", resource.GroupVersionResource.Version)
	fmt.Println()

	// Default fields
	if len(resource.DefaultFields) > 0 {
		fmt.Printf("Default Fields: %s\n", strings.Join(resource.DefaultFields, ", "))
		fmt.Println()
	}

	// All fields
	fmt.Println("All Fields:")
	fmt.Println(strings.Repeat("─", 80))
	fmt.Printf("%-20s %-10s %s\n", "FIELD", "TYPE", "DESCRIPTION")
	fmt.Println(strings.Repeat("─", 80))

	// Sort fields by name for consistent output
	var fieldNames []string
	for name := range resource.Fields {
		fieldNames = append(fieldNames, name)
	}
	sort.Strings(fieldNames)

	for _, fieldName := range fieldNames {
		fieldDef := resource.Fields[fieldName]

		// Build field display with aliases
		fieldDisplay := fieldName
		if len(fieldDef.Aliases) > 0 {
			fieldDisplay += fmt.Sprintf(" (%s)", strings.Join(fieldDef.Aliases, ", "))
		}

		// Truncate description if too long
		description := fieldDef.Description
		if len(description) > 45 {
			description = description[:42] + "..."
		}

		fmt.Printf("%-20s %-10s %s\n", fieldDisplay, fieldDef.Type, description)
	}

	fmt.Println(strings.Repeat("─", 80))
	fmt.Printf("Total: %d fields\n", len(resource.Fields))

	return nil
}

// AllResources prints a summary of all available resources
func AllResources(reg *registry.Registry) {
	resources := reg.ListResources()

	fmt.Println("Available Resources:")
	fmt.Println(strings.Repeat("─", 80))
	fmt.Printf("%-20s %-15s %s\n", "RESOURCE", "SCOPE", "ALIASES")
	fmt.Println(strings.Repeat("─", 80))

	// Sort by name
	sort.Slice(resources, func(i, j int) bool {
		return resources[i].Name < resources[j].Name
	})

	for _, res := range resources {
		scope := "namespaced"
		if !res.Namespaced {
			scope = "cluster-scoped"
		}

		aliases := ""
		if len(res.Aliases) > 0 {
			aliases = strings.Join(res.Aliases, ", ")
			if len(aliases) > 40 {
				aliases = aliases[:37] + "..."
			}
		}

		fmt.Printf("%-20s %-15s %s\n", res.Name, scope, aliases)
	}

	fmt.Println(strings.Repeat("─", 80))
	fmt.Printf("Total: %d resources\n", len(resources))
	fmt.Println()
	fmt.Println("Use 'kselect --describe <resource>' or 'DESCRIBE <resource>' for detailed information")
}
