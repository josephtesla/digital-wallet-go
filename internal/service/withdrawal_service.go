package service

import (
	"context"
	"fmt"
	"time"

	"github.com/josephtesla/digital-wallet-go/internal/infra"
	"github.com/josephtesla/digital-wallet-go/internal/models"
	"github.com/josephtesla/digital-wallet-go/internal/repository"
	"github.com/josephtesla/digital-wallet-go/internal/utils"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type WithdrawalService interface {
	InitializeWithdrawal(ctx context.Context, userID, walletID, bankAccountID string, amount decimal.Decimal) (*infra.InitiateTransferResponse, error)
	ProcessWithdrawalWebhook(ctx context.Context, webhookData map[string]interface{}) error
	GetWithdrawalHistory(ctx context.Context, walletID string) ([]*models.Payment, error)
}

type withdrawalService struct {
	walletRepo      repository.WalletRepository
	ledgerRepo      repository.LedgerRepository
	paymentRepo     repository.PaymentRepository
	bankAccountRepo repository.BankAccountRepository
	idempotencyRepo repository.IdempotencyRepository
	paystackClient  *infra.PaystackClient
	logger          *zap.Logger
}

func NewWithdrawalService(walletRepo repository.WalletRepository, ledgerRepo repository.LedgerRepository, paymentRepo repository.PaymentRepository, bankAccountRepo repository.BankAccountRepository, idempotencyRepo repository.IdempotencyRepository, paystackClient *infra.PaystackClient, logger *zap.Logger) WithdrawalService {
	return &withdrawalService{
		walletRepo:      walletRepo,
		ledgerRepo:      ledgerRepo,
		paymentRepo:     paymentRepo,
		bankAccountRepo: bankAccountRepo,
		idempotencyRepo: idempotencyRepo,
		paystackClient:  paystackClient,
		logger:          logger,
	}
}

func (s *withdrawalService) InitializeWithdrawal(ctx context.Context, userID, walletID, bankAccountID string, amount decimal.Decimal) (*infra.InitiateTransferResponse, error) {
	// Validate amount
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, utils.NewInvalidAmountError(amount.String())
	}

	// Get wallet
	wallet, err := s.walletRepo.GetByID(walletID)
	if err != nil {
		s.logger.Error("Wallet not found", zap.String("walletID", walletID), zap.Error(err))
		return nil, utils.NewWalletNotFoundError(walletID)
	}

	// Check sufficient balance
	if wallet.Balance.LessThan(amount) {
		return nil, utils.NewInsufficientFundsError(fmt.Sprintf("Available: %s, Required: %s", wallet.Balance.String(), amount.String()))
	}

	// Get bank account
	bankAccount, err := s.bankAccountRepo.GetByID(bankAccountID)
	if err != nil {
		s.logger.Error("Bank account not found", zap.String("bankAccountID", bankAccountID), zap.Error(err))
		return nil, utils.NewBankAccountNotFoundError(bankAccountID)
	}

	if !bankAccount.IsVerified {
		return nil, utils.NewInvalidBankAccountError("Bank account is not verified")
	}

	// Generate unique reference
	reference := fmt.Sprintf("WTH_%d_%s", time.Now().Unix(), uuid.New().String()[:8])

	// Create payment record
	payment := &models.Payment{
		UserID:    uuid.MustParse(userID),
		WalletID:  uuid.MustParse(walletID),
		Reference: reference,
		Amount:    amount,
		Currency:  wallet.Currency,
		Type:      models.PaymentTypeWithdrawal,
		Status:    models.PaymentStatusPending,
	}

	if err := s.paymentRepo.Create(payment); err != nil {
		s.logger.Error("Failed to create payment record", zap.Error(err))
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}

	// Create transfer recipient if not exists (you might want to cache this)
	recipientReq := &infra.CreateTransferRecipientRequest{
		Type:          "nuban",
		Name:          bankAccount.AccountName,
		AccountNumber: bankAccount.AccountNumber,
		BankCode:      bankAccount.BankCode,
		Currency:      "NGN",
	}

	recipientResp, err := s.paystackClient.CreateTransferRecipient(recipientReq)
	if err != nil {
		s.logger.Error("Failed to create transfer recipient", zap.Error(err))
		return nil, fmt.Errorf("failed to create transfer recipient: %w", err)
	}

	// Initiate transfer
	transferReq := &infra.InitiateTransferRequest{
		Source:    "balance",
		Amount:    amount.Mul(decimal.NewFromInt(100)).IntPart(), // Convert to kobo
		Recipient: recipientResp.Data.RecipientCode,
		Reason:    fmt.Sprintf("Withdrawal to %s", bankAccount.AccountName),
		Reference: reference,
	}

	transferResp, err := s.paystackClient.InitiateTransfer(transferReq)
	if err != nil {
		s.logger.Error("Failed to initiate transfer", zap.Error(err))
		return nil, fmt.Errorf("failed to initiate transfer: %w", err)
	}

	// Update payment with Paystack reference
	payment.PaystackReference = transferResp.Data.TransferCode
	if err := s.paymentRepo.Update(payment); err != nil {
		s.logger.Error("Failed to update payment with Paystack reference", zap.Error(err))
		return nil, fmt.Errorf("failed to update payment with Paystack reference: %w", err)
	}

	// Deduct amount from wallet immediately (pending status)
	newBalance := wallet.Balance.Sub(amount)
	if err := s.walletRepo.UpdateBalance(walletID, newBalance); err != nil {
		s.logger.Error("Failed to update wallet balance", zap.Error(err))
		return nil, fmt.Errorf("failed to update wallet balance: %w", err)
	}

	// Create ledger transaction
	ledgerTransaction := &models.LedgerTransaction{
		Reference:   reference,
		Description: fmt.Sprintf("Withdrawal to %s", bankAccount.AccountName),
		Amount:      amount,
		Currency:    wallet.Currency,
		Type:        models.TransactionTypeWithdrawal,
		Status:      models.TransactionStatusPending,
	}

	if err := s.ledgerRepo.CreateTransaction(ledgerTransaction); err != nil {
		s.logger.Error("Failed to create ledger transaction", zap.Error(err))
		return nil, fmt.Errorf("failed to create ledger transaction: %w", err)
	}

	// Create ledger entry for wallet debit
	ledgerEntry := &models.LedgerEntry{
		TransactionID: ledgerTransaction.ID,
		AccountID:     payment.WalletID,
		AccountType:   models.AccountTypeWallet,
		DebitAmount:   amount,
		CreditAmount:  decimal.Zero,
		Currency:      payment.Currency,
		Description:   fmt.Sprintf("Withdrawal debit from wallet %s", payment.WalletID.String()),
	}

	if err := s.ledgerRepo.CreateEntry(ledgerEntry); err != nil {
		s.logger.Error("Failed to create ledger entry", zap.Error(err))
		return nil, fmt.Errorf("failed to create ledger entry: %w", err)
	}

	s.logger.Info("Withdrawal initialized successfully",
		zap.String("reference", reference),
		zap.String("walletID", walletID),
		zap.String("amount", amount.String()),
	)

	return transferResp, nil
}

func (s *withdrawalService) ProcessWithdrawalWebhook(ctx context.Context, webhookData map[string]interface{}) error {
	// Extract transfer code from webhook data
	transferCode, ok := webhookData["transfer_code"].(string)
	if !ok {
		return fmt.Errorf("invalid webhook data: missing transfer_code")
	}

	// Get payment by Paystack reference
	payment, err := s.paymentRepo.GetByPaystackReference(transferCode)
	if err != nil {
		s.logger.Error("Payment not found", zap.String("transferCode", transferCode), zap.Error(err))
		return fmt.Errorf("payment not found: %w", err)
	}

	// Update payment status based on webhook data
	status, ok := webhookData["status"].(string)
	if !ok {
		return fmt.Errorf("invalid webhook data: missing status")
	}

	switch status {
	case "success":
		payment.Status = models.PaymentStatusSuccess
		now := time.Now()
		payment.PaidAt = &now

		// Update ledger transaction status
		if err := s.ledgerRepo.UpdateTransactionStatus(payment.Reference, models.TransactionStatusCompleted); err != nil {
			s.logger.Error("Failed to update ledger transaction status", zap.Error(err))
			return fmt.Errorf("failed to update ledger transaction status: %w", err)
		}

		s.logger.Info("Withdrawal completed successfully",
			zap.String("reference", payment.Reference),
			zap.String("amount", payment.Amount.String()),
		)

	case "failed":
		payment.Status = models.PaymentStatusFailed

		// Refund amount to wallet
		wallet, err := s.walletRepo.GetByID(payment.WalletID.String())
		if err != nil {
			s.logger.Error("Wallet not found", zap.String("walletID", payment.WalletID.String()), zap.Error(err))
			return fmt.Errorf("wallet not found: %w", err)
		}

		newBalance := wallet.Balance.Add(payment.Amount)
		if err := s.walletRepo.UpdateBalance(payment.WalletID.String(), newBalance); err != nil {
			s.logger.Error("Failed to refund wallet balance", zap.Error(err))
			return fmt.Errorf("failed to refund wallet balance: %w", err)
		}

		// Update ledger transaction status
		if err := s.ledgerRepo.UpdateTransactionStatus(payment.Reference, models.TransactionStatusFailed); err != nil {
			s.logger.Error("Failed to update ledger transaction status", zap.Error(err))
			return fmt.Errorf("failed to update ledger transaction status: %w", err)
		}

		s.logger.Info("Withdrawal failed, amount refunded",
			zap.String("reference", payment.Reference),
			zap.String("amount", payment.Amount.String()),
		)
	}

	// Update payment record
	if err := s.paymentRepo.Update(payment); err != nil {
		s.logger.Error("Failed to update payment record", zap.Error(err))
		return fmt.Errorf("failed to update payment record: %w", err)
	}

	return nil
}

func (s *withdrawalService) GetWithdrawalHistory(ctx context.Context, walletID string) ([]*models.Payment, error) {
	payments, err := s.paymentRepo.GetByUserID(walletID) // Note: This should be filtered by wallet and type
	if err != nil {
		s.logger.Error("Failed to get withdrawal history", zap.String("walletID", walletID), zap.Error(err))
		return nil, fmt.Errorf("failed to get withdrawal history: %w", err)
	}

	// Filter for withdrawal payments
	var withdrawals []*models.Payment
	for _, payment := range payments {
		if payment.Type == models.PaymentTypeWithdrawal {
			withdrawals = append(withdrawals, payment)
		}
	}

	return withdrawals, nil
}
