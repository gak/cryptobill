package cryptobill

import (
	"errors"
	"strings"
)

type Currency string

var Currencies = map[string]Currency{
	"AUD": Currency("AUD"),

	"BTC":       Currency("BTC"),
	"ETH":       Currency("ETH"),
	"BCH":       Currency("BCH"),
	"LTC":       Currency("LTC"),
	"XRP":       Currency("XRP"),
	"LIGHTNING": Currency("LIGHTNING"),
}

func NewCurrencyFromString(s string) (Currency, error) {
	s = strings.ToUpper(s)

	if c, exists := Currencies[s]; exists {
		return c, nil
	} else {
		return c, errors.New("unknown currency: " + s)
	}
}
