---
sidebar_position: 1
---

# Spatial Consistency at Scale

Pahlawan Pangan uses a **Geo-Sharded Actor Model** to manage hundreds of thousands of concurrent food surplus claims across Indonesia without data corruption or "double-claiming."

## The S2 Geometry Engine

Instead of standard latitude/longitude bounding boxes which suffer from "poles distortion" and performance degradation at scale, we utilize **Google's S2 Geometry Library**.

- **Sharding Level**: 13 (approx. 1.2km² per cell).
- **Indexing**: All surplus items are indexed in **Redis GEORADIUS** for sub-10ms retrieval.
- **Consistency**: Each S2 cell acts as a logical "actor." Claims within a cell are processed with optimistic concurrency control.

## Distributed Locking Strategy

To prevent the "Thundering Herd" problem when a 90% discount is announced at a popular bakery:

1.  **Optimistic Lock**: The system attempts to claim the item in the database with a conditional `WHERE status = 'available'`.
2.  **Idempotency Key**: Every claim requires a unique client-generated key (UUID v7) to prevent duplicate processing on network retries.
3.  **Atomic Decrement**: Stock levels are decremented atomically using PostgreSQL `Check` constraints.

## Implementation Details

The core matching logic is isolated in `internal/matching/engine.go`, utilizing a non-blocking worker pool to ensure high throughput even under extreme load (10M+ TPS).

---
*Reference: [ARCHITECTURE.md](../../ARCHITECTURE.md) for full system specifications.*
