package mock_quoter

import (
	"github.com/hihoak/currency-api/internal/pkg/logger"
	"github.com/hihoak/currency-api/internal/pkg/models"
	"math/rand"
	"sync"
	"time"
)

type Quote struct {
	logg *logger.Logger
	mu *sync.RWMutex
	quotes map[models.Currencies]map[models.Currencies]float64
}

func New(logg *logger.Logger) *Quote {
	return &Quote{
		quotes: map[models.Currencies]map[models.Currencies]float64{
			models.RUB: {
				models.USD: 0.016,
				models.EUR: 0.016,
				models.JPY: 2.31,
				models.CHF: 0.016,
				models.GBP: 0.014,
				models.CNY: 0.12,
			},
			models.USD: {
				models.RUB: 60.85,
				models.EUR: 0.97,
			},
			models.EUR: {
				models.USD: 1.03,
				models.RUB: 62.95,
			},
			models.CHF: {
				models.RUB: 63.73,
			},
			models.CNY: {
				models.RUB: 8.55,
			},
			models.JPY: {
				models.RUB: 0.43,
			},
			models.GBP: {
				models.RUB: 72.34,
			},
		},
		mu: &sync.RWMutex{},
		logg: logg,
	}
}

func (q *Quote) Start() {
	go func() {
		for {
			q.mu.Lock()
			for from, currencies := range q.quotes {
				for to := range currencies {
					q.quotes[from][to] += q.quotes[from][to] / 500 * (rand.Float64() - 0.5)
				}
			}
			q.mu.Unlock()
			time.Sleep(time.Second * 5)
			q.logg.Debug().Msgf("new courses: %v", q.quotes)
		}
	}()
}

func (q *Quote) GetQuote(from string, to string) (float64, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.quotes[models.Currencies(from)][models.Currencies(to)], nil
}
