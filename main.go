package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/etrubenok/make-trades-registry/fetchers"
	"github.com/etrubenok/make-trades-registry/types"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"

	"github.com/gocql/gocql"
)

// GetLatestSymbolsSnapshot gets the latest available snapshot about the input
// exchanges or all of them if the input slice 'exchanges' is empty
func GetLatestSymbolsSnapshot(exchanges []string) (*types.ExchangesSymbols, error) {
	// for _, e := range exchanges {
	// 	f, err := fetchers.FetcherFactory(e)
	// 	if err != nil {
	// 		glog.Errorf("GetLatestSymbolsSnapshot: cannot get a fetcher for exchange '%s' due to error %s", e, err)
	// 		return nil, err
	// 	}
	// 	exchangeSymbols, err := f.FetchSymbols()
	// 	if err != nil {
	// 		glog.Errorf("GetLatestSymbolsSnapshot: cannot fetch symbols from the ")
	// 	}
	// }
	return nil, fmt.Errorf("not implem ented yet")
}

func getSymbols(c *gin.Context) {
	filter := c.Request.URL.Query().Get("exchanges")
	exchanges := []string{}
	if filter != "" {
		exchanges = strings.Split(filter, "@")
	}
	symbolsSnapshot, err := GetLatestSymbolsSnapshot(exchanges)
	if err != nil {
		glog.Errorf("getSymbols: cannot get symbols due to error %s", err)
	}
	c.JSON(http.StatusForbidden, symbolsSnapshot)
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

	session, err := cluster.CreateSession()
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
				err := importer.SaveSymbolsSnapshots(time.Now().Unix(), &s)
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
