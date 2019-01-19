package main

import (
	"github.com/gocql/gocql"
	"github.com/golang/glog"

	"github.com/scylladb/gocqlx"
	"github.com/scylladb/gocqlx/qb"

	"github.com/etrubenok/make-trades-registry/types"
)

// DBImporter is an interface for importing data into the database
type DBImporter interface {
	SaveSymbolsSnapshots(snapshot *types.ExchangesSymbols) error
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
func (d *DBImporterImpl) SaveSymbolsSnapshots(snapshots *types.ExchangesSymbols) error {
	for _, e := range snapshots.Exchanges {
		err := d.SaveSymbols(&e)
		if err != nil {
			glog.Errorf("SaveSymbolsSnapshots: cannot save symbols of exchange id '%d' into DB due to error %s", e.ExchangeID, err)
			return err
		}
	}
	return nil
}

// SaveSymbols saves the symbols for one exchange
func (d *DBImporterImpl) SaveSymbols(exchangeSymbols *types.ExchangeSymbols) error {
	stmt, names := qb.Insert("maketrades.symbols_snapshots").Columns("year",
		"month",
		"day",
		"exchange_id",
		"snapshot_time",
		"symbols").ToCql()
	query := gocqlx.Query(d.session.Query(stmt), names).BindStruct(exchangeSymbols)

	if err := query.ExecRelease(); err != nil {
		glog.Errorf("SaveSymbols: cannot insert symbols for exchange id '%d' into the DB due to error %s", exchangeSymbols.ExchangeID, err)
		return err
	}
	return nil

}
