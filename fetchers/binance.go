package fetchers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/etrubenok/make-trades-registry/types"
	"github.com/etrubenok/make-trades-types/registry"
	"github.com/golang/glog"
)

// BinanceFetcher implements all the fetcher functions for Binance
type BinanceFetcher struct {
}

// NewBinanceFetcher instantiates BinanceFetcher object
func NewBinanceFetcher() Fetcher {
	f := BinanceFetcher{}
	return &f
}

// FetchSymbols fetches symbols from Binance
func (f *BinanceFetcher) FetchSymbols() (*types.ExchangeSymbols, error) {
	raw, err := f.GetBinanceExchangeInfo("https://api.binance.com/api/v1/exchangeInfo")
	if err != nil {
		glog.Errorf("BinanceFetcher.FetchSymbols: cannot fetch symbols due to error %s", err)
		return nil, err
	}
	_, symbols, err := f.GetListOfSymbolsAndTime(raw)
	if err != nil {
		glog.Errorf("BinanceFetcher.FetchSymbols: cannot get symbols due to error %s", err)
		return nil, err
	}
	exchangeID, err := registry.GetExchangeID("binance")
	if err != nil {
		glog.Errorf("BinanceFetcher.FetchSymbols: cannot get exchangeID for 'binance' due to error %s", err)
		return nil, err
	}
	r := types.ExchangeSymbols{
		ExchangeID:   exchangeID,
		SnapshotTime: time.Now().UnixNano() / int64(time.Millisecond),
		Symbols:      symbols,
	}

	year, month, day := GetYearMonthDay(r.SnapshotTime)
	glog.V(1).Infof("BinanceFetcher.FetchSymbols: year: %d, month: %d, day: %d", year, month, day)

	r.Year = year
	r.Month = month
	r.Day = day

	return &r, nil
}

// ConvertToJSON converts string to JSON struct
func ConvertToJSON(message string) (map[string]interface{}, error) {
	m := make(map[string]interface{})

	d := json.NewDecoder(bytes.NewBuffer([]byte(message)))
	d.UseNumber()
	if err := d.Decode(&m); err != nil {
		e := fmt.Errorf("ConvertToJSON: cannot unmarshal message %s to JSON due to error %s", message, err)
		glog.Error(e)
		return m, e
	}

	return m, nil
}

// GetBinanceExchangeInfo requests the exchange information from Binance
func (f *BinanceFetcher) GetBinanceExchangeInfo(url string) (map[string]interface{}, error) {
	resp, err := http.Get(url)
	if err != nil {
		glog.Errorf("GetBinanceExchangeInfo: cannot get the exchange information from Binance due to error %s", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("GetBinanceExchangeInfo: cannot get the exchange information response from Binance due to error %s", err)
		return nil, err
	}
	response, err := ConvertToJSON(string(body))
	if err != nil {
		glog.Errorf("GetBinanceExchangeInfo: cannot get the response %s as JSON due to error  %s", body, err)
		return nil, err
	}
	glog.V(1).Infof("GetBinanceExchangeInfo: response: %v", response)
	return response, nil
}

// GetListOfSymbolsAndTime extracts the symbols information from the response
func (f *BinanceFetcher) GetListOfSymbolsAndTime(response map[string]interface{}) (time.Time, []types.SymbolInfo, error) {
	tNumber, ok := response["serverTime"].(json.Number)
	if !ok {
		return time.Unix(0, 0), nil, fmt.Errorf("GetListOfSymbolsAndTime cannot extract 'serverTime' from the response %v", response)
	}
	serverTime, _ := strconv.ParseInt(string(tNumber), 10, 64)

	rawSymbols, ok := response["symbols"].([]interface{})
	if !ok {
		return time.Unix(0, 0), nil, fmt.Errorf("GetListOfSymbolsAndTime cannot extract 'symbols' from the response %v", response)

	}
	symbols := make([]types.SymbolInfo, 0)
	for _, s := range rawSymbols {
		symbol := types.SymbolInfo{}
		sMapped := s.(map[string]interface{})
		symbol.Symbol, ok = sMapped["symbol"].(string)
		if !ok {
			glog.Errorf("GetListOfSymbolsAndTime cannot extract 'symbol' from the symbol JSON %v", sMapped)
			continue
		}
		symbol.Status, ok = sMapped["status"].(string)
		if !ok {
			glog.Errorf("GetListOfSymbolsAndTime cannot extract 'status' from the symbol JSON %v", sMapped)
		}
		symbol.BaseAsset, ok = sMapped["baseAsset"].(string)
		if !ok {
			glog.Errorf("GetListOfSymbolsAndTime cannot extract 'baseAsset' from the symbol JSON %v", sMapped)
		}
		baseAssetPrecision, ok := sMapped["baseAssetPrecision"].(json.Number)
		if !ok {
			glog.Errorf("GetListOfSymbolsAndTime cannot extract 'baseAssetPrecision' from the symbol JSON %v", sMapped)
		}
		symbol.BaseAssetPrecision, _ = baseAssetPrecision.Int64()

		symbol.QuoteAsset, ok = sMapped["quoteAsset"].(string)
		if !ok {
			glog.Errorf("GetListOfSymbolsAndTime cannot extract 'quoteAsset' from the symbol JSON %v", sMapped)
		}

		quotePrecision, ok := sMapped["quotePrecision"].(json.Number)
		if !ok {
			glog.Errorf("GetListOfSymbolsAndTime cannot extract 'quotePrecision' from the symbol JSON %v", sMapped)
		}
		symbol.QuotePrecision, _ = quotePrecision.Int64()

		orderTypes, ok := sMapped["orderTypes"].([]interface{})
		if !ok {
			glog.Errorf("GetListOfSymbolsAndTime cannot extract 'orderTypes' from the symbol JSON %v", sMapped)
		}
		for _, ot := range orderTypes {
			symbol.OrderTypes = append(symbol.OrderTypes, ot.(string))
		}

		symbol.IcebergAllowed, ok = sMapped["icebergAllowed"].(bool)
		if !ok {
			glog.Errorf("GetListOfSymbolsAndTime cannot extract 'icebergAllowed' from the symbol JSON %v", sMapped)
		}
		symbols = append(symbols, symbol)
	}
	return time.Unix(0, serverTime*int64(time.Millisecond)), symbols, nil
}
