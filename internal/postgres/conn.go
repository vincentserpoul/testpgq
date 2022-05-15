package postgres

import (
	"context"
	"fmt"
	"pgq/internal/configuration"
	"pgq/internal/db"

	"github.com/jackc/pgx/v4/pgxpool"
)

func DSNFromConfig(dbConf *configuration.Database) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s pool_max_conns=%d pool_min_conns=%d",
		dbConf.Host, dbConf.Port, dbConf.Username, dbConf.Password, dbConf.DatabaseName, dbConf.SSLMode, 10, 10,
	)
}

func New(ctx context.Context, dbConf *configuration.Database) (*pgxpool.Pool, *db.Queries, error) {
	poolConf, err := pgxpool.ParseConfig(DSNFromConfig(dbConf))
	if err != nil {
		return nil, nil, fmt.Errorf("opening parsing config: %w", err)
	}

	dbConn, err := pgxpool.ConnectConfig(ctx, poolConf)

	if err != nil {
		return nil, nil, fmt.Errorf("opening conn: %w", err)
	}

	if err := dbConn.Ping(ctx); err != nil {
		return nil, nil, fmt.Errorf("ping: %w", err)
	}

	return dbConn, db.New(dbConn), nil
}
