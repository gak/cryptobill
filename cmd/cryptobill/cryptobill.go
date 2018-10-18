package main

import (
	"fmt"
	"github.com/gak/cryptobill"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/alecthomas/kong"
)

type Quote struct {
	Fiat   cryptobill.Currency `arg`
	Amount cryptobill.Amount   `arg`
	Filter []string            `help:"Filter by cryptocurrency, e.g. BTC,ETH"`
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

	sortByCryptoAndValue(result)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight|tabwriter.Debug)
	for _, quote := range result {

		if !m.showQuote(quote) {
			continue
		}

		fmt.Fprintf(w, "%v\t%v\t%5.5f\t\n", quote.Service.ShortName(), quote.Pair.Crypto, quote.Conversion.Crypto)
	}
	w.Flush()
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
