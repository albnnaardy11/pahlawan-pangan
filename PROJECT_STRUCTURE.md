# Project Structure

```
pahlawan-pangan/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/
│   ├── api/
│   │   └── handler.go              # HTTP API handlers
│   ├── matching/
│   │   ├── engine.go               # Core matching logic
│   │   └── engine_test.go          # Unit tests
│   ├── messaging/
│   │   └── nats.go                 # NATS JetStream publisher
│   └── outbox/
│       └── outbox.go               # Transactional outbox pattern
├── db/
│   └── schema.sql                  # PostgreSQL schema with PostGIS
├── k8s/
│   ├── deployment.yaml             # K8s deployment + HPA + PDB
│   └── redis-cluster.yaml          # Redis StatefulSet
├── observability/
│   ├── prometheus-config.yaml      # Prometheus scrape config
│   ├── alert-rules.yaml            # Alerting rules
│   └── otel-collector-config.yaml  # OpenTelemetry Collector
├── docker-compose.yaml             # Local development stack
├── Dockerfile                      # Multi-stage production build
├── Makefile                        # Common tasks
├── go.mod                          # Go dependencies
├── go.sum                          # Dependency checksums
├── README.md                       # Project overview
├── ARCHITECTURE.md                 # Architecture deep dive
├── TECHNICAL_DESIGN.md             # Detailed technical design
└── DOCS.md                         # Quick reference
```

## Key Files

### Core Implementation
- **`internal/matching/engine.go`**: Geo-spatial matching with circuit breaker
- **`internal/outbox/outbox.go`**: Transactional outbox for guaranteed delivery
- **`internal/api/handler.go`**: REST API with OpenTelemetry tracing

### Infrastructure
- **`k8s/deployment.yaml`**: HPA with custom metrics, PDB for HA
- **`db/schema.sql`**: Partitioned tables with PostGIS for geo-queries
- **`docker-compose.yaml`**: Full local stack (Postgres, Redis, NATS, Jaeger)

### Observability
- **`observability/prometheus-config.yaml`**: Metrics collection
- **`observability/alert-rules.yaml`**: SLA monitoring alerts
- **`observability/otel-collector-config.yaml`**: Distributed tracing

## Quick Commands

```bash
# Start local development
make compose-up
make run

# Run tests
make test

# Build and deploy
make docker-build
make k8s-deploy

# View observability
make metrics    # Prometheus
make traces     # Jaeger
make grafana    # Dashboards
```
