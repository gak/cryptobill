package cryptobill

import (
	"encoding/json"
	"fmt"
	"github.com/alecthomas/repr"
	"github.com/pkg/errors"
	"os"
	"text/tabwriter"
)

type Bill struct {
	Name string `arg`
	BPAY BPAY   `cmd`
	EFT  EFT    `cmd`
}

type Bills map[string]*Bill

var billPath = "bills.json"

func (cb *CryptoBill) LoadBills() (Bills, error) {
	if _, err := os.Stat(billPath); os.IsNotExist(err) {
		err = cb.SaveBills(Bills{})
		if err != nil {
			return nil, errors.Wrap(err, "save bills")
		}
	}

	fp, err := os.Open(billPath)
	if err != nil {
		return nil, errors.Wrap(err, "open bills.json")
	}

	entries := Bills{}
	err = json.NewDecoder(fp).Decode(&entries)
	if err != nil {
		return nil, errors.Wrap(err, "decode json from bills.json")
	}

	return entries, nil
}

func (cb *CryptoBill) SaveBills(entries Bills) error {
	fp, err := os.Create(billPath)
	if err != nil {
		return errors.Wrap(err, "can't create bills.json")
	}

	encoder := json.NewEncoder(fp)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(entries)
	if err != nil {
		return errors.Wrap(err, "can't encode bills.json")
	}

	return nil
}

func (cb *CryptoBill) AddBill(entry *Bill) error {
	entries, err := cb.LoadBills()
	if err != nil {
		return errors.Wrap(err, "load bills")
	}

	entries[entry.Name] = entry

	err = cb.SaveBills(entries)
	if err != nil {
		return errors.Wrap(err, "save bills")
	}

	return nil
}

func (cb *CryptoBill) ListBills() error {
	bills, err := cb.LoadBills()
	if err != nil {
		return errors.Wrap(err, "load bill")
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight|tabwriter.Debug)

	for _, bill := range bills {
		_, err = fmt.Fprintf(
			w, "%v\t%v\t%v\t\n",
			bill.Name,
			repr.String(bill.EFT),
			repr.String(bill.BPAY),
		)
		if err != nil {
			return errors.Wrap(err, "fprint")
		}
	}

	err = w.Flush()
	if err != nil {
		return errors.Wrap(err, "flush")
	}

	return nil
}

func (cb *CryptoBill) GetBill(name string) (*Bill, error) {
	bills, err := cb.LoadBills()
	if err != nil {
		return nil, errors.Wrap(err, "load bill")
	}

	bill, ok := bills[name]
	if ok {
		return bill, nil
	} else {
		return nil, errors.New("no such bill")
	}
}
