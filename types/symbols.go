package types

// SymbolInfo type contains information about one symbol
type SymbolInfo struct {
	Symbol             string   `json:"symbol" cql:"symbol"`
	Status             string   `json:"status" cql:"status"`
	BaseAsset          string   `json:"baseAsset" cql:"asset"`
	BaseAssetPrecision int64    `json:"baseAssetPrecision" cql:"asset_precision"`
	QuoteAsset         string   `json:"quoteAsset" cql:"quote"`
	QuotePrecision     int64    `json:"quotePrecision" cql:"quote_precision"`
	OrderTypes         []string `json:"quoteTypes" cql:"order_types"`
	IcebergAllowed     bool     `json:"icebergAllowed" cql:"iceberg_allowed"`
}

// ExchangeSymbols type contains information about symbols of an exchange
type ExchangeSymbols struct {
	Year         int          `cql:"year"`
	Month        int          `cql:"month"`
	Day          int          `cql:"day"`
	ExchangeID   int          `cql:"exchange_id"`
	SnapshotTime int64        `cql:"snapshot_time"`
	Symbols      []SymbolInfo `json:"symbols" cql:"symbols"`
}

// ExchangesSymbols type contains information about symbols of several exchanges
type ExchangesSymbols struct {
	Exchanges []ExchangeSymbols `json:"exchanges"`
}
