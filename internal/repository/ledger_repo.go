package repository

import (
	"github.com/josephtesla/digital-wallet-go/internal/models"

	"gorm.io/gorm"
)

type ledgerRepository struct {
	db *gorm.DB
}

func NewLedgerRepository(db *gorm.DB) LedgerRepository {
	return &ledgerRepository{db: db}
}

func (r *ledgerRepository) CreateTransaction(transaction *models.LedgerTransaction) error {
	return r.db.Create(transaction).Error
}

func (r *ledgerRepository) GetTransactionByReference(reference string) (*models.LedgerTransaction, error) {
	var transaction models.LedgerTransaction
	err := r.db.Preload("Entries").First(&transaction, "reference = ?", reference).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *ledgerRepository) CreateEntry(entry *models.LedgerEntry) error {
	return r.db.Create(entry).Error
}

func (r *ledgerRepository) GetEntriesByTransactionID(transactionID string) ([]*models.LedgerEntry, error) {
	var entries []*models.LedgerEntry
	err := r.db.Where("transaction_id = ?", transactionID).Find(&entries).Error
	if err != nil {
		return nil, err
	}
	return entries, nil
}

func (r *ledgerRepository) GetAccountBalance(accountID string, accountType models.AccountType) (interface{}, error) {
	var result struct {
		DebitTotal  float64 `json:"debit_total"`
		CreditTotal float64 `json:"credit_total"`
		Balance     float64 `json:"balance"`
	}

	err := r.db.Model(&models.LedgerEntry{}).
		Select("COALESCE(SUM(debit_amount), 0) as debit_total, COALESCE(SUM(credit_amount), 0) as credit_total").
		Where("account_id = ? AND account_type = ?", accountID, accountType).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	// Calculate balance based on account type
	switch accountType {
	case models.AccountTypeAsset, models.AccountTypeWallet:
		result.Balance = result.DebitTotal - result.CreditTotal
	case models.AccountTypeLiability, models.AccountTypeEquity, models.AccountTypeRevenue:
		result.Balance = result.CreditTotal - result.DebitTotal
	case models.AccountTypeExpense:
		result.Balance = result.DebitTotal - result.CreditTotal
	default:
		result.Balance = result.DebitTotal - result.CreditTotal
	}

	return result, nil
}

func (r *ledgerRepository) UpdateTransactionStatus(reference string, status models.TransactionStatus) error {
	return r.db.Model(&models.LedgerTransaction{}).
		Where("reference = ?", reference).
		Update("status", status).Error
}
