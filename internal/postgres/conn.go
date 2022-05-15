package postgres

import (
	"context"
	"fmt"
	"pgq/internal/configuration"
	"pgq/internal/db"

	"github.com/jackc/pgx/v4/pgxpool"
)

func ConnectStringFromConfig(dbConf *configuration.Database) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		dbConf.Host, dbConf.Port, dbConf.Username, dbConf.Password, dbConf.DatabaseName, dbConf.SSLMode,
	)
}

func New(ctx context.Context, dbConf *configuration.Database) (*pgxpool.Pool, *db.Queries, error) {
	psqlInfo := ConnectStringFromConfig(dbConf)

	dbConn, err := pgxpool.Connect(ctx, psqlInfo)
	if err != nil {
		return nil, nil, fmt.Errorf("opening conn: %w", err)
	}

	if err := dbConn.Ping(ctx); err != nil {
		return nil, nil, fmt.Errorf("ping: %w", err)
	}

	return dbConn, db.New(dbConn), nil
}
