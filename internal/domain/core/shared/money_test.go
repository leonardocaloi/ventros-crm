package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMoney_Valid(t *testing.T) {
	tests := []struct {
		name      string
		amount    float64
		currency  Currency
		wantCents int64
	}{
		{
			name:      "zero BRL",
			amount:    0.0,
			currency:  BRL,
			wantCents: 0,
		},
		{
			name:      "10.50 USD",
			amount:    10.50,
			currency:  USD,
			wantCents: 1050,
		},
		{
			name:      "99.99 EUR",
			amount:    99.99,
			currency:  EUR,
			wantCents: 9999,
		},
		{
			name:      "1.01 GBP",
			amount:    1.01,
			currency:  GBP,
			wantCents: 101,
		},
		{
			name:      "large amount",
			amount:    1000000.00,
			currency:  USD,
			wantCents: 100000000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewMoney(tt.amount, tt.currency)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantCents, m.Cents())
			assert.Equal(t, tt.currency, m.Currency())
		})
	}
}

func TestNewMoney_Invalid(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		currency Currency
		wantErr  error
	}{
		{
			name:     "negative amount",
			amount:   -10.50,
			currency: USD,
			wantErr:  ErrMoneyNegative,
		},
		{
			name:     "invalid currency",
			amount:   10.00,
			currency: Currency("INVALID"),
			wantErr:  ErrMoneyInvalidCurrency,
		},
		{
			name:     "empty currency",
			amount:   10.00,
			currency: Currency(""),
			wantErr:  ErrMoneyInvalidCurrency,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewMoney(tt.amount, tt.currency)
			assert.Error(t, err)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestNewMoneyFromCents(t *testing.T) {
	tests := []struct {
		name     string
		cents    int64
		currency Currency
		wantAmt  float64
	}{
		{
			name:     "1050 cents = 10.50",
			cents:    1050,
			currency: USD,
			wantAmt:  10.50,
		},
		{
			name:     "0 cents",
			cents:    0,
			currency: BRL,
			wantAmt:  0.00,
		},
		{
			name:     "1 cent",
			cents:    1,
			currency: EUR,
			wantAmt:  0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewMoneyFromCents(tt.cents, tt.currency)
			assert.NoError(t, err)
			assert.Equal(t, tt.cents, m.Cents())
			assert.Equal(t, tt.wantAmt, m.Amount())
		})
	}
}

func TestNewMoneyFromCents_Invalid(t *testing.T) {
	tests := []struct {
		name     string
		cents    int64
		currency Currency
		wantErr  error
	}{
		{
			name:     "negative cents",
			cents:    -100,
			currency: USD,
			wantErr:  ErrMoneyNegative,
		},
		{
			name:     "invalid currency",
			cents:    100,
			currency: Currency("XYZ"),
			wantErr:  ErrMoneyInvalidCurrency,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewMoneyFromCents(tt.cents, tt.currency)
			assert.Error(t, err)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestZero(t *testing.T) {
	m := Zero(BRL)
	assert.Equal(t, int64(0), m.Cents())
	assert.Equal(t, BRL, m.Currency())
	assert.True(t, m.IsZero())
	assert.False(t, m.IsPositive())
}

func TestMoney_Amount(t *testing.T) {
	tests := []struct {
		name    string
		cents   int64
		wantAmt float64
	}{
		{
			name:    "1050 cents = 10.50",
			cents:   1050,
			wantAmt: 10.50,
		},
		{
			name:    "100 cents = 1.00",
			cents:   100,
			wantAmt: 1.00,
		},
		{
			name:    "1 cent = 0.01",
			cents:   1,
			wantAmt: 0.01,
		},
		{
			name:    "0 cents = 0.00",
			cents:   0,
			wantAmt: 0.00,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, _ := NewMoneyFromCents(tt.cents, USD)
			assert.Equal(t, tt.wantAmt, m.Amount())
		})
	}
}

func TestMoney_Add(t *testing.T) {
	tests := []struct {
		name      string
		m1        Money
		m2        Money
		wantCents int64
		wantErr   error
	}{
		{
			name:      "10.50 + 5.25 = 15.75",
			m1:        mustMoney(10.50, USD),
			m2:        mustMoney(5.25, USD),
			wantCents: 1575,
			wantErr:   nil,
		},
		{
			name:      "0 + 10 = 10",
			m1:        Zero(BRL),
			m2:        mustMoney(10.00, BRL),
			wantCents: 1000,
			wantErr:   nil,
		},
		{
			name:      "different currencies",
			m1:        mustMoney(10.00, USD),
			m2:        mustMoney(10.00, BRL),
			wantCents: 0,
			wantErr:   ErrMoneyDifferentCurrency,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.m1.Add(tt.m2)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCents, result.Cents())
			}
		})
	}
}

func TestMoney_Subtract(t *testing.T) {
	tests := []struct {
		name      string
		m1        Money
		m2        Money
		wantCents int64
		wantErr   error
	}{
		{
			name:      "10.50 - 5.25 = 5.25",
			m1:        mustMoney(10.50, USD),
			m2:        mustMoney(5.25, USD),
			wantCents: 525,
			wantErr:   nil,
		},
		{
			name:      "10 - 0 = 10",
			m1:        mustMoney(10.00, BRL),
			m2:        Zero(BRL),
			wantCents: 1000,
			wantErr:   nil,
		},
		{
			name:      "10 - 10 = 0",
			m1:        mustMoney(10.00, EUR),
			m2:        mustMoney(10.00, EUR),
			wantCents: 0,
			wantErr:   nil,
		},
		{
			name:      "negative result",
			m1:        mustMoney(5.00, USD),
			m2:        mustMoney(10.00, USD),
			wantCents: 0,
			wantErr:   ErrMoneyNegative,
		},
		{
			name:      "different currencies",
			m1:        mustMoney(10.00, USD),
			m2:        mustMoney(5.00, EUR),
			wantCents: 0,
			wantErr:   ErrMoneyDifferentCurrency,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.m1.Subtract(tt.m2)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCents, result.Cents())
			}
		})
	}
}

func TestMoney_Multiply(t *testing.T) {
	tests := []struct {
		name      string
		m         Money
		factor    float64
		wantCents int64
		wantErr   error
	}{
		{
			name:      "10.00 * 2 = 20.00",
			m:         mustMoney(10.00, USD),
			factor:    2.0,
			wantCents: 2000,
			wantErr:   nil,
		},
		{
			name:      "10.50 * 1.5 = 15.75",
			m:         mustMoney(10.50, USD),
			factor:    1.5,
			wantCents: 1575,
			wantErr:   nil,
		},
		{
			name:      "10.00 * 0 = 0.00",
			m:         mustMoney(10.00, BRL),
			factor:    0.0,
			wantCents: 0,
			wantErr:   nil,
		},
		{
			name:      "negative factor",
			m:         mustMoney(10.00, USD),
			factor:    -2.0,
			wantCents: 0,
			wantErr:   ErrMoneyNegative,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.m.Multiply(tt.factor)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCents, result.Cents())
			}
		})
	}
}

func TestMoney_IsZero(t *testing.T) {
	assert.True(t, Zero(USD).IsZero())
	assert.True(t, mustMoney(0.00, BRL).IsZero())
	assert.False(t, mustMoney(0.01, EUR).IsZero())
	assert.False(t, mustMoney(10.00, GBP).IsZero())
}

func TestMoney_IsPositive(t *testing.T) {
	assert.False(t, Zero(USD).IsPositive())
	assert.False(t, mustMoney(0.00, BRL).IsPositive())
	assert.True(t, mustMoney(0.01, EUR).IsPositive())
	assert.True(t, mustMoney(10.00, GBP).IsPositive())
}

func TestMoney_GreaterThan(t *testing.T) {
	tests := []struct {
		name    string
		m1      Money
		m2      Money
		want    bool
		wantErr error
	}{
		{
			name:    "10 > 5",
			m1:      mustMoney(10.00, USD),
			m2:      mustMoney(5.00, USD),
			want:    true,
			wantErr: nil,
		},
		{
			name:    "5 > 10 = false",
			m1:      mustMoney(5.00, USD),
			m2:      mustMoney(10.00, USD),
			want:    false,
			wantErr: nil,
		},
		{
			name:    "10 > 10 = false",
			m1:      mustMoney(10.00, USD),
			m2:      mustMoney(10.00, USD),
			want:    false,
			wantErr: nil,
		},
		{
			name:    "different currencies",
			m1:      mustMoney(10.00, USD),
			m2:      mustMoney(5.00, BRL),
			want:    false,
			wantErr: ErrMoneyDifferentCurrency,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.m1.GreaterThan(tt.m2)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestMoney_LessThan(t *testing.T) {
	tests := []struct {
		name    string
		m1      Money
		m2      Money
		want    bool
		wantErr error
	}{
		{
			name:    "5 < 10",
			m1:      mustMoney(5.00, USD),
			m2:      mustMoney(10.00, USD),
			want:    true,
			wantErr: nil,
		},
		{
			name:    "10 < 5 = false",
			m1:      mustMoney(10.00, USD),
			m2:      mustMoney(5.00, USD),
			want:    false,
			wantErr: nil,
		},
		{
			name:    "10 < 10 = false",
			m1:      mustMoney(10.00, USD),
			m2:      mustMoney(10.00, USD),
			want:    false,
			wantErr: nil,
		},
		{
			name:    "different currencies",
			m1:      mustMoney(5.00, USD),
			m2:      mustMoney(10.00, EUR),
			want:    false,
			wantErr: ErrMoneyDifferentCurrency,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.m1.LessThan(tt.m2)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestMoney_Equals(t *testing.T) {
	m1 := mustMoney(10.50, USD)
	m2 := mustMoney(10.50, USD)
	m3 := mustMoney(10.50, BRL)
	m4 := mustMoney(5.00, USD)

	assert.True(t, m1.Equals(m2))
	assert.False(t, m1.Equals(m3)) // different currency
	assert.False(t, m1.Equals(m4)) // different amount
}

func TestMoney_String(t *testing.T) {
	tests := []struct {
		name string
		m    Money
		want string
	}{
		{
			name: "USD 10.50",
			m:    mustMoney(10.50, USD),
			want: "USD 10.50",
		},
		{
			name: "BRL 99.99",
			m:    mustMoney(99.99, BRL),
			want: "BRL 99.99",
		},
		{
			name: "EUR 0.00",
			m:    Zero(EUR),
			want: "EUR 0.00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.m.String())
		})
	}
}

func TestMoney_Format(t *testing.T) {
	tests := []struct {
		name string
		m    Money
		want string
	}{
		{
			name: "USD with symbol",
			m:    mustMoney(10.50, USD),
			want: "$10.50",
		},
		{
			name: "BRL with symbol",
			m:    mustMoney(99.99, BRL),
			want: "R$99.99",
		},
		{
			name: "EUR with symbol",
			m:    mustMoney(50.00, EUR),
			want: "€50.00",
		},
		{
			name: "GBP with symbol",
			m:    mustMoney(25.75, GBP),
			want: "£25.75",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.m.Format())
		})
	}
}

func TestCurrency_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		currency Currency
		want     bool
	}{
		{
			name:     "USD is valid",
			currency: USD,
			want:     true,
		},
		{
			name:     "BRL is valid",
			currency: BRL,
			want:     true,
		},
		{
			name:     "EUR is valid",
			currency: EUR,
			want:     true,
		},
		{
			name:     "GBP is valid",
			currency: GBP,
			want:     true,
		},
		{
			name:     "XYZ is invalid",
			currency: Currency("XYZ"),
			want:     false,
		},
		{
			name:     "empty is invalid",
			currency: Currency(""),
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.currency.IsValid())
		})
	}
}

// Helper function for tests
func mustMoney(amount float64, currency Currency) Money {
	m, err := NewMoney(amount, currency)
	if err != nil {
		panic(err)
	}
	return m
}
