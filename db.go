package main

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/golang/glog"

	"github.com/scylladb/gocqlx"
	"github.com/scylladb/gocqlx/qb"

	"github.com/etrubenok/make-trades-registry/types"
)

// DBImporter is an interface for importing data into the database
type DBImporter interface {
	SaveSymbolsSnapshots(timestamp int64, snapshot *types.ExchangesSymbols) error
}

// DBImporterImpl is an implementation of DBImporter interface
type DBImporterImpl struct {
	session *gocql.Session
}

// NewDBImporter instantiates object of DBImporter interface (DBImporterImpl class)
func NewDBImporter(session *gocql.Session) DBImporter {
	d := DBImporterImpl{
		session: session}
	return &d
}

// SaveSymbolsSnapshots saves symbols snapshots into the database
func (d *DBImporterImpl) SaveSymbolsSnapshots(timestamp int64, snapshots *types.ExchangesSymbols) error {
	for _, e := range snapshots.Exchanges {
		for _, s := range e.Symbols {
			err := d.SaveSymbol(timestamp, e.Exchange, &s)
			if err != nil {
				glog.Errorf("SaveSymbolsSnapshots: cannot save symbol %v of exchange %s into DB due to error %s", s, e.Exchange, err)
				return err
			}
		}
	}
	return nil
}

// GetYearMonthDay gets year (YYYY), month (M) and day (D) from a given timestamp in UTC
func (d *DBImporterImpl) GetYearMonthDay(timestamp int64) (int, int, int) {
	t := time.Unix(0, timestamp*int64(time.Millisecond)).UTC()
	return t.Year(), int(t.Month()), t.Day()
}

// GetExchangeID returns the id for the given exchange name
func (d *DBImporterImpl) GetExchangeID(exchange string) (int, error) {
	if exchange == "binance" {
		return 1, nil
	}
	return 0, fmt.Errorf("GetExchangeID: exchange: '%s' is not known", exchange)
}

// SaveSymbol saves the symbols for one exchange
func (d *DBImporterImpl) SaveSymbol(timestamp int64, exchange string, symbolInfo *types.SymbolInfo) error {
	year, month, day := d.GetYearMonthDay(timestamp)
	glog.V(1).Infof("SaveSymbol: year: %d, month: %d, day: %d", year, month, day)
	exchangeID, err := d.GetExchangeID(exchange)
	if err != nil {
		glog.Errorf("SaveSymbol: cannot get exchangeId for exchange %s due to error %s", exchange, err)
		return err
	}
	glog.V(1).Infof("SaveSymbol: exchangeID: %d", exchangeID)

	symbolInfo.Year = year
	symbolInfo.Month = month
	symbolInfo.Day = day
	symbolInfo.MarketID = exchangeID
	symbolInfo.SnapshotTime = timestamp

	stmt, names := qb.Insert("maketrades.symbols_snapshots").Columns("year",
		"month",
		"day",
		"market_id",
		"snapshot_time",
		"symbol",
		"status",
		"asset",
		"asset_precision",
		"quote",
		"quote_precision",
		"order_types",
		"iceberg_allowed").ToCql()
	query := gocqlx.Query(d.session.Query(stmt), names).BindStruct(symbolInfo)

	if err := query.ExecRelease(); err != nil {
		glog.Errorf("SaveSymbol: cannot insert symbol info %v into the DB due to error %s", symbolInfo, err)
		return err
	}
	return nil

}
