package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitDB(ctx context.Context) (masterDB, slaveDB *pgxpool.Pool, err error) {
	masterDSN, slaveDSN, err := loadDSN()
	if err != nil {
		log.Fatalf("storage.InitDB: Error loading DSN: %v", err)
	}

	masterDB, err = pgxpool.New(ctx, masterDSN)
	if err != nil {
		return nil, nil, fmt.Errorf("storage.InitDB: failed to create connection pool: %w", err)
	}

	if err = masterDB.Ping(ctx); err != nil {
		return nil, nil, fmt.Errorf("storage.InitDB: failed to ping database: %w", err)
	}
	if slaveDSN == "" {
		log.Println("Database connection established...")
		return masterDB, nil, nil
	}
	slaveDB, err = pgxpool.New(ctx, slaveDSN)
	if err != nil {
		return nil, nil, fmt.Errorf("storage.InitDB: failed to create slave connection pool: %w", err)
	}

	if err = slaveDB.Ping(ctx); err != nil {
		return nil, nil, fmt.Errorf("storage.InitDB: failed to ping slave database: %w", err)
	}
	log.Println("Databases connection established...")

	return masterDB, slaveDB, nil
}
