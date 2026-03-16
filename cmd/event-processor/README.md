

## Queue System

Uses Amazon SQS.

Queue messages should contain only identifiers:

{
"event_id": "...",
"endpoint_id": "...",
"attempt": 1
}

Workers fetch event + endpoint from database.

---

## Delivery Worker

Workers must:

1. read job from SQS
2. load event and endpoint
3. send HTTP POST to endpoint
4. record result
5. retry on failure using exponential backoff
6. move to dead-letter state after max attempts

Retry schedule:

attempt 1 immediate
attempt 2 30 seconds
attempt 3 2 minutes
attempt 4 10 minutes
attempt 5 1 hour

Use SQS delay for retries.

---

## Webhook Delivery

Webhook request format:

POST endpoint_url

Headers:

Content-Type: application/json
X-Webhook-Id
X-Webhook-Event
X-Webhook-Timestamp
X-Webhook-Signature

Signature must be HMAC SHA256:

HMAC(secret, timestamp + payload)

Payload format:

{
"id": "evt_xxx",
"type": "invoice.paid",
"created_at": "...",
"data": { ... }
}

Success is defined as any 2xx response.

---

## Delivery Channel Abstraction

Even though only webhooks are supported now, the system must be built using a **delivery channel interface**.

Example:

type DeliveryChannel interface {
Deliver(ctx context.Context, event Event, endpoint Endpoint) error
}

Implement:

WebhookChannel

Future implementations could include:

QueueChannel
EmailChannel
KafkaChannel

Workers should call the interface rather than hardcoding HTTP logic.

---

## Storage and Retention

Events must be stored for 30 days.

Implement a scheduled cleanup job to remove old events.

---

## Observability

Include structured logging.

Track:

delivery attempts
latency
failures
failures