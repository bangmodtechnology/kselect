# kselect Examples

ไฟล์ตัวอย่าง Kubernetes resources สำหรับทดสอบ kselect queries

## Setup

```bash
# สร้าง namespace สำหรับทดสอบ
kubectl create namespace kselect-demo

# Deploy ทุก resources
kubectl apply -f examples/

# ลบทิ้งเมื่อเสร็จ
kubectl delete namespace kselect-demo
```

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
