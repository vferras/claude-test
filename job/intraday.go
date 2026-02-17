package job

import (
	"database/sql"
	"log"
	"time"

	"claude-test/marketstack"
	"claude-test/model"
)

type IntradayJob struct {
	db      *sql.DB
	client  *marketstack.Client
	symbols []string
}

func NewIntradayJob(db *sql.DB, client *marketstack.Client, symbols []string) *IntradayJob {
	return &IntradayJob{
		db:      db,
		client:  client,
		symbols: symbols,
	}
}

func (j *IntradayJob) Run() {
	log.Println("Intraday job: starting fetch")

	today := lastTradingDay(time.Now())
	prices, err := j.client.FetchIntraday(j.symbols, today, today)
	if err != nil {
		log.Printf("Intraday job: fetch error: %v", err)
		return
	}

	log.Printf("Intraday job: fetched %d prices", len(prices))

	for _, p := range prices {
		if err := j.upsert(p); err != nil {
			log.Printf("Intraday job: upsert error for %s: %v", p.Symbol, err)
		}
	}

	log.Println("Intraday job: completed")
}

func (j *IntradayJob) Backfill(from, to time.Time) {
	log.Printf("Intraday backfill: fetching %s to %s", from.Format("2006-01-02"), to.Format("2006-01-02"))

	prices, err := j.client.FetchIntraday(j.symbols, from, to)
	if err != nil {
		log.Printf("Intraday backfill: fetch error: %v", err)
		return
	}

	log.Printf("Intraday backfill: fetched %d prices", len(prices))

	for _, p := range prices {
		if err := j.upsert(p); err != nil {
			log.Printf("Intraday backfill: upsert error for %s on %s: %v", p.Symbol, p.Date, err)
		}
	}

	log.Println("Intraday backfill: completed")
}

func (j *IntradayJob) upsert(p model.IntradayPrice) error {
	query := `
		INSERT INTO intraday_prices (symbol, date, interval, open, high, low, close, volume, exchange)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (symbol, date, interval) DO UPDATE SET
			open = EXCLUDED.open,
			high = EXCLUDED.high,
			low = EXCLUDED.low,
			close = EXCLUDED.close,
			volume = EXCLUDED.volume,
			exchange = EXCLUDED.exchange,
			fetched_at = NOW()`

	_, err := j.db.Exec(query, p.Symbol, p.Date, p.Interval, p.Open, p.High, p.Low, p.Close, p.Volume, p.Exchange)
	return err
}

func (j *IntradayJob) Schedule() {
	go j.Run()

	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day(), 22, 0, 0, 0, now.Location())
			if !next.After(now) {
				next = next.Add(24 * time.Hour)
			}

			duration := next.Sub(now)
			log.Printf("Intraday job: next run scheduled in %s (at %s)", duration.Round(time.Second), next.Format(time.RFC3339))

			timer := time.NewTimer(duration)
			<-timer.C

			j.Run()
		}
	}()
}
