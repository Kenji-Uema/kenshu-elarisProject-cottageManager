# cottageManager

Owns the cottage catalog, availability calculations, and booking lifecycle.

## What It Does

- serves cottage catalog and availability HTTP endpoints
- creates and cancels bookings in MongoDB
- publishes invoice-generation requests when a booking is created
- consumes payment confirmations and publishes booking confirmations back to the guest communication bus

## Interfaces

- HTTP: `/cottages`, `/cottage/:name`, `/cottage/type/:cottageType/available-dates`, `/cottage/:name/booking`
- probes: `/healthz`, `/readyz`
- Swagger UI: `/swagger/index.html`
- RabbitMQ publishers for invoice generation and booking confirmation
- RabbitMQ consumer for payment confirmations

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
