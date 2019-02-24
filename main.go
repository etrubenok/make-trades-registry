package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/etrubenok/make-trades-registry/fetchers"
	"github.com/etrubenok/make-trades-registry/types"
	"github.com/etrubenok/make-trades-types/registry"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"

	"github.com/gocql/gocql"
)

var session *gocql.Session

// GetPreviousDate returns year, month, day of the previous day from the currentTime
func GetPreviousDate(currentTime time.Time) (int, int, int) {
	t := currentTime.AddDate(0, 0, -1).UnixNano() / int64(time.Millisecond)
	year, month, day := fetchers.GetYearMonthDay(t)
	glog.Infof("GetPreviousDate: previous day (year: %d, month: %d, day: %d)",
		year, month, day)
	return year, month, day
}

// GetYearMonthDay returns year, month and day for a given date in format 'yyyy-mm-dd'
func GetYearMonthDay(yyyymmddDate string) (int, int, int, error) {
	t, err := time.Parse("2006-01-02", yyyymmddDate)
	if err != nil {
		glog.Errorf("GetYearMonthDay: cannot parse date string in format 'yyyy-mm-dd' (%s) due to error %s", yyyymmddDate, err)
		return 0, 0, 0, err
	}
	return t.Year(), int(t.Month()), t.Day(), nil
}

// GetSymbolsSnapshot gets symbols snapshot on a date
func GetSymbolsSnapshot(exchanges []string, getDate func() (int, int, int, error)) (*types.APIExchangesSymbols, error) {

	exchnageIDs := make([]int, 0)
	for _, e := range exchanges {
		exchangeID, err := registry.GetExchangeID(e)
		if err != nil {
			glog.Errorf("GetLatestSymbolsSnapshot: cannot get exchange id for exchange '%s' due to error %s", e, err)
			return nil, err
		}
		exchnageIDs = append(exchnageIDs, exchangeID)
	}
	l := NewDBLoader(session)
	exchangesSymbols, err := l.LoadSymbolsSnapshots(exchnageIDs, getDate)

	if err != nil {
		glog.Errorf("GetLatestSymbolsSnapshot: LoadSymbolsSnapshots failed to load the symbols for exchnages %v due to error %s", exchanges, err)
		return nil, err
	}
	resp, err := types.ConvertExchangeSymbolsToAPIResponse(exchangesSymbols)
	if err != nil {
		glog.Errorf("ConvertExchangeSymbolsToAPIResponse: cannot convert to API response due to error %s", err)
		return nil, err
	}
	return resp, nil
}

// GetAllExchanges returns all the supported exchanges
func GetAllExchanges() []string {
	return []string{"binance", "bitfinex"}
}

func getSymbols(c *gin.Context) {
	filter := c.Request.URL.Query().Get("exchanges")
	exchanges := []string{}
	if filter != "" {
		exchanges = strings.Split(filter, "@")
	}
	if filter == "" {
		exchanges = GetAllExchanges()
	}

	date := c.Request.URL.Query().Get("date")
	var symbolsSnapshot *types.APIExchangesSymbols
	var err error
	if date != "" {
		symbolsSnapshot, err = GetSymbolsSnapshot(exchanges, func() (int, int, int, error) {
			year, month, day, err := GetYearMonthDay(date)
			if err != nil {
				glog.Errorf("getSymbols: cannot get year, month and day from string 'yyyy-mm-dd'(%s) due to error '%s'",
					date, err)
				return 0, 0, 0, nil
			}
			return year, month, day, nil
		})
	} else {
		symbolsSnapshot, err = GetSymbolsSnapshot(exchanges, func() (int, int, int, error) {
			t := time.Now().UnixNano() / int64(time.Millisecond)
			year, month, day := fetchers.GetYearMonthDay(t)
			return year, month, day, nil
		})
	}
	if err != nil {
		glog.Errorf("getSymbols: cannot get symbols for exchanges '%v' and date '%s' due to error '%s'",
			exchanges,
			date,
			err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "server error"})
		return
	}
	c.JSON(http.StatusOK, symbolsSnapshot)
}

func main() {
	flag.Parse()

	cluster := gocql.NewCluster("do-trade-scylla-scylladb-0.do-trade-scylla-scylladb",
		"do-trade-scylla-scylladb-1.do-trade-scylla-scylladb",
		"do-trade-scylla-scylladb-2.do-trade-scylla-scylladb")
	cluster.Keyspace = "maketrades"
	cluster.Consistency = gocql.Quorum
	cluster.RetryPolicy = &gocql.ExponentialBackoffRetryPolicy{
		NumRetries: 10,
		Min:        10 * time.Millisecond,
		Max:        2 * time.Second}

	var err error
	session, err = cluster.CreateSession()
	if err != nil {
		glog.Errorf("main: cannot create a session to the DB due to error %s", err)
		panic(err)
	}
	defer session.Close()

	snapshots := make(chan types.ExchangesSymbols)
	job := fetchers.NewFetchJob()
	job.Init([]string{"binance"}, snapshots)

	importer := NewDBImporter(session)
	go func() {
		for {
			s := <-snapshots
			glog.Infof("main: snapshot 'binance' symbols number %d", len(s.Exchanges[0].Symbols))
			if len(s.Exchanges) == 0 {
				glog.Errorf("main: snapshot is empty")
			} else {
				err := importer.SaveSymbolsSnapshots(&s)
				if err != nil {
					glog.Errorf("main: cannot import snapshots for %d exchange(s) into the database due to error %s", len(s.Exchanges), err)
					panic(err)
				}
			}
		}
	}()

	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.GET("/symbols", getSymbols)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			glog.Errorf("listen: %s", err)
		}
	}()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	run := true
	for run == true {
		select {
		case sig := <-sigchan:
			glog.Infof("Caught signal %v: terminating", sig)
			run = false
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		glog.Errorf("Server Shutdown: %s", err)
	}
	glog.Infof("Server exiting")
}
