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
	Message string `json:"Message"`
	Items   struct {
		Currencies []CurrencyDetail `json:"CurrencyDetails"`
	} `json:"Items"`
}

type CurrencyDetail struct {
	ShortForm string `json:"ShortForm"`
	Type      string `json:"Type"`
}

type OrderBookResponse struct {
	HighestBuy float64 `json:"HighestBuy"`
}

func NewPaidByCoins() Service {
	return &PaidByCoins{}
}

func (*PaidByCoins) Name() string {
	return "Paid By Coins"
}

func (*PaidByCoins) Website() string {
	panic("implement me")
}

func (pbp *PaidByCoins) Quote(cb *CryptoBill, from Currency, amount Amount) ([]QuoteResult, error) {
	currencies, err := pbp.getCurrencies(cb)
	if err != nil {
		return nil, errors.Wrap(err, "get currencies")
	}

	var results []QuoteResult
	for _, currency := range currencies.Items.Currencies {
		book, err := pbp.orderBook(cb, currency.Type)
		if err != nil {
			return nil, err
		}

		to, err := NewCurrencyFromString(currency.ShortForm)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		fmt.Println(currency, book)
		result := QuoteResult{
			Service:    pbp,
			Pair:       Pair{from, to},
			Conversion: Conversion{amount, amount / Amount(book.HighestBuy)},
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

	fmt.Println(currencies)

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
