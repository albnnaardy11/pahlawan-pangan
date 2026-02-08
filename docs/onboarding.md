---
id: onboarding
title: 🚀 Onboarding Guide
---

# Welcome to Pahlawan Pangan Engineering!

Follow this guide to ship your first PR in under 30 minutes.

## 1. Prerequisites
Ensure you have:
*   Go 1.22+
*   Docker & Docker Compose
*   Make

## 2. Setup Environment
```bash
git clone https://github.com/albnnaardy11/pahlawan-pangan.git
cd pahlawan-pangan
make setup  # Installs hooks and dependencies
docker-compose up -d # Spins up Postgres, NATS, Redis
```

## 3. Run the Playground
We don't just run tests; we run simulations.
```bash
go test -v ./tests/pentest_simulation_test.go
```
If you see "SYSTEM SECURE", you are ready.

## 4. Architecture orientation
*   **Hot Path**: Check `internal/auth/usecase/auth_usecase.go`. Notice the `unsafe` conversions.
*   **Data Integrity**: Check `internal/outbox`. This is our "Holy Grail" of consistency.

## 5. Your First Task
Find a "Good First Issue" tagged with `complexity:low`. Usually, it involves adding a new structured log field or updating a metric in `internal/middleware`.

**Happy Coding!**
