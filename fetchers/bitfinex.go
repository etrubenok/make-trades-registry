package fetchers

import (
	"fmt"
	"strings"
	"time"

	bitfinex "github.com/bitfinexcom/bitfinex-api-go/v1"
	"github.com/etrubenok/make-trades-registry/types"
	"github.com/etrubenok/make-trades-types/registry"
	"github.com/golang/glog"
)

// BitfinexFetcher implements all the fetcher functions for Bitfinex
type BitfinexFetcher struct {
}

// NewBitfinexFetcher instantiates BitfinexFetcher object
func NewBitfinexFetcher() Fetcher {
	f := BitfinexFetcher{}
	return &f
}

// FetchSymbols fetches symbols from Bitfinex
func (f *BitfinexFetcher) FetchSymbols() (*types.ExchangeSymbols, error) {
	pairs, err := bitfinex.NewClient().Pairs.AllDetailed()

	symbols, err := f.ConvertSymbols(pairs)
	if err != nil {
		glog.Errorf("BitfinexFetcher.FetchSymbols: cannot convert symbols due to error %s", err)
		return nil, err
	}
	exchangeID, err := registry.GetExchangeID("bitfinex")
	if err != nil {
		glog.Errorf("BitfinexFetcher.FetchSymbols: cannot get exchangeID for 'bitfinex' due to error %s", err)
		return nil, err
	}
	r := types.ExchangeSymbols{
		ExchangeID:   exchangeID,
		SnapshotTime: time.Now().UnixNano() / int64(time.Millisecond),
		Symbols:      symbols,
	}

	year, month, day := GetYearMonthDay(r.SnapshotTime)
	glog.V(1).Infof("BitfinexFetcher.FetchSymbols: year: %d, month: %d, day: %d", year, month, day)

	r.Year = year
	r.Month = month
	r.Day = day

	return &r, nil
}

// ConvertSymbols converts Bitfinex pairs into the make trades symbols
func (f *BitfinexFetcher) ConvertSymbols(pairs []bitfinex.Pair) ([]types.SymbolInfo, error) {
	symbols := make([]types.SymbolInfo, 0)
	fPairs := make(map[string]bool)
	for _, p := range pairs {
		s := types.SymbolInfo{
			Symbol:             fmt.Sprintf("t%s", strings.ToUpper(p.Pair)),
			BaseAssetPrecision: int64(p.PricePrecision),
			QuotePrecision:     int64(p.PricePrecision)}
		symbols = append(symbols, s)

		if p.Margin {
			// ASSUMPTION: pair name len is 6 and format is [one][two]
			pair1 := fmt.Sprintf("f%s", strings.ToUpper(string([]byte(p.Pair)[3:])))
			fPairs[pair1] = true

			pair2 := fmt.Sprintf("f%s", strings.ToUpper(string([]byte(p.Pair)[:3])))
			fPairs[pair2] = true
		}
	}
	for k := range fPairs {
		s := types.SymbolInfo{
			Symbol:             k,
			BaseAssetPrecision: int64(8),
			QuotePrecision:     int64(8)}
		symbols = append(symbols, s)
	}
	return symbols, nil
}
