# kselect

SQL-like query tool for Kubernetes resources.

```bash
kselect name,status,ip FROM pod WHERE namespace=default
```

## Why kselect?

ถ้าคุณเคยใช้ `kubectl get pods` แล้วต้อง pipe ผ่าน `grep`, `awk`, `jq` หลายชั้นเพื่อหาข้อมูลที่ต้องการ — kselect ช่วยให้ทำสิ่งเดียวกันด้วยคำสั่งเดียว

**ปัญหาที่เจอบ่อย:**

```bash
# แบบเดิม: หา pod ที่ restart เยอะใน production
kubectl get pods -n production -o json | jq -r '.items[] | select(.status.containerStatuses[0].restartCount > 5) | [.metadata.name, .status.containerStatuses[0].restartCount] | @tsv' | sort -k2 -rn | head -5

# แบบ kselect: อ่านง่าย เขียนเร็ว
kselect name,restarts FROM pod WHERE namespace=production AND restarts > 5 ORDER BY restarts DESC LIMIT 5
```

**แนวคิดหลัก:**
- ใช้ syntax ที่คุ้นเคยแบบ SQL — ไม่ต้องจำ JSONPath หรือ jq syntax
- เลือกเฉพาะ field ที่ต้องการ ไม่ต้องดูข้อมูลทั้งหมด
- Filter, Sort, Aggregate ได้ในคำสั่งเดียว
- รองรับ JOIN ข้ามประเภท resource (เช่น pod กับ service)
- ขยาย resource ใหม่ผ่าน YAML plugin ได้โดยไม่ต้องแก้โค้ด

## How It Works

```
kselect name,status FROM pod WHERE namespace=default AND status!=Running ORDER BY name LIMIT 10
       ├──────────┘      └──┘     └──────────────────────────────────┘         └───────┘    └───┘
       fields          resource              conditions                       sorting     limit
```

kselect แปลง query เป็นการเรียก Kubernetes API โดยอัตโนมัติ:

1. **Parse** — แยก query ออกเป็น fields, resource, conditions
2. **Registry** — หา resource definition (GVR + field-to-JSONPath mapping)
3. **Execute** — เรียก K8s API ผ่าน dynamic client เพื่อดึง resource
4. **Filter** — กรองผลลัพธ์ตาม WHERE conditions ฝั่ง client
5. **Transform** — JOIN, Aggregate, Sort, Paginate
6. **Output** — แสดงผลในรูปแบบที่เลือก (table, json, yaml, csv)

## Quick Start

```bash
# ดู resources และ fields ที่รองรับ
kselect -list

# Query แรก
kselect name,status,ip FROM pod WHERE namespace=default

# Export เป็น JSON
kselect name,status FROM pod -o json

# Monitor แบบ real-time
kselect name,status,restarts FROM pod --watch
```

## Installation

### From Source

```bash
git clone https://github.com/bangmodtechnology/kselect.git
cd kselect
make build
sudo mv kselect /usr/local/bin/
```

### Or use `make install`

```bash
make install
```

## Prerequisites

- Go 1.21+
- `~/.kube/config` configured with cluster access

## Usage

```
kselect [flags] fields FROM resource [WHERE conditions] [ORDER BY field] [LIMIT n]
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-o` | Output format: `table`, `json`, `yaml`, `csv`, `wide` | `table` |
| `-list` | List available resources and fields | |
| `-plugins` | Directory containing plugin YAML files | |
| `-watch` | Watch mode: continuously refresh results | |
| `-interval` | Watch refresh interval | `2s` |
| `-version` | Show version | |

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

# All fields
kselect * FROM pod WHERE namespace=default
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

```bash
$ kselect namespace, COUNT(*) as pod_count FROM pod GROUP BY namespace ORDER BY pod_count DESC
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
$ kselect status, COUNT(*) as count FROM pod WHERE namespace=default GROUP BY status
```
```
STATUS     COUNT
Running    9
Pending    2
Failed     1

3 resource(s) found.
```

```bash
$ kselect DISTINCT status FROM pod WHERE namespace=default
```
```
STATUS
Running
Pending
Failed

3 resource(s) found.
```

```bash
# Average restarts
kselect namespace, AVG(restarts) as avg_restarts FROM pod GROUP BY namespace

# Multiple aggregates
kselect namespace, COUNT(*) as total, SUM(restarts) as restarts FROM pod GROUP BY namespace

# HAVING filter
kselect namespace, COUNT(*) as count FROM pod GROUP BY namespace HAVING count > 10
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

## Available Resources

| Resource | Aliases | Key Fields |
|----------|---------|------------|
| pod | pods, po | name, namespace, status, ip, node, image, restarts, cpu.req, cpu.limit, mem.req, mem.limit, age, labels |
| deployment | deployments, deploy | name, namespace, replicas, ready, available, updated, image, strategy, age, labels |
| service | services, svc | name, namespace, type, cluster-ip, external-ip, port, targetport, selector, age |
| ingress | ingresses, ing | name, namespace, class, host, address, age |
| configmap | configmaps, cm | name, namespace, data-keys, age |
| secret | secrets | name, namespace, type, data-keys, age |
| serviceaccount | serviceaccounts, sa | name, namespace, secrets, age |

Run `kselect -list` to see all available resources and fields.

## Real-World Use Cases

### Troubleshooting: หา pod ที่มีปัญหา

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
# Pod ที่ restart บ่อย
kselect name,restarts,status FROM pod WHERE restarts > 10 ORDER BY restarts DESC

# Pod ที่ใช้ image เก่า
kselect name,image FROM pod WHERE image LIKE '%:latest%'
```

### Capacity Planning: ดู resource usage

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
# Export เป็น CSV เพื่อวิเคราะห์ต่อ
kselect name,cpu.req,mem.req FROM pod -o csv > resources.csv
```

### Networking: ตรวจสอบ service/ingress

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
# ดู ingress ทั้งหมดพร้อม host
kselect name,host,class FROM ingress -o wide

# หา service ที่เชื่อม pod
kselect pod.name,pod.ip,svc.name,svc.port \
  FROM pod INNER JOIN service svc ON pod.label.app = svc.selector.app
```

### Security: ตรวจสอบ secrets/configmaps

```bash
# Secret ทุกประเภท
kselect name,type FROM secret WHERE namespace=production

# ConfigMap ที่มี key เยอะ
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
make test       # Run tests
make deps       # go mod tidy
make clean      # Clean build artifacts
make release    # Cross-compile for linux/darwin/windows
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
