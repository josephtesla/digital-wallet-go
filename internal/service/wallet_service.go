package service

import (
	"context"
	"fmt"

	"github.com/josephtesla/digital-wallet-go/internal/models"
	"github.com/josephtesla/digital-wallet-go/internal/repository"
	"github.com/josephtesla/digital-wallet-go/internal/utils"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type WalletService interface {
	CreateWallet(ctx context.Context, userID string, currency string) (*models.Wallet, error)
	GetWallet(ctx context.Context, walletID string) (*models.Wallet, error)
	GetUserWallets(ctx context.Context, userID string) ([]*models.Wallet, error)
	GetBalance(ctx context.Context, walletID string) (*utils.Money, error)
	UpdateBalance(ctx context.Context, walletID string, newBalance decimal.Decimal) error
}

type walletService struct {
	walletRepo repository.WalletRepository
	ledgerRepo repository.LedgerRepository
	userRepo   repository.UserRepository
	logger     *zap.Logger
}

func NewWalletService(walletRepo repository.WalletRepository, ledgerRepo repository.LedgerRepository, userRepo repository.UserRepository, logger *zap.Logger) WalletService {
	return &walletService{
		walletRepo: walletRepo,
		ledgerRepo: ledgerRepo,
		userRepo:   userRepo,
		logger:     logger,
	}
}

func (s *walletService) CreateWallet(ctx context.Context, userID string, currency string) (*models.Wallet, error) {
	// Validate user exists
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		s.logger.Error("User not found", zap.String("userID", userID), zap.Error(err))
		return nil, utils.NewUserNotFoundError(userID)
	}

	// Check if user already has a wallet for this currency
	existingWallets, err := s.walletRepo.GetByUserID(userID)
	if err != nil {
		s.logger.Error("Failed to get existing wallets", zap.String("userID", userID), zap.Error(err))
		return nil, fmt.Errorf("failed to check existing wallets: %w", err)
	}

	for _, wallet := range existingWallets {
		if wallet.Currency == currency {
			return nil, fmt.Errorf("user already has a wallet for currency %s", currency)
		}
	}

	// Create wallet
	wallet := &models.Wallet{
		UserID:   uuid.MustParse(userID),
		Currency: currency,
		Balance:  decimal.Zero,
		IsActive: true,
	}

	if err := s.walletRepo.Create(wallet); err != nil {
		s.logger.Error("Failed to create wallet", zap.String("userID", userID), zap.Error(err))
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	// Create initial ledger entry for wallet creation
	ledgerEntry := &models.LedgerEntry{
		AccountID:    wallet.ID,
		AccountType:  models.AccountTypeWallet,
		DebitAmount:  decimal.Zero,
		CreditAmount: decimal.Zero,
		Currency:     currency,
		Description:  fmt.Sprintf("Wallet created for user %s", user.Email),
	}

	if err := s.ledgerRepo.CreateEntry(ledgerEntry); err != nil {
		s.logger.Error("Failed to create initial ledger entry", zap.String("walletID", wallet.ID.String()), zap.Error(err))
		// Note: In production, you might want to rollback the wallet creation here
	}

	s.logger.Info("Wallet created successfully", zap.String("walletID", wallet.ID.String()), zap.String("userID", userID))

	return wallet, nil
}

func (s *walletService) GetWallet(ctx context.Context, walletID string) (*models.Wallet, error) {
	wallet, err := s.walletRepo.GetByID(walletID)
	if err != nil {
		s.logger.Error("Wallet not found", zap.String("walletID", walletID), zap.Error(err))
		return nil, utils.NewWalletNotFoundError(walletID)
	}

	return wallet, nil
}

func (s *walletService) GetUserWallets(ctx context.Context, userID string) ([]*models.Wallet, error) {
	wallets, err := s.walletRepo.GetByUserID(userID)
	if err != nil {
		s.logger.Error("Failed to get user wallets", zap.String("userID", userID), zap.Error(err))
		return nil, fmt.Errorf("failed to get user wallets: %w", err)
	}

	return wallets, nil
}

func (s *walletService) GetBalance(ctx context.Context, walletID string) (*utils.Money, error) {
	wallet, err := s.walletRepo.GetByID(walletID)
	if err != nil {
		s.logger.Error("Wallet not found", zap.String("walletID", walletID), zap.Error(err))
		return nil, utils.NewWalletNotFoundError(walletID)
	}

	money := utils.NewMoneyFromNaira(wallet.Balance, wallet.Currency)
	return &money, nil
}

func (s *walletService) UpdateBalance(ctx context.Context, walletID string, newBalance decimal.Decimal) error {
	if err := s.walletRepo.UpdateBalance(walletID, newBalance); err != nil {
		s.logger.Error("Failed to update wallet balance", zap.String("walletID", walletID), zap.Error(err))
		return fmt.Errorf("failed to update wallet balance: %w", err)
	}

	s.logger.Info("Wallet balance updated", zap.String("walletID", walletID), zap.String("newBalance", newBalance.String()))

	return nil
}
