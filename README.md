# Pay Your Dues - a Go Debt Tracker Application

A comprehensive debt tracking application built with Go, PostgreSQL, and Gin framework. Track your debts, money lent to others, and send notifications via email, SMS, or Facebook Messenger.

## Features

- **User Authentication**: Secure JWT-based authentication
- **Debt Management**: Track money you owe and money owed to you
- **Multiple Debt Lists**: Organize debts into different categories
- **Notifications**: Send reminders via email, SMS, or Facebook Messenger
- **User Settings**: Customize notification preferences and default currency
- **RESTful API**: Clean and well-documented API endpoints

## Tech Stack

- **Backend**: Go 1.21+
- **Framework**: Gin (HTTP web framework)
- **Database**: PostgreSQL
- **Authentication**: JWT tokens
- **Password Hashing**: bcrypt
- **Logging**: Zerolog
- **Hot Reloading**: Air

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 12 or higher
- Air (for development hot reloading)

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

Edit `.env` with your database credentials:

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
```

## Running the Application

### Development (with hot reloading)

```bash
air
```

### Production

```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`

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
│   └── server/          # Main application entry point
├── internal/
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
