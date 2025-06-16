package main

import (
	"database/sql"
	"encoding/json"
	"go-test/internal/logger"
	"go-test/internal/utils"
	"log"
	"os"
	"sync"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
)

type Batcher struct {
	mu     sync.Mutex
	buffer []logger.Event
}

func (b *Batcher) Add(e logger.Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buffer = append(b.buffer, e)
}

func (b *Batcher) Flush(db *sql.DB) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.buffer) == 0 {
		return
	}

	tx, err := db.Begin()
	if err != nil {
		log.Printf("failed to begin tx: %v", err)
		return
	}

	stmt, err := tx.Prepare("INSERT INTO goods_log (id, project_id, action, timestamp) VALUES (?, ?, ?, ?)")
	if err != nil {
		log.Printf("prepare failed: %v", err)
		_ = tx.Rollback()
		return
	}

	for _, e := range b.buffer {
		_, err := stmt.Exec(e.ID, e.ProjectID, e.Action, e.Timestamp)
		if err != nil {
			log.Printf("insert failed: %v", err)
		}
	}

	_ = stmt.Close()
	if err := tx.Commit(); err != nil {
		log.Printf("commit failed: %v", err)
	}

	log.Printf("flushed %d logs to ClickHouse", len(b.buffer))
	b.buffer = nil
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	clickhouseDSN := os.Getenv("CLICKHOUSE_DSN")
	natsURL := os.Getenv("NATS_URL")
	topic := os.Getenv("NATS_LOG_TOPIC")

	db, err := utils.RetryConnectToClickhouse(clickhouseDSN, 10, 2*time.Second)
	if err != nil {
		log.Fatalf("failed to connect to ClickHouse: %v", err)
	}
	defer db.Close()

	nc, err := utils.RetryConnectToNATS(natsURL, 10, 2*time.Second)
	if err != nil {
		log.Fatalf("failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	batcher := &Batcher{}
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {
			batcher.Flush(db)
		}
	}()

	_, err = nc.Subscribe(topic, func(msg *nats.Msg) {
		var e logger.Event
		if err := json.Unmarshal(msg.Data, &e); err != nil {
			log.Printf("failed to unmarshal event: %v", err)
			return
		}
		batcher.Add(e)
	})
	if err != nil {
		log.Fatalf("failed to subscribe: %v", err)
	}

	log.Println("consumer started and listening for logs...")
	select {}
}
