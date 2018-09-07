package cryptobill

import "net/http"

type CryptoBill struct {
	HttpClient *http.Client
}

type Service interface {
	Name() string
	Website() string
	Quote(cb *CryptoBill, from Currency, amount Amount) ([]QuoteResult, error)
}

var Services = []Service{
	NewLivingRoom(),
	NewPaidByCoins(),
	NewBit2Bill(),
}

type Amount float64

type Pair struct {
	From, To Currency
}

type Conversion struct {
	From, To Amount
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
