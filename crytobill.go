package cryptobill

import "net/http"

type CryptoBill struct {
	HttpClient *http.Client
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
