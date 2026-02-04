# 🚀 QUICK START GUIDE - Pahlawan Pangan

## Prerequisites
- Go 1.22+
- Docker & Docker Compose
- kubectl (for K8s deployment)
- Make (optional, for convenience)

---

## 🏃 5-Minute Local Setup

### Step 1: Start Infrastructure
```bash
cd c:\dev\food-waste
docker-compose up -d
```

**What this does**:
- ✅ PostgreSQL + PostGIS (port 5432)
- ✅ Redis Stack (port 6379)
- ✅ NATS JetStream (port 4222)
- ✅ Jaeger (port 16686)
- ✅ Prometheus (port 9090)
- ✅ Grafana (port 3000)

### Step 2: Install Dependencies
```bash
go mod download
```

### Step 3: Run Database Migrations
```bash
# Set environment variable
$env:DATABASE_URL="postgres://admin:secret123@localhost:5432/pahlawan?sslmode=disable"

# Run migrations
psql $env:DATABASE_URL -f db/schema.sql
```

### Step 4: Start the Server
```bash
# Set all environment variables
$env:DATABASE_URL="postgres://admin:secret123@localhost:5432/pahlawan?sslmode=disable"
$env:REDIS_URL="redis://localhost:6379"
$env:NATS_URL="nats://localhost:4222"
$env:OTEL_EXPORTER_OTLP_ENDPOINT="localhost:4317"
$env:OTEL_SERVICE_NAME="matching-engine"

# Run server
go run cmd/server/main.go
```

**Expected output**:
```
Starting API server on :8080
Starting metrics server on :9090
```

### Step 5: Test the API
```bash
# Health check
curl http://localhost:8080/health/live

# Post a surplus
curl -X POST http://localhost:8080/api/v1/surplus \
  -H "Content-Type: application/json" \
  -d '{
    "provider_id": "provider-123",
    "lat": -6.2088,
    "lon": 106.8456,
    "quantity_kgs": 50.0,
    "food_type": "Rice",
    "expiry_time": "2026-02-05T12:00:00Z"
  }'
```

### Step 6: View Observability
- **Jaeger Traces**: http://localhost:16686
- **Prometheus Metrics**: http://localhost:9090
- **Grafana Dashboards**: http://localhost:3000 (admin/admin)

---

## 🧪 Running Tests

```bash
# Unit tests
go test ./...

# With coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Benchmarks
go test -bench=. ./internal/matching/
```

**Expected benchmark results**:
```
BenchmarkMatchNGO-8    5000    250000 ns/op    1200 B/op    15 allocs/op
```

---

## 🐳 Docker Deployment

### Build Image
```bash
docker build -t pahlawan-pangan:latest .
```

### Run Container
```bash
docker run -p 8080:8080 -p 9090:9090 \
  -e DATABASE_URL="postgres://admin:secret123@host.docker.internal:5432/pahlawan?sslmode=disable" \
  -e REDIS_URL="redis://host.docker.internal:6379" \
  -e NATS_URL="nats://host.docker.internal:4222" \
  pahlawan-pangan:latest
```

---

## ☸️ Kubernetes Deployment

### Prerequisites
- Kubernetes cluster (minikube, kind, or cloud)
- kubectl configured

### Deploy

```bash
# Create namespace
kubectl create namespace pahlawan-pangan

# Deploy Redis Cluster
kubectl apply -f k8s/redis-cluster.yaml

# Wait for Redis to be ready
kubectl wait --for=condition=ready pod -l app=redis-cluster -n pahlawan-pangan --timeout=300s

# Deploy Application
kubectl apply -f k8s/deployment.yaml

# Check status
kubectl get pods -n pahlawan-pangan
kubectl get hpa -n pahlawan-pangan
kubectl get pdb -n pahlawan-pangan
```

### Port Forward for Testing
```bash
# API
kubectl port-forward -n pahlawan-pangan svc/matching-engine 8080:80

# Metrics
kubectl port-forward -n pahlawan-pangan svc/matching-engine 9090:9090
```

### Scale Manually (for testing)
```bash
kubectl scale deployment matching-engine -n pahlawan-pangan --replicas=20
```

---

## 📊 Monitoring

### View Metrics
```bash
# Prometheus
open http://localhost:9090

# Example queries:
# - histogram_quantile(0.95, rate(surplus_claim_latency_seconds_bucket[5m]))
# - rate(food_waste_prevented_tons_total[1h])
# - matching_engine_saturation_ratio
```

### View Traces
```bash
# Jaeger
open http://localhost:16686

# Search for:
# - Service: matching-engine
# - Operation: MatchNGO
# - Tags: surplus.id=<id>
```

### View Logs
```bash
# Docker Compose
docker-compose logs -f

# Kubernetes
kubectl logs -f -n pahlawan-pangan -l app=matching-engine
```

---

## 🔧 Troubleshooting

### Issue: "Database connection refused"
```bash
# Check if PostgreSQL is running
docker-compose ps postgres

# Check logs
docker-compose logs postgres

# Restart
docker-compose restart postgres
```

### Issue: "NATS connection failed"
```bash
# Check NATS status
docker-compose ps nats

# Test connection
nats-cli server ping nats://localhost:4222
```

### Issue: "Redis GEORADIUS not available"
```bash
# Ensure using Redis Stack (not vanilla Redis)
docker exec -it pahlawan-redis redis-cli INFO modules

# Should show: module:name=ReJSON, module:name=search
```

### Issue: "HPA not scaling"
```bash
# Check metrics server
kubectl get apiservice v1beta1.metrics.k8s.io

# Check custom metrics
kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1

# View HPA status
kubectl describe hpa matching-engine-hpa -n pahlawan-pangan
```

---

## 🎯 Common Tasks

### Add Test Data
```bash
# Create providers
curl -X POST http://localhost:8080/api/v1/surplus \
  -H "Content-Type: application/json" \
  -d '{
    "provider_id": "restaurant-1",
    "lat": -6.2088,
    "lon": 106.8456,
    "quantity_kgs": 25.0,
    "food_type": "Vegetables",
    "expiry_time": "2026-02-05T18:00:00Z"
  }'
```

### Query Database
```bash
# Connect to PostgreSQL
docker exec -it pahlawan-postgres psql -U admin -d pahlawan

# Example queries:
SELECT COUNT(*) FROM surplus;
SELECT * FROM surplus WHERE status = 'available';
SELECT * FROM outbox_events WHERE published = false;
```

### Clear Data
```bash
# Truncate tables
docker exec -it pahlawan-postgres psql -U admin -d pahlawan -c "TRUNCATE surplus, outbox_events CASCADE;"
```

---

## 📚 Next Steps

1. **Read Documentation**:
   - `ARCHITECTURE.md` - High-level overview
   - `TECHNICAL_DESIGN.md` - Deep dive
   - `EXECUTIVE_SUMMARY.md` - Complete summary

2. **Explore Code**:
   - `internal/matching/engine.go` - Core logic
   - `internal/api/handler.go` - API endpoints
   - `internal/outbox/outbox.go` - Event handling

3. **Customize**:
   - Add authentication middleware
   - Implement actual routing API (OSRM/Google Maps)
   - Add mobile app integration
   - Implement ML-based matching

4. **Deploy to Production**:
   - Set up managed Kubernetes (GKE/EKS/AKS)
   - Configure managed PostgreSQL (Cloud SQL/RDS/Azure DB)
   - Set up Redis Enterprise
   - Configure monitoring (Datadog/New Relic)

---

## 🆘 Getting Help

- **Documentation**: See `README.md` and other docs
- **Issues**: Check error logs in Docker/K8s
- **Metrics**: Use Prometheus to diagnose performance
- **Traces**: Use Jaeger to debug request flow

---

## ✅ Verification Checklist

After setup, verify:

- [ ] API responds to health checks
- [ ] Can post surplus successfully
- [ ] Events appear in outbox table
- [ ] Metrics visible in Prometheus
- [ ] Traces visible in Jaeger
- [ ] HPA shows current metrics (K8s only)
- [ ] Redis GEORADIUS works
- [ ] PostgreSQL partitions created

---

**You're ready to save the world from food waste! 🌍**
