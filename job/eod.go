package job

import (
	"database/sql"
	"log"
	"time"

	"claude-test/marketstack"
	"claude-test/model"
)

type EODJob struct {
	db      *sql.DB
	client  *marketstack.Client
	symbols []string
}

func NewEODJob(db *sql.DB, client *marketstack.Client, symbols []string) *EODJob {
	return &EODJob{
		db:      db,
		client:  client,
		symbols: symbols,
	}
}

func (j *EODJob) Run() {
	log.Println("EOD job: starting fetch")

	today := lastTradingDay(time.Now())
	prices, err := j.client.FetchEOD(j.symbols, today)
	if err != nil {
		log.Printf("EOD job: fetch error: %v", err)
		return
	}

	log.Printf("EOD job: fetched %d prices", len(prices))

	for _, p := range prices {
		if err := j.upsert(p); err != nil {
			log.Printf("EOD job: upsert error for %s: %v", p.Symbol, err)
		}
	}

	log.Println("EOD job: completed")
}

func (j *EODJob) upsert(p model.EODPrice) error {
	query := `
		INSERT INTO eod_prices (symbol, date, open, high, low, close, volume, adj_close, exchange)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (symbol, date) DO UPDATE SET
			open = EXCLUDED.open,
			high = EXCLUDED.high,
			low = EXCLUDED.low,
			close = EXCLUDED.close,
			volume = EXCLUDED.volume,
			adj_close = EXCLUDED.adj_close,
			exchange = EXCLUDED.exchange,
			fetched_at = NOW()`

	_, err := j.db.Exec(query, p.Symbol, p.Date, p.Open, p.High, p.Low, p.Close, p.Volume, p.AdjClose, p.Exchange)
	return err
}

func lastTradingDay(t time.Time) time.Time {
	switch t.Weekday() {
	case time.Sunday:
		return t.AddDate(0, 0, -2)
	case time.Saturday:
		return t.AddDate(0, 0, -1)
	default:
		return t
	}
}

func (j *EODJob) Schedule() {
	// Run immediately on startup
	go j.Run()

	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day(), 22, 0, 0, 0, now.Location())
			if !next.After(now) {
				next = next.Add(24 * time.Hour)
			}

			duration := next.Sub(now)
			log.Printf("EOD job: next run scheduled in %s (at %s)", duration.Round(time.Second), next.Format(time.RFC3339))

			timer := time.NewTimer(duration)
			<-timer.C

			j.Run()
		}
	}()
}
