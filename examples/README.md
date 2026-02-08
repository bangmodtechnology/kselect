# kselect Examples

Example Kubernetes resources for testing kselect queries.

## Setup

```bash
# Deploy all resources (files are numbered for correct apply order)
kubectl apply -f examples/

# Clean up when done
kubectl delete namespace kselect-demo
```

## What Gets Created

| Resource | Count | Notes |
|----------|-------|-------|
| Namespace | 1 | `kselect-demo` |
| ConfigMaps | 3 | app-config, nginx-config, redis-config |
| Secrets | 3 | db-credentials, api-keys, tls-secret |
| ServiceAccounts | 2 | backend-sa, worker-sa |
| Deployments | 5 | nginx(3), backend(2), worker(1), redis(1), postgres(1) |
| Pods (standalone) | 4 | debug, sidecar, failing, batch-job |
| Services | 5 | LB, ClusterIP, NodePort |
| Ingresses | 2 | main-ingress, api-ingress |

Total: ~12 pods (low resource requests, suitable for local clusters)

## Query Examples

```bash
# --- Basic ---
kselect name,status,ip FROM pod WHERE namespace=kselect-demo
kselect name,replicas,ready FROM deployment WHERE namespace=kselect-demo
kselect name,type,cluster-ip,port FROM service WHERE namespace=kselect-demo

# --- WHERE conditions ---
kselect name,status FROM pod WHERE namespace=kselect-demo AND status=Running
kselect name FROM pod WHERE namespace=kselect-demo AND name LIKE 'nginx-%'
kselect name,restarts FROM pod WHERE namespace=kselect-demo AND restarts > 0

# --- ORDER BY / LIMIT ---
kselect name,cpu.req,mem.req FROM pod WHERE namespace=kselect-demo ORDER BY name
kselect name,restarts FROM pod WHERE namespace=kselect-demo ORDER BY restarts DESC LIMIT 3

# --- Aggregation ---
kselect status, COUNT(*) as count FROM pod WHERE namespace=kselect-demo GROUP BY status
kselect namespace, COUNT(*) as pod_count FROM pod GROUP BY namespace ORDER BY pod_count DESC LIMIT 5

# --- Output formats ---
kselect name,status FROM pod WHERE namespace=kselect-demo -o json
kselect name,status FROM pod WHERE namespace=kselect-demo -o yaml
kselect name,status,ip FROM pod WHERE namespace=kselect-demo -o csv > pods.csv
kselect name,image FROM pod WHERE namespace=kselect-demo -o wide

# --- Watch mode ---
kselect name,status,restarts FROM pod WHERE namespace=kselect-demo --watch
```
