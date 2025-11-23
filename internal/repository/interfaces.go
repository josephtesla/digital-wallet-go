package repository

import (
	"github.com/josephtesla/digital-wallet-go/internal/models"

	"gorm.io/gorm"
)

type Repositories struct {
	Wallet      WalletRepository
	Ledger      LedgerRepository
	Payment     PaymentRepository
	BankAccount BankAccountRepository
	Idempotency IdempotencyRepository
	User        UserRepository
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Wallet:      NewWalletRepository(db),
		Ledger:      NewLedgerRepository(db),
		Payment:     NewPaymentRepository(db),
		BankAccount: NewBankAccountRepository(db),
		Idempotency: NewIdempotencyRepository(db),
		User:        NewUserRepository(db),
	}
}

type WalletRepository interface {
	Create(wallet *models.Wallet) error
	GetByID(id string) (*models.Wallet, error)
	GetByUserID(userID string) ([]*models.Wallet, error)
	UpdateBalance(id string, balance interface{}) error
	Update(wallet *models.Wallet) error
	Delete(id string) error
}

type LedgerRepository interface {
	CreateTransaction(transaction *models.LedgerTransaction) error
	GetTransactionByReference(reference string) (*models.LedgerTransaction, error)
	CreateEntry(entry *models.LedgerEntry) error
	GetEntriesByTransactionID(transactionID string) ([]*models.LedgerEntry, error)
	GetAccountBalance(accountID string, accountType models.AccountType) (interface{}, error)
	UpdateTransactionStatus(reference string, status models.TransactionStatus) error
}

type PaymentRepository interface {
	Create(payment *models.Payment) error
	GetByReference(reference string) (*models.Payment, error)
	GetByPaystackReference(reference string) (*models.Payment, error)
	UpdateStatus(reference string, status models.PaymentStatus) error
	Update(payment *models.Payment) error
	GetByUserID(userID string) ([]*models.Payment, error)
}

type BankAccountRepository interface {
	Create(account *models.BankAccount) error
	GetByID(id string) (*models.BankAccount, error)
	GetByUserID(userID string) ([]*models.BankAccount, error)
	Update(account *models.BankAccount) error
	Delete(id string) error
	VerifyAccount(id string) error
}

type IdempotencyRepository interface {
	Create(key *models.Idempotency) error
	GetByKey(key string) (*models.Idempotency, error)
	UpdateResponse(key string, response string, status string) error
}

type UserRepository interface {
	Create(user *models.User) error
	GetByID(id string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	Delete(id string) error
}
