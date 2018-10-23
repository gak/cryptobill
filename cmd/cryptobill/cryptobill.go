package main

import (
	"encoding/json"
	"fmt"
	"github.com/gak/cryptobill"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/alecthomas/kong"
)

type Quote struct {
	Fiat          cryptobill.Currency `arg`
	Amount        cryptobill.Amount   `arg`
	Filter        []string            `help:"Filter by cryptocurrency, e.g. BTC,ETH"`
	Services      []string            `help:"Filter by service, e.g. BPC,LROS"`
	NoConvertBack bool
}

type CLI struct {
	Quote Quote `cmd`
}

type Main struct {
	cb  *cryptobill.CryptoBill
	cli CLI
}

func main() {
	m := Main{
		cb: cryptobill.NewCryptoBill(),
	}

	ctx := kong.Parse(&m.cli)
	switch ctx.Command() {
	case "quote <fiat> <amount>":
		m.quote(&m.cli.Quote)
	default:
		panic(ctx.Command())
	}
}

func (m *Main) quote(q *Quote) {
	result, err := m.cb.Quote(q.Fiat, q.Amount)
	if err != nil {
		panic(err)
	}

	lookup := map[cryptobill.Currency]cryptobill.Amount{}

	if q.NoConvertBack {
		sortByCryptoAndValue(result)

	} else {
		lookup, err = m.fetchExchange(result)
		if err != nil {
			panic(err)
		}

		sortByFiatValue(result, lookup)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight|tabwriter.Debug)
	for _, quote := range result {

		if !m.showQuote(quote) {
			continue
		}

		fmt.Fprintf(
			w, "%v\t%v\t%5.5f\t",
			quote.Service.ShortName(),
			quote.Pair.Crypto,
			quote.Conversion.Crypto,
		)

		if !q.NoConvertBack {
			fmt.Fprintf(
				w, "%5.5f\t%2.3f%%\t",
				lookup[quote.Pair.Crypto]*quote.Conversion.Crypto,
				lookup[quote.Pair.Crypto]*quote.Conversion.Crypto/quote.Conversion.Fiat*100-100,
			)
		}

		fmt.Fprintf(w, "\n")
	}
	w.Flush()
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

		fmt.Println(last)

		lookup[quote.Pair.Crypto] = cryptobill.Amount(last)
	}

	fmt.Println(lookup)

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
