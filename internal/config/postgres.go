package config

import (
	"context"
	"fmt"
	"starter/internal/app/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

//go:generate mockery --name DBPool
type DBPool interface {
	Ping(ctx context.Context) error
	Begin(ctx context.Context) (pgx.Tx, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Close()
}

var _ DBPool = (*pgxpool.Pool)(nil)

func getConnectionURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s&pool_max_conns=%s&pool_min_conns=%s&pool_max_conn_lifetime=%s&pool_max_conn_idle_time=%s",
		utils.GetEnvAsString("DB_USER", "postgres"),
		utils.GetEnvAsString("DB_PASSWORD", "postgres"),
		utils.GetEnvAsString("DB_HOST", "127.0.0.1"),
		utils.GetEnvAsString("DB_PORT", "5432"),
		utils.GetEnvAsString("DB_NAME", "wtbbe_dev"),
		utils.GetEnvAsString("DB_SSL_MODE", "disable"),
		utils.GetEnvAsString("DB_MAX_POOL_CONNECTIONS", "10"),
		utils.GetEnvAsString("DB_MIN_POOL_CONNECTIONS", "1"),
		utils.GetEnvAsString("DB_MAX_CONN_LIFETIME", "30m"),
		utils.GetEnvAsString("DB_MAX_CONN_IDLE_TIME", "10m"),
	)
}

func ConnectDB() DBPool {
	// Construct the connection string
	connectionURL := getConnectionURL()

	// Connect to the database
	dbConnection, err := pgxpool.New(context.Background(), connectionURL)

	if err != nil {
		logrus.Fatalf("Unable to connect to database: %v\n", err)
		return nil
	}

	err = dbConnection.Ping(context.Background())
	if err != nil {
		logrus.Fatalf("Unable to ping the database: %v\n", err)
		return nil
	}
	return dbConnection
}
