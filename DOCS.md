# Pahlawan Pangan: Hyper-Scale Food Surplus Platform
## Architectural Blueprint & Global Strategy

### 1. Distributed Spatial Consistency & The Matching Engine
**Challenge**: Handling the "Thundering Herd" of 10,000+ simultaneous restaurant closings.

#### Geo-Sharded Actor Model
To maintain sub-second latency and spatial consistency, we utilize a **Geo-Sharded Actor Model** (implemented via `Proto.Actor` or `Dapr` on Kubernetes).
- **Spatial Sharding**: The globe is divided into S2 Cells (Level 12-15). Each cell has a dedicated **Shard Actor**.
- **Conflict Resolution**: When multiple NGOs claim the same surplus, the Shard Actor serves as the single source of truth for that specific location, processing claims sequentially to ensure **Strong Local Consistency**.
- **The Buffer Layer**: Redis `GEORADIUS` stores the "Hot State" (available surplus within the last 2 hours). PostGIS serves as the durable "Cold State" for historical analysis and long-term search.

#### CAP Theorem Strategy
- **Choice: AP (Availability / Partition Tolerance)** with **Conflict-Free State**.
- **Reasoning**: It is better to show an NGO food that *might* be gone (and handle the "already claimed" error gracefully) than to show no food during a network partition. We use **Optimistic Locking** on the `surplus_id` in the database to prevent double-claiming if an Actor fails over.

---

### 2. Event-Driven Reliability (EDA)
#### Transactional Outbox Pattern
To guarantee that every surplus post results in a notification, we implement the **Transactional Outbox Pattern**:
1.  **Atomicity**: The `surplus` write and an `outbox_event` (JSON) are committed in a single PostgreSQL transaction.
2.  **Relay**: A high-performance Change Data Capture (CDC) service (e.g., Debezium) or a dedicated Go poller reads the outbox and publishes to **NATS JetStream**.
3.  **Deduplication**: NATS JetStream's sequence numbers ensure exactly-once delivery to the Matching Microservice.

#### Resiliency & Re-routing
- **TTL Deadlines**: Every surplus has an `ExpiryTime`.
- **Automated Escalation**: If the primary matched NGO does not "Acknowledge" within 5 minutes, the flow is:
  - Actor triggers `RE_MATCH_EVENT`.
  - Engine excludes previous NGO and finds the next nearest.
  - Push notification sent to the new NGO.
- **DLQ Strategy**: Failed matching attempts are sent to a **Geo-DLQ**, where manual intervention or heuristic adjustments can save the food (e.g., lowering the barrier for general food banks).

---

### 3. Advanced Observability & Self-Healing
#### Distributed Tracing (OpenTelemetry)
Every request carries a `trace_id`. We trace: `Provider App -> Envoy Proxy -> Matching Service -> Redis -> Redis Response -> Push Service -> NGO App`.
- **Context Propagation**: Using W3C Trace Context.

#### Custom Prometheus Metrics
- `surplus_claim_latency_seconds`: (Histogram) Time from post to first claim.
- `food_waste_prevented_tons_total`: (Counter) Sum of all QuantityKgs of claimed surplus.
- `matching_engine_saturation_ratio`: (Gauge) Monitor worker pool saturation to trigger HPA.

---

### 4. Infrastructure & Global Scale
#### Kubernetes (K8s) Configuration
- **Horizontal Pod Autoscaler (HPA)**: Scaled based on `matching_engine_saturation_ratio` (custom metric via Prometheus Adapter) rather than CPU, allowing pre-emptive scaling before the 10 PM peak.
- **PodDisruptionBudget (PDB)**: `minAvailable: 80%` to ensure high availability during node upgrades or failures.

#### Database Sharding (PostgreSQL)
We handle 100TB+ data using **Citus** or manual sharding:
- **Shard Key**: `geo_region_id` (Hash-based) + `created_at` (Time-based partitioning).
- **Optimization**: Use `pg_partman` to automatically manage daily/monthly partitions, moving older data to cheaper storage (S3/GCS) via `Foreign Data Wrappers (FDW)`.

---

### 5. Code Challenge Implementation
The provided Go implementation (see `internal/matching/engine.go`) features:
- **Non-blocking Select**: Prevents goroutine leaks during timeout.
- **Circuit Breaker**: Protects external Routing APIs (OSRM/Google Maps).
- **Haversine Fallback**: Zero-latency fallback when external systems fail.
- **High-Performance Serialization**: Using `segmentio/encoding/json` for minimal GC impact.
