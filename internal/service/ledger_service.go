package service

import (
	"context"
	"fmt"

	"github.com/josephtesla/digital-wallet-go/internal/models"
	"github.com/josephtesla/digital-wallet-go/internal/repository"

	"go.uber.org/zap"
)

type LedgerService interface {
	CreateTransaction(ctx context.Context, transaction *models.LedgerTransaction, entries []*models.LedgerEntry) error
	GetTransactionByReference(ctx context.Context, txnRef string) (*models.LedgerTransaction, error)
	GetAccountBalance(ctx context.Context, accountID string) (int64, error)
	ValidateDoubleEntry(ctx context.Context, entries []*models.LedgerEntry) error
}

type ledgerService struct {
	ledgerRepo repository.LedgerRepository
	logger     *zap.Logger
}

func NewLedgerService(ledgerRepo repository.LedgerRepository, logger *zap.Logger) LedgerService {
	return &ledgerService{
		ledgerRepo: ledgerRepo,
		logger:     logger,
	}
}

func (s *ledgerService) CreateTransaction(ctx context.Context, transaction *models.LedgerTransaction, entries []*models.LedgerEntry) error {
	// Validate double-entry bookkeeping
	if err := s.ValidateDoubleEntry(ctx, entries); err != nil {
		s.logger.Error("Double-entry validation failed", zap.Error(err))
		return fmt.Errorf("double-entry validation failed: %w", err)
	}

	// Create transaction
	if err := s.ledgerRepo.CreateTransaction(transaction); err != nil {
		s.logger.Error("Failed to create ledger transaction", zap.Error(err))
		return fmt.Errorf("failed to create ledger transaction: %w", err)
	}

	// Create entries
	for _, entry := range entries {
		entry.TransactionID = transaction.ID
		if err := s.ledgerRepo.CreateEntry(entry); err != nil {
			s.logger.Error("Failed to create ledger entry", zap.Error(err))
			return fmt.Errorf("failed to create ledger entry: %w", err)
		}
	}

	s.logger.Info("Ledger transaction created successfully", zap.String("txnRef", transaction.TxnRef))
	return nil
}

func (s *ledgerService) GetTransactionByReference(ctx context.Context, txnRef string) (*models.LedgerTransaction, error) {
	transaction, err := s.ledgerRepo.GetTransactionByReference(txnRef)
	if err != nil {
		s.logger.Error("Transaction not found", zap.String("txnRef", txnRef), zap.Error(err))
		return nil, fmt.Errorf("transaction not found: %w", err)
	}

	return transaction, nil
}

func (s *ledgerService) GetAccountBalance(ctx context.Context, accountID string) (int64, error) {
	balance, err := s.ledgerRepo.GetAccountBalance(accountID)
	if err != nil {
		s.logger.Error("Failed to get account balance", zap.String("accountID", accountID), zap.Error(err))
		return 0, fmt.Errorf("failed to get account balance: %w", err)
	}

	return balance.(int64), nil
}

func (s *ledgerService) ValidateDoubleEntry(ctx context.Context, entries []*models.LedgerEntry) error {
	var totalDebit, totalCredit int64

	for _, entry := range entries {
		if entry.Amount > 0 {
			totalDebit += entry.Amount
		} else {
			totalCredit += -entry.Amount // Convert negative to positive
		}
	}

	if totalDebit != totalCredit {
		return fmt.Errorf("double-entry validation failed: debit total (%d) != credit total (%d)", totalDebit, totalCredit)
	}

	return nil
}
