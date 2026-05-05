# AP2 Assignment 3 — Medical Scheduling Platform

## What changed compared to Assignment 2

- Both services now use golang-migrate for schema management — no raw DDL in application code
- Connection strings read from environment variables 
- After each successful write, services publish a domain event to NATS
- New Notification Service subscribes to all three event subjects and logs structured JSON

## Broker Choice: NATS 

Chosen: NATS Core

Reason: simpler setup (single binary, no Docker required for dev), zero configuration, fire-and-forget pub/sub is sufficient for a notification service that only logs events. RabbitMQ would be needed if guaranteed delivery, persistent queues, or dead-letter handling were required in production.

## Environment Variables

### Doctor Service

DATABASE_URL=host=localhost port=5432 user=postgres password=postgres dbname=doctor sslmode=disable 
NATS_URL=nats://localhost:4222
grpcPORT=50051

### Appointment Service

DATABASE_URL=host=localhost port=5432 user=postgres password=postgres dbname=appointment sslmode=disable 
NATS_URL=nats://localhost:4222 
DOCTOR_SERVICE_ADDR=localhost:50051 

### Notification Service

NATS_URL=nats://localhost:4222

## Migration Instructions

Migrations run automatically on service startup.

To run manually using golang-migrate CLI:

# Apply
migrate -path ./doctor-service/migrations -database "postgres://postgres:postgres@localhost:5432/doctor?sslmode=disable" up

# Rollback
migrate -path ./doctor-service/migrations -database "postgres://postgres:postgres@localhost:5432/doctor?sslmode=disable" down
`

## Service Startup Order

1. Start PostgreSQL and NATS first
2. Start Doctor Service 
3. Start Appointment Service
4. Start Notification Service

# Terminal 1

cd ..
cd nats-server
./nats-server.exe

# Terminal 2
cd doctor-service 
go run ./cmd/doctorService

# Terminal 3
cd appointment-service 
go run ./cmd/appointmentService 

# Terminal 4
cd notification-service 
go run ./cmd/notificationService


## Event Contract

doctors.created - Doctor Service - CreateDoctor success - event_type, occurred_at, id, full_name, specialization, email 
appointments.created - Appointment Service - CreateAppointment success - event_type, occurred_at, id, title, doctor_id, status 
appointments.status_updated - Appointment Service - UpdateAppointmentStatus success - event_type, occurred_at, id, old_status, new_status 

## grpcurl Commands + Expected Notification Logs

# By cmd
cmd

# 1. Doctor service

## Create doctor
grpcurl -plaintext -proto doctor-service/proto/doctor.proto -d "{\"full_name\":\"Asylbek\",\"specialization\":\"Dentist\",\"email\":\"asylbest@gmail.com\"}" localhost:50051 doctor.DoctorService/CreateDoctor

## Get doctor by id
grpcurl -plaintext -proto doctor-service/proto/doctor.proto -d "{\"id\":\"ID_OF_DOCTOR\"}" localhost:50051 doctor.DoctorService/GetDoctor

## Get list of doctors
grpcurl -plaintext -proto doctor-service/proto/doctor.proto -d "{}" localhost:50051 doctor.DoctorService/ListDoctors

# 2. Appointment service

## Create appointment
grpcurl -plaintext -proto appointment-service/proto/appointment.proto -d "{\"title\":\"Teath carries\",\"description\":\"First visit\",\"doctor_id\":\"ID_OF_DOCTOR\"}" localhost:50052 appointment.AppointmentService/CreateAppointment

## Get appointment by id
grpcurl -plaintext -proto appointment-service/proto/appointment.proto -d "{\"id\":\"ID_OF_APPOINTMENT\"}" localhost:50052 appointment.AppointmentService/GetAppointment

## Get list of all appointments
grpcurl -plaintext -proto appointment-service/proto/appointment.proto -d "{}" localhost:50052 appointment.AppointmentService/ListAppointments

## Updata status of appointment
grpcurl -plaintext -proto appointment-service/proto/appointment.proto -d "{\"id\":\"ID_OF_APPOINTMENT\",\"status\":\"in_progress\"}" localhost:50052 appointment.AppointmentService/UpdateAppointmentStatus

## Consistency Trade-offs

Because broker publishing is best-effort, a process crash between DB commit and publish will losethe event. 
Solutions:
- Outbox Pattern: write events to a DB table in the same transaction, then a background worker publishes them
- NATS JetStream: persistent, at-least-once delivery with acknowledgements
- RabbitMQ publisher confirms: broker confirms each message was persisted before ack

## NATS vs RabbitMQ

### NATS Core   
Persistence: None — fire-and-forget 
Setup: Single binary, zero config 
Delivery guarantee: At most once 
Use case: Stateless notifications, low-latency 

### RabbitMQ
Persistence: Queue-level durability with disk persistence 
Setup: Docker image, management UI, exchanges + queues config 
Delivery guarantee: At least once (with publisher confirms + consumer acks) 
Use case: Financial transactions, critical event pipelines 

Choose NATS when: simplicity and speed matter, losing occasional events is acceptable.  
Choose RabbitMQ when: guaranteed delivery is required (payments, audit logs, compliance).

# Architecture Diagram
![Architecture Diagram](image.png)