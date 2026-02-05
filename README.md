# Pahlawan Pangan

🌍 **Global-Scale Food Waste Solution Platform**

Connecting surplus food providers with NGOs in real-time at sub-second latency.

## 🎯 Mission
Solve the "Last Mile Logistics" of food surplus distribution at a scale of **10M+ transactions/day**.

## 🏗️ Architecture Highlights

### Core Technologies
- **Language**: Go 1.22+
- **Database**: PostgreSQL + PostGIS (Citus for sharding)
- **Cache**: Redis Cluster (GEORADIUS)
- **Messaging**: NATS JetStream
- **Orchestration**: Kubernetes
- **Observability**: OpenTelemetry + Prometheus + Jaeger

### Key Features
✅ **Geo-Sharded Actor Model** for spatial consistency  
✅ **Transactional Outbox Pattern** for guaranteed delivery  
✅ **Circuit Breaker** with Haversine fallback  
✅ **Custom Metrics HPA** for predictive scaling  
✅ **Distributed Tracing** with OpenTelemetry  
✅ **Sub-second latency** at global scale  

## 📊 Performance Targets

| Metric | Target | P95 | P99 |
|--------|--------|-----|-----|
| Claim Latency | <500ms | <800ms | <1.5s |
| Throughput | 10M/day | - | - |
| Availability | 99.95% | - | - |

## 🚀 Quick Start

### Prerequisites
- Go 1.22+
- Docker & Docker Compose
- Kubernetes cluster (local: kind/minikube)

### Local Development

```bash
# Clone repository
git clone https://github.com/albnnaardy11/pahlawan-pangan
cd pahlawan-pangan

# Install dependencies
go mod download

# Start infrastructure (PostgreSQL, Redis, NATS)
docker-compose up -d

# Run migrations
psql $DATABASE_URL < db/schema.sql

# Start server
go run cmd/server/main.go
```

### Environment Variables

```bash
export DATABASE_URL="postgres://user:pass@localhost:5432/pahlawan?sslmode=disable"
export REDIS_URL="redis://localhost:6379"
export NATS_URL="nats://localhost:4222"
export OTEL_EXPORTER_OTLP_ENDPOINT="localhost:4317"
export OTEL_SERVICE_NAME="matching-engine"
```

## 🧪 Testing

```bash
# Unit tests
go test ./...

# Integration tests
go test -tags=integration ./...

# Load test (requires k6)
k6 run tests/load/surplus_post.js
```

## 📦 Deployment

### Kubernetes

```bash
# Create namespace
kubectl create namespace pahlawan-pangan

# Deploy infrastructure
kubectl apply -f k8s/redis-cluster.yaml
kubectl apply -f k8s/postgres.yaml
kubectl apply -f k8s/nats.yaml

# Deploy application
kubectl apply -f k8s/deployment.yaml

# Verify
kubectl get pods -n pahlawan-pangan
```

### Docker

```bash
# Build image
docker build -t pahlawan-pangan:latest .

# Run container
docker run -p 8080:8080 -p 9090:9090 \
  -e DATABASE_URL=$DATABASE_URL \
  pahlawan-pangan:latest
```

## 📖 Documentation

- [Architecture Deep Dive](ARCHITECTURE.md)
- [API Documentation](docs/API.md)
- [Database Schema](db/schema.sql)
- [Observability Guide](docs/OBSERVABILITY.md)

## 🔍 Monitoring

### Prometheus Metrics
- `surplus_claim_latency_seconds` - Claim processing time
- `food_waste_prevented_tons_total` - Total food saved
- `matching_engine_saturation_ratio` - Worker pool utilization

### Dashboards
- Grafana: http://localhost:3000
- Jaeger: http://localhost:16686
- Prometheus: http://localhost:9090

## 🛡️ Security

- JWT authentication (RS256)
- mTLS for service-to-service communication
- Rate limiting (100 req/min per provider)
- PII encryption (AES-256)

## 🌟 Impact

**Potential**: Save **1M tons** of food annually, feeding **10M+ people** globally.

## 📄 License

MIT License - see [LICENSE](LICENSE)

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md)

## 📧 Contact

- Email: engineering@pahlawan-pangan.org
- Slack: [Join our community](https://slack.pahlawan-pangan.org)

---

**Built with ❤️ by the Pahlawan Pangan Engineering Team**
