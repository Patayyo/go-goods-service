package utils

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

func RetryConnectToPostgres(dsn string, attempts int, delay time.Duration) (*sql.DB, error) {
	var db *sql.DB
	var err error

	for i := 0; i < attempts; i++ {
		db, err = sql.Open("postgres", dsn)
		if err == nil && db.Ping() == nil {
			fmt.Println("Successfully connected to PostgreSQL")
			return db, nil
		}
		log.Printf("Waiting for PostgreSQL (%d/%d): %v", i+1, attempts, err)
		time.Sleep(delay)
	}
	return nil, fmt.Errorf("could not connect to PostgreSQL: %v", err)
}

func RetryConnectToNATS(url string, attempts int, delay time.Duration) (*nats.Conn, error) {
	var nc *nats.Conn
	var err error

	for i := 0; i < attempts; i++ {
		nc, err = nats.Connect(url)
		if err == nil {
			fmt.Println("Successfully connected to NATS")
			return nc, nil
		}
		log.Printf("Waiting for NATS (%d/%d): %v", i+1, attempts, err)
		time.Sleep(delay)
	}
	return nil, fmt.Errorf("could not connect to NATS: %v", err)
}

func RetryConnectToClickhouse(dsn string, attempts int, delay time.Duration) (*sql.DB, error) {
	var db *sql.DB
	var err error

	for i := 0; i < attempts; i++ {
		db, err = sql.Open("clickhouse", dsn)
		if err == nil && db.Ping() == nil {
			log.Println("Successfully connected to ClickHouse")
			return db, nil
		}
		log.Printf("Waiting for ClickHouse (%d/%d): %v", i+1, attempts, err)
		time.Sleep(delay)
	}
	return nil, fmt.Errorf("could not connect to ClickHouse: %v", err)
}
