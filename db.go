package main

import (
	"time"

	"github.com/etrubenok/make-trades-registry/fetchers"
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

// DBLoader is an interface for the DB reading operations
type DBLoader interface {
	LoadSymbolsSnapshots(exchangeIDs []int) (*types.ExchangesSymbols, error)
}

// DBLoaderImpl is an implementation of DBLoader interface
type DBLoaderImpl struct {
	session *gocql.Session
}

// NewDBLoader instantiates object of DBLoader interface (DBLoaderImpl class)
func NewDBLoader(session *gocql.Session) DBLoader {
	l := DBLoaderImpl{
		session: session}
	return &l
}

// LoadSymbolsSnapshots loads the symbols for the exchangeIDs
func (l *DBLoaderImpl) LoadSymbolsSnapshots(exchangeIDs []int) (*types.ExchangesSymbols, error) {
	r := types.ExchangesSymbols{
		Exchanges: make([]types.ExchangeSymbols, 0),
	}
	t := time.Now().UnixNano() / int64(time.Millisecond)
	year, month, day := fetchers.GetYearMonthDay(t)
	// TODO: if a new day the snapshot may be missing - need to be taken from the previous day
	glog.V(1).Infof("LoadSymbolsSnapshots.FetchSymbols: year: %d, month: %d, day: %d", year, month, day)

	for _, e := range exchangeIDs {
		symbols, err := l.LoadSymbols(year, month, day, e)
		if err != nil {
			glog.Errorf("LoadSymbolsSnapshots: cannot load symbols of exchange id '%d' from DB due to error %s", e, err)
			return nil, err
		}
		r.Exchanges = append(r.Exchanges, *symbols)
	}
	return nil, nil
}

// LoadSymbols loads the latest snapshot of symbols for a given exchnage from DB
func (l *DBLoaderImpl) LoadSymbols(year, month, day, exchangeID int) (*types.ExchangeSymbols, error) {
	var symbols types.ExchangeSymbols
	stmt, names := qb.Select("maketrades.symbols_snapshots").Where(qb.Eq("year"), qb.Eq("month"), qb.Eq("day"), qb.Eq("exchnage_id")).OrderBy("snaphost_time", qb.DESC).Limit(1).ToCql()
	q := gocqlx.Query(session.Query(stmt), names).BindMap(qb.M{
		"year": year, "month": month, "day": day, "exchange_id": exchangeID,
	})
	if err := q.GetRelease(&symbols); err != nil {
		glog.Errorf("LoadSymbols: cannot load the last snapshots with symbols for exchnage id '%d' due to error %s",
			exchangeID,
			err)
		return nil, err
	}
	return &symbols, nil
}
