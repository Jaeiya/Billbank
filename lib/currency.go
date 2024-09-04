package lib

import (
	"fmt"
	"strconv"
	"strings"
)

type CurrencyType = uint8

const (
	USD = CurrencyType(iota)
	EUR
	JPY
)

var (
	ErrCurrencyFloat    = fmt.Errorf("failed to parse input as float")
	ErrCurrencyInt      = fmt.Errorf("failed to parse input as int")
	ErrUSDDollar        = fmt.Errorf("not a valid dollar amount")
	ErrUSDCents         = fmt.Errorf("too precise; round to cents only")
	ErrCurrencyType     = fmt.Errorf("using an unsupported currency type")
	ErrCurrencyConflict = fmt.Errorf("currency types do not match")
)

type Currency struct {
	amount       int
	currencyType CurrencyType
}

func NewCurrency(amount string, ct CurrencyType) *Currency {
	c := &Currency{}
	switch ct {
	case USD:
		if amount != "" {
			err := c.Add(amount)
			if err != nil {
				panic(err)
			}
		}
	default:
		panic(ErrCurrencyType)

	}
	return c
}

func (c *Currency) Add(amount string) error {
	switch c.currencyType {

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
		panic(ErrCurrencyType)

	}
	return nil
}

func (c *Currency) AddCurrency(currencies ...*Currency) {
	for _, c2 := range currencies {
		if c.currencyType != c2.currencyType {
			panic(ErrCurrencyConflict)
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

func (c *Currency) SubtractCurrency(currencies ...*Currency) {
	for _, c2 := range currencies {
		if c.currencyType != c2.currencyType {
			panic(ErrCurrencyConflict)
		}
		c.amount -= c2.amount
	}
}

func (c *Currency) String() string {
	switch c.currencyType {
	case USD:
		return fmt.Sprintf("$%d.%02d", c.amount/100, c.amount%100)
	default:
		panic(ErrCurrencyType)
	}
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
