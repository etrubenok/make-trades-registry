package types

import (
	"github.com/etrubenok/make-trades-types/registry"
	"github.com/golang/glog"
)

// APISymbolInfo type contains information about one symbol
type APISymbolInfo struct {
	Symbol string `json:"symbol"`
	Status string `json:"status"`
	Asset  string `json:"asset"`
	Quote  string `json:"quote"`
}

// APIExchangeSymbols type contains information about symbols of an exchange
type APIExchangeSymbols struct {
	Exchange     string          `json:"exchange"`
	SnapshotTime int64           `json:"snapshot_time"`
	Symbols      []APISymbolInfo `json:"symbols"`
}

// APIExchangesSymbols type contains information about symbols of several exchanges
type APIExchangesSymbols struct {
	Exchanges []APIExchangeSymbols `json:"exchanges"`
}

// ConvertExchangeSymbolsToAPIResponse converts the given exchangesSymbols in DB format into API responce format
func ConvertExchangeSymbolsToAPIResponse(exchangesSymbols *ExchangesSymbols) (*APIExchangesSymbols, error) {
	r := APIExchangesSymbols{
		Exchanges: make([]APIExchangeSymbols, len(exchangesSymbols.Exchanges))}
	for i, e := range exchangesSymbols.Exchanges {
		apiExchangeSymbols, err := convertExchangeSymbols(&e)
		if err != nil {
			glog.Errorf("ConvertExchangeSymbolsToAPIResponse: cannot convert exchange symbols for exchange id '%d' due to error %s",
				e.ExchangeID, err)
			return nil, err
		}
		r.Exchanges[i] = *apiExchangeSymbols
	}
	return &r, nil
}

func convertExchangeSymbols(exchangeSymbols *ExchangeSymbols) (*APIExchangeSymbols, error) {
	exchange, err := registry.GetExchangeNameByID(exchangeSymbols.ExchangeID)
	if err != nil {
		glog.Errorf("convertExchangeSymbols: cannot get exchange name by id '%d' due to error %s",
			exchangeSymbols.ExchangeID, err)
		return nil, err
	}
	e := APIExchangeSymbols{
		Exchange:     exchange,
		SnapshotTime: exchangeSymbols.SnapshotTime,
		Symbols:      make([]APISymbolInfo, len(exchangeSymbols.Symbols))}

	for i, s := range exchangeSymbols.Symbols {
		apiSymbolInfo, err := convertSymbolInfo(exchange, &s)
		if err != nil {
			glog.Errorf("convertExchangeSymbols: cannot convert symbol '%v' into API format due to error %s",
				s, err)
			return nil, err
		}
		e.Symbols[i] = *apiSymbolInfo
	}
	return &e, nil
}

func convertSymbolInfo(exchange string, symbolInfo *SymbolInfo) (*APISymbolInfo, error) {
	s := APISymbolInfo{
		Symbol: exchange + "-" + symbolInfo.Symbol,
		Status: symbolInfo.Status,
		Asset:  symbolInfo.BaseAsset,
		Quote:  symbolInfo.QuoteAsset}
	return &s, nil
}
