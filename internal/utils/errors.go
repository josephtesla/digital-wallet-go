package utils

import (
	"fmt"
)

// DomainError represents a business logic error
type DomainError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *DomainError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Predefined error codes
const (
	ErrCodeInsufficientFunds      = "INSUFFICIENT_FUNDS"
	ErrCodeWalletNotFound         = "WALLET_NOT_FOUND"
	ErrCodeUserNotFound           = "USER_NOT_FOUND"
	ErrCodeInvalidAmount          = "INVALID_AMOUNT"
	ErrCodeInvalidCurrency        = "INVALID_CURRENCY"
	ErrCodeDuplicateTransaction   = "DUPLICATE_TRANSACTION"
	ErrCodeTransactionNotFound    = "TRANSACTION_NOT_FOUND"
	ErrCodePaymentFailed          = "PAYMENT_FAILED"
	ErrCodeBankAccountNotFound    = "BANK_ACCOUNT_NOT_FOUND"
	ErrCodeInvalidBankAccount     = "INVALID_BANK_ACCOUNT"
	ErrCodeIdempotencyKeyExists   = "IDEMPOTENCY_KEY_EXISTS"
	ErrCodeLedgerBalanceMismatch  = "LEDGER_BALANCE_MISMATCH"
	ErrCodeInvalidTransactionType = "INVALID_TRANSACTION_TYPE"
)

// Error constructors
func NewInsufficientFundsError(details string) *DomainError {
	return &DomainError{
		Code:    ErrCodeInsufficientFunds,
		Message: "Insufficient funds for this transaction",
		Details: details,
	}
}

func NewWalletNotFoundError(walletID string) *DomainError {
	return &DomainError{
		Code:    ErrCodeWalletNotFound,
		Message: "Wallet not found",
		Details: walletID,
	}
}

func NewUserNotFoundError(userID string) *DomainError {
	return &DomainError{
		Code:    ErrCodeUserNotFound,
		Message: "User not found",
		Details: userID,
	}
}

func NewInvalidAmountError(amount string) *DomainError {
	return &DomainError{
		Code:    ErrCodeInvalidAmount,
		Message: "Invalid amount provided",
		Details: amount,
	}
}

func NewInvalidCurrencyError(currency string) *DomainError {
	return &DomainError{
		Code:    ErrCodeInvalidCurrency,
		Message: "Invalid currency provided",
		Details: currency,
	}
}

func NewDuplicateTransactionError(reference string) *DomainError {
	return &DomainError{
		Code:    ErrCodeDuplicateTransaction,
		Message: "Transaction with this reference already exists",
		Details: reference,
	}
}

func NewTransactionNotFoundError(reference string) *DomainError {
	return &DomainError{
		Code:    ErrCodeTransactionNotFound,
		Message: "Transaction not found",
		Details: reference,
	}
}

func NewPaymentFailedError(details string) *DomainError {
	return &DomainError{
		Code:    ErrCodePaymentFailed,
		Message: "Payment processing failed",
		Details: details,
	}
}

func NewBankAccountNotFoundError(accountID string) *DomainError {
	return &DomainError{
		Code:    ErrCodeBankAccountNotFound,
		Message: "Bank account not found",
		Details: accountID,
	}
}

func NewInvalidBankAccountError(details string) *DomainError {
	return &DomainError{
		Code:    ErrCodeInvalidBankAccount,
		Message: "Invalid bank account details",
		Details: details,
	}
}

func NewIdempotencyKeyExistsError(key string) *DomainError {
	return &DomainError{
		Code:    ErrCodeIdempotencyKeyExists,
		Message: "Request with this idempotency key already processed",
		Details: key,
	}
}

func NewLedgerBalanceMismatchError(details string) *DomainError {
	return &DomainError{
		Code:    ErrCodeLedgerBalanceMismatch,
		Message: "Ledger balance mismatch detected",
		Details: details,
	}
}

func NewInvalidTransactionTypeError(transactionType string) *DomainError {
	return &DomainError{
		Code:    ErrCodeInvalidTransactionType,
		Message: "Invalid transaction type",
		Details: transactionType,
	}
}

