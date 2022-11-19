package exchanger

import (
	"github.com/hihoak/currency-api/internal/pkg/models"
	"sync"
)

var allSupportedCurrencies = []models.Currencies{models.RUB, models.EUR, models.USD}

type CurrenciesQuotes struct {
	Data map[models.Currencies]float64
	mu *sync.RWMutex
}

func NewCurrenciesQuotes(currency models.Currencies) *CurrenciesQuotes {
	data := make(map[models.Currencies]float64)
	for _, c := range allSupportedCurrencies {
		if c != currency {
			data[c] = 0.0
		}
	}

	return &CurrenciesQuotes{
		Data: data,
		mu: &sync.RWMutex{},
	}
}

func (c *CurrenciesQuotes) Update(to models.Currencies, quote float64) {
	c.mu.Lock()
	c.Data[to] = quote
	c.mu.Unlock()
}

func (c *CurrenciesQuotes) Get(to models.Currencies) float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Data[to]
}
