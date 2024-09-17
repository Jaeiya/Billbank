package lib

import (
	"fmt"
	"strconv"
	"strings"
)

type CurrencyCode uint8

const (
	USD = CurrencyCode(iota)
	EUR
	JPY
)

var (
	ErrCurrencyFloat = fmt.Errorf("failed to parse input as float")
	ErrCurrencyInt   = fmt.Errorf("failed to parse input as int")
	ErrUSDDollar     = fmt.Errorf("not a valid dollar amount")
	ErrUSDCents      = fmt.Errorf("too precise; round to cents only")
	ErrCurrency      = fmt.Errorf("using an unsupported currency")
	ErrCurrencyKind  = fmt.Errorf("cannot combine currencies of different kinds")
)

type Currency struct {
	amount int
	code   CurrencyCode
}

func NewCurrency(amount string, code CurrencyCode) Currency {
	c := Currency{}
	switch code {
	case USD:
		if amount != "" {
			err := c.Add(amount)
			if err != nil {
				panic(err)
			}
		}
	default:
		panic(ErrCurrency)

	}
	return c
}

func (c *Currency) Add(amount string) error {
	switch c.code {

	case USD:
		err := verifyUSDAmount(amount)
		if err != nil {
			return err
		}

		intAmount, err := toUSDCents(amount)
		if err != nil {
			return err
		}
		c.amount += intAmount

	default:
		panic(ErrCurrency)

	}
	return nil
}

func (c *Currency) AddCurrency(currencies ...Currency) {
	for _, c2 := range currencies {
		if c.code != c2.code {
			panic(ErrCurrencyKind)
		}
		c.amount += c2.amount
	}
}

func (c *Currency) Subtract(amount string) error {
	err := c.Add("-" + amount)
	if err != nil {
		return err
	}
	return nil
}

func (c *Currency) SubtractCurrency(currencies ...Currency) {
	for _, c2 := range currencies {
		if c.code != c2.code {
			panic(ErrCurrencyKind)
		}
		c.amount -= c2.amount
	}
}

func (c *Currency) Set(amount string) error {
	switch c.code {
	case USD:
		if err := verifyUSDAmount(amount); err != nil {
			return err
		}

		cents, err := toUSDCents(amount)
		if err != nil {
			return err
		}
		c.amount = cents

	default:
		panic(ErrCurrencyKind)
	}

	return nil
}

func (c *Currency) SetCurrency(c2 Currency) {
	if c.code != c2.code {
		panic(ErrCurrencyKind)
	}
	c.amount = c2.amount
}

func (c Currency) String() string {
	switch c.code {
	case USD:
		return fmt.Sprintf("$%d.%02d", c.amount/100, c.amount%100)
	default:
		panic(ErrCurrency)
	}
}

/*
ToInt returns the total amount in the lowest denomination
for that currency. Should only be used for storage.
*/
func (c *Currency) ToInt() int {
	return c.amount
}

/*
LoadAmount loads an amount as the smallest denomination for the
currency. For example, if the currency is USD, then the amount
is assumed to be in Cents, not Dollars.
*/
func (c *Currency) LoadAmount(amount int) Currency {
	c.amount = amount
	return *c
}

/*
ToCurrency tries to return 'v' as a Currency
*/
func ToCurrency(v any) (Currency, error) {
	if _, ok := v.(Currency); ok {
		return v.(Currency), nil
	}
	return Currency{}, fmt.Errorf("not a currency")
}

func verifyUSDAmount(amount string) error {
	if !strings.Contains(amount, ".") {
		if _, err := strconv.ParseInt(amount, 10, 64); err != nil {
			return ErrUSDDollar
		}
		return nil
	}

	if _, err := strconv.ParseFloat(amount, 64); err != nil {
		return ErrCurrencyFloat
	}

	s := strings.Split(amount, ".")[1]
	if len(s) > 2 {
		return ErrUSDCents
	}
	return nil
}

func toUSDCents(amount string) (int, error) {
	if !strings.Contains(amount, ".") {
		intAmount, err := strconv.Atoi(amount)
		if err != nil {
			return 0, ErrCurrencyInt
		}
		return intAmount * 100, nil
	}

	// Add hundredths placeholder
	parts := strings.Split(amount, ".")
	if len(parts[1]) == 1 {
		parts[1] += "0"
	}
	amount = parts[0] + parts[1]

	intAmount, err := strconv.Atoi(amount)
	if err != nil {
		return 0, ErrCurrencyInt
	}
	return intAmount, nil
}
