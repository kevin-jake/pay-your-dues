# Pay Your Dues - a Go Debt Tracker Application

A comprehensive debt tracking application built with Go, PostgreSQL, and Gin framework. Track your debts, money lent to others, and send notifications via email, SMS, or Facebook Messenger.

## Features

- **User Authentication**: Secure JWT-based authentication
- **Debt Management**: Track money you owe and money owed to you
- **Multiple Debt Lists**: Organize debts into different categories
- **Notifications**: Send reminders via email, SMS, Slack, Telegram, and Discord (processed asynchronously via RabbitMQ)
- **User Settings**: Customize notification preferences and default currency
- **RESTful API**: Clean and well-documented API endpoints

## Tech Stack

- **Backend**: Go 1.24+
- **Framework**: Gin (HTTP web framework)
- **Database**: PostgreSQL
- **Message Queue**: RabbitMQ (notification delivery)
- **Authentication**: JWT tokens
- **Password Hashing**: bcrypt
- **Logging**: Zerolog
- **Hot Reloading**: Air

## Prerequisites

- Go 1.24 or higher
- PostgreSQL 12 or higher
- RabbitMQ 3.x (for notification delivery)
- Docker and Docker Compose (recommended for local PostgreSQL and RabbitMQ)
- Air (optional, for API hot reloading)

## Installation

### 1. Clone the repository

```bash
git clone <repository-url>
cd pay-your-dues
```

### 2. Install dependencies

```bash
go mod download
```

### 3. Install Air (for development)

```bash
go install github.com/air-verse/air@latest
```

### 4. Set up PostgreSQL

Create a database for the application:

```sql
CREATE DATABASE pay_your_dues;
```

### 5. Configure environment variables

Copy the example environment file and update it with your settings:

```bash
cp env.example .env
```

Edit `.env` with your database, RabbitMQ, and notification settings. See `env.example` for the full list. Minimum required values:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=pay_your_dues
DB_SSL_MODE=disable

SERVER_PORT=8080
SERVER_HOST=localhost

JWT_SECRET=your-secret-key-here
JWT_EXPIRY=24h

LOG_LEVEL=debug

# RabbitMQ (required for notification delivery)
RABBITMQ_URL=amqp://guest:guest@localhost:5672/

# Notification worker health endpoint
NOTIFICATION_WORKER_PORT=8081
```

## Running the Application

The app runs as **two Go processes**:

| Process | Command | Purpose |
|---------|---------|---------|
| **API** | `cmd/server` | HTTP API, auth, debts, settings |
| **Notification worker** | `cmd/notification-worker` | Schedules due reminders and delivers notifications via RabbitMQ |

Both processes share the same PostgreSQL database and `.env` file.

### 1. Start infrastructure (PostgreSQL + RabbitMQ)

Using Docker Compose (see [README-docker.md](README-docker.md) for details):

```bash
docker compose up -d postgres rabbitmq
```

- PostgreSQL: `localhost:5432`
- RabbitMQ AMQP: `localhost:5672`
- RabbitMQ management UI: `http://localhost:15672` (guest / guest)

### 2. Start the API server

**Development (with hot reloading):**

```bash
air
```

**Manual / production:**

```bash
go run cmd/server/main.go
# or
go build -o bin/api ./cmd/server && ./bin/api
```

API: `http://localhost:8080`

When `APP_ENV=development`, the API automatically loads `test-data.sql` on startup (skipped if already seeded). Demo accounts:

| Email | Password | Notes |
|-------|----------|-------|
| `alice@dev.local` | `password123` | Primary demo user with 10 debt scenarios |
| `bob@dev.local` | `password123` | Creditor with cross-user debts |
| `carol@dev.local` | `devpassword` | SMS-focused custom settings |

### 3. Start the notification worker

In a **separate terminal**:

```bash
go run cmd/notification-worker/main.go
# or
go build -o bin/notification-worker ./cmd/notification-worker && ./bin/notification-worker
```

Worker health check: `http://localhost:8081/health`

The worker:

- Polls PostgreSQL for due `pending` notifications and publishes them to RabbitMQ
- Consumes jobs from RabbitMQ and sends email, SMS, Slack, Telegram, and Discord notifications
- Runs Telegram long polling in development when `TELEGRAM_POLLING_MODE=true`

If RabbitMQ is unavailable, the API still starts but uses a no-op publisher (notifications are scheduled in the DB but not delivered until the worker is running).

## API Endpoints

### Authentication

- `POST /api/auth/register` - Register a new user
- `POST /api/auth/login` - Login user

### Protected Endpoints (require JWT token)

- `GET /api/health` - Health check

## Database Schema

The application uses the following main tables:

- **users**: User accounts and authentication
- **user_settings**: User preferences and notification settings
- **debt_lists**: Collections of debt items
- **debt_items**: Individual debt records
- **notifications**: Notification history and tracking

## Development

### Project Structure

```
├── cmd/
│   ├── server/              # API HTTP server
│   └── notification-worker/ # Notification scheduler + RabbitMQ consumer
├── internal/
│   ├── messaging/       # RabbitMQ publisher, consumer, topology
│   ├── config/          # Configuration management
│   ├── handlers/        # HTTP handlers
│   ├── middleware/      # HTTP middleware
│   ├── models/          # Data models
│   ├── database/        # Database connection and GORM setup
│   ├── services/        # Business logic
│   ├── repository/      # Data access layer
│   ├── domain/          # Domain entities and interfaces
│   └── mocks/           # Mock implementations for testing
├── tests/
│   ├── unit/            # Unit tests for services and business logic
│   └── integration/     # Integration tests for complete workflows
├── pkg/                 # Public packages
└── scripts/             # Utility scripts
```

### Running Tests

SQLite-backed integration tests require CGO (`gcc` installed, `CGO_ENABLED=1`).

```bash
go test ./...
```

**Specific Test Categories**
You can run specific types of tests:

```bash
# Unit tests only (services and business logic)
./run_tests.sh unit

# Integration tests (complete user-contact-debt workflows)
./run_tests.sh integration

# All tests
./run_tests.sh all

# Performance benchmarks
./run_tests.sh performance

# Race condition detection
./run_tests.sh race

# Coverage report only
./run_tests.sh coverage
```

**Manual Test Execution**
You can also run tests manually using Go commands:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -v -coverprofile=coverage.out ./...

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Run specific test package
go test ./tests/unit
go test ./tests/integration

# Run with race detection
go test -race ./...

# Run benchmarks
go test -bench=. -benchmem ./...
```

### Database Schema

The application uses GORM's auto-migration feature to automatically create and update the database schema based on the model definitions. No manual migrations are required.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License.
