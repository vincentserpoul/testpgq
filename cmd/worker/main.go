package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"pgq/internal/configuration"
	"pgq/internal/db"
	"pgq/internal/postgres"
	"sync"
	"time"

	"github.com/induzo/otelinit"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const workerPoolSize = 10
const batchSize = 1000

func main() {
	// context
	ctx := context.Background()

	// configuration
	currEnv := "local"
	if e := os.Getenv("APP_ENVIRONMENT"); e != "" {
		currEnv = e
	}

	cfg, err := configuration.GetConfig(currEnv)
	if err != nil {
		if errors.Is(err, configuration.MissingBaseConfigError{}) {
			log.Printf("getConfig: %v", err)

			return
		}

		log.Printf("getConfig: %v", err)
	}

	// logging
	if cfg.Application.PrettyLog {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	shutdown, err := otelinit.InitProvider(
		ctx,
		"simple-gohttp",
		otelinit.WithGRPCTraceExporter(
			ctx,
			fmt.Sprintf(
				"%s:%d",
				cfg.Observability.Collector.Host,
				cfg.Observability.Collector.Port,
			),
		),
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize opentelemetry")

		return
	}

	defer func() {
		if err := shutdown(); err != nil {
			log.Error().Err(err).Msg("failed to shutdown")
		}
	}()

	// querier
	dbPool, q, err := postgres.New(ctx, &cfg.Database)
	if err != nil {
		log.Warn().Err(err).Msg("postgres")

		return
	}
	defer dbPool.Close()

	// Set up cancellation context and waitgroup
	ctx, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()

	wg := &sync.WaitGroup{}

	// start err logger
	errC := make(chan error)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case err := <-errC:
				log.Error().Err(err).Msg("error")
			}
		}
	}()

	// Start workers
	for i := 0; i < workerPoolSize; i++ {
		dbConn, err := dbPool.Acquire(ctx)
		if err != nil {
			log.Error().Err(err).Msg("failed to acquire db connection")
		}

		wg.Add(1)
		go processQueue(ctx, wg, errC, q, dbConn)
	}

	log.Info().Msgf("processing")
	wg.Wait()
	cancelFunc()

	log.Info().Msgf("done! waiting for all in flights to finish")
	time.Sleep(3 * time.Second)

	log.Info().Msg("no more work to do, shutting down!")
}

func processQueue(ctx context.Context, wg *sync.WaitGroup, errC chan error, q *db.Queries, dbConn *pgxpool.Conn) {
	defer func() {
		dbConn.Release()
		wg.Done()
	}()

	// start done listener
	doneC := make(chan struct{}, 1)

	for {
		select {
		case <-ctx.Done():
			return
		case <-doneC:
			return
		default:
			if err := processBatch(ctx, errC, doneC, q, dbConn); err != nil {
				errC <- fmt.Errorf("processQueue: %w", err)
			}
		}
	}
}

func processBatch(ctx context.Context, errC chan error, doneC chan struct{}, q *db.Queries, dbConn *pgxpool.Conn) error {
	// create tx
	tx, err := dbConn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("impossible to start tx: %w", err)
	}

	var errTx error
	defer func() {
		if errTx != nil {
			if errRB := tx.Rollback(ctx); errRB != nil {
				errC <- errRB
			}

			return
		}

		if errCM := tx.Commit(ctx); errCM != nil {
			errC <- errCM
		}
	}()

	q = q.WithTx(tx)

	// select x from queue
	items, err := q.BatchGet(ctx, batchSize)
	if err != nil {
		errTx = err
		errC <- fmt.Errorf("BatchGet: %w", err)

		return errTx
	}

	if len(items) == 0 {
		doneC <- struct{}{}

		return nil
	}

	tbu := make([]int32, len(items))
	for i, item := range items {
		tbu[i] = item.ID
	}

	bResult := q.SetProcessed(ctx, tbu)
	defer bResult.Close()

	bResult.Exec(func(_ int, err error) {
		if err != nil {
			errTx = err
			errC <- fmt.Errorf("SetProcessed: %w", err)
		}
	})

	return nil
}
