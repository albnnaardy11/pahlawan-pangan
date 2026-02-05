---
sidebar_position: 1
---

# Security & Compliance

Pahlawan Pangan is built with **Tier-1 Tech Giant** security standards to protect user data and ensure the integrity of the food rescue ecosystem.

## Zero Trust Architecture

Our infrastructure follows the Zero Trust principle:

1.  **mTLS (Mutual TLS)**: All internal communication between microservices (e.g., API Gateway to Matching Engine) is encrypted using mTLS via the Kubernetes Service Mesh.
2.  **JWT Authentication**: Secure API access using RS256 signatures.
3.  **Role-Based Access (RBAC)**: Fine-grained permissions for Providers, NGOs, Citizens, and Couriers.

## Data Protection & PII

- **Encryption at Rest**: All PostgreSQL data is encrypted using AES-256.
- **PII Hashing**: Personal Identifiable Information (Names, Phone Numbers) is obfuscated in audit logs.
- **Geography Privacy**: Citizen locations are "fuzzed" in analytics to prevent exact tracking while maintaining matching accuracy.

## Resiliency Headers

Service-to-service calls are protected by:
- **Rate Limiting**: Per-API-Key quotas managed at the gateway.
- **Circuit Breakers**: Automatic failover to stale cache if the primary matching engine is unreachable.
- **Timeouts**: Strict 1s p99 latency targets enforced.

```go
// Internal Security Middleware Example
r.Use(middleware.Timeout(30 * time.Second))
r.Use(cors.Handler(cors.Options{
    AllowedOrigins: []string{"https://*"},
    AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
}))
```
