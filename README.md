# Webhook Delivery Platform

A multi-tenant webhook delivery SaaS built in Go. Reliably delivers webhooks with automatic retries, HMAC-SHA256 signing, delivery tracking, and structured observability. Uses Amazon SQS for durable job queuing and PostgreSQL for storage.

**Architecture is designed so webhooks are the only delivery channel today, but adding new channels (SQS, Kafka, email) requires only implementing the `delivery.Channel` interface.**

## Architecture

```
Event Ingestion API → Event Store → Router → Delivery Jobs → SQS Queue → Workers → Webhook Endpoints

             ┌─────────────┐
             │   Customer   │
             └──────┬───────┘
                    │
           POST /v1/events
          (Bearer API_KEY)
                    │
            ┌───────▼────────┐
            │   API Server   │  ← chi router, auth middleware
            │  (cmd/api)     │
            └───────┬────────┘
                    │
         ┌──────────┼──────────┐
         │          │          │
    1. Validate  2. Store   3. Route
       Tenant      Event      Event
                    │          │
                    ▼          ▼
              ┌──────────┐  ┌──────────────┐
              │PostgreSQL │  │   Router     │  ← subscription matching
              │  events   │  │ exact/wild/* │
              └──────────┘  └──────┬───────┘
                                   │
                          4. Create Deliveries
                          5. Enqueue SQS Jobs
                                   │
                            ┌──────▼───────┐
                            │  Amazon SQS  │
                            │   Queue      │
                            └──────┬───────┘
                                   │
                        ┌──────────▼──────────┐
                        │   Delivery Workers  │  ← cmd/worker
                        │  (N goroutines)     │
                        └──────────┬──────────┘
                                   │
                          delivery.Channel
                          interface dispatch
                                   │
                        ┌──────────▼──────────┐
                        │  WebhookChannel     │  ← HTTP POST + HMAC signing
                        │  (future: SQS,      │
                        │   Kafka, Email...)   │
                        └──────────┬──────────┘
                                   │
                            ┌──────▼───────┐
                            │   Customer   │
                            │   Endpoint   │
                            └──────────────┘
```

## Project Structure

```
├── cmd/
│   ├── api/main.go          # API server entry point
│   └── worker/main.go       # Delivery worker entry point
├── main.go                  # Combined API + worker (dev convenience)
├── internal/
│   ├── config/              # Environment-based configuration
│   ├── db/                  # PostgreSQL pool, migrations, data access (store)
│   ├── delivery/            # DeliveryChannel interface
│   ├── events/              # HTTP API handlers
│   ├── models/              # Core domain types
│   ├── queue/               # Amazon SQS wrapper
│   ├── router/              # Event → subscription matching → delivery creation
│   └── webhooks/            # WebhookChannel implementation (HMAC-signed HTTP POST)
├── migrations/              # SQL migration files
├── scripts/                 # LocalStack init scripts
├── Dockerfile               # Multi-binary Docker build
├── docker-compose.yml       # API + Worker + Postgres + LocalStack
└── Makefile                 # Build/run/test/deploy targets
```

## API Endpoints

### Create Tenant
```
POST /v1/tenants
{"name": "Acme Corp"}
→ {"id": "...", "name": "Acme Corp", "api_key": "whk_..."}
```

### Register Webhook Endpoint
```
POST /v1/webhooks
Authorization: Bearer <API_KEY>
{"url": "https://example.com/webhook", "events": ["invoice.paid", "invoice.*"]}
→ {"id": "...", "url": "...", "secret": "whsec_...", "created_at": "..."}
```

### Publish Event
```
POST /v1/events
Authorization: Bearer <API_KEY>
{"event_type": "invoice.paid", "data": {"invoice_id": "inv_123", "amount": 9900}}
→ {"id": "...", "event_type": "invoice.paid", "deliveries": 2}
```

## Subscription Matching

The router supports three subscription patterns:

| Pattern | Example | Matches |
|---------|---------|---------|
| Exact | `invoice.paid` | Only `invoice.paid` |
| Wildcard | `invoice.*` | `invoice.paid`, `invoice.created`, etc. |
| Global | `*` | Every event |

Matching is done via an efficient single SQL query with indexed lookups.

## Delivery & Retries

Retry schedule with exponential backoff:

| Attempt | Delay |
|---------|-------|
| 1 | Immediate |
| 2 | 30 seconds |
| 3 | 2 minutes |
| 4 | 10 minutes |
| 5 | 1 hour |

After max attempts, delivery moves to `dead_letter` status.

SQS `DelaySeconds` is used for retries. For delays > 15 minutes, re-enqueue with remaining delay.

## Webhook Request Format

```
POST <endpoint_url>
Content-Type: application/json
X-Webhook-Id: <event_id>
X-Webhook-Event: invoice.paid
X-Webhook-Timestamp: <unix_timestamp>
X-Webhook-Signature: v1=<hmac_sha256_hex>

{"id": "evt_...", "type": "invoice.paid", "created_at": "...", "data": {...}}
```

Signature: `HMAC-SHA256(secret, timestamp + payload)`

## Quick Start

```bash
# Start infrastructure
make infra-up

# Run combined API + worker
make run

# Or run separately
make run-api
make run-worker

# Run with Docker
make docker-up
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | API server port |
| `DATABASE_URL` | `postgres://...localhost/webhooks` | PostgreSQL DSN |
| `SQS_REGION` | `us-east-1` | AWS region |
| `SQS_QUEUE_URL` | `http://localhost:4566/...` | SQS queue URL |
| `SQS_ENDPOINT_URL` | `` | Custom SQS endpoint (LocalStack) |
| `SQS_ACCESS_KEY` | `` | Static AWS access key |
| `SQS_SECRET_KEY` | `` | Static AWS secret key |
| `WORKER_CONCURRENCY` | `10` | Number of worker goroutines |
| `EVENT_RETENTION_DAYS` | `30` | Days to retain events |

## Key Architectural Decisions

1. **Separate subscriptions table** — Enables efficient indexed lookups (exact, wildcard prefix, global) without scanning all endpoints.

2. **Delivery channel interface** — Workers dispatch via `delivery.Channel` rather than hardcoding HTTP. Adding a new channel (SQS, Kafka, email) only requires implementing the interface and registering it in the channel map.

3. **SQS for job queuing** — Durable, scalable, and supports native `DelaySeconds` for retries. Messages contain only IDs; workers hydrate from the database for consistency.

4. **Separate API/worker binaries** — Can scale API servers and workers independently. Combined binary available for convenience.

5. **30-day event retention** — Background cleanup job runs hourly, using cascade deletes to clean up deliveries and attempts.
