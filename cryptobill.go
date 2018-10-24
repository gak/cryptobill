package cryptobill

import (
	"errors"
	"net/http"
	"net/http/cookiejar"
	"strings"
)

type CryptoBill struct {
	HttpClient *http.Client
}

type Service interface {
	Name() string
	ShortName() string
	Website() string
	Quote(cb *CryptoBill, amount Amount, fiat Currency) ([]QuoteResult, error)
	PayBPAY(cb *CryptoBill, bpay *BPAYInfo, crypto Currency, auth string) (*PayResult, error)
}

var Services = []Service{
	NewLivingRoom(),
	NewPaidByCoins(),
	NewBit2Bill(),
}

type Amount float64

type Pair struct {
	Fiat, Crypto Currency
}

type Conversion struct {
	Fiat, Crypto Amount
}

type QuoteResult struct {
	Service    Service
	Pair       Pair
	Conversion Conversion
}

type BPAYInfo struct {
	BillerCode    int
	BillerName    string
	BillerAccount string
	FiatCurrency  Currency
	FiatAmount    Amount
}

type PayResult struct {
	Address string
	Amount  Amount
}

func NewCryptoBill() *CryptoBill {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}

	return &CryptoBill{
		HttpClient: &http.Client{Jar: jar},
	}
}

func (cb *CryptoBill) PayBPAY(serviceName string, crypto Currency, bpay *BPAYInfo, auth string) (*PayResult, error) {
	for _, s := range Services {
		if strings.EqualFold(s.ShortName(), serviceName) {
			return s.PayBPAY(cb, bpay, crypto, auth)
		}
	}

	return nil, errors.New("unknown service: " + serviceName)
}
