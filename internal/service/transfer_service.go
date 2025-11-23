package service

import (
	"context"
	"fmt"
	"time"

	"github.com/josephtesla/digital-wallet-go/internal/models"
	"github.com/josephtesla/digital-wallet-go/internal/repository"
	"github.com/josephtesla/digital-wallet-go/internal/utils"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type TransferService interface {
	TransferFunds(ctx context.Context, fromWalletID, toWalletID string, amount decimal.Decimal, description string) (*models.LedgerTransaction, error)
	GetTransferHistory(ctx context.Context, walletID string) ([]*models.LedgerTransaction, error)
}

type transferService struct {
	walletRepo  repository.WalletRepository
	ledgerRepo  repository.LedgerRepository
	paymentRepo repository.PaymentRepository
	redisClient *redis.Client
	logger      *zap.Logger
}

func NewTransferService(walletRepo repository.WalletRepository, ledgerRepo repository.LedgerRepository, paymentRepo repository.PaymentRepository, redisClient *redis.Client, logger *zap.Logger) TransferService {
	return &transferService{
		walletRepo:  walletRepo,
		ledgerRepo:  ledgerRepo,
		paymentRepo: paymentRepo,
		redisClient: redisClient,
		logger:      logger,
	}
}

func (s *transferService) TransferFunds(ctx context.Context, fromWalletID, toWalletID string, amount decimal.Decimal, description string) (*models.LedgerTransaction, error) {
	// Validate amount
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, utils.NewInvalidAmountError(amount.String())
	}

	// Get source wallet
	fromWallet, err := s.walletRepo.GetByID(fromWalletID)
	if err != nil {
		s.logger.Error("Source wallet not found", zap.String("fromWalletID", fromWalletID), zap.Error(err))
		return nil, utils.NewWalletNotFoundError(fromWalletID)
	}

	// Get destination wallet
	toWallet, err := s.walletRepo.GetByID(toWalletID)
	if err != nil {
		s.logger.Error("Destination wallet not found", zap.String("toWalletID", toWalletID), zap.Error(err))
		return nil, utils.NewWalletNotFoundError(toWalletID)
	}

	// Check if wallets have same currency
	if fromWallet.Currency != toWallet.Currency {
		return nil, fmt.Errorf("cannot transfer between different currencies")
	}

	// Check sufficient balance
	if fromWallet.Balance.LessThan(amount) {
		return nil, utils.NewInsufficientFundsError(fmt.Sprintf("Available: %s, Required: %s", fromWallet.Balance.String(), amount.String()))
	}

	// Create lock key for atomic transfer
	lockKey := fmt.Sprintf("transfer:%s:%s", fromWalletID, toWalletID)
	lock := utils.NewRedisLock(s.redisClient, lockKey, 30*time.Second)

	// Acquire lock
	acquired, err := lock.AcquireWithRetry(ctx, 3, 1*time.Second)
	if err != nil {
		s.logger.Error("Failed to acquire transfer lock", zap.Error(err))
		return nil, fmt.Errorf("failed to acquire transfer lock: %w", err)
	}
	if !acquired {
		return nil, fmt.Errorf("failed to acquire transfer lock after retries")
	}

	defer func() {
		if err := lock.Release(ctx); err != nil {
			s.logger.Error("Failed to release transfer lock", zap.Error(err))
		}
	}()

	// Generate transaction reference
	reference := fmt.Sprintf("TXN_%d_%s", time.Now().Unix(), uuid.New().String()[:8])

	// Create ledger transaction
	transaction := &models.LedgerTransaction{
		Reference:   reference,
		Description: description,
		Amount:      amount,
		Currency:    fromWallet.Currency,
		Type:        models.TransactionTypeTransfer,
		Status:      models.TransactionStatusPending,
	}

	if err := s.ledgerRepo.CreateTransaction(transaction); err != nil {
		s.logger.Error("Failed to create ledger transaction", zap.Error(err))
		return nil, fmt.Errorf("failed to create ledger transaction: %w", err)
	}

	// Create debit entry (from wallet)
	debitEntry := &models.LedgerEntry{
		TransactionID: transaction.ID,
		AccountID:     fromWallet.ID,
		AccountType:   models.AccountTypeWallet,
		DebitAmount:   amount,
		CreditAmount:  decimal.Zero,
		Currency:      fromWallet.Currency,
		Description:   fmt.Sprintf("Transfer to wallet %s", toWalletID),
	}

	if err := s.ledgerRepo.CreateEntry(debitEntry); err != nil {
		s.logger.Error("Failed to create debit entry", zap.Error(err))
		return nil, fmt.Errorf("failed to create debit entry: %w", err)
	}

	// Create credit entry (to wallet)
	creditEntry := &models.LedgerEntry{
		TransactionID: transaction.ID,
		AccountID:     toWallet.ID,
		AccountType:   models.AccountTypeWallet,
		DebitAmount:   decimal.Zero,
		CreditAmount:  amount,
		Currency:      toWallet.Currency,
		Description:   fmt.Sprintf("Transfer from wallet %s", fromWalletID),
	}

	if err := s.ledgerRepo.CreateEntry(creditEntry); err != nil {
		s.logger.Error("Failed to create credit entry", zap.Error(err))
		return nil, fmt.Errorf("failed to create credit entry: %w", err)
	}

	// Update wallet balances
	newFromBalance := fromWallet.Balance.Sub(amount)
	newToBalance := toWallet.Balance.Add(amount)

	if err := s.walletRepo.UpdateBalance(fromWalletID, newFromBalance); err != nil {
		s.logger.Error("Failed to update source wallet balance", zap.Error(err))
		return nil, fmt.Errorf("failed to update source wallet balance: %w", err)
	}

	if err := s.walletRepo.UpdateBalance(toWalletID, newToBalance); err != nil {
		s.logger.Error("Failed to update destination wallet balance", zap.Error(err))
		return nil, fmt.Errorf("failed to update destination wallet balance: %w", err)
	}

	// Update transaction status to completed
	if err := s.ledgerRepo.UpdateTransactionStatus(reference, models.TransactionStatusCompleted); err != nil {
		s.logger.Error("Failed to update transaction status", zap.Error(err))
		return nil, fmt.Errorf("failed to update transaction status: %w", err)
	}

	transaction.Status = models.TransactionStatusCompleted

	s.logger.Info("Transfer completed successfully",
		zap.String("reference", reference),
		zap.String("fromWalletID", fromWalletID),
		zap.String("toWalletID", toWalletID),
		zap.String("amount", amount.String()),
	)

	return transaction, nil
}

func (s *transferService) GetTransferHistory(ctx context.Context, walletID string) ([]*models.LedgerTransaction, error) {
	// This would typically involve querying ledger transactions where the wallet is involved
	// For now, we'll return a simple implementation
	// In a real implementation, you'd query the ledger entries to find transactions involving this wallet

	s.logger.Info("Getting transfer history", zap.String("walletID", walletID))

	// Placeholder implementation - you would implement proper querying here
	return []*models.LedgerTransaction{}, nil
}
