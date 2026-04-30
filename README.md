# AP2 Assignment 3 — Medical Scheduling Platform

## What changed compared to Assignment 2

- Both services now use **golang-migrate** for schema management — no raw DDL in application code
- Connection strings read from environment variables (no hardcoding)
- After each successful write, services publish a domain event to **NATS**
- New **Notification Service** subscribes to all three event subjects and logs structured JSON

## Broker Choice: NATS (Core)

**Chosen: NATS Core**

Reason: simpler setup (single binary, no Docker required for dev), zero configuration, fire-and-forget
pub/sub is sufficient for a notification service that only logs events. RabbitMQ would be needed
if guaranteed delivery, persistent queues, or dead-letter handling were required in production.

## Environment Variables

### Doctor Service
| Variable | Default | Description |
|---|---|---|
| DATABASE_URL | host=localhost port=5432 user=postgres password=postgres dbname=doctor sslmode=disable | PostgreSQL DSN |
| NATS_URL | nats://localhost:4222 | NATS broker URL |

### Appointment Service
| Variable | Default | Description |
|---|---|---|
| DATABASE_URL | host=localhost port=5432 user=postgres password=postgres dbname=appointment sslmode=disable | PostgreSQL DSN |
| NATS_URL | nats://localhost:4222 | NATS broker URL |
| DOCTOR_SERVICE_ADDR | localhost:50051 | Doctor Service gRPC address |

### Notification Service
| Variable | Default | Description |
|---|---|---|
| NATS_URL | nats://localhost:4222 | NATS broker URL |

## Infrastructure Setup

```bash
# Start PostgreSQL
docker run -d --name postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 \
  postgres:16

# Create databases
docker exec -it postgres psql -U postgres -c "CREATE DATABASE doctor;"
docker exec -it postgres psql -U postgres -c "CREATE DATABASE appointment;"

# Start NATS
docker run -d --name nats -p 4222:4222 nats:latest
```

## Migration Instructions

Migrations run automatically on service startup.

To run manually using golang-migrate CLI:
```bash
# Apply
migrate -path ./doctor-service/migrations -database "postgres://postgres:postgres@localhost:5432/doctor?sslmode=disable" up

# Rollback
migrate -path ./doctor-service/migrations -database "postgres://postgres:postgres@localhost:5432/doctor?sslmode=disable" down
```

## Service Startup Order

1. Start PostgreSQL and NATS first
2. Start Doctor Service (it must be up before Appointment Service connects)
3. Start Appointment Service
4. Start Notification Service

```bash
# Terminal 1
cd doctor-service && go run ./cmd/doctor-service

# Terminal 2
cd appointment-service && go run ./cmd/appointment-service  # wait for doctor-service

# Terminal 3
cd notification-service && go run ./cmd/notification-service
```

## Event Contract

| Subject | Publisher | Trigger | Fields |
|---|---|---|---|
| doctors.created | Doctor Service | CreateDoctor success | event_type, occurred_at, id, full_name, specialization, email |
| appointments.created | Appointment Service | CreateAppointment success | event_type, occurred_at, id, title, doctor_id, status |
| appointments.status_updated | Appointment Service | UpdateAppointmentStatus success | event_type, occurred_at, id, old_status, new_status |

## grpcurl Commands + Expected Notification Logs

```bash
# 1. Create a doctor
grpcurl -plaintext -d '{"full_name":"Dr. Aisha Seitkali","specialization":"Cardiology","email":"a.seitkali@clinic.kz"}' \
  localhost:50051 doctor.DoctorService/CreateDoctor

# Notification Service prints:
{"time":"2026-05-01T10:23:44Z","subject":"doctors.created","event":{"event_type":"doctors.created","occurred_at":"...","id":"<uuid>","full_name":"Dr. Aisha Seitkali","specialization":"Cardiology","email":"a.seitkali@clinic.kz"}}

# 2. Create an appointment (use the id from step 1)
grpcurl -plaintext -d '{"title":"Initial cardiac consultation","description":"First visit","doctor_id":"<uuid from step 1>"}' \
  localhost:50052 appointment.AppointmentService/CreateAppointment

# Notification Service prints:
{"time":"...","subject":"appointments.created","event":{"event_type":"appointments.created","occurred_at":"...","id":"<uuid>","title":"Initial cardiac consultation","doctor_id":"<uuid>","status":"new"}}

# 3. Update appointment status (use appointment id from step 2)
grpcurl -plaintext -d '{"id":"<appointment uuid>","status":"in_progress"}' \
  localhost:50052 appointment.AppointmentService/UpdateAppointmentStatus

# Notification Service prints:
{"time":"...","subject":"appointments.status_updated","event":{"event_type":"appointments.status_updated","occurred_at":"...","id":"<uuid>","old_status":"new","new_status":"in_progress"}}
```

## Consistency Trade-offs

Because broker publishing is **best-effort**, a process crash between DB commit and publish will lose
the event. Solutions:
- **Outbox Pattern**: write events to a DB table in the same transaction, then a background worker publishes them
- **NATS JetStream**: persistent, at-least-once delivery with acknowledgements
- **RabbitMQ publisher confirms**: broker confirms each message was persisted before ack

## NATS vs RabbitMQ

| | NATS Core | RabbitMQ |
|---|---|---|
| Persistence | None — fire-and-forget | Queue-level durability with disk persistence |
| Setup | Single binary, zero config | Docker image, management UI, exchanges + queues config |
| Delivery guarantee | At most once | At least once (with publisher confirms + consumer acks) |
| Use case | Stateless notifications, low-latency | Financial transactions, critical event pipelines |

Choose NATS when: simplicity and speed matter, losing occasional events is acceptable.  
Choose RabbitMQ when: guaranteed delivery is required (payments, audit logs, compliance).