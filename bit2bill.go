package cryptobill

import (
	"encoding/json"
	"github.com/pkg/errors"
	"strings"
)

type Bit2Bill struct{}

func NewBit2Bill() Service {
	return &Bit2Bill{}
}

func (*Bit2Bill) Name() string {
	return "Bit2Bill"
}

func (*Bit2Bill) ShortName() string {
	return "B2B"
}

func (*Bit2Bill) Website() string {
	panic("implement me")
}

func (bb *Bit2Bill) Quote(cb *CryptoBill, from Currency, amount Amount) ([]QuoteResult, error) {
	url := "https://www.bit2bill.com.au/api/rate"
	resp, err := cb.HttpClient.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "request failed")
	}

	rates := map[string]float64{}
	err = json.NewDecoder(resp.Body).Decode(&rates)
	if err != nil {
		return nil, errors.Wrap(err, "can't decode "+url)
	}

	var results []QuoteResult
	for k, v := range rates {
		// We're expecting the keys to look like "BTCRate", etc.
		to, err := NewCurrencyFromString(strings.TrimSuffix(k, "Rate"))
		if err != nil {
			return nil, err
		}

		result := QuoteResult{
			Service:    bb,
			Pair:       Pair{from, to},
			Conversion: Conversion{amount, amount / Amount(v)},
		}
		results = append(results, result)
	}

	return results, nil
}
