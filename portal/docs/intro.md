---
sidebar_position: 1
---

# Introduction

Welcome to the **Pahlawan Pangan** technical documentation. 

Pahlawan Pangan is a high-performance food redistribution platform designed for national-scale operation in Indonesia. Built with Go and a suite of distributed systems tools, it handles the complex logistics of connecting food surplus with those in need.

## Core Mission

Our mission is to achieve **Zero Hunger** and **Zero Waste** by bridging the gap between surplus providers and NGOs/citizens through real-time, low-latency matching.

## Key Pillars

- **Scalability**: Designed for 10M+ daily transactions.
- **Reliability**: Transactional Outbox and NATS JetStream ensure exactly-once processing.
- **Intelligence**: Pahlawan-AI predicts waste before it happens.
- **Community**: RT/RW-based group buys to democratize access.

## Technical Stack

- **Backend**: Go 1.23+
- **Geo-Spatial**: S2 Geometry + PostGIS + Redis Cluster
- **Events**: NATS JetStream
- **Observability**: OpenTelemetry + Prometheus + Jaeger
- **Infrastructure**: Kubernetes (HPA, PDB, StatefulSets)

## Getting Started

To explore the API, check out our [API Reference](/api). For local development, refer to the `README.md` in the root repository.
