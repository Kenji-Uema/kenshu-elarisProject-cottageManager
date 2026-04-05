# cottageManager

Owns cottages, availability, and bookings.

## Responsibilities

- expose HTTP APIs for cottages and availability
- persist bookings
- attach bookings to cottages
- publish invoice creation requests
- react to payment confirmation events and confirm bookings

## Interfaces

- HTTP API for cottage and booking operations
- RabbitMQ publisher for create-invoice requests
- RabbitMQ consumer for payment confirmations
- RabbitMQ publisher for booking confirmation notifications

## Run

```sh
go run .
```

## Build

```sh
make build
make docker-build
```

## Configuration

Configuration is environment-driven. See:

- `internal/config/config.go`
- `internal/config/rabbitmq_config.go`

Important families:

- service HTTP: `SERVICE_*`, timeout settings
- MongoDB: `MONGO_*`
- RabbitMQ: `RABBITMQ_*`
- publishers: `CREATE_INVOICE_*`, `BOOKING_CONFIRMATION_*`
- consumer: `PAYMENT_CONFIRMED_*`
- telemetry: `OTEL_EXPORTER_OTLP_*`

## Docs

- Swagger assets live in `docs/`

## Entry points

- `main.go`
- `internal/app/booking_service.go`
- `internal/app/communication_service.go`
