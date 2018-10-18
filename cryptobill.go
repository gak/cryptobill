package cryptobill

import "net/http"

type CryptoBill struct {
	HttpClient *http.Client
}

type Service interface {
	Name() string
	ShortName() string
	Website() string
	Quote(cb *CryptoBill, fiat Currency, amount Amount) ([]QuoteResult, error)
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

func NewCryptoBill() *CryptoBill {
	return &CryptoBill{
		HttpClient: &http.Client{},
	}
}
