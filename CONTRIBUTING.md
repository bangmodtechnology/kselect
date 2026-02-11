# Contributing to kselect

Thank you for your interest in contributing to kselect! This guide will help you get started.

## 🚀 Quick Start

1. **Fork and clone the repository**
   ```bash
   git clone https://github.com/yourusername/kselect.git
   cd kselect
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Build and test**
   ```bash
   make build
   make test
   ```

4. **Run locally**
   ```bash
   ./kselect name,status FROM pod
   ```

## 🏗️ Project Structure

```
kselect/
├── cmd/kselect/          # Main entry point
│   └── main.go           # CLI setup, flag parsing
├── pkg/
│   ├── parser/           # SQL-like query parser
│   │   ├── query.go      # Query struct and parsing logic
│   │   └── condition.go  # WHERE clause parsing
│   ├── registry/         # Resource definitions
│   │   ├── registry.go   # Resource registry
│   │   ├── pod.go        # Pod resource definition
│   │   └── ...           # Other resources
│   ├── executor/         # K8s API interaction
│   │   ├── executor.go   # Query execution
│   │   ├── join.go       # JOIN implementation
│   │   ├── aggregate.go  # GROUP BY/aggregations
│   │   └── watch.go      # Watch mode
│   ├── output/           # Output formatters
│   │   ├── formatter.go  # Format interface
│   │   ├── table.go      # Table format
│   │   └── color.go      # ANSI color support
│   ├── validator/        # Query validation
│   │   ├── validator.go  # Validation logic
│   │   └── fuzzy.go      # Fuzzy matching
│   ├── completion/       # Shell completion
│   ├── describe/         # DESCRIBE command
│   └── repl/             # Interactive REPL
├── test/                 # Integration tests
└── plugins/              # YAML resource plugins (optional)
```

## 📝 How to Contribute

### Adding a New Built-in Resource

1. **Create resource file** in `pkg/registry/`

   Example: `pkg/registry/configmap.go`
   ```go
   package registry

   import (
       "k8s.io/apimachinery/pkg/runtime/schema"
   )

   func init() {
       GetGlobalRegistry().Register(&ResourceDefinition{
           Name:    "configmap",
           Aliases: []string{"configmaps", "cm"},
           GroupVersionResource: schema.GroupVersionResource{
               Group:    "",
               Version:  "v1",
               Resource: "configmaps",
           },
           Namespaced:    true,
           DefaultFields: []string{"name", "namespace", "age"},
           Fields: map[string]FieldDefinition{
               "name": {
                   Name:        "name",
                   JSONPath:    "{.metadata.name}",
                   Description: "ConfigMap name",
                   Type:        "string",
               },
               "namespace": {
                   Name:        "namespace",
                   Aliases:     []string{"ns"},
                   JSONPath:    "{.metadata.namespace}",
                   Description: "Namespace",
                   Type:        "string",
               },
               // Add more fields...
           },
       })
   }
   ```

2. **Import the resource** in `pkg/registry/registry.go`
   ```go
   import (
       _ "github.com/bangmodtechnology/kselect/pkg/registry"
   )
   ```

3. **Add tests** in `pkg/registry/registry_test.go`

4. **Update documentation** in README.md

### Adding a New Output Format

1. **Create formatter** in `pkg/output/`

   Example: `pkg/output/markdown.go`
   ```go
   package output

   import "fmt"

   type MarkdownFormatter struct{}

   func (f *MarkdownFormatter) Print(results []map[string]interface{}, fields []string) error {
       // Print header
       fmt.Print("| ")
       for _, field := range fields {
           fmt.Printf("%s | ", field)
       }
       fmt.Println()

       // Print separator
       fmt.Print("| ")
       for range fields {
           fmt.Print("--- | ")
       }
       fmt.Println()

       // Print rows
       for _, row := range results {
           fmt.Print("| ")
           for _, field := range fields {
               fmt.Printf("%v | ", row[field])
           }
           fmt.Println()
       }
       return nil
   }
   ```

2. **Register format** in `pkg/output/formatter.go`
   ```go
   const (
       FormatMarkdown Format = "markdown"
   )

   func NewFormatter(format Format) Formatter {
       switch format {
       case FormatMarkdown:
           return &MarkdownFormatter{}
       // ...
       }
   }
   ```

3. **Add tests** in `pkg/output/formatter_test.go`

### Creating a YAML Plugin (No Code Changes!)

1. **Create YAML file** in `plugins/` directory

   Example: `plugins/certificate.yaml`
   ```yaml
   name: certificate
   aliases:
     - certificates
     - cert
   group: cert-manager.io
   version: v1
   resource: certificates
   namespaced: true
   defaultFields:
     - name
     - ready
     - secret
     - issuer
   fields:
     - name: name
       jsonPath: "{.metadata.name}"
       description: "Certificate name"
       type: string

     - name: ready
       jsonPath: "{.status.conditions[?(@.type=='Ready')].status}"
       description: "Ready status"
       type: string

     - name: secret
       jsonPath: "{.spec.secretName}"
       description: "Secret name"
       type: string

     - name: issuer
       jsonPath: "{.spec.issuerRef.name}"
       description: "Issuer name"
       type: string
   ```

2. **Use the plugin**
   ```bash
   kselect name,ready FROM certificate --plugins ./plugins/
   ```

## 🧪 Testing

### Unit Tests

```bash
# Run all tests
make test

# Run tests for a specific package
go test -v ./pkg/parser/
go test -v ./pkg/validator/

# Run with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Integration Tests

```bash
go test -v ./test/
```

### Manual Testing

```bash
# Build and test locally
make build
./kselect --dry-run name,status FROM pod

# Test with real cluster
./kselect name,status FROM pod -n default
./kselect --interactive
```

## 📋 Pull Request Guidelines

1. **Create a feature branch**
   ```bash
   git checkout -b feature/my-new-feature
   ```

2. **Make your changes**
   - Write clean, readable code
   - Follow existing code style
   - Add tests for new features
   - Update documentation if needed

3. **Run tests**
   ```bash
   make test
   ```

4. **Commit with clear messages**
   ```bash
   git add .
   git commit -m "feat: add support for CronJob resource"
   ```

   Commit message format:
   - `feat:` new feature
   - `fix:` bug fix
   - `docs:` documentation changes
   - `test:` test additions/changes
   - `refactor:` code refactoring
   - `chore:` maintenance tasks

5. **Push and create PR**
   ```bash
   git push origin feature/my-new-feature
   ```
   Then create a Pull Request on GitHub.

## 🎨 Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Use meaningful variable names
- Add comments for complex logic
- Keep functions small and focused
- Prefer composition over inheritance

## 🐛 Reporting Bugs

1. **Check existing issues** first
2. **Create a new issue** with:
   - Clear title and description
   - Steps to reproduce
   - Expected vs actual behavior
   - kselect version (`kselect --version`)
   - Kubernetes version
   - Operating system

## 💡 Feature Requests

1. **Check ROADMAP.md** to see if it's already planned
2. **Create an issue** with:
   - Clear use case
   - Example syntax/usage
   - Why it's useful

## 🔍 Code Review Process

1. Maintainers will review your PR
2. Address feedback and update PR
3. Once approved, it will be merged
4. Your contribution will be in the next release!

## 📚 Resources

- [ROADMAP.md](ROADMAP.md) - Planned features
- [CLAUDE.md](CLAUDE.md) - Architecture guide
- [Go Documentation](https://golang.org/doc/)
- [Kubernetes Client Go](https://github.com/kubernetes/client-go)

## 📞 Questions?

- Open a [GitHub Discussion](https://github.com/bangmodtechnology/kselect/discussions)
- Create an [Issue](https://github.com/bangmodtechnology/kselect/issues)

---

Thank you for contributing! 🙏
