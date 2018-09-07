package cryptobill

type Service interface {
	Name() string
	Website() string
	Quote(cb *CryptoBill, from Currency, amount Amount) ([]QuoteResult, error)
}

var Services = []Service{
	NewLivingRoom(),
	NewPaidByCoins(),
	NewBit2Bill(),
}
