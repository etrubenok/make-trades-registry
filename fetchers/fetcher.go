package fetchers

import (
	"fmt"
	"time"

	"github.com/golang/glog"

	"github.com/etrubenok/make-trades-registry/types"
)

// GetYearMonthDay gets year (YYYY), month (M) and day (D) from a given timestamp in UTC
func GetYearMonthDay(timestamp int64) (int, int, int) {
	t := time.Unix(0, timestamp*int64(time.Millisecond)).UTC()
	return t.Year(), int(t.Month()), t.Day()
}

// Fetcher is the interface for all the fetchers
type Fetcher interface {
	FetchSymbols() (*types.ExchangeSymbols, error)
}

// FetcherFactory creates a required fetcher based on the exchange name
func FetcherFactory(exchange string) (Fetcher, error) {
	switch exchange {
	case "binance":
		return NewBinanceFetcher(), nil
	case "bitfinex":
		return NewBitfinexFetcher(), nil
	default:
		return nil, fmt.Errorf("FetcherFactory: exchange '%s' is not supported", exchange)
	}
}

// FetchJob is the interface for fetch job
type FetchJob interface {
	Init(exchanges []string, results chan<- types.ExchangesSymbols)
}

// FetchJobImpl is an implementation of FetchJob
type FetchJobImpl struct {
	exchanges []string
	ticker    *time.Ticker
}

// NewFetchJob instantiates a fetch job
func NewFetchJob() FetchJob {
	f := FetchJobImpl{}
	return &f
}

// Init initialises the fetch job with the list of exchanges
func (j *FetchJobImpl) Init(exchanges []string, results chan<- types.ExchangesSymbols) {
	j.exchanges = exchanges
	j.ticker = time.NewTicker(30 * time.Minute)
	go j.FetchExchangesSymbols(results)
}

// FetchExchangesSymbols executes fetching across all the exchanges
func (j *FetchJobImpl) FetchExchangesSymbols(results chan<- types.ExchangesSymbols) {
	for {
		select {
		case <-j.ticker.C:
			exchangeChan := make(chan types.ExchangeSymbols)
			errorChan := make(chan error)
			for _, e := range j.exchanges {
				go j.FetchExchange(e, exchangeChan, errorChan)
			}
			var errors []error
			var exchangesSymbols = types.ExchangesSymbols{
				Exchanges: make([]types.ExchangeSymbols, 0),
			}
			i := 0
			for i < len(j.exchanges) {
				select {
				case msg := <-exchangeChan:
					exchangesSymbols.Exchanges = append(exchangesSymbols.Exchanges, msg)
					i++
				case err := <-errorChan:
					glog.Errorf("FetchExchangesSymbols: got error %s", err)
					errors = append(errors, err)
					i++
				}
			}
			results <- exchangesSymbols
		}
	}
}

// FetchExchange executes fetching from the specified exchange
func (j *FetchJobImpl) FetchExchange(exchange string, exchangeChan chan<- types.ExchangeSymbols, errorChan chan<- error) {
	fetcher, err := FetcherFactory(exchange)
	if err != nil {
		glog.Errorf("FetchExchange: cannot instantiate a fetcher for exchange '%s' due to error '%s'", exchange, err)
		errorChan <- err
		return
	}
	exSymbols, err := fetcher.FetchSymbols()
	if err != nil {
		glog.Errorf("FetchExchange: cannot fetch from exchange '%s' due to error '%s'", exchange, err)
		errorChan <- err
		return
	}
	exchangeChan <- *exSymbols
}
