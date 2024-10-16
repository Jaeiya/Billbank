package lib

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type CurrencyFunc = int

type MockAddSubtract struct {
	should            string
	amountsAdded      []string
	totalAdded        string
	amountsSubtracted []string
	totalSubtracted   string
}

type MockAddSubCurrency struct {
	should             string
	currencyAdded      []*Currency
	totalAdded         string
	initialCurrency    *Currency
	currencySubtracted []*Currency
	totalSubtracted    string
}

func TestCurrency(t *testing.T) {
	mockTable := []MockAddSubtract{
		{
			should: "add dollars and cents together",
			amountsAdded: []string{
				"1.33", "2.12", "10.11",
			},
			totalAdded: "$13.56",
		},
		{
			should: "add cents without thousandths placeholder",
			amountsAdded: []string{
				".2", ".4", ".1",
			},
			totalAdded: "$0.70",
		},
		{
			should: "add dollars and cents without thousandths placeholder",
			amountsAdded: []string{
				"1.2", "2.4", "7.1",
			},
			totalAdded: "$10.70",
		},
		{
			should: "add all possible fractional amounts",
			amountsAdded: []string{
				"1.20", ".2", ".1", "7.20", "5.32", "2.1", ".75", "8.50",
			},
			totalAdded: "$25.37",
		},
		{
			should: "add whole dollars",
			amountsAdded: []string{
				"1", "3", "7", "2",
			},
			totalAdded: "$13.00",
		},
		{
			should: "subtract dollars and cents together",
			amountsAdded: []string{
				"20.32",
			},
			totalAdded: "$20.32",
			amountsSubtracted: []string{
				"5.32", "1.00",
			},
			totalSubtracted: "$14.00",
		},
		{
			should: "subtract all possible dollar fractional amounts",
			amountsAdded: []string{
				"25.37",
			},
			totalAdded: "$25.37",
			amountsSubtracted: []string{
				"1.20", ".2", ".1", "7.20", "5.32", "2.1", ".75", "8.50",
			},
			totalSubtracted: "$0.00",
		},
		{
			should: "subtract whole dollars",
			amountsAdded: []string{
				"13.00",
			},
			totalAdded: "$13.00",
			amountsSubtracted: []string{
				"1", "3", "7", "2",
			},
			totalSubtracted: "$0.00",
		},
	}

	for _, mock := range mockTable {
		t.Run("should "+mock.should, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			c := NewCurrency("", USD)

			for _, amount := range mock.amountsAdded {
				err := c.Add(amount)
				require.NoError(t, err, "there should be no parsing errors")
			}
			if len(mock.amountsAdded) > 0 {
				a.Equal(mock.totalAdded, c.String())
			}

			for _, amount := range mock.amountsSubtracted {
				err := c.Subtract(amount)
				a.NoError(err, "there should be no parsing errors")
			}
			if len(mock.amountsSubtracted) > 0 {
				a.Equal(mock.totalSubtracted, c.String())
			}
		})
	}

	t.Run("should set currency", func(t *testing.T) {
		t.Parallel()
		c := NewCurrency("", USD)
		a := assert.New(t)

		c.Set("1.56")
		a.Equal(c.String(), "$1.56")

		c.SetCurrency(NewCurrency(".5", USD))
		a.Equal(c.String(), "$0.50")
	})

	t.Run("should error with non-number USD amount", func(t *testing.T) {
		t.Parallel()
		c := NewCurrency("", USD)
		a := assert.New(t)
		a.ErrorIs(c.Add("hello there"), ErrUSDDollar)
	})

	t.Run("should error with invalid float USD amount", func(t *testing.T) {
		t.Parallel()
		c := NewCurrency("", USD)
		a := assert.New(t)
		a.ErrorIs(c.Add("t.t"), ErrCurrencyFloat)
	})

	t.Run("should panic with unsupported currency", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)
		a.PanicsWithError(ErrCurrency.Error(), func() { NewCurrency("", EUR) })

		c := NewCurrency("", USD)
		c.code = EUR
		c2 := NewCurrency("", USD)
		a.PanicsWithError(ErrCurrencyKind.Error(), func() { c.AddCurrency(c2) })
		a.PanicsWithError(ErrCurrencyKind.Error(), func() { c.SubtractCurrency(c2) })
		a.PanicsWithError(ErrCurrencyKind.Error(), func() { c.SetCurrency(c2) })
	})

	t.Run("should error with micro-cent USD amount", func(t *testing.T) {
		t.Parallel()
		c := NewCurrency("", USD)
		a := assert.New(t)
		a.ErrorIs(c.Add("24.232384"), ErrUSDCents)
	})

	t.Run("should error if conversion to cents is non-int USD amount", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)
		_, err := toUSDCents("abc")
		a.ErrorIs(err, ErrCurrencyInt)

		_, err = toUSDCents("a.b")
		a.ErrorIs(err, ErrCurrencyInt)
	})
}
