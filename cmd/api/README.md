

## API Endpoints

### Create Tenant

POST /v1/tenants

Returns tenant id and API key.

---

### Register Webhook Endpoint

POST /v1/webhooks

Body:

{
"url": "[https://example.com/webhook](https://example.com/webhook)",
"events": ["invoice.paid", "invoice.*"]
}

Response should include:

id
url
secret
created_at

The secret will be used to sign webhook requests.

---

### Publish Event

POST /v1/events

Headers:

Authorization: Bearer TENANT_API_KEY

Body:

{
"event_type": "invoice.paid",
"data": { ... }
}

The API must:

1. validate tenant
2. store event
3. trigger router
4. enqueue delivery jobs

---

## Router Design (Important)

The router must efficiently determine which subscriptions match an event.

Avoid scanning the entire subscription table.

Implement a strategy that supports fast lookup:

Maintain indexed subscription groups such as:

exact event subscriptions
wildcard prefix subscriptions
global subscriptions

Suggested approach:

1. exact match lookup
2. prefix wildcard lookup
3. global subscriptions

Example for event "invoice.paid":

match:

invoice.paid
invoice.*
*

Implement database indexes to support this.

Router must return a list of endpoint_ids.

For each endpoint:

create a delivery record and enqueue a delivery job.