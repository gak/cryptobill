package cryptobill

import "github.com/hashicorp/go-multierror"

func (cb *CryptoBill) Quote(info *FiatInfo) ([]QuoteResult, error) {
	var results []QuoteResult
	var errors error
	for _, s := range Services {
		result, err := s.Quote(cb, info)
		if err != nil {
			errors = multierror.Append(errors, err)
			continue
		}

		results = append(results, result...)
	}
	return results, errors
}
