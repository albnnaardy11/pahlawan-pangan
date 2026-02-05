# Pahlawan Pangan - The Food Redistribution Engine

This is a high-performance food redistribution platform. It is designed to
actually work at scale (10M+ transactions/day, for those who care about 
benchmarks) without falling over or making your database cry.

## What is it?

Pahlawan Pangan is a geo-spatial matching engine written in Go. Its job is to 
connect surplus food (from restaurants, hotels, etc.) with people and 
organizations (NGOs, citizens) who actually need it.

Unlike many "prototypes" that claim to solve food waste, this one handles 
the boring but critical stuff: distributed consistency, race conditions in
geo-spatial claims, and real-time logistics. It’s built for Indonesia-wide 
scale, covering 38 provinces from day one.

## Why use it?

If you want a system that:
1.  Handles 10M+ daily transactions with sub-second latency.
2.  Actually understands geography (S2 Geometry, not just flat math).
3.  Doesn't lose data when your network flakes out (Transactional Outbox).
4.  Predicts waste before it happens using AI, because reacting is too slow.
5.  Includes a dynamic pricing engine, because money matters.

Then this might be for you. If you want a slow, centralized PHP app that 
breaks when 100 people use it, look elsewhere.

## Requirements

If you can't run these, you're doing it wrong:
- Go 1.23+ (Current stable toolchain)
- Docker & Docker Compose (For the infra stack)
- Kubernetes (For when you actually go live)
- PostGIS (Because standard SQL isn't enough for maps)

## Getting Started

I like `make`. You should too.

```bash
# Get the infra up (Postgres, Redis, NATS)
docker-compose up -d

# Build and run the thing
make run
```

If you don't have `make`, you can use `go run cmd/server/main.go`, but fix 
your environment.

## The "Super-App" Bits (Unicorn stuff)

I've added features that most apps miss:
- **Pahlawan-Market**: Dynamic pricing using exponential decay. Prices drop 
  automatically as food approaches expiry. It works.
- **Pahlawan-AI**: Predictive analytics. It looks at weather and history to 
  tell a restaurant they'll have 15kg of waste before they even tahu.
- **Pahlawan-Express**: Logistic hooks for courier services.
- **Pahlawan-Comm**: Group buying for neighborhoods (RT/RW) to kill delivery
  fees.

## Technical Specs (The real meat)

- **Sharding**: S2 Geometry Level 13. High granularity, zero hotspots.
- **Messaging**: NATS JetStream. Fast, durable, and doesn't suck like rabbit.
- **Telemetry**: OpenTelemetry throughout. If it's slow, you'll see why in 
  Jaeger.
- **Resilience**: Circuit Breakers with Haversine fallbacks. The system stays
  up even when the routing API goes down.

## Contributing

Don't send me crap. Write tests. Ensure `go fmt` is happy. If you break the 
matching logic, I'll probably revert your PR without reading it.

## License

MIT. Do whatever you want with it, just don't blame me if you use it 
wrong.

---
*Pahlawan Pangan: It's just code. But it's code that actually works.*
