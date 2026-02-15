package model

import "time"

type EODPrice struct {
	ID       int64     `json:"-"`
	Symbol   string    `json:"symbol"`
	Date     string    `json:"date"`
	Open     float64   `json:"open"`
	High     float64   `json:"high"`
	Low      float64   `json:"low"`
	Close    float64   `json:"close"`
	Volume   int64     `json:"volume"`
	AdjClose float64   `json:"adj_close"`
	Exchange string    `json:"exchange"`
	FetchedAt time.Time `json:"-"`
}
