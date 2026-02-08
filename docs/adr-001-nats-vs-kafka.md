---
id: adr-001
title: ADR 001: Choosing NATS JetStream over Kafka
status: accepted
date: 2026-02-08
---

# Context
We needed a distributed messaging system for our Outbox pattern implementation. The requirements were:
1. Low latency for real-time food rescue updates.
2. "At-least-once" delivery guarantee.
3. Lightweight operational complexity (we are a small DevOps team).
4. Go-native ecosystem integration.

# Decision
We chose **NATS JetStream**.

# Consequences
### Positive
*   **Operational Simplicity**: NATS is a single binary 15MB, compared to Kafka's JVM + Zookeeper/KRaft requirement.
*   **Performance**: NATS demonstrates lower p99 latency for our message payload sizes (<4KB).
*   **Subject-Based Addressing**: Allows dynamic wildcard subscriptions (e.g., `food.rescue.*.urgent`), which meshes perfectly with our regional notification logic.

### Negative
*   **Ecosystem**: Smaller ecosystem than Kafka for connectors (e.g., Kafka Connect).
*   **Retention**: Disk retention policies are less granular than Kafka's tiered storage.

# Mitigation
We mitigate the retention issue by archiving cold events to S3 via a dedicated "Archiver Worker".
