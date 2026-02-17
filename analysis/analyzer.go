package analysis

import (
	"database/sql"
	"fmt"
	"math"
	"sort"
	"time"
)

type SymbolScore struct {
	Rank        int     `json:"rank"`
	Symbol      string  `json:"symbol"`
	Score       float64 `json:"score"`
	Momentum    float64 `json:"momentum"`
	Volatility  float64 `json:"volatility"`
	VolumeTrend float64 `json:"volume_trend"`
	LatestClose float64 `json:"latest_close"`
}

type AnalysisResult struct {
	Date     string        `json:"date"`
	Rankings []SymbolScore `json:"rankings"`
}

type dailyRow struct {
	date   time.Time
	close  float64
	volume int64
}

func Analyze(db *sql.DB) (*AnalysisResult, error) {
	symbols, err := getSymbols(db)
	if err != nil {
		return nil, err
	}

	var scores []SymbolScore
	for _, sym := range symbols {
		rows, err := getHistory(db, sym)
		if err != nil {
			return nil, fmt.Errorf("querying %s: %w", sym, err)
		}
		if len(rows) == 0 {
			continue
		}
		scores = append(scores, computeRawScore(sym, rows))
	}

	if len(scores) == 0 {
		return &AnalysisResult{Date: time.Now().Format("2006-01-02"), Rankings: []SymbolScore{}}, nil
	}

	normalize(scores)

	sort.Slice(scores, func(i, j int) bool { return scores[i].Score > scores[j].Score })
	for i := range scores {
		scores[i].Rank = i + 1
	}

	return &AnalysisResult{
		Date:     time.Now().Format("2006-01-02"),
		Rankings: scores,
	}, nil
}

func getSymbols(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT DISTINCT symbol FROM eod_prices ORDER BY symbol")
	if err != nil {
		return nil, fmt.Errorf("querying symbols: %w", err)
	}
	defer rows.Close()

	var symbols []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		symbols = append(symbols, s)
	}
	return symbols, rows.Err()
}

func getHistory(db *sql.DB, symbol string) ([]dailyRow, error) {
	rows, err := db.Query(
		`SELECT date, close, volume FROM eod_prices WHERE symbol = $1 ORDER BY date ASC`,
		symbol,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []dailyRow
	for rows.Next() {
		var r dailyRow
		if err := rows.Scan(&r.date, &r.close, &r.volume); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, rows.Err()
}

func computeRawScore(symbol string, rows []dailyRow) SymbolScore {
	latest := rows[len(rows)-1]
	oldest := rows[0]

	// Momentum: % change from oldest to latest close
	var momentum float64
	if oldest.close != 0 {
		momentum = (latest.close - oldest.close) / oldest.close
	}

	// Volatility: standard deviation of daily returns (lower is better)
	var volatility float64
	if len(rows) > 1 {
		var returns []float64
		for i := 1; i < len(rows); i++ {
			if rows[i-1].close != 0 {
				returns = append(returns, (rows[i].close-rows[i-1].close)/rows[i-1].close)
			}
		}
		volatility = stddev(returns)
	}

	// Volume trend: recent volume (last 5 days or fewer) vs overall average
	var volumeTrend float64
	var totalVol int64
	for _, r := range rows {
		totalVol += r.volume
	}
	avgVol := float64(totalVol) / float64(len(rows))

	recentN := 5
	if len(rows) < recentN {
		recentN = len(rows)
	}
	var recentVol int64
	for _, r := range rows[len(rows)-recentN:] {
		recentVol += r.volume
	}
	recentAvg := float64(recentVol) / float64(recentN)

	if avgVol > 0 {
		volumeTrend = recentAvg / avgVol
	}

	return SymbolScore{
		Symbol:      symbol,
		Momentum:    momentum,
		Volatility:  volatility,
		VolumeTrend: volumeTrend,
		LatestClose: latest.close,
	}
}

func normalize(scores []SymbolScore) {
	// Min-max normalize each factor, then compute weighted score
	// Momentum: higher is better
	// Volatility: lower is better (invert after normalization)
	// Volume trend: higher is better

	momMin, momMax := minMax(scores, func(s SymbolScore) float64 { return s.Momentum })
	volMin, volMax := minMax(scores, func(s SymbolScore) float64 { return s.Volatility })
	vtMin, vtMax := minMax(scores, func(s SymbolScore) float64 { return s.VolumeTrend })

	for i := range scores {
		momNorm := minMaxNorm(scores[i].Momentum, momMin, momMax)
		volNorm := 1.0 - minMaxNorm(scores[i].Volatility, volMin, volMax) // invert: low vol = high score
		vtNorm := minMaxNorm(scores[i].VolumeTrend, vtMin, vtMax)

		scores[i].Score = math.Round((0.4*momNorm+0.3*volNorm+0.3*vtNorm)*100) / 100
	}
}

func minMax(scores []SymbolScore, f func(SymbolScore) float64) (float64, float64) {
	min, max := f(scores[0]), f(scores[0])
	for _, s := range scores[1:] {
		v := f(s)
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max
}

func minMaxNorm(val, min, max float64) float64 {
	if max == min {
		return 0.5
	}
	return (val - min) / (max - min)
}

func stddev(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	var sum float64
	for _, v := range vals {
		sum += v
	}
	mean := sum / float64(len(vals))
	var sqSum float64
	for _, v := range vals {
		d := v - mean
		sqSum += d * d
	}
	return math.Sqrt(sqSum / float64(len(vals)))
}
