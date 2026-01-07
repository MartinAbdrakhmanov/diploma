package storage

import (
	"fmt"
	"os"

	"github.com/lpernett/godotenv"
)

func loadDSN() (masterDSN, slaveDSN string, err error) {
	err = godotenv.Load()
	if err != nil {
		return "", "", fmt.Errorf("failed to load .env file: %w", err)
	}

	postgresUser := os.Getenv("POSTGRES_USER")
	if postgresUser == "" {
		return "", "", fmt.Errorf("POSTGRES_USER is not set in .env file")
	}
	postgresPswd := os.Getenv("POSTGRES_PASSWORD")
	if postgresPswd == "" {
		return "", "", fmt.Errorf("POSTGRES_PASSWORD is not set in .env file")
	}
	postgresHost := os.Getenv("POSTGRES_HOST")
	if postgresHost == "" {
		return "", "", fmt.Errorf("POSTGRES_HOST is not set in .env file")
	}
	postgresPort := os.Getenv("POSTGRES_PORT")
	if postgresPort == "" {
		return "", "", fmt.Errorf("POSTGRES_PORT is not set in .env file")
	}
	postgresDB := os.Getenv("POSTGRES_DB")
	if postgresDB == "" {
		return "", "", fmt.Errorf("POSTGRES_DB is not set in .env file")
	}
	replPort := os.Getenv("REPL_PORT")
	if replPort != "" {
		slaveDSN = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			postgresUser, postgresPswd, postgresHost, replPort, postgresDB)
	}

	masterDSN = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		postgresUser, postgresPswd, postgresHost, postgresPort, postgresDB)

	return masterDSN, slaveDSN, nil
}
