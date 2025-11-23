package utils

import (
	"strconv"

	"github.com/shopspring/decimal"
)

// Money represents monetary amounts in kobo (smallest currency unit)
type Money struct {
	Amount   decimal.Decimal `json:"amount"`
	Currency string          `json:"currency"`
}

// NewMoney creates a new Money instance from kobo amount
func NewMoney(koboAmount int64, currency string) Money {
	return Money{
		Amount:   decimal.NewFromInt(koboAmount).Div(decimal.NewFromInt(100)),
		Currency: currency,
	}
}

// NewMoneyFromNaira creates a new Money instance from naira amount
func NewMoneyFromNaira(nairaAmount decimal.Decimal, currency string) Money {
	return Money{
		Amount:   nairaAmount,
		Currency: currency,
	}
}

// ToKobo converts the amount to kobo (smallest currency unit)
func (m Money) ToKobo() int64 {
	return m.Amount.Mul(decimal.NewFromInt(100)).IntPart()
}

// ToNaira returns the amount in naira
func (m Money) ToNaira() decimal.Decimal {
	return m.Amount
}

// Add adds another Money amount to this one
func (m Money) Add(other Money) Money {
	if m.Currency != other.Currency {
		panic("Cannot add amounts with different currencies")
	}
	return Money{
		Amount:   m.Amount.Add(other.Amount),
		Currency: m.Currency,
	}
}

// Subtract subtracts another Money amount from this one
func (m Money) Subtract(other Money) Money {
	if m.Currency != other.Currency {
		panic("Cannot subtract amounts with different currencies")
	}
	return Money{
		Amount:   m.Amount.Sub(other.Amount),
		Currency: m.Currency,
	}
}

// IsZero checks if the amount is zero
func (m Money) IsZero() bool {
	return m.Amount.IsZero()
}

// IsPositive checks if the amount is positive
func (m Money) IsPositive() bool {
	return m.Amount.IsPositive()
}

// IsNegative checks if the amount is negative
func (m Money) IsNegative() bool {
	return m.Amount.IsNegative()
}

// GreaterThan checks if this amount is greater than another
func (m Money) GreaterThan(other Money) bool {
	if m.Currency != other.Currency {
		panic("Cannot compare amounts with different currencies")
	}
	return m.Amount.GreaterThan(other.Amount)
}

// GreaterThanOrEqual checks if this amount is greater than or equal to another
func (m Money) GreaterThanOrEqual(other Money) bool {
	if m.Currency != other.Currency {
		panic("Cannot compare amounts with different currencies")
	}
	return m.Amount.GreaterThanOrEqual(other.Amount)
}

// LessThan checks if this amount is less than another
func (m Money) LessThan(other Money) bool {
	if m.Currency != other.Currency {
		panic("Cannot compare amounts with different currencies")
	}
	return m.Amount.LessThan(other.Amount)
}

// LessThanOrEqual checks if this amount is less than or equal to another
func (m Money) LessThanOrEqual(other Money) bool {
	if m.Currency != other.Currency {
		panic("Cannot compare amounts with different currencies")
	}
	return m.Amount.LessThanOrEqual(other.Amount)
}

// Equal checks if this amount is equal to another
func (m Money) Equal(other Money) bool {
	if m.Currency != other.Currency {
		panic("Cannot compare amounts with different currencies")
	}
	return m.Amount.Equal(other.Amount)
}

// String returns the string representation of the amount
func (m Money) String() string {
	return m.Amount.String() + " " + m.Currency
}

// Format formats the amount with the specified precision
func (m Money) Format(precision int32) string {
	return m.Amount.Round(precision).String() + " " + m.Currency
}

// ParseKoboAmount parses a string amount in kobo to int64
func ParseKoboAmount(amountStr string) (int64, error) {
	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return amount, nil
}

// ParseNairaAmount parses a string amount in naira to decimal
func ParseNairaAmount(amountStr string) (decimal.Decimal, error) {
	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		return decimal.Zero, err
	}
	return amount, nil
}

