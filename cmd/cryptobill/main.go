package main

import (
	"encoding/json"
	"fmt"
	"github.com/alecthomas/repr"
	"github.com/gak/cryptobill"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/alecthomas/kong"
	"strings"
)

type Quote struct {
	cryptobill.FiatInfo `cmd`
	Filter              []string `help:"Filter by cryptocurrency, e.g. BTC,ETH"`
	Services            []string `help:"Filter by service, e.g. BPC,LROS"`
	NoConvertBack       bool
}

type Add struct {
	BPAY struct {
		Name string `arg`
		cryptobill.BPAY
	} `cmd help:"Add BPAY a bill to be used with \"pay\". e.g. \"add bpay mybill 12345 998877\""`

	EFT struct {
		Name string `arg`
		cryptobill.EFT
	} `cmd`
}

type List struct{}

type Pay struct {
	Name string `arg`
	cryptobill.PayInfoService
}

type CLI struct {
	Quote Quote `cmd`
	Add   Add   `cmd`
	List  List  `cmd help:"List your different bills."`
	Pay   Pay   `cmd help:"Prepare a payment and retrieve an address to send crypto to."`
}

type Main struct {
	cb  *cryptobill.CryptoBill
	cli CLI
}

func main() {
	var err error

	m := Main{
		cb: cryptobill.NewCryptoBill(),
	}

	ctx := kong.Parse(&m.cli)
	switch ctx.Command() {
	case "quote <amount> <fiat>":
		err = m.quote(&m.cli.Quote)
	case "list":
		err = m.cb.ListBills()
	case "add bpay <name> <code> <account>":
		err = m.cb.AddBill(entry(&m.cli.Add))
	case "add eft <name> <bsb> <account-number> <account-name>":
		err = m.cb.AddBill(entry(&m.cli.Add))
	case "pay <name> <amount> <fiat> <crypto> <service>":
		err = m.pay(&m.cli.Pay)
	default:
		panic("unknown command: " + ctx.Command())
	}

	if err != nil {
		panic(err)
	}
}

func entry(add *Add) *cryptobill.Bill {
	if add.BPAY.Name != "" {
		return &cryptobill.Bill{
			Name: add.BPAY.Name,
			BPAY: add.BPAY.BPAY,
		}
	} else if add.EFT.Name != "" {
		return &cryptobill.Bill{
			Name: add.EFT.Name,
			EFT:  add.EFT.EFT,
		}
	} else {
		panic("unknown bill type")
	}
}

func (m *Main) pay(pay *Pay) error {
	bill, err := m.cb.GetBill(pay.Name)
	if err != nil {
		return errors.Wrap(err, "get bill")
	}

	if bill.BPAY != (cryptobill.BPAY{}) {
		payBPAY := cryptobill.PayBPAY{
			PayInfoService: pay.PayInfoService,
			BPAY:           bill.BPAY,
		}
		result, err := m.cb.PayBPAY(&payBPAY)
		if err != nil {
			return errors.Wrap(err, "pay bpay")
		}

		repr.Println(result)
	} else if bill.EFT != (cryptobill.EFT{}) {
		payEFT := cryptobill.PayEFT{
			PayInfoService: pay.PayInfoService,
			EFT:           bill.EFT,
		}
		result, err := m.cb.PayEFT(&payEFT)
		if err != nil {
			return errors.Wrap(err, "pay bpay")
		}

		repr.Println(result)
	} else {
		return errors.New("bill error, could not find data")
	}

	return nil
}

func (m *Main) quote(q *Quote) error {
	result, err := m.cb.Quote(&q.FiatInfo)
	if err != nil {
		return errors.Wrap(err, "quote")
	}

	lookup := map[cryptobill.Currency]cryptobill.Amount{}

	if q.NoConvertBack {
		sortByCryptoAndValue(result)

	} else {
		lookup, err = m.fetchExchange(result)
		if err != nil {
			return errors.Wrap(err, "quote")
		}

		sortByFiatValue(result, lookup)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight|tabwriter.Debug)
	for _, quote := range result {

		if !m.showQuote(quote) {
			continue
		}

		_, err := fmt.Fprintf(
			w, "%v\t%v\t%5.5f\t",
			quote.Service.ShortName(),
			quote.Pair.Crypto,
			quote.Conversion.Crypto,
		)
		if err != nil {
			return errors.Wrap(err, "fprintf")
		}

		if !q.NoConvertBack {
			_, err = fmt.Fprintf(
				w, "%5.5f\t%2.3f%%\t",
				lookup[quote.Pair.Crypto]*quote.Conversion.Crypto,
				lookup[quote.Pair.Crypto]*quote.Conversion.Crypto/quote.Conversion.Fiat*100-100,
			)
			if err != nil {
				return errors.Wrap(err, "fprintf")
			}
		}

		_, err = fmt.Fprintf(w, "\n")
		if err != nil {
			return errors.Wrap(err, "fprintf")
		}
	}

	err = w.Flush()
	if err != nil {
		return errors.Wrap(err, "flush")
	}

	return nil
}

func sortByFiatValue(result []cryptobill.QuoteResult, lookup map[cryptobill.Currency]cryptobill.Amount) {
	sort.Slice(result, func(i, j int) bool {
		vi := result[i].Conversion.Crypto * lookup[result[i].Pair.Crypto]
		vj := result[j].Conversion.Crypto * lookup[result[j].Pair.Crypto]
		return vi < vj
	})
}

func sortByCryptoAndValue(result []cryptobill.QuoteResult) {
	sort.Slice(result, func(i, j int) bool {
		ci := result[i].Pair.Crypto
		cj := result[j].Pair.Crypto

		ai := result[i].Conversion.Crypto
		aj := result[j].Conversion.Crypto

		if ci == cj {
			return ai < aj
		} else {
			return ci < cj
		}
	})
}

func (m *Main) showQuote(quote cryptobill.QuoteResult) bool {
	for _, nope := range notSupported {
		if strings.EqualFold(nope, string(quote.Pair.Crypto)) {
			return false
		}
	}

	if len(m.cli.Quote.Filter) == 0 {
		return true
	}

	showQuote := false
	for _, filterStr := range m.cli.Quote.Filter {
		filter, err := cryptobill.NewCurrencyFromString(filterStr)
		if err != nil {
			panic(err)
		}
		if filter == quote.Pair.Crypto {
			showQuote = true
		}
	}

	return showQuote
}

var notSupported = []string{
	"SDB",
	"BTX",
	"XEM",
	"DCR",
	"STEEM",
	"DCR",
	"SBD",
	"DOGE",
	"ETC",
	"OMG",
	"DASH",
	"LIGHTNING",
	"PIVX",
}

func (m *Main) fetchExchange(result []cryptobill.QuoteResult) (map[cryptobill.Currency]cryptobill.Amount, error) {
	lookup := map[cryptobill.Currency]cryptobill.Amount{}

	for _, quote := range result {
		if !m.showQuote(quote) {
			continue
		}

		if _, ok := lookup[quote.Pair.Crypto]; ok {
			continue
		}

		symbol := quote.Pair.Crypto + quote.Pair.Fiat
		last, err := bitcoinAverage(m.cb, string(symbol))
		if err != nil {
			return nil, err
		}

		lookup[quote.Pair.Crypto] = cryptobill.Amount(last)
	}

	return lookup, nil
}

type BitcoinAverageResponse struct {
	Last float64
}

func bitcoinAverage(cb *cryptobill.CryptoBill, symbol string) (float64, error) {
	req, err := http.NewRequest("GET", "https://apiv2.bitcoinaverage.com/indices/global/ticker/"+symbol, nil)

	if err != nil {
		return 0, errors.Wrap(err, "request builder")
	}

	resp, err := cb.HttpClient.Do(req)
	if err != nil {
		return 0, errors.Wrap(err, "server request")
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, errors.Wrap(err, "reading body")
	}

	decoded := BitcoinAverageResponse{}
	err = json.Unmarshal(body, &decoded)
	if err != nil {
		return 0, errors.Wrap(err, "decoding body to json: "+string(body))
	}

	return decoded.Last, nil
}
