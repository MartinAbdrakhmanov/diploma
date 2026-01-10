package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitDB(ctx context.Context) (masterDB *pgxpool.Pool, err error) {
	masterDSN, err := loadDSN()
	if err != nil {
		log.Fatalf("storage.InitDB: Error loading DSN: %v", err)
	}

	masterDB, err = pgxpool.New(ctx, masterDSN)
	if err != nil {
		return nil, fmt.Errorf("storage.InitDB: failed to create connection pool: %w", err)
	}

	if err = masterDB.Ping(ctx); err != nil {
		return nil, fmt.Errorf("storage.InitDB: failed to ping database: %w", err)
	}

	log.Println("Database connection established...")
	return masterDB, nil

}
