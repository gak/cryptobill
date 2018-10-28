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
	Quote(cb *CryptoBill, info *FiatInfo) ([]QuoteResult, error)
	PayBPAY(cb *CryptoBill, bpay *PayBPAY) (*PayResult, error)
	PayEFT(cb *CryptoBill, eft *PayEFT) (*PayResult, error)
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

type PayResult struct {
	Address string
	Amount  Amount
}

type FiatInfo struct {
	Amount Amount   `arg help:"Fiat amount"`
	Fiat   Currency `arg help:"Fiat type, e.g. AUD"`
}

type PayInfo struct {
	FiatInfo
	Crypto Currency `arg help:"Cryptocurrency to spend"`
}

type PayInfoService struct {
	PayInfo
	Service string `arg help:"Service, e.g. PBC"`

	Auth string `required help:"For now only for your PBC email address."`
}

type PayBPAY struct {
	PayInfoService
	BPAY
}

type PayEFT struct {
	PayInfoService
	EFT
}

type BPAY struct {
	Code int `arg`

	// Populated dynamically
	Name string

	Account string `arg`
}

type EFT struct {
	BSB           int    `arg`
	AccountNumber int    `arg`
	AccountName   string `arg`
	Remitter      string `help:"Shown on the receiving bank statement."`
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

func (cb *CryptoBill) PayBPAY(bpay *PayBPAY) (*PayResult, error) {
	for _, s := range Services {
		if strings.EqualFold(s.ShortName(), bpay.Service) {
			return s.PayBPAY(cb, bpay)
		}
	}

	return nil, errors.New("unknown service: " + bpay.Service)
}

