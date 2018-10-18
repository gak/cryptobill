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
	"STEEM":     Currency("STEEM"),
	"PIVX":      Currency("PIVX"),
	"ZEC":       Currency("ZEC"),
	"ETC":       Currency("ETC"),
	"XMR":       Currency("XMR"),
	"DASH":      Currency("DASH"),
	"DOGE":      Currency("DOGE"),
	"BTX":       Currency("BTX"),
	"XEM":       Currency("XEM"),
	"SBD":       Currency("SBD"),
	"LIGHTNING": Currency("LIGHTNING"),
	"DCR":       Currency("DCR"),
}

func NewCurrencyFromString(s string) (Currency, error) {
	s = strings.ToUpper(s)

	if c, exists := Currencies[s]; exists {
		return c, nil
	} else {
		return c, errors.New("unknown currency: " + s)
	}
}
