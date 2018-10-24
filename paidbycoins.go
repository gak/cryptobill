package cryptobill

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/alecthomas/repr"
	"github.com/nu7hatch/gouuid"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
)

type PaidByCoins struct {
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

func (pbc *PaidByCoins) Quote(cb *CryptoBill, amount Amount, fiat Currency) ([]QuoteResult, error) {
	currencies, err := pbc.getCurrencies(cb)
	if err != nil {
		return nil, errors.Wrap(err, "get currencies")
	}

	var results []QuoteResult
	for _, currency := range currencies.Items.CurrencyDetails {
		crypto, err := NewCurrencyFromString(currency.ShortForm)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		exch, err := pbc.exchangeRate(cb, crypto)
		if err != nil {
			return nil, err
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

func (pbc *PaidByCoins) PayBPAY(cb *CryptoBill, bpay *BPAYInfo, crypto Currency, email string) (*PayResult, error) {
	exchResp, err := pbc.exchangeRate(cb, crypto)
	if err != nil {
		return nil, errors.Wrap(err, "exchangeRate")
	}

	currencies, err := pbc.getCurrencies(cb)
	if err != nil {
		return nil, errors.Wrap(err, "getCurrencies")
	}

	var currencyDetail *CurrencyDetail
	for _, c := range currencies.Items.CurrencyDetails {
		if c.ShortForm == string(crypto) {
			currencyDetail = &c
			break
		}
	}
	if currencyDetail == nil {
		return nil, errors.New("unknown crypto currency: " + string(crypto))
	}

	err = pbc.fillBillerName(cb, bpay)
	if err != nil {
		return nil, errors.Wrap(err, "fill biller name")
	}

	txReq, err := newTxReq(exchResp, bpay, currencyDetail, email)
	if err != nil {
		return nil, errors.Wrap(err, "newTxReq")
	}

	txAddResp, err := pbc.transactionAdd(cb, txReq)
	if err != nil {
		return nil, errors.Wrap(err, "transactionAdd")
	}

	repr.Println(txAddResp)

	//payResult := pbc.makePayResult(txAddResp)

	return nil, nil
}

type VerifyEmailResponse struct {
	Message    string
	IsVerified bool
}

func (pbc *PaidByCoins) verifyEmail(cb *CryptoBill, email string) error {
	url := "https://api.paidbycoins.com/email/veml?email=" + email
	resp, err := pbc.request(cb, "GET", url, nil)
	if err != nil {
		return errors.Wrap(err, "request")
	}

	verify := VerifyEmailResponse{}
	err = json.NewDecoder(resp.Body).Decode(&verify)
	if err != nil {
		return errors.Wrap(err, "decoding json")
	}

	if verify.Message != "" {
		return errors.New(verify.Message)
	}

	repr.Println(verify)

	return nil
}

type VerifyPinRequest struct {
	Email, Pin string
}

func (pbc *PaidByCoins) verifyPin(cb *CryptoBill, email, pin string) error {
	vpr := VerifyPinRequest{
		Email: email,
		Pin:   pin,
	}

	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(vpr)
	if err != nil {
		return errors.Wrap(err, "encoding json")
	}

	url := "https://api.paidbycoins.com/email/vep"
	resp, err := pbc.request(cb, "POST", url, body)
	if err != nil {
		return errors.Wrap(err, "request")
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "readall")
	}

	state := string(data)
	if state != "true" {
		return errors.New("unexpected body: " + state)
	}

	return nil
}

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

type OrderBookResponse struct {
	HighestBuy float64
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

type ExchangeRateResponse struct {
	PrimaryCurrency   string
	SecondaryCurrency string
	Price             float64
	ExchgID           int
	RTXVal            float64
}

func (pbc *PaidByCoins) exchangeRate(cb *CryptoBill, crypto Currency) (*ExchangeRateResponse, error) {
	url := fmt.Sprintf("https://api.paidbycoins.com/tran/exchgrate/%v", crypto)
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

type TransactionAddRequest struct {
	BillerCode               int
	BillerName               string
	RefCode                  string
	EnteredAmount            float64
	CurrencyType             string
	EnteredCurrency          string
	CurrencyExchRate         float64
	TotalAmount              float64
	Email                    string
	HasEmail                 bool
	SessionID                string
	AlternateAddress         string
	TransactionServiceAmount int
	RTXVal                   float64
	QuoteExchgID             int
}

type TransactionAddResponse struct {
	Message     string
	ToAddress   string
	TotalAmount float64
}

func (pbc *PaidByCoins) transactionAdd(cb *CryptoBill, txReq *TransactionAddRequest) (*TransactionAddResponse, error) {
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(txReq)
	if err != nil {
		return nil, errors.Wrap(err, "encoding json")
	}

	url := fmt.Sprintf("https://api.paidbycoins.com/tran/add")
	resp, err := pbc.request(cb, "POST", url, body)
	if err != nil {
		return nil, errors.Wrap(err, "request")
	}

	exch := &TransactionAddResponse{}
	err = json.NewDecoder(resp.Body).Decode(exch)
	if err != nil {
		return nil, errors.Wrap(err, "decoding json from "+url)
	}

	if exch.Message != "" {
		return nil, errors.New(exch.Message)
	}

	return exch, nil
}

func (pbc *PaidByCoins) fillBillerName(cb *CryptoBill, info *BPAYInfo) error {
	url := fmt.Sprintf("https://api.paidbycoins.com/common/biller/%v", info.BillerCode)
	resp, err := pbc.request(cb, "GET", url, nil)
	if err != nil {
		return errors.Wrap(err, "request")
	}

	err = json.NewDecoder(resp.Body).Decode(&info.BillerName)
	if err != nil {
		return errors.Wrap(err, "decoding json from "+url)
	}

	return nil
}

// Helpers

func newTxReq(exchResp *ExchangeRateResponse, bpay *BPAYInfo, currencyDetail *CurrencyDetail, email string) (*TransactionAddRequest, error) {
	sessionId, err := uuid.NewV4()
	if err != nil {
		return nil, errors.Wrap(err, "uuid")
	}

	totalAmount := float64(bpay.FiatAmount) / exchResp.Price
	//totalAmount = math.Round(totalAmount*1e5) / 1e5

	tranReq := &TransactionAddRequest{
		SessionID: sessionId.String(),

		HasEmail: true,
		Email:    email,

		BillerCode:      bpay.BillerCode,
		BillerName:      bpay.BillerName,
		RefCode:         bpay.BillerAccount,
		EnteredCurrency: string(bpay.FiatCurrency),
		EnteredAmount:   float64(bpay.FiatAmount),

		CurrencyExchRate: exchResp.Price,
		RTXVal:           exchResp.RTXVal,
		QuoteExchgID:     exchResp.ExchgID,
		TotalAmount:      totalAmount,
		CurrencyType:     currencyDetail.Type,
	}

	return tranReq, nil
}

func (pbc *PaidByCoins) request(cb *CryptoBill, method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, errors.Wrap(err, "new request")
	}

	req.Header.Set("Content-type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3440.106 Safari/537.36")
	req.Header.Set("sw", "sexir")

	return cb.HttpClient.Do(req)
}

func (pbc *PaidByCoins) makePayResult(response *TransactionAddResponse) interface{} {
	panic("asdf")
}
