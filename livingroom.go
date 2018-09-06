package cryptobill

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"strings"
)

type LivingRoom struct{}

type QuoteResponse map[string]float64

func NewLivingRoom() Service {
	return &LivingRoom{}
}

func (lros *LivingRoom) Name() string {
	return "Living Room of Satoshi"
}

func (lros *LivingRoom) Website() string {
	return ""
}

func (lros *LivingRoom) Quote(cb *CryptoBill, from Currency, amount Amount) ([]QuoteResult, error) {
	req, err := http.NewRequest("GET", "https://www.livingroomofsatoshi.com/api/v1/current_rates", nil)
	if err != nil {
		return nil, errors.Wrap(err, "request builder")
	}

	resp, err := cb.HttpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "server request")
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "reading body")
	}

	decoded := QuoteResponse{}
	err = json.Unmarshal(body, &decoded)
	if err != nil {
		return nil, errors.Wrap(err, "decoding body to json")
	}

	var results []QuoteResult
	for pair, quoted := range decoded {
		bits := strings.Split(pair, "_")
		from, err := NewCurrencyFromString(bits[0])
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		to, err := NewCurrencyFromString(bits[1])
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		qr := QuoteResult{
			lros,
			Pair{from, to},
			Conversion{amount, amount / Amount(quoted)},
		}
		results = append(results, qr)
	}

	return results, nil
}
