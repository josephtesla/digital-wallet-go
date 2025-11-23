package service

import (
	"github.com/josephtesla/digital-wallet-go/internal/infra"
	"github.com/josephtesla/digital-wallet-go/internal/repository"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type Services struct {
	Wallet      WalletService
	Transfer    TransferService
	Deposit     DepositService
	Withdrawal  WithdrawalService
	Ledger      LedgerService
	Idempotency IdempotencyService
}

func NewServices(repos *repository.Repositories, redisClient *redis.Client, paystackClient *infra.PaystackClient, logger *zap.Logger) *Services {
	return &Services{
		Wallet:      NewWalletService(repos.Wallet, repos.Ledger, repos.User, logger),
		Transfer:    NewTransferService(repos.Wallet, repos.Ledger, repos.Payment, redisClient, logger),
		Deposit:     NewDepositService(repos.Wallet, repos.Ledger, repos.Payment, repos.Idempotency, paystackClient, logger),
		Withdrawal:  NewWithdrawalService(repos.Wallet, repos.Ledger, repos.Payment, repos.BankAccount, repos.Idempotency, paystackClient, logger),
		Ledger:      NewLedgerService(repos.Ledger, logger),
		Idempotency: NewIdempotencyService(repos.Idempotency, logger),
	}
}
