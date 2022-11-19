package mock_quoter

import (
	"github.com/hihoak/currency-api/internal/pkg/models"
	"math/rand"
	"sync"
	"time"
)

type Quote struct {
	mu *sync.RWMutex
	quotes map[models.Currencies]map[models.Currencies]float64
}

func New() *Quote {
	return &Quote{
		quotes: map[models.Currencies]map[models.Currencies]float64{
			models.RUB: {
				models.USD: 0.016,
				models.EUR: 0.016,
			},
			models.USD: {
				models.RUB: 60.85,
				models.EUR: 0.97,
			},
			models.EUR: {
				models.USD: 1.03,
				models.RUB: 62.95,
			},
		},
		mu: &sync.RWMutex{},
	}
}

func (q *Quote) Start() {
	go func() {
		for {
			q.mu.Lock()
			for from, currencies := range q.quotes {
				for to := range currencies {
					q.quotes[from][to] += (rand.Float64() - 0.35) / 5
				}
			}
			q.mu.Unlock()
			time.Sleep(time.Second)
		}
	}()
}

func (q *Quote) GetQuote(from string, to string) (float64, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.quotes[models.Currencies(from)][models.Currencies(to)], nil
}
