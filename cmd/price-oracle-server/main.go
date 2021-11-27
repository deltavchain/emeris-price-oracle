package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/allinbits/emeris-price-oracle/price-oracle/config"
	"github.com/allinbits/emeris-price-oracle/price-oracle/database"
	"github.com/allinbits/emeris-price-oracle/price-oracle/rest"
	"github.com/allinbits/emeris-price-oracle/price-oracle/sql"
	"github.com/allinbits/emeris-price-oracle/utils/logging"
	"github.com/allinbits/emeris-price-oracle/utils/store"
)

var Version = "not specified"

func main() {
	config, err := config.Read()
	if err != nil {
		panic(err)
	}

	logger := logging.New(logging.LoggingConfig{
		LogPath: config.LogPath,
		Debug:   config.Debug,
	})

	logger.Infow("price-oracle-server", "version", Version)

	db, err := sql.NewDB(config.DatabaseConnectionURL)
	if err != nil {
		logger.Fatal(err)
	}

	storeHandler, err := database.NewStoreHandler(db)
	if err != nil {
		logger.Fatal(err)
	}
	ri, err := store.NewClient(config.RedisUrl)
	if err != nil {
		logger.Panicw("unable to start redis client", "error", err)
	}

	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg.Add(2)
	go func() {
		defer wg.Done()
		database.StartAggregate(ctx, storeHandler, logger, config, 5)
	}()
	go func() {
		defer wg.Done()
		database.StartSubscription(ctx, storeHandler, logger, config)
	}()

	restServer := rest.NewServer(
		storeHandler,
		ri,
		logger,
		config,
	)
	go func() {
		if err := restServer.Serve(config.ListenAddr); err != nil {
			logger.Panicw("rest http server error", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	cancel()
	wg.Wait()
	logger.Info("Shutting down server...")
}
