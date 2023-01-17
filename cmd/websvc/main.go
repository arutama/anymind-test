package main

import (
	"anymind/src/api"
	"anymind/src/httpapi"
	"anymind/src/persistence"
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type appCfg struct {
	port  int
	pgdsn string
}

// loadCfg will initialize configuration from env var.
func loadCfg() *appCfg {
	viper.AutomaticEnv()
	viper.AllowEmptyEnv(false)

	cfg := &appCfg{
		port:  viper.GetInt("HTTP_PORT"),
		pgdsn: viper.GetString("PG_DSN"),
	}

	return cfg
}

func main() {
	cfg := loadCfg()
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	db, err := sql.Open("pgx", cfg.pgdsn)
	if err != nil {
		logger.Fatal("unable to connect db", zap.Error(err))
	}
	defer db.Close()

	persistenceSvc := persistence.NewService(
		db,
		persistence.WithLogger(logger))

	apiSvc := api.NewService(
		persistenceSvc,
		api.WithLogger(logger))

	httpService := httpapi.NewService(
		apiSvc,
		httpapi.WithLogger(logger),
		httpapi.WithListenPort(8080))

	var svcRunning sync.WaitGroup

	svcRunning.Add(1)
	go func() {
		defer svcRunning.Done()
		httpService.Start(ctx)
		<-ctx.Done()
		cancel()
	}()

	defer func() {
		cancel()
		svcRunning.Wait()
	}()

	// wait for interrupt or terminate signal
	waiter := make(chan os.Signal, 1)
	signal.Notify(waiter, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-waiter:
		logger.Info("catch exit signal")
	case <-ctx.Done():
	}
}
