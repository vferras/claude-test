package model

import "time"

type IntradayPrice struct {
	ID        int64     `json:"-"`
	Symbol    string    `json:"symbol"`
	Date      string    `json:"date"`
	Interval  string    `json:"interval"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    int64     `json:"volume"`
	Exchange  string    `json:"exchange"`
	FetchedAt time.Time `json:"-"`
}
