package shared

import (
	"errors"
	"fmt"
)

var (
	ErrMoneyNegative          = errors.New("money amount cannot be negative")
	ErrMoneyInvalidCurrency   = errors.New("invalid currency code")
	ErrMoneyDifferentCurrency = errors.New("cannot operate on different currencies")
)

type Currency string

const (
	USD Currency = "USD"
	BRL Currency = "BRL"
	EUR Currency = "EUR"
	GBP Currency = "GBP"
)

func (c Currency) IsValid() bool {
	switch c {
	case USD, BRL, EUR, GBP:
		return true
	default:
		return false
	}
}

type Money struct {
	cents    int64
	currency Currency
}

func NewMoney(amount float64, currency Currency) (Money, error) {
	if !currency.IsValid() {
		return Money{}, ErrMoneyInvalidCurrency
	}

	cents := int64(amount * 100)

	if cents < 0 {
		return Money{}, ErrMoneyNegative
	}

	return Money{
		cents:    cents,
		currency: currency,
	}, nil
}

func NewMoneyFromCents(cents int64, currency Currency) (Money, error) {
	if !currency.IsValid() {
		return Money{}, ErrMoneyInvalidCurrency
	}

	if cents < 0 {
		return Money{}, ErrMoneyNegative
	}

	return Money{
		cents:    cents,
		currency: currency,
	}, nil
}

func Zero(currency Currency) Money {
	return Money{
		cents:    0,
		currency: currency,
	}
}

func (m Money) Cents() int64 {
	return m.cents
}

func (m Money) Currency() Currency {
	return m.currency
}

func (m Money) Amount() float64 {
	return float64(m.cents) / 100.0
}

func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, ErrMoneyDifferentCurrency
	}

	return Money{
		cents:    m.cents + other.cents,
		currency: m.currency,
	}, nil
}

func (m Money) Subtract(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, ErrMoneyDifferentCurrency
	}

	newCents := m.cents - other.cents
	if newCents < 0 {
		return Money{}, ErrMoneyNegative
	}

	return Money{
		cents:    newCents,
		currency: m.currency,
	}, nil
}

func (m Money) Multiply(factor float64) (Money, error) {
	if factor < 0 {
		return Money{}, ErrMoneyNegative
	}

	newCents := int64(float64(m.cents) * factor)

	return Money{
		cents:    newCents,
		currency: m.currency,
	}, nil
}

func (m Money) IsZero() bool {
	return m.cents == 0
}

func (m Money) IsPositive() bool {
	return m.cents > 0
}

func (m Money) GreaterThan(other Money) (bool, error) {
	if m.currency != other.currency {
		return false, ErrMoneyDifferentCurrency
	}
	return m.cents > other.cents, nil
}

func (m Money) LessThan(other Money) (bool, error) {
	if m.currency != other.currency {
		return false, ErrMoneyDifferentCurrency
	}
	return m.cents < other.cents, nil
}

func (m Money) Equals(other Money) bool {
	return m.cents == other.cents && m.currency == other.currency
}

func (m Money) String() string {
	return fmt.Sprintf("%s %.2f", m.currency, m.Amount())
}

func (m Money) Format() string {
	symbol := m.currencySymbol()
	return fmt.Sprintf("%s%.2f", symbol, m.Amount())
}

func (m Money) currencySymbol() string {
	switch m.currency {
	case USD:
		return "$"
	case BRL:
		return "R$"
	case EUR:
		return "€"
	case GBP:
		return "£"
	default:
		return string(m.currency) + " "
	}
}
