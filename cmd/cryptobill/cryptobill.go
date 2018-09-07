package main

import (
	"fmt"
	"github.com/gak/cryptobill"

	"github.com/alecthomas/kong"
)

type Quote struct {
	From   cryptobill.Currency `arg`
	Amount cryptobill.Amount   `arg`
}

var CLI struct {
	Quote Quote `cmd`
}

type Main struct {
	cb *cryptobill.CryptoBill
}

func main() {
	m := Main{
		cb: cryptobill.NewCryptoBill(),
	}

	ctx := kong.Parse(&CLI)
	switch ctx.Command() {
	case "quote <from> <amount>":
		m.quote(&CLI.Quote)
	default:
		panic(ctx.Command())
	}
}

func (m *Main) quote(q *Quote) {
	result, err := m.cb.Quote(q.From, q.Amount)
	if err != nil {
		panic(err)
	}

	for _, quote := range result {
		fmt.Println(quote.Service.Name(), quote.Pair.To, quote.Conversion.To)
	}
}
