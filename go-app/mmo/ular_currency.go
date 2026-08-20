package mmo

// CurrencyService spends coins for cosmetics only in later phases.
// Coins never grant answers, dice, extra steps, snake immunity, or wins.
type CurrencyService struct {
	Store *ProgressStore
}

func NewCurrencyService(store *ProgressStore) *CurrencyService {
	return &CurrencyService{Store: store}
}

func (c *CurrencyService) Balance(userID string) int {
	if c == nil || c.Store == nil {
		return 0
	}
	p, ok := c.Store.Get(userID)
	if !ok {
		return 0
	}
	return p.Coins
}

func (c *CurrencyService) Spend(userID string, amount int, ref string) (CoinTransaction, string) {
	if c == nil || c.Store == nil {
		return CoinTransaction{}, "currency tidak siap"
	}
	return c.Store.SpendCoins(userID, amount, ref)
}
