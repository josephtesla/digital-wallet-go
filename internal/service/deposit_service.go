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

type DepositService interface {
	InitializeDeposit(ctx context.Context, userID, walletID string, amount decimal.Decimal, email string) (*infra.InitializeTransactionResponse, error)
	VerifyDeposit(ctx context.Context, reference string) (*models.Payment, error)
	ProcessWebhook(ctx context.Context, webhookData map[string]interface{}) error
}

type depositService struct {
	walletRepo      repository.WalletRepository
	ledgerRepo      repository.LedgerRepository
	paymentRepo     repository.PaymentRepository
	idempotencyRepo repository.IdempotencyRepository
	paystackClient  *infra.PaystackClient
	logger          *zap.Logger
}

func NewDepositService(walletRepo repository.WalletRepository, ledgerRepo repository.LedgerRepository, paymentRepo repository.PaymentRepository, idempotencyRepo repository.IdempotencyRepository, paystackClient *infra.PaystackClient, logger *zap.Logger) DepositService {
	return &depositService{
		walletRepo:      walletRepo,
		ledgerRepo:      ledgerRepo,
		paymentRepo:     paymentRepo,
		idempotencyRepo: idempotencyRepo,
		paystackClient:  paystackClient,
		logger:          logger,
	}
}

func (s *depositService) InitializeDeposit(ctx context.Context, userID, walletID string, amount decimal.Decimal, email string) (*infra.InitializeTransactionResponse, error) {
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

	// Generate unique reference
	reference := fmt.Sprintf("DEP_%d_%s", time.Now().Unix(), uuid.New().String()[:8])

	// Create payment record
	payment := &models.Payment{
		UserID:    uuid.MustParse(userID),
		WalletID:  uuid.MustParse(walletID),
		Reference: reference,
		Amount:    amount,
		Currency:  wallet.Currency,
		Type:      models.PaymentTypeDeposit,
		Status:    models.PaymentStatusPending,
	}

	if err := s.paymentRepo.Create(payment); err != nil {
		s.logger.Error("Failed to create payment record", zap.Error(err))
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}

	// Initialize Paystack transaction
	paystackReq := &infra.InitializeTransactionRequest{
		Amount:      amount.Mul(decimal.NewFromInt(100)).IntPart(), // Convert to kobo
		Email:       email,
		Reference:   reference,
		CallbackURL: "https://your-domain.com/webhooks/paystack",
	}

	paystackResp, err := s.paystackClient.InitializeTransaction(paystackReq)
	if err != nil {
		s.logger.Error("Failed to initialize Paystack transaction", zap.Error(err))
		return nil, fmt.Errorf("failed to initialize Paystack transaction: %w", err)
	}

	// Update payment with Paystack reference
	payment.PaystackReference = paystackResp.Data.Reference
	if err := s.paymentRepo.Update(payment); err != nil {
		s.logger.Error("Failed to update payment with Paystack reference", zap.Error(err))
		return nil, fmt.Errorf("failed to update payment with Paystack reference: %w", err)
	}

	s.logger.Info("Deposit initialized successfully",
		zap.String("reference", reference),
		zap.String("walletID", walletID),
		zap.String("amount", amount.String()),
	)

	return paystackResp, nil
}

func (s *depositService) VerifyDeposit(ctx context.Context, reference string) (*models.Payment, error) {
	// Get payment record
	payment, err := s.paymentRepo.GetByReference(reference)
	if err != nil {
		s.logger.Error("Payment not found", zap.String("reference", reference), zap.Error(err))
		return nil, utils.NewTransactionNotFoundError(reference)
	}

	// Verify with Paystack
	verifyResp, err := s.paystackClient.VerifyTransaction(reference)
	if err != nil {
		s.logger.Error("Failed to verify Paystack transaction", zap.Error(err))
		return nil, fmt.Errorf("failed to verify Paystack transaction: %w", err)
	}

	// Update payment status
	if verifyResp.Data.Status == "success" {
		payment.Status = models.PaymentStatusSuccess
		now := time.Now()
		payment.PaidAt = &now
		payment.GatewayResponse = verifyResp.Data.GatewayResponse

		// Update wallet balance
		wallet, err := s.walletRepo.GetByID(payment.WalletID.String())
		if err != nil {
			s.logger.Error("Wallet not found", zap.String("walletID", payment.WalletID.String()), zap.Error(err))
			return nil, utils.NewWalletNotFoundError(payment.WalletID.String())
		}

		newBalance := wallet.Balance.Add(payment.Amount)
		if err := s.walletRepo.UpdateBalance(payment.WalletID.String(), newBalance); err != nil {
			s.logger.Error("Failed to update wallet balance", zap.Error(err))
			return nil, fmt.Errorf("failed to update wallet balance: %w", err)
		}

		// Create ledger transaction
		ledgerTransaction := &models.LedgerTransaction{
			Reference:   reference,
			Description: fmt.Sprintf("Deposit via Paystack - %s", reference),
			Amount:      payment.Amount,
			Currency:    payment.Currency,
			Type:        models.TransactionTypeDeposit,
			Status:      models.TransactionStatusCompleted,
		}

		if err := s.ledgerRepo.CreateTransaction(ledgerTransaction); err != nil {
			s.logger.Error("Failed to create ledger transaction", zap.Error(err))
			return nil, fmt.Errorf("failed to create ledger transaction: %w", err)
		}

		// Create ledger entry for wallet credit
		ledgerEntry := &models.LedgerEntry{
			TransactionID: ledgerTransaction.ID,
			AccountID:     payment.WalletID,
			AccountType:   models.AccountTypeWallet,
			DebitAmount:   decimal.Zero,
			CreditAmount:  payment.Amount,
			Currency:      payment.Currency,
			Description:   fmt.Sprintf("Deposit credit to wallet %s", payment.WalletID.String()),
		}

		if err := s.ledgerRepo.CreateEntry(ledgerEntry); err != nil {
			s.logger.Error("Failed to create ledger entry", zap.Error(err))
			return nil, fmt.Errorf("failed to create ledger entry: %w", err)
		}

		s.logger.Info("Deposit verified and processed successfully",
			zap.String("reference", reference),
			zap.String("amount", payment.Amount.String()),
		)
	} else {
		payment.Status = models.PaymentStatusFailed
		payment.GatewayResponse = verifyResp.Data.GatewayResponse
	}

	// Update payment record
	if err := s.paymentRepo.Update(payment); err != nil {
		s.logger.Error("Failed to update payment record", zap.Error(err))
		return nil, fmt.Errorf("failed to update payment record: %w", err)
	}

	return payment, nil
}

func (s *depositService) ProcessWebhook(ctx context.Context, webhookData map[string]interface{}) error {
	// Extract reference from webhook data
	reference, ok := webhookData["reference"].(string)
	if !ok {
		return fmt.Errorf("invalid webhook data: missing reference")
	}

	// Process the webhook by verifying the transaction
	_, err := s.VerifyDeposit(ctx, reference)
	if err != nil {
		s.logger.Error("Failed to process webhook", zap.String("reference", reference), zap.Error(err))
		return fmt.Errorf("failed to process webhook: %w", err)
	}

	s.logger.Info("Webhook processed successfully", zap.String("reference", reference))
	return nil
}
