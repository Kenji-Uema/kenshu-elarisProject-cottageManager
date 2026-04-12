# Cottage Manager

Owns the cottage catalog, availability calculations, and booking lifecycle.

## Main Docs

See the main project documentation: <https://github.com/Kenji-Uema/kenshu-elarisProject-docs>

## What It Does

- serves cottage catalog and availability HTTP endpoints
- creates and cancels bookings in MongoDB
- publishes invoice-generation requests when a booking is created
- consumes payment confirmations and publishes booking confirmations back to the guest communication bus

## RabbitMQ Specification

This service declares two exchanges and one consumer queue at startup.

### Publisher: create invoice

- exchange env: `CREATE_INVOICE_EXCHANGE_*`
- routing key: `booking.create-invoice`
- exchange defaults from env:
  - name: `ex.invoice.generate`
  - kind: `direct`
- serialization: raw protobuf
- AMQP `content_type`: `application/protobuf`
- AMQP header `message_type`: `cottageManager.invoice.CreateInvoicePaymentRequest`
- delivery mode: persistent

Message example:

```json
{
  "idempotencyKey": "69d56ce351d7f8bbb753b70b",
  "bookingId": "69d56ce351d7f8bbb753b70b",
  "payerId": "69d56ce3543361444126758c",
  "issuedAt": "2026-04-07T20:45:23Z",
  "dueAt": "2026-04-09T20:45:23Z",
  "payer": {
    "name": "Mohamed Wu",
    "email": "Mohamed.Wu@test.com",
    "documentNumber": "449379519",
    "billingAddress": "8806 South Canyonshire, Apt 581, Colorado Springs, Nevada 51093"
  },
  "booking": {
    "cottageName": "Barbara Karst",
    "nights": 3,
    "numberOfGuests": 2,
    "valuePerNight": {
      "amount": "20000",
      "currency": "USD"
    }
  },
  "total": {
    "amount": "60000",
    "currency": "USD"
  },
  "taxTotal": {
    "amount": "3000",
    "currency": "USD"
  },
  "discountTotal": {
    "amount": "0",
    "currency": "USD"
  }
}
```

Notes:

- `amount` fields are minor units, so `20000` means `USD 200.00`.
- `dueAt` is `48h` after `issuedAt`.
- `taxTotal` is `5%` of `total`.
- `idempotencyKey` is the booking ID.

### Consumer: payment confirmed

- queue env: `PAYMENT_CONFIRMED_QUEUE_*`
- binding env: `PAYMENT_CONFIRMED_BINDING_*`
- consume env: `PAYMENT_CONFIRMED_CONSUME_*`
- queue default name: `q.cottage-manager.payment-confirmed`
- binding exchange default: `ex.payment`
- binding routing key default: `booking.*.confirmation`
- serialization expected by the service: raw protobuf
- expected AMQP `content_type`: `application/protobuf`
- expected AMQP header `message_type`: `cottageManager.payment.v1.PaymentConfirmation`
- acknowledgements: manual ack/nack

Consumed message example:

```json
{
  "id": "pay_01jrj0m4xw7xj2l6v0w9m3r9af",
  "bookingId": "69d56ce351d7f8bbb753b70b",
  "payerId": "69d56ce3543361444126758c",
  "invoiceNumber": "INV-20250410-99c18f",
  "receiptNumber": "RCPT-20250410-7a1c2e",
  "confirmedAt": "2026-04-07T20:45:23Z"
}
```

Consumer behavior:

- invalid protobuf payload: `nack(requeue=false)`
- invalid `bookingId`: `nack(requeue=false)`
- transient repository or publish failure: `nack(requeue=true)`
- success path:
  - update booking status to `CONFIRMED`
  - load the booking from MongoDB
  - publish a booking confirmation event for the guest
  - `ack`

### Publisher: booking confirmation

- exchange env: `BOOKING_CONFIRMATION_EXCHANGE_*`
- routing key pattern: `guest.<guestId>`
- exchange defaults from env:
  - name: `ex.communication`
  - kind: `direct`
- serialization: raw protobuf
- AMQP `content_type`: `application/protobuf`
- AMQP header `message_type`: `cottageManager.invoice.BookingConfirmedNotificationEvent`
- delivery mode: persistent

Published message example:

```json
{
  "id": "pay_01jrj0m4xw7xj2l6v0w9m3r9af",
  "bookingId": "69d56ce351d7f8bbb753b70b",
  "bookingStatus": "BOOKING_STATUS_CONFIRMED",
  "guest": {
    "guestId": "69d56ce3543361444126758c"
  },
  "booking": {
    "cottageName": "Barbara Karst",
    "checkIn": "2025-04-15T00:00:00Z",
    "checkOut": "2025-04-18T00:00:00Z",
    "nights": 3,
    "numberOfGuests": 2
  },
  "confirmedAt": "2026-04-07T20:45:23Z"
}
```

Notes:

- `id` is copied from the payment confirmation event.
- `guest.name`, `guest.email`, `guest.phone`, and `totalPaid` are currently not populated by this service.
- the routing key is built as `guest.` plus the booking `mainGuest` id from MongoDB.

## HTTP Examples

Assume the service is running locally:

```sh
export BASE_URL=http://localhost:8080
```

List all cottages:

```sh
curl -sS "$BASE_URL/cottages"
```

Example `200 OK` response:

```json
[
  {
    "name": "Barbara Karst",
    "type": "seaside",
    "details": {
      "description": "<p>\"Barbara Karst\" is a romantic tropical-style cottage inspired by the vibrant bougainvillea that surrounds it.</p>",
      "view": "seaside",
      "furniture_description": "<p>The interior follows a coastal minimalist aesthetic.</p>",
      "bathroom_description": "<p>The bathroom embraces a spa-like tropical design.</p>",
      "amenities_description": "<p>Guests staying in Barbara Karst enjoy complimentary tropical breakfast served on the private veranda.</p>"
    },
    "photos": [
      "Barbara_Karst_room.png",
      "Barbara_Karst_bathroom.png",
      "Barbara_Karst_view.png"
    ],
    "price_per_night": 200,
    "bookings": []
  }
]
```

Get one cottage by name:

```sh
curl -sS "$BASE_URL/cottage/Barbara%20Karst"
```

Example `200 OK` response:

```json
{
  "name": "Barbara Karst",
  "type": "seaside",
  "details": {
    "description": "<p>\"Barbara Karst\" is a romantic tropical-style cottage inspired by the vibrant bougainvillea that surrounds it.</p>",
    "view": "seaside",
    "furniture_description": "<p>The interior follows a coastal minimalist aesthetic.</p>",
    "bathroom_description": "<p>The bathroom embraces a spa-like tropical design.</p>",
    "amenities_description": "<p>Guests staying in Barbara Karst enjoy complimentary tropical breakfast served on the private veranda.</p>"
  },
  "photos": [
    "Barbara_Karst_room.png",
    "Barbara_Karst_bathroom.png",
    "Barbara_Karst_view.png"
  ],
  "price_per_night": 200,
  "bookings": []
}
```

Filter cottages by view:

```sh
curl -sS "$BASE_URL/cottage/view/seaside"
```

Example `200 OK` response:

```json
[
  {
    "name": "Barbara Karst",
    "type": "seaside",
    "details": {
      "description": "<p>\"Barbara Karst\" is a romantic tropical-style cottage inspired by the vibrant bougainvillea that surrounds it.</p>",
      "view": "seaside",
      "furniture_description": "<p>The interior follows a coastal minimalist aesthetic.</p>",
      "bathroom_description": "<p>The bathroom embraces a spa-like tropical design.</p>",
      "amenities_description": "<p>Guests staying in Barbara Karst enjoy complimentary tropical breakfast served on the private veranda.</p>"
    },
    "photos": [
      "Barbara_Karst_room.png",
      "Barbara_Karst_bathroom.png",
      "Barbara_Karst_view.png"
    ],
    "price_per_night": 200,
    "bookings": []
  },
  {
    "name": "Golden Glow",
    "type": "seaside",
    "details": {
      "description": "<p>Golden Glow captures the warmth and optimism of a tropical sunset.</p>",
      "view": "seaside",
      "furniture_description": "<p>Stepping inside Golden Glow, guests are welcomed by a space that feels naturally illuminated.</p>",
      "bathroom_description": "<p>The bathroom offers a warm, earthy retreat inspired by the sunlit coast.</p>",
      "amenities_description": "<p>Every detail in Golden Glow is designed for slow, sunlit comfort.</p>"
    },
    "photos": [
      "Golden_Glow_room.png",
      "Golden_Glow_bathroom.png",
      "Golden_Glow_view.png"
    ],
    "price_per_night": 200,
    "bookings": []
  }
]
```

Get availability for one cottage in a date range:

```sh
curl -sS "$BASE_URL/cottage/Barbara%20Karst/available-dates?from=2026-05-01&to=2026-05-10"
```

Example `200 OK` response:

```json
{
  "cottage_name": "Barbara Karst",
  "available_periods": [
    {
      "check_in": "2026-05-01T00:00:00Z",
      "check_out": "2026-05-10T00:00:00Z"
    }
  ]
}
```

Get availability for every cottage of a type in a date range:

```sh
curl -sS "$BASE_URL/cottage/type/seaside/available-dates?from=2026-05-01&to=2026-05-10"
```

Example `200 OK` response:

```json
[
  {
    "Name": "Barbara Karst",
    "Periods": [
      {
        "CheckIn": "2026-05-01T00:00:00Z",
        "CheckOut": "2026-05-10T00:00:00Z"
      }
    ]
  },
  {
    "Name": "Juanita Hatten",
    "Periods": [
      {
        "CheckIn": "2026-05-01T00:00:00Z",
        "CheckOut": "2026-05-10T00:00:00Z"
      }
    ]
  }
]
```

Create a booking:

```sh
curl -sS -X POST "$BASE_URL/cottage/Barbara%20Karst/booking" \
  -H "Content-Type: application/json" \
  -d '{
    "mainGuest": "69d56ce3543361444126758c",
    "numberOfGuests": 2,
    "checkInDate": "2025-04-15",
    "checkOutDate": "2025-04-18",
    "guestName": "Mohamed Wu",
    "guestEmail": "Mohamed.Wu@test.com",
    "guestDocument": "449379519",
    "billingAddress": "8806 South Canyonshire, Apt 581, Colorado Springs, Nevada 51093"
  }'
```

Example `200 OK` response:

```json
{
  "message": "Thank you for choosing us. Your booking registered, soon you will receive your invoice",
  "bookingId": "69d56ce351d7f8bbb753b70b",
  "info": {
    "cottageName": "Barbara Karst",
    "numberOfGuests": 2,
    "checkInDate": "2025-04-15",
    "checkOutDate": "2025-04-18"
  }
}
```

Cancel a booking:

```sh
curl -sS -X DELETE "$BASE_URL/cottage/Barbara%20Karst/booking/69d56ce351d7f8bbb753b70b"
```

Use the `bookingId` returned by the create-booking response.

Example `200 OK` response:

```json
{
  "message": "IsValidBooking removed successfully"
}
```

Liveness probe:

```sh
curl -i "$BASE_URL/healthz"
```

Example `200 OK` response:

```http
HTTP/1.1 200 OK
```

Readiness probe:

```sh
curl -i "$BASE_URL/readyz"
```

Example `200 OK` response:

```http
HTTP/1.1 200 OK
```

## Local Commands

```sh
go run .
go build .
go test ./...
make generate
make docker-build
```

## Minimum Env To Start

Optional vars with defaults, such as `SERVICE_NAME`, `VERSION`, collection names, and timeout values, are omitted here.

```sh
SERVICE_HOST=0.0.0.0
SERVICE_PORT=8080

MONGO_INITDB_ROOT_USERNAME=<mongo user>
MONGO_INITDB_ROOT_PASSWORD=<mongo password>
MONGO_HOST=<mongo host>
MONGO_DATABASE=cottages

RABBITMQ_USERNAME=<rabbit user>
RABBITMQ_PASSWORD=<rabbit password>
RABBITMQ_HOST=<rabbit host>
RABBITMQ_PORT=5672

CREATE_INVOICE_EXCHANGE_NAME=ex.invoice.generate
CREATE_INVOICE_EXCHANGE_KIND=direct

BOOKING_CONFIRMATION_EXCHANGE_NAME=ex.communication
BOOKING_CONFIRMATION_EXCHANGE_KIND=direct

PAYMENT_CONFIRMED_QUEUE_NAME=q.cottage-manager.payment-confirmed
PAYMENT_CONFIRMED_BINDING_EXCHANGE_NAME=ex.payment
PAYMENT_CONFIRMED_BINDING_ROUTING_KEY=booking.*.confirmation

OTEL_EXPORTER_OTLP_ENDPOINT=<otel host>
OTEL_EXPORTER_OTLP_GRPC_PORT=4317
OTEL_EXPORTER_OTLP_HEALTH_PORT=13133
OTEL_EXPORTER_OTLP_INSECURE=true
```

## Configuration

Configuration is environment-driven. Start with:

- `internal/config/config.go`
- `internal/config/rabbitmq_config.go`

Important groups:

- service HTTP: `SERVICE_*`
- MongoDB: `MONGO_*`
- RabbitMQ connection: `RABBITMQ_*`
- invoice publisher: `CREATE_INVOICE_*`
- payment consumer: `PAYMENT_CONFIRMED_*`
- guest communication publisher: `BOOKING_CONFIRMATION_*`
- telemetry: `OTEL_EXPORTER_OTLP_*`

## Key Files

- `main.go`
- `internal/app/booking_service.go`
- `internal/app/availability_service.go`
- `internal/app/communication_service.go`
- `internal/transport/http/server.go`
- `docs/`
