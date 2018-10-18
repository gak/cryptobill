package main

import (
	"fmt"
	"github.com/gak/cryptobill"
	"os"
	"text/tabwriter"

	"github.com/alecthomas/kong"
)

type Quote struct {
	From   cryptobill.Currency `arg`
	Amount cryptobill.Amount   `arg`
	To     []string            `help:"Filter by coin, e.g. BTC,ETH"`
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
	case "quote <from> <amount>":
		m.quote(&m.cli.Quote)
	default:
		panic(ctx.Command())
	}
}

func (m *Main) quote(q *Quote) {
	result, err := m.cb.Quote(q.From, q.Amount)
	if err != nil {
		panic(err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight|tabwriter.Debug)
	for _, quote := range result {

		if !m.showQuote(quote) {
			continue
		}

		fmt.Fprintf(w, "%v\t%v\t%5.5f\t\n", quote.Service.ShortName(), quote.Pair.To, quote.Conversion.To)
	}
	w.Flush()
}

func (m *Main) showQuote(quote cryptobill.QuoteResult) bool {
	showQuote := false
	for _, to := range m.cli.Quote.To {
		toCur, err := cryptobill.NewCurrencyFromString(to)
		if err != nil {
			panic(err)
		}
		if toCur == quote.Pair.To {
			showQuote = true
		}
	}
	return showQuote
}
