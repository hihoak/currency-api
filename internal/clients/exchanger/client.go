package exchanger

import (
	"context"
	"github.com/hihoak/currency-api/internal/pkg/logger"
	"github.com/hihoak/currency-api/internal/pkg/models"
	"sync"
	"time"
)

type Storager interface {
	SaveCourses(ctx context.Context, timeNow time.Time, fromCurrency, toCurrency models.Currencies, course float64) error
}

type Quoter interface {
	GetQuote(from string, to string) (float64, error)
}

type Exchage struct {
	mu *sync.Mutex
	currentCourses map[models.Currencies]*CurrenciesQuotes

	ticker *time.Ticker

	logg *logger.Logger
	quoter Quoter

	storage Storager

	doneChan <-chan struct{}
}

func New(ctx context.Context, logg *logger.Logger, quoter Quoter, storage Storager) *Exchage {
	currentCourses := map[models.Currencies]*CurrenciesQuotes{
		models.RUB: NewCurrenciesQuotes(models.RUB),
		models.EUR: NewCurrenciesQuotes(models.EUR),
		models.USD: NewCurrenciesQuotes(models.USD),
	}

	return &Exchage{
		logg: logg,
		quoter: quoter,
		currentCourses: currentCourses,
		storage: storage,

		ticker: time.NewTicker(time.Second * 10),

		doneChan: ctx.Done(),
	}
}

func (e *Exchage) Start() {
	go func() {
		for {
			select {
			case <-e.ticker.C:
				timeNow := time.Now()
				wg := sync.WaitGroup{}
				for currency, currentCourse := range e.currentCourses {
					for toCurrency := range currentCourse.Data {
						wg.Add(1)
						go func(from, to models.Currencies) {
							defer wg.Done()
							newQuote, err := e.quoter.GetQuote(string(from), string(to))
							if err != nil {
								e.logg.Error().Err(err).Msgf("failed to get quote")
								return
							}
							e.currentCourses[from].Update(to, newQuote)
							if err := e.storage.SaveCourses(context.Background(), timeNow, from, to, newQuote); err != nil {
								e.logg.Error().Err(err).Msgf("failed to save courses to DB")
							}
						}(currency, toCurrency)
					}
				}
				e.logg.Debug().Msgf("Exchage: successfully update courses: %v", e.currentCourses)
				wg.Wait()
			case <-e.doneChan:
				e.logg.Info().Msg("stop consuming quotes...")
				return
			}
		}
	}()
}

func (e *Exchage) GetCourse(from, to models.Currencies) CourseInfo {
	return e.currentCourses[from].Get(to)
}
