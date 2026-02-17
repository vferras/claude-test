package db

import (
	"database/sql"
	"fmt"
)

func Migrate(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS eod_prices (
		id            SERIAL PRIMARY KEY,
		symbol        TEXT NOT NULL,
		date          DATE NOT NULL,
		open          NUMERIC(12,4),
		high          NUMERIC(12,4),
		low           NUMERIC(12,4),
		close         NUMERIC(12,4),
		volume        BIGINT,
		adj_close     NUMERIC(12,4),
		exchange      TEXT,
		fetched_at    TIMESTAMPTZ DEFAULT NOW(),
		UNIQUE(symbol, date)
	);`

	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("running eod migration: %w", err)
	}

	intradayQuery := `
	CREATE TABLE IF NOT EXISTS intraday_prices (
		id            SERIAL PRIMARY KEY,
		symbol        TEXT NOT NULL,
		date          TIMESTAMPTZ NOT NULL,
		interval      TEXT NOT NULL,
		open          NUMERIC(12,4),
		high          NUMERIC(12,4),
		low           NUMERIC(12,4),
		close         NUMERIC(12,4),
		volume        BIGINT,
		exchange      TEXT,
		fetched_at    TIMESTAMPTZ DEFAULT NOW(),
		UNIQUE(symbol, date, interval)
	);`

	if _, err := db.Exec(intradayQuery); err != nil {
		return fmt.Errorf("running intraday migration: %w", err)
	}

	return nil
}
