# kselect

SQL-like query tool for Kubernetes resources.

```bash
kselect name,status,ip FROM pod -n default
```

## Why kselect?

If you've ever used `kubectl get pods` and had to pipe through `grep`, `awk`, `jq` multiple times to find the data you need — kselect lets you do the same thing in a single command.

**The problem:**

```bash
# Traditional: find pods with high restarts in production
kubectl get pods -n production -o json | jq -r '.items[] | select(.status.containerStatuses[0].restartCount > 5) | [.metadata.name, .status.containerStatuses[0].restartCount] | @tsv' | sort -k2 -rn | head -5

# With kselect: readable and fast
kselect name,restarts FROM pod WHERE namespace=production AND restarts > 5 ORDER BY restarts DESC LIMIT 5
```

**Key concepts:**
- Familiar SQL-like syntax — no need to memorize JSONPath or jq syntax
- Select only the fields you need
- Filter, Sort, Aggregate in a single command
- JOIN across resource types (e.g. pod with service)
- Extend with new resources via YAML plugins without modifying code

## ✨ Features

### 🔍 **Query Validation**
Validate queries before execution with smart suggestions for typos:

```bash
kselect --dry-run nam FROM pod
# Error: Field 'nam' not found in resource 'pod'
# Did you mean: name?

kselect --dry-run name FROM pode
# Error: Resource 'pode' not found
# Did you mean: pod, node, or pods?
```

### 💬 **Interactive REPL Mode**
Query interactively without repeating `kselect`:

```bash
kselect --interactive
kselect> name,status FROM pod WHERE ns=default
kselect> DESCRIBE pod
kselect> namespace,COUNT as total FROM pod GROUP BY namespace
kselect> \exit
```

### 📋 **Resource Discovery**
Explore available resources and their fields:

```bash
# Describe a resource schema
kselect --describe pod
kselect DESCRIBE deployment

# List all available resources
kselect --list
```

### 🎨 **Rich Output**
- **Color-coded status fields** (Running=green, Pending=yellow, Failed=red)
- **Multiple formats:** table, json, yaml, csv, wide
- **Auto-detects TTY** for appropriate output
- **Watch mode** for real-time monitoring

### 🚀 **Advanced SQL Features**
- **Subqueries:** `WHERE name IN kselect name FROM deployment`
- **JOINs:** INNER, LEFT, RIGHT JOIN across resources
- **Aggregations:** COUNT, SUM, AVG, MIN, MAX with GROUP BY
- **HAVING clause:** Filter aggregated results
- **DISTINCT:** Remove duplicate rows
- **Field aliases:** Use `ns` for `namespace`, etc.

### 🧪 **Production Ready**
- 78%+ test coverage
- 66 integration test cases
- Comprehensive query validation
- Fuzzy matching for error suggestions

## How It Works

```
kselect name,status FROM pod WHERE namespace=default AND status!=Running ORDER BY name LIMIT 10
       ├──────────┘      └──┘     └──────────────────────────────────┘         └───────┘    └───┘
       fields          resource              conditions                       sorting     limit
```

kselect automatically translates your query into Kubernetes API calls:

1. **Parse** — Break the query into fields, resource, and conditions
2. **Registry** — Look up the resource definition (GVR + field-to-JSONPath mapping)
3. **Execute** — Call the K8s API via dynamic client to fetch resources
4. **Filter** — Apply WHERE conditions client-side
5. **Transform** — JOIN, Aggregate, Sort, Paginate
6. **Output** — Display results in the chosen format (table, json, yaml, csv)

## Quick Start

```bash
# List supported resources and fields
kselect --list

# Describe a resource to see available fields
kselect --describe pod

# Your first query (uses current kube context namespace)
kselect name,status,ip FROM pod

# Specify namespace
kselect name,status FROM pod -n production

# All namespaces
kselect name,namespace,status FROM pod -A

# Validate query before running
kselect --dry-run name,status FROM pod WHERE ns=default

# Interactive mode (REPL)
kselect --interactive

# Export as JSON
kselect name,status FROM pod -o json

# Real-time monitoring
kselect name,status,restarts FROM pod --watch
```

## Installation

### Homebrew (macOS and Linux)

```bash
# Add tap
brew tap bangmodtech/tap

# Install
brew install kselect

# Verify installation
kselect --version
```

### Install Script (One-liner)

```bash
curl -sSL https://raw.githubusercontent.com/bangmodtechnology/kselect/master/install.sh | sh
```

### Download Binary

Download pre-built binaries from [GitHub Releases](https://github.com/bangmodtechnology/kselect/releases/latest):

```bash
# macOS (Apple Silicon)
curl -L https://github.com/bangmodtechnology/kselect/releases/latest/download/kselect-darwin-arm64.tar.gz | tar xz
sudo mv kselect /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/bangmodtechnology/kselect/releases/latest/download/kselect-darwin-amd64.tar.gz | tar xz
sudo mv kselect /usr/local/bin/

# Linux (x86_64)
curl -L https://github.com/bangmodtechnology/kselect/releases/latest/download/kselect-linux-amd64.tar.gz | tar xz
sudo mv kselect /usr/local/bin/

# Linux (ARM64)
curl -L https://github.com/bangmodtechnology/kselect/releases/latest/download/kselect-linux-arm64.tar.gz | tar xz
sudo mv kselect /usr/local/bin/
```

### From Source

```bash
git clone https://github.com/bangmodtechnology/kselect.git
cd kselect
make build
sudo mv kselect /usr/local/bin/
```

Or use `make install`:

```bash
make install
```

### Shell Completion

After installation, enable shell completion:

```bash
# Bash
echo 'source <(kselect completion bash)' >> ~/.bashrc
source ~/.bashrc

# Zsh
echo 'source <(kselect completion zsh)' >> ~/.zshrc
source ~/.zshrc
```

## Prerequisites

- `~/.kube/config` configured with cluster access
- Go 1.21+ (only for building from source)

## Usage

```
kselect [flags] fields FROM resource [WHERE conditions] [ORDER BY field] [LIMIT n]
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-n` | Namespace (like `kubectl -n`) | current context |
| `-A` | All namespaces (like `kubectl -A`) | |
| `-o` | Output format: `table`, `json`, `yaml`, `csv`, `wide` | `table` |
| `--interactive` | Interactive REPL mode | |
| `--dry-run` | Validate query without executing | |
| `--describe` | Describe resource schema (e.g., `--describe pod`) | |
| `--list` | List available resources and fields | |
| `--plugins` | Directory containing plugin YAML files | |
| `--watch` | Watch mode: continuously refresh results | |
| `--interval` | Watch refresh interval | `2s` |
| `--no-color` | Disable color output | auto-detect TTY |
| `--version` | Show version | |

### Namespace Resolution

Priority: `-A` > `-n` flag > `WHERE namespace=` > current kube context namespace

## Query Examples

### Basic Queries

```bash
$ kselect name,status,ip FROM pod WHERE namespace=default
```
```
NAME                        STATUS    IP
nginx-frontend-7c9b5d8f4-abc12   Running   10.244.0.15
nginx-frontend-7c9b5d8f4-def34   Running   10.244.0.16
backend-api-5d4f8b7c9-ghi56      Running   10.244.1.22
redis-6f7b8c9d0-jkl78            Running   10.244.1.23
postgres-8a9b0c1d2-mno90         Running   10.244.2.10

5 resource(s) found.
```

```bash
$ kselect name,replicas,ready,available FROM deployment WHERE namespace=default
```
```
NAME               REPLICAS   READY   AVAILABLE
nginx-frontend     3          3       3
backend-api        2          2       2
worker             1          1       1
redis              1          1       1
postgres           1          1       1

5 resource(s) found.
```

```bash
# Services with ports
kselect name,type,cluster-ip,port FROM service WHERE namespace=default

# Default fields (use * or omit fields entirely)
kselect * FROM pod WHERE ns=default
kselect FROM pod WHERE ns=default
```

### WHERE Conditions

```bash
$ kselect name,status FROM pod WHERE namespace=default AND status=Running
```
```
NAME                                STATUS
nginx-frontend-7c9b5d8f4-abc12     Running
nginx-frontend-7c9b5d8f4-def34     Running
backend-api-5d4f8b7c9-ghi56        Running

3 resource(s) found.
```

```bash
$ kselect name,image FROM pod WHERE name LIKE 'nginx-%'
```
```
NAME                                IMAGE
nginx-frontend-7c9b5d8f4-abc12     nginx:1.25-alpine
nginx-frontend-7c9b5d8f4-def34     nginx:1.25-alpine
nginx-frontend-7c9b5d8f4-xyz99     nginx:1.25-alpine

3 resource(s) found.
```

```bash
# OR
kselect name,status FROM pod WHERE status=Running OR status=Pending

# NOT LIKE
kselect name FROM pod WHERE image NOT LIKE '%latest%'

# IN
kselect name,status FROM pod WHERE status IN ('Running','Pending','Failed')

# Comparison operators
kselect name,restarts FROM pod WHERE restarts > 5
```

### Sorting & Pagination

```bash
$ kselect name,restarts FROM pod ORDER BY restarts DESC LIMIT 5
```
```
NAME                                RESTARTS
backend-api-5d4f8b7c9-ghi56        12
worker-9e8f7d6c5-pqr12             8
redis-6f7b8c9d0-jkl78              3
nginx-frontend-7c9b5d8f4-abc12     1
postgres-8a9b0c1d2-mno90           0

5 resource(s) found.
```

```bash
# Sort ascending
kselect name,status FROM pod ORDER BY name

# Multiple sort fields
kselect namespace,name FROM pod ORDER BY namespace ASC, name ASC

# Limit with offset (pagination)
kselect name,status FROM pod LIMIT 10 OFFSET 20
```

### Aggregation

Aggregate functions support **shell-safe syntax** — no parentheses needed:

| SQL Syntax | Shell-Safe Syntax | Meaning |
|------------|-------------------|---------|
| `COUNT(*)` | `COUNT as alias` | Count all rows |
| `COUNT(field)` | `COUNT.field as alias` | Count non-null values |
| `SUM(field)` | `SUM.field as alias` | Sum values |
| `AVG(field)` | `AVG.field as alias` | Average values |

> **Why?** Shells interpret `*` and `()` as glob patterns and subshells.
> Use the shell-safe syntax to avoid quoting issues.

```bash
$ kselect namespace, COUNT as pod_count FROM pod -A GROUP BY namespace ORDER BY pod_count DESC
```
```
NAMESPACE        POD_COUNT
default          12
kube-system      8
production       6
monitoring       4

4 resource(s) found.
```

```bash
$ kselect status, COUNT as count FROM pod -n default GROUP BY status
```
```
STATUS     COUNT
Running    9
Pending    2
Failed     1

3 resource(s) found.
```

```bash
$ kselect DISTINCT status FROM pod -n default
```
```
STATUS
Running
Pending
Failed

3 resource(s) found.
```

```bash
# Average restarts (dot notation)
kselect namespace, AVG.restarts as avg_restarts FROM pod GROUP BY namespace

# Multiple aggregates
kselect namespace, COUNT as total, SUM.restarts as restarts FROM pod GROUP BY namespace

# HAVING filter
kselect namespace, COUNT as count FROM pod GROUP BY namespace HAVING count > 10
```

### JOIN

```bash
# Join pods with services
kselect pod.name,pod.ip,svc.name,svc.cluster-ip \
  FROM pod \
  INNER JOIN service svc ON pod.label.app = svc.selector.app \
  WHERE pod.namespace=default

# Left join deployments with pods
kselect deploy.name,deploy.replicas,pod.name,pod.status \
  FROM deployment deploy \
  LEFT JOIN pod ON deploy.selector.matchLabels.app = pod.label.app \
  WHERE deploy.namespace=production
```

### Output Formats

**Table** (default):
```bash
$ kselect name,status FROM pod WHERE namespace=default LIMIT 3
```
```
NAME                                STATUS
nginx-frontend-7c9b5d8f4-abc12     Running
backend-api-5d4f8b7c9-ghi56        Running
redis-6f7b8c9d0-jkl78              Running

3 resource(s) found.
```

**JSON**:
```bash
$ kselect name,status FROM pod WHERE namespace=default LIMIT 2 -o json
```
```json
[
  {
    "name": "nginx-frontend-7c9b5d8f4-abc12",
    "status": "Running"
  },
  {
    "name": "backend-api-5d4f8b7c9-ghi56",
    "status": "Running"
  }
]
```

**CSV** (export to file):
```bash
$ kselect name,status,ip FROM pod -o csv > pods.csv
$ cat pods.csv
name,status,ip
nginx-frontend-7c9b5d8f4-abc12,Running,10.244.0.15
backend-api-5d4f8b7c9-ghi56,Running,10.244.1.22
```

```bash
# YAML
kselect name,status FROM pod -o yaml

# Wide (no column truncation)
kselect name,image FROM pod -o wide
```

### Watch Mode

```bash
# Watch pods in real-time
kselect name,status,restarts FROM pod WHERE namespace=default --watch

# Custom refresh interval
kselect name,ready FROM deployment --watch --interval 5s
```

### Interactive Mode (REPL)

```bash
$ kselect --interactive
kselect> name,status FROM pod WHERE ns=default LIMIT 3
NAME                                STATUS
nginx-frontend-7c9b5d8f4-abc12     Running
backend-api-5d4f8b7c9-ghi56        Running
redis-6f7b8c9d0-jkl78              Running

3 resource(s) found.

kselect> DESCRIBE pod
Resource: pod (aliases: pods, po)
Scope: Namespaced

Default Fields: name, status, ip, node, restarts, age

All Fields:
  name         - Pod name [string]
  namespace    - Namespace (alias: ns) [string]
  status       - Pod phase [string]
  ip           - Pod IP [string]
  ...

kselect> namespace, COUNT as total FROM pod GROUP BY namespace
NAMESPACE      TOTAL
default        12
kube-system    8

kselect> \exit
Goodbye!
```

### Query Validation

```bash
# Validate query before execution
$ kselect --dry-run name,invalid FROM pod
Validation error: Field 'invalid' not found in resource 'pod'

# Typo detection with suggestions
$ kselect --dry-run nam FROM pod
Validation error: Field 'nam' not found in resource 'pod'
Did you mean: name?

$ kselect --dry-run name FROM pode
Validation error: Resource 'pode' not found
Did you mean: pod, node, or pods?

# Check query is valid
$ kselect --dry-run name,status FROM pod WHERE ns=default
Query validation passed ✓
Resource: pod
Fields: name, status
```

### Resource Discovery

```bash
# List all available resources
$ kselect --list
Available Resources:

  pod (pods, po)
  Default: name, status, ip, node, restarts, age
  Fields:
    - name          Pod name [string]
    - namespace     Namespace (alias: ns) [string]
    - status        Pod phase [string]
    ...

# Describe specific resource
$ kselect --describe deployment
Resource: deployment (aliases: deployments, deploy)
Scope: Namespaced

Default Fields: name, replicas, ready, available, age

All Fields:
  name         - Deployment name [string]
  namespace    - Namespace (alias: ns) [string]
  replicas     - Desired replicas [number]
  ready        - Ready replicas [number]
  ...

# Alternative syntax
$ kselect DESCRIBE service
```

## Available Resources

| Resource | Aliases | Default Fields | All Fields |
|----------|---------|----------------|------------|
| pod | pods, po | name, status, ip, node, restarts, age | + namespace, image, cpu.req, cpu.limit, mem.req, mem.limit, cpu.req-m, cpu.limit-m, mem.req-mi, mem.limit-mi, labels |
| deployment | deployments, deploy | name, replicas, ready, available, age | + namespace, updated, image, strategy, cpu.req, cpu.limit, mem.req, mem.limit, cpu.req-m, cpu.limit-m, mem.req-mi, mem.limit-mi, labels |
| daemonset | daemonsets, ds | name, desired, current, ready, available, age | + namespace, updated, misscheduled, image, selector, cpu.req, cpu.limit, mem.req, mem.limit, cpu.req-m, cpu.limit-m, mem.req-mi, mem.limit-mi, labels |
| statefulset | statefulsets, sts | name, replicas, ready, age | + namespace, current, updated, image, servicename, cpu.req, cpu.limit, mem.req, mem.limit, cpu.req-m, cpu.limit-m, mem.req-mi, mem.limit-mi, labels |
| job | jobs | name, completions, succeeded, failed, age | + namespace, active, parallelism, backofflimit, image, cpu.req, cpu.limit, mem.req, mem.limit, cpu.req-m, cpu.limit-m, mem.req-mi, mem.limit-mi, labels |
| cronjob | cronjobs, cj | name, schedule, suspend, active, last-schedule, age | + namespace, last-success, concurrency, image, cpu.req, cpu.limit, mem.req, mem.limit, cpu.req-m, cpu.limit-m, mem.req-mi, mem.limit-mi, labels |
| service | services, svc | name, type, cluster-ip, port, age | + namespace, external-ip, targetport, selector |
| ingress | ingresses, ing | name, class, host, address, age | + namespace |
| configmap | configmaps, cm | name, data-keys, age | + namespace |
| secret | secrets | name, type, age | + namespace, data-keys |
| serviceaccount | serviceaccounts, sa | name, secrets, age | + namespace |
| node | nodes, no | name, status, roles, version, internal-ip, age | + external-ip, os, kernel, container-runtime, cpu, memory, pods, arch, labels |
| gateway | gateways, gw | name, class, addresses, programmed, age | + namespace, listeners, labels |
| networkpolicy | netpol | name, pod-selector, policy-types, age | + namespace, ingress-rules, egress-rules, labels |
| poddisruptionbudget | pdb, pdbs | name, min-available, max-unavailable, current-healthy, age | + namespace, desired-healthy, disruptions-allowed, expected-pods, labels |
| resourcequota | quota, quotas | name, hard, used, age | + namespace, scopes, labels |
| role | roles | name, rules, age | + namespace, labels |
| rolebinding | rolebindings | name, role-ref, subjects, age | + namespace, labels |
| clusterrole | clusterroles | name, rules, age | + aggregation-rule, labels |
| clusterrolebinding | clusterrolebindings | name, role-ref, subjects, age | + labels |

### Normalized Fields for Aggregation

Workload resources (pod, deployment, daemonset, statefulset, job, cronjob) support normalized numeric fields for accurate `SUM` and `AVG` aggregations:

| Field | Unit | Description |
|-------|------|-------------|
| `cpu.req-m` | millicores | CPU requests normalized to millicores |
| `cpu.limit-m` | millicores | CPU limits normalized to millicores |
| `mem.req-mi` | MiB | Memory requests normalized to MiB |
| `mem.limit-mi` | MiB | Memory limits normalized to MiB |

```bash
# Example: Total CPU limits by namespace
kselect ns, SUM.cpu.limit-m as total_cpu FROM pod GROUP BY ns -A

# Example: Average memory requests across deployments
kselect ns, AVG.mem.req-mi as avg_mem FROM deployment GROUP BY ns -A
```

`*` or omitting fields will display the Default Fields for that resource.

Run `kselect -list` to see all available resources and fields.

## Field Aliases

WHERE conditions support Kubernetes API short names:

| Alias | Full Name |
|-------|-----------|
| ns | namespace |

```bash
# Both are equivalent
kselect * FROM pod WHERE ns=default
kselect * FROM pod WHERE namespace=default
```

Aliases work in WHERE, ORDER BY, GROUP BY, and HAVING clauses.

## Shell Quoting

Shells like zsh and bash interpret `*` and `()` as special characters. kselect provides **shell-safe syntax** so you never need to quote:

```bash
# Problem: shell interprets * and () as glob/subshell
kselect namespace, COUNT(*) as total FROM pod    # zsh: no matches found: COUNT(*)
kselect namespace, COUNT() as total FROM pod     # zsh: interprets () as function def

# Solution: use shell-safe syntax (no parens needed)
kselect namespace, COUNT as total FROM pod -A GROUP BY namespace
kselect namespace, SUM.restarts as total FROM pod GROUP BY namespace

# Quoting also works if you prefer SQL syntax
kselect namespace, 'COUNT(*)' as total FROM pod -A GROUP BY namespace

# Note: bare * for fields works fine (kselect handles shell-expanded filenames gracefully)
kselect * FROM pod
```

## Real-World Use Cases

### Troubleshooting: Find problematic pods

```bash
$ kselect name,status,node FROM pod WHERE namespace=production AND status!=Running
```
```
NAME                           STATUS              NODE
payment-worker-broken-abc12    CrashLoopBackOff    node-3
migration-job-xyz99            Error               node-1
pending-pod-def34              Pending             <none>

3 resource(s) found.
```

```bash
# Pods with frequent restarts
kselect name,restarts,status FROM pod WHERE restarts > 10 ORDER BY restarts DESC

# Pods using outdated images
kselect name,image FROM pod WHERE image LIKE '%:latest%'
```

### Capacity Planning: View resource usage

```bash
$ kselect name,cpu.req,mem.req,cpu.limit,mem.limit FROM pod WHERE namespace=production LIMIT 4
```
```
NAME                           CPU.REQ   MEM.REQ   CPU.LIMIT   MEM.LIMIT
nginx-frontend-7c9b5d8f4-a     100m      64Mi      200m        128Mi
backend-api-5d4f8b7c9-g         250m      256Mi     500m        512Mi
worker-9e8f7d6c5-p               200m      128Mi     400m        256Mi
postgres-8a9b0c1d2-m             250m      256Mi     500m        512Mi

4 resource(s) found.
```

```bash
$ kselect name,replicas,ready FROM deployment WHERE ready!=replicas
```
```
NAME               REPLICAS   READY
backend-api        3          2
worker             2          0

2 resource(s) found.
```

```bash
# Export as CSV for further analysis
kselect name,cpu.req,mem.req FROM pod -o csv > resources.csv
```

### Networking: Inspect services and ingresses

```bash
$ kselect name,type,cluster-ip,port FROM service WHERE namespace=default
```
```
NAME               TYPE           CLUSTER-IP       PORT
nginx-service      LoadBalancer   10.96.100.10     80,443
backend-service    ClusterIP      10.96.100.20     8080
redis-service      ClusterIP      10.96.100.30     6379
postgres-service   ClusterIP      10.96.100.40     5432
debug-nodeport     NodePort       10.96.100.50     8080

5 resource(s) found.
```

```bash
# List all ingresses with hosts
kselect name,host,class FROM ingress -o wide

# Find services connected to pods
kselect pod.name,pod.ip,svc.name,svc.port \
  FROM pod INNER JOIN service svc ON pod.label.app = svc.selector.app
```

### Infrastructure: View nodes and gateways

```bash
# List all nodes
$ kselect * FROM node
```
```
NAME        STATUS   ROLES          VERSION   INTERNAL-IP    AGE
node-1      Ready    control-plane  v1.29.1   192.168.1.10   30d
node-2      Ready    worker         v1.29.1   192.168.1.11   30d
node-3      Ready    worker         v1.29.1   192.168.1.12   25d

3 resource(s) found.
```

```bash
# Node capacity
kselect name,cpu,memory,pods FROM no

# Gateway API resources
kselect * FROM gw WHERE ns=default
```

### Security: Inspect secrets and configmaps

```bash
# All secret types
kselect name,type FROM secret WHERE namespace=production

# ConfigMaps with many keys
kselect name,data-keys FROM configmap WHERE namespace=default -o wide
```

## Plugin System

Extend kselect with custom resource definitions via YAML:

```yaml
# plugins/certificate.yaml
name: certificate
aliases:
  - certificates
  - cert
group: cert-manager.io
version: v1
resource: certificates
fields:
  name:
    jsonpath: "{.metadata.name}"
    type: string
    description: Certificate name
  ready:
    jsonpath: "{.status.conditions[?(@.type=='Ready')].status}"
    type: string
    description: Ready status
  secret:
    jsonpath: "{.spec.secretName}"
    type: string
    description: Secret name
  issuer:
    jsonpath: "{.spec.issuerRef.name}"
    type: string
    description: Issuer name
```

Load plugins:

```bash
kselect --plugins=./plugins name,ready,issuer FROM certificate WHERE namespace=default
```

## Development

```bash
make build      # Build binary
make test       # Run tests (all packages)
make deps       # go mod tidy
make clean      # Clean build artifacts
make install    # Build and install to /usr/local/bin
```

### Testing

kselect has comprehensive test coverage:

```bash
# Run all tests
make test

# Run specific package tests
go test -v ./pkg/parser/
go test -v ./pkg/validator/
go test -v ./test/

# Run with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Test Coverage:**
- Parser: 79.5% coverage
- Validator: 78.1% coverage
- Integration: 66 test cases
- Edge cases & complex scenarios

### Project Structure

```
kselect/
├── cmd/kselect/          # Main CLI entry point
├── pkg/
│   ├── parser/           # SQL-like query parser
│   ├── validator/        # Query validation with fuzzy matching
│   ├── registry/         # Resource definitions
│   ├── executor/         # K8s API interaction
│   ├── output/           # Output formatters
│   ├── completion/       # Shell completion
│   ├── describe/         # DESCRIBE command
│   └── repl/             # Interactive REPL
└── test/                 # Integration tests
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on:

- Adding new resources
- Adding new output formats
- Creating YAML plugins
- Running tests
- Submitting pull requests

Quick contribution examples:

```bash
# Add a new built-in resource
# 1. Create pkg/registry/myresource.go
# 2. Register with init() function
# 3. Add tests

# Add a YAML plugin (no code needed!)
# 1. Create plugins/myresource.yaml
# 2. Define fields with JSONPath
# 3. Use with --plugins flag
```

## Community

- **Issues:** [GitHub Issues](https://github.com/bangmodtechnology/kselect/issues)
- **Discussions:** [GitHub Discussions](https://github.com/bangmodtechnology/kselect/discussions)
- **Contributing:** [CONTRIBUTING.md](CONTRIBUTING.md)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
