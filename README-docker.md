# Docker Setup

Docker Compose runs the backing services for Pay Your Dues locally: PostgreSQL, RabbitMQ, and LocalStack (S3).

The Go applications (`cmd/server` and `cmd/notification-worker`) run on the host and connect to these containers.

## Quick Start

### Start all services

```bash
docker compose up -d
```

### Start only what notifications need

```bash
docker compose up -d postgres rabbitmq
```

### Stop services

```bash
docker compose down
```

### Stop and remove all data

```bash
docker compose down -v
```

## Services

| Service | Container | Ports | Purpose |
|---------|-----------|-------|---------|
| PostgreSQL | `pay-your-dues-postgres` | `5432` | Application database |
| RabbitMQ | `pay-your-dues-rabbitmq` | `5672`, `15672` | Notification message queue |
| LocalStack | `pay-your-dues-localstack` | `4566` | Local S3 for receipt uploads |

## Configuration

### PostgreSQL

- **Host**: `localhost`
- **Port**: `5432`
- **Database**: `pay_your_dues`
- **Username**: `postgres`
- **Password**: `postgres`

Connection string:

```
postgresql://postgres:postgres@localhost:5432/pay_your_dues
```

Match these values in your `.env`:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=pay_your_dues
DB_SSL_MODE=disable
```

### RabbitMQ

- **AMQP URL**: `amqp://guest:guest@localhost:5672/`
- **Management UI**: [http://localhost:15672](http://localhost:15672) (login: `guest` / `guest`)

Match in your `.env`:

```env
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
RABBITMQ_EXCHANGE=notifications.topic
RABBITMQ_SEND_QUEUE=notifications.send
```

Queues and the topic exchange are created automatically when the API or notification worker starts.

### LocalStack (S3)

Used for receipt photo storage in development. Requires additional S3 variables in `.env` — see `env.example`.

## Running the application

After Docker services are up:

**Terminal 1 — API:**

```bash
go run cmd/server/main.go
```

**Terminal 2 — Notification worker:**

```bash
go run cmd/notification-worker/main.go
```

| Endpoint | URL |
|----------|-----|
| API | `http://localhost:8080` |
| API health | `http://localhost:8080/health` |
| Worker health | `http://localhost:8081/health` |
| RabbitMQ UI | `http://localhost:15672` |

## How notifications flow

```
API (cmd/server)
  → writes pending notifications to PostgreSQL
  → publishes immediate jobs to RabbitMQ

Notification worker (cmd/notification-worker)
  → scheduler: claims due rows from PostgreSQL → publishes to RabbitMQ
  → consumer: reads from RabbitMQ → sends email/SMS/webhooks → updates status in PostgreSQL
```

Both processes must be running for end-to-end notification delivery.

## Customization

### Environment variables

Edit the `environment` section in `docker-compose.yml` to change database credentials:

```yaml
environment:
  POSTGRES_DB: your_database_name
  POSTGRES_USER: your_username
  POSTGRES_PASSWORD: your_password
```

### Initialization script

Edit `init.sql` to add database extensions, schemas, or seed data on first PostgreSQL startup.

## Troubleshooting

### Check container status

```bash
docker compose ps
```

### View logs

```bash
docker compose logs postgres
docker compose logs rabbitmq
docker compose logs localstack
```

### Connect to PostgreSQL

```bash
docker compose exec postgres psql -U postgres -d pay_your_dues
```

### Verify RabbitMQ is accepting connections

Open the management UI at `http://localhost:15672` or check the **Queues** tab for `notifications.send` after starting the worker.

### Reset database

```bash
docker compose down -v
docker compose up -d postgres rabbitmq
```

### API starts but notifications are not sent

1. Confirm RabbitMQ is running: `docker compose ps rabbitmq`
2. Confirm the notification worker is running: `curl http://localhost:8081/health`
3. Confirm `RABBITMQ_URL` in `.env` points to `localhost:5672`
4. Check worker logs for SMTP/Twilio/webhook configuration errors
