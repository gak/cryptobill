package cryptobill

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net/http"
)

type PaidByCoins struct{}

type CurrenciesResponse struct {
	Message string
	Items   struct {
		CurrencyDetails []CurrencyDetail
	}
}

type CurrencyDetail struct {
	// "BTC", etc.
	ShortForm string

	// "BitcoinCash", etc. Used for calling other endpoints.
	Type string

	// An added charge on top of the order book cost.
	TransactionCharge float64
	BrokeragePercent  float64
	GSTPercent        float64
}

type OrderBookResponse struct {
	HighestBuy float64
}

// {"PrimaryCurrency":"BTC","SecondaryCurrency":"AUD","Price":8874.90,"ExchgID":1,"RTXVal":102.00000000
type ExchangeRateResponse struct {
	PrimaryCurrency   string
	SecondaryCurrency string
	Price             float64
	ExchgID           int
	RTXVal            float64
}

func NewPaidByCoins() Service {
	return &PaidByCoins{}
}

func (*PaidByCoins) Name() string {
	return "Paid By Coins"
}

func (*PaidByCoins) ShortName() string {
	return "PBC"
}

func (*PaidByCoins) Website() string {
	panic("implement me")
}

func (pbc *PaidByCoins) Quote(cb *CryptoBill, fiat Currency, amount Amount) ([]QuoteResult, error) {
	currencies, err := pbc.getCurrencies(cb)
	if err != nil {
		return nil, errors.Wrap(err, "get currencies")
	}

	var results []QuoteResult
	for _, currency := range currencies.Items.CurrencyDetails {
		exch, err := pbc.exchangeRate(cb, currency.ShortForm)
		if err != nil {
			return nil, err
		}

		crypto, err := NewCurrencyFromString(currency.ShortForm)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		finalAmount := amount / Amount(exch.Price)
		result := QuoteResult{
			Service:    pbc,
			Pair:       Pair{fiat, crypto},
			Conversion: Conversion{amount, finalAmount},
		}
		results = append(results, result)
	}

	return results, nil
}

func (pbc *PaidByCoins) getCurrencies(cb *CryptoBill) (*CurrenciesResponse, error) {
	url := "https://api.paidbycoins.com/tran/details"
	resp, err := pbc.request(cb, "GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "request")
	}

	currencies := CurrenciesResponse{}
	err = json.NewDecoder(resp.Body).Decode(&currencies)
	if err != nil {
		return nil, errors.Wrap(err, "decoding json from "+url)
	}

	if currencies.Message != "" {
		return nil, errors.New(currencies.Message)
	}

	return &currencies, nil
}

func (pbc *PaidByCoins) orderBook(cb *CryptoBill, currency string) (*OrderBookResponse, error) {
	url := fmt.Sprintf("https://api.paidbycoins.com/tran/obook/%v", currency)
	resp, err := pbc.request(cb, "GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "request")
	}

	book := &OrderBookResponse{}
	err = json.NewDecoder(resp.Body).Decode(book)
	if err != nil {
		return nil, errors.Wrap(err, "decoding json from "+url)
	}

	return book, nil
}

func (pbc *PaidByCoins) exchangeRate(cb *CryptoBill, currency string) (*ExchangeRateResponse, error) {
	url := fmt.Sprintf("https://api.paidbycoins.com/tran/exchgrate/%v", currency)
	resp, err := pbc.request(cb, "GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "request")
	}

	exch := &ExchangeRateResponse{}
	err = json.NewDecoder(resp.Body).Decode(exch)
	if err != nil {
		return nil, errors.Wrap(err, "decoding json from "+url)
	}

	return exch, nil
}

func (pbc *PaidByCoins) request(cb *CryptoBill, method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, errors.Wrap(err, "new request")
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3440.106 Safari/537.36")
	req.Header.Set("sw", "sexir")

	return cb.HttpClient.Do(req)
}
