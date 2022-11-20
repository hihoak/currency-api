package exchanger

import (
	"github.com/hihoak/currency-api/internal/pkg/models"
	"sync"
)

var allSupportedCurrencies = []models.Currencies{models.RUB, models.EUR, models.USD}

type CourseInfo struct {
	Value float64 `json:"value"`
	IsIncreasing bool `json:"is_increasing"`
}

type CurrenciesQuotes struct {
	Data map[models.Currencies]CourseInfo
	mu *sync.RWMutex
}

func NewCurrenciesQuotes(currency models.Currencies) *CurrenciesQuotes {
	data := make(map[models.Currencies]CourseInfo)
	for _, c := range allSupportedCurrencies {
		if c != currency {
			data[c] = CourseInfo{
				Value: 0.0,
				IsIncreasing: false,
			}
		}
	}

	return &CurrenciesQuotes{
		Data: data,
		mu: &sync.RWMutex{},
	}
}

func (c *CurrenciesQuotes) Update(to models.Currencies, quote float64) {
	c.mu.Lock()
	oldValue := c.Data[to].Value
	c.Data[to] = CourseInfo{
		Value: quote,
		IsIncreasing: quote > oldValue,
	}
	c.mu.Unlock()
}

func (c *CurrenciesQuotes) Get(to models.Currencies) CourseInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Data[to]
}
