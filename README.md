# Digital Wallet Backend

A production-ready digital wallet backend built with Go, Gin, PostgreSQL, Redis, and Paystack integration. Supports deposits, transfers, withdrawals, and double-entry bookkeeping.

## 🏗️ Architecture

The project follows a clean, layered architecture:

```
cmd/server/          # Entry point
internal/
  ├── api/           # HTTP handlers & middleware
  ├── service/       # Business logic
  ├── repository/    # Data access layer
  ├── models/        # GORM models
  ├── infra/         # Infrastructure (DB, Redis, Config)
  └── utils/         # Helpers (errors, responses, money)
db/                  # Database (migrations auto-generated)
test/integration/    # Integration tests
```

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- PostgreSQL 15
- Redis 7

### Setup

1. **Clone and setup environment:**
   ```bash
   cp .env.example .env
   ```

2. **Start services:**
   ```bash
   make docker-up
   ```

3. **Run the application:**
   ```bash
   make run
   ```

The server will start on `http://localhost:8080`

### Available Commands

```bash
make help              # Show all available commands
make build            # Build the application
make run              # Run locally
make test             # Run unit tests
make test-integration # Run integration tests
make docker-up        # Start Docker services
make docker-down      # Stop Docker services
make fmt              # Format code
make lint             # Run linter
```

## 🗄️ Database Migrations

**Migrations are automatically handled by GORM!** 

When the application starts, it automatically:
1. Creates all tables from Go models using `AutoMigrate`
2. Seeds initial system accounts (settlement, revenue, pending_withdrawal)

**No manual SQL files needed.** All schema is defined in `/internal/models/` and auto-generated on startup.

To add new models:
1. Create the model struct in `internal/models/`
2. Add it to the `Migrate()` function in `db/migrations.go`
3. Restart the application

## 💰 Data Models

### Core Entities

#### User
```go
- ID, Email, Password
- Links to one Wallet
- Can have multiple BankAccounts
```

#### Wallet
```go
- ID, UserID, Currency (NGN)
- Balance (in kobo - smallest unit)
- Status (active, frozen, closed)
```

#### Payment
```go
- Tracks deposits & withdrawals via Paystack
- PaystackRef for reconciliation
- Status: pending, success, failed
```

#### Ledger (Double-Entry Bookkeeping)
```go
LedgerTransaction
- ID, Reference, Amount, Type, Status

LedgerEntry (always in pairs - debit/credit)
- ID, TransactionID, AccountID, Amount
```

#### BankAccount
```go
- Links user to bank details
- RecipientCode from Paystack
```

#### SystemAccount
```go
- Internal accounts: settlement, revenue, pending_withdrawal
- For ledger balancing
```

## 🔄 Key Features

### 1. Double-Entry Bookkeeping
Every transaction creates balanced entries:
- Transfer: Wallet1 -50 → Wallet2 +50
- Deposit: Wallet +100 ← Settlement -100
- Withdrawal: Wallet -100 → Bank +100

### 2. Idempotency
Prevents duplicate operations via:
- Idempotency keys in request headers
- Cached responses stored in DB
- Perfect for webhook replays

### 3. Atomic Transfers
Uses Redis distributed locks to ensure:
- No race conditions
- Consistent ledger entries
- Wallet balance always reflects ledger

### 4. Paystack Integration
- **Deposits**: Initialize → Paystack checkout → Webhook verification
- **Withdrawals**: Create recipient → Initiate transfer → Webhook confirmation

## 📡 API Endpoints

### Wallets
```
POST   /api/v1/wallets              # Create wallet
GET    /api/v1/wallets/:id          # Get wallet details
GET    /api/v1/wallets/:id/balance  # Get balance
```

### Transfers
```
POST   /api/v1/transfers            # Transfer between wallets
GET    /api/v1/transfers/:walletId/history
```

### Deposits
```
POST   /api/v1/deposits/init        # Initialize Paystack deposit
GET    /api/v1/deposits/verify/:reference
```

### Withdrawals
```
POST   /api/v1/withdrawals/init     # Initialize withdrawal
GET    /api/v1/withdrawals/:walletId/history
```

### Webhooks
```
POST   /api/v1/webhooks/paystack    # Paystack event webhooks
```

## 🧪 Testing

Integration tests use testcontainers for real PostgreSQL & Redis:

```bash
make test-integration
```

Tests include:
- ✅ Wallet creation & balance retrieval
- ✅ Transfer atomicity & balance consistency
- ✅ Deposit initialization & Paystack webhook
- ✅ Withdrawal flow & payout completion
- ✅ Insufficient funds scenarios

## 🔐 Security Features

- **JWT Authentication** (middleware ready)
- **Idempotency** for safe retries
- **Distributed Locking** for atomic operations
- **Structured Logging** for audit trails
- **Error Handling** with domain-specific errors

## 📝 Environment Variables

See `.env.example` for all required variables:
- `DATABASE_URL` - PostgreSQL connection
- `REDIS_URL` - Redis connection
- `PAYSTACK_SECRET_KEY` - Paystack API key
- `LOG_LEVEL` - debug, info, warn, error
- `PORT` - Server port (default 8080)

## 🐳 Docker Support

### Development
```bash
docker-compose up -d        # Start all services
docker-compose down         # Stop all services
docker-compose logs -f api  # View logs
```

### Testing
```bash
docker-compose -f docker-compose.test.yml up -d
# Run tests
docker-compose -f docker-compose.test.yml down
```

## 📚 Project Structure Details

```
internal/
├── api/
│   ├── router.go                    # Route definitions
│   ├── handlers/                    # Request handlers
│   │   ├── wallet_handler.go
│   │   ├── transfer_handler.go
│   │   ├── deposit_handler.go
│   │   ├── withdraw_handler.go
│   │   └── webhook_handler.go
│   └── middleware/
│       ├── auth.go                  # JWT authentication
│       ├── idempotency.go           # Request deduplication
│       ├── logger.go                # Request logging
│       └── recovery.go              # Panic recovery
│
├── service/
│   ├── wallet_service.go            # Wallet operations
│   ├── transfer_service.go          # Atomic transfers with locking
│   ├── deposit_service.go           # Paystack deposits
│   ├── withdrawal_service.go        # Paystack withdrawals
│   ├── ledger_service.go            # Double-entry validation
│   └── idempotency_service.go       # Deduplication logic
│
├── repository/
│   ├── interfaces.go                # Repository contracts
│   ├── wallet_repo.go               # Wallet queries
│   ├── ledger_repo.go               # Ledger queries
│   ├── payment_repo.go              # Payment queries
│   ├── bank_account_repo.go         # Bank account queries
│   ├── idempotency_repo.go          # Idempotency queries
│   └── tx.go                        # Transaction helpers
│
├── models/                          # GORM models (auto-migrated)
├── infra/
│   ├── config.go                    # Configuration
│   ├── db.go                        # GORM setup
│   ├── redis.go                     # Redis client
│   ├── logger.go                    # Zap logger
│   └── paystack_client.go           # Paystack REST client
│
└── utils/
    ├── response.go                  # API response helpers
    ├── errors.go                    # Domain errors
    ├── money.go                     # Amount helpers (kobo/naira)
    └── lock.go                      # Redis locking

db/
└── migrations.go                    # AutoMigrate + Seed functions
```

## 🎯 Next Steps

1. **Implement JWT Auth**: Uncomment auth middleware in router
2. **Add Rate Limiting**: Implement Gin rate limiter
3. **Add Audit Logs**: Log all financial operations
4. **Implement Health Checks**: DB & Redis health endpoints
5. **Add Metrics**: Prometheus metrics for monitoring
6. **Implement Transactions UI**: Frontend for wallet operations

## 📄 License

MIT
