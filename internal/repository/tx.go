package repository

import (
	"context"

	"github.com/josephtesla/digital-wallet-go/internal/models"

	"gorm.io/gorm"
)

// Transaction represents a database transaction
type Transaction struct {
	tx *gorm.DB
}

// NewTransaction creates a new transaction
func NewTransaction(db *gorm.DB) *Transaction {
	return &Transaction{tx: db.Begin()}
}

// Begin starts a new transaction
func (t *Transaction) Begin() error {
	t.tx = t.tx.Begin()
	return t.tx.Error
}

// Commit commits the transaction
func (t *Transaction) Commit() error {
	return t.tx.Commit().Error
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback() error {
	return t.tx.Rollback().Error
}

// GetDB returns the underlying GORM DB instance
func (t *Transaction) GetDB() *gorm.DB {
	return t.tx
}

// TransactionManager handles database transactions
type TransactionManager struct {
	db *gorm.DB
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(db *gorm.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

// WithTransaction executes a function within a database transaction
func (tm *TransactionManager) WithTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
	tx := tm.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// TransactionalRepository wraps repository operations in transactions
type TransactionalRepository struct {
	db *gorm.DB
}

// NewTransactionalRepository creates a new transactional repository
func NewTransactionalRepository(db *gorm.DB) *TransactionalRepository {
	return &TransactionalRepository{db: db}
}

// ExecuteInTransaction executes multiple repository operations in a single transaction
func (tr *TransactionalRepository) ExecuteInTransaction(ctx context.Context, operations ...func(*gorm.DB) error) error {
	tx := tr.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	for _, operation := range operations {
		if err := operation(tx); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// Helper functions for common transaction patterns

// CreateWalletWithLedger creates a wallet and initial ledger entry in a transaction
func CreateWalletWithLedger(tx *gorm.DB, wallet *models.Wallet, ledgerEntry *models.LedgerEntry) error {
	if err := tx.Create(wallet).Error; err != nil {
		return err
	}

	ledgerEntry.AccountID = wallet.ID
	return tx.Create(ledgerEntry).Error
}

// CreateTransactionWithEntries creates a ledger transaction and its entries in a transaction
func CreateTransactionWithEntries(tx *gorm.DB, transaction *models.LedgerTransaction, entries []*models.LedgerEntry) error {
	if err := tx.Create(transaction).Error; err != nil {
		return err
	}

	for _, entry := range entries {
		entry.TransactionID = transaction.ID
		if err := tx.Create(entry).Error; err != nil {
			return err
		}
	}

	return nil
}

// UpdateWalletBalanceWithLedger updates wallet balance and creates ledger entry in a transaction
func UpdateWalletBalanceWithLedger(tx *gorm.DB, walletID string, newBalance interface{}, ledgerEntry *models.LedgerEntry) error {
	if err := tx.Model(&models.Wallet{}).Where("id = ?", walletID).Update("balance", newBalance).Error; err != nil {
		return err
	}

	ledgerEntry.AccountID = walletID
	return tx.Create(ledgerEntry).Error
}
