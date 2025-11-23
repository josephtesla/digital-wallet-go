# Database Migration Guide

## 🎯 Why No SQL Files?

This project uses **GORM AutoMigrate**, which means:
- ✅ Database schema is defined entirely in Go code (`internal/models/`)
- ✅ Tables are auto-created/updated on application startup
- ✅ No manual SQL files to maintain
- ✅ Type-safe schema management

## 📝 How It Works

### On Application Startup

1. **db/migrations.go** calls `Migrate(db, logger)`
2. GORM's `AutoMigrate()` reads each model struct
3. For each model:
   - If table doesn't exist → Create it
   - If table exists → Update columns/indexes as needed
4. `Seed(db, logger)` creates initial system accounts

### Model → Table Mapping

Each Go struct field becomes a database column:

```go
type Wallet struct {
    ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
    UserID    uuid.UUID `gorm:"type:uuid;index"`
    Currency  string    `gorm:"size:3;default:'NGN'"`
    Balance   int64     `gorm:"not null;default:0"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

Becomes:
```sql
CREATE TABLE wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    currency VARCHAR(3) DEFAULT 'NGN',
    balance BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
CREATE INDEX idx_wallets_user_id ON wallets(user_id);
```

## 🔧 How to Add a New Table

### Step 1: Create the Model
Add a new file in `internal/models/`:

```go
// internal/models/example.go
package models

import (
    "time"
    "github.com/google/uuid"
)

type Example struct {
    ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
    Name      string    `gorm:"not null"`
    CreatedAt time.Time
}

func (Example) TableName() string {
    return "examples"
}
```

### Step 2: Register in Migrations
Update `db/migrations.go`:

```go
func Migrate(db *gorm.DB, logger *zap.Logger) error {
    // ...
    if err := db.AutoMigrate(
        &models.User{},
        &models.Example{},  // ← Add here
        // ...
    ); err != nil {
        return err
    }
    return nil
}
```

### Step 3: Restart Application
```bash
make run
# or
make docker-up
```

That's it! The table is created automatically.

## 📊 GORM Tags Reference

Common tags used in this project:

```go
type Example struct {
    // Primary Key
    ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

    // Columns
    Name    string    `gorm:"not null"`
    Email   string    `gorm:"uniqueIndex"`
    Status  string    `gorm:"type:varchar(20);default:'active'"`
    Balance int64     `gorm:"not null;default:0"`

    // Timestamps
    CreatedAt time.Time
    UpdatedAt time.Time

    // Index
    UserID uuid.UUID `gorm:"type:uuid;index"`

    // Foreign Key
    User User `gorm:"foreignKey:UserID"`
}
```

## 🚀 Development Workflow

### First Run
```bash
cp .env.example .env
make docker-up
make run
```
→ All tables created automatically ✅

### Add New Feature with DB Changes
```go
// 1. Create model
// internal/models/new_feature.go

// 2. Register in migrations
// db/migrations.go

// 3. Restart
make run
```
→ New table appears automatically ✅

### Modify Existing Model
```go
// 1. Update struct in internal/models/
// Add/remove/modify fields

// 2. Restart
make run
```
→ Table schema updated automatically ✅

## ⚠️ Important Notes

### Safe Changes (Auto-applied)
- ✅ Add new column
- ✅ Add new index
- ✅ Change default value
- ✅ Add NOT NULL constraint to new columns

### Dangerous Changes (Review Needed)
- ❌ Remove column (data loss)
- ❌ Rename column (data loss)
- ❌ Add NOT NULL to existing nullable column (migration needed)
- ❌ Change column type (might fail)

For dangerous changes:
1. Write raw SQL migration script
2. Execute manually before code changes
3. Then update the model

### Testing
```bash
make test-integration
```
Uses testcontainers - fresh DB each time, so migrations tested automatically.

## 📚 Resources

- [GORM Migration Docs](https://gorm.io/docs/migration.html)
- [GORM Tags Reference](https://gorm.io/docs/models.html)
- [PostgreSQL Types](https://www.postgresql.org/docs/current/datatype.html)

## 🔄 Resetting Database (Dev Only)

```bash
make migrate-down
# This drops all tables

make docker-down
make docker-up
make run
# This recreates everything from scratch
```

---

**Summary**: No more SQL files! Just update Go models, restart, and GORM handles the rest. 🚀
