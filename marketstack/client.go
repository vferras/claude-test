package marketstack

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"claude-test/model"
)

const (
	baseURL        = "http://api.marketstack.com/v1/eod"
	intradayURL    = "http://api.marketstack.com/v1/intraday"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

type apiResponse struct {
	Data []apiEOD `json:"data"`
}

type apiIntradayResponse struct {
	Data []apiIntraday `json:"data"`
}

type apiIntraday struct {
	Open     float64 `json:"open"`
	High     float64 `json:"high"`
	Low      float64 `json:"low"`
	Close    float64 `json:"close"`
	Volume   float64 `json:"volume"`
	Symbol   string  `json:"symbol"`
	Exchange string  `json:"exchange"`
	Date     string  `json:"date"`
}

type apiEOD struct {
	Open     float64 `json:"open"`
	High     float64 `json:"high"`
	Low      float64 `json:"low"`
	Close    float64 `json:"close"`
	Volume   float64 `json:"volume"`
	AdjClose float64 `json:"adj_close"`
	Symbol   string  `json:"symbol"`
	Exchange string  `json:"exchange"`
	Date     string  `json:"date"`
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) FetchEOD(symbols []string, date time.Time) ([]model.EODPrice, error) {
	return c.FetchEODRange(symbols, date, date)
}

func (c *Client) FetchEODRange(symbols []string, from, to time.Time) ([]model.EODPrice, error) {
	fromStr := from.Format("2006-01-02")
	toStr := to.Format("2006-01-02")

	params := url.Values{}
	params.Set("access_key", c.apiKey)
	params.Set("symbols", strings.Join(symbols, ","))
	params.Set("date_from", fromStr)
	params.Set("date_to", toStr)
	params.Set("limit", "1000")

	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("requesting marketstack API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("marketstack API returned status %d", resp.StatusCode)
	}

	var apiResp apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decoding marketstack response: %w", err)
	}

	var prices []model.EODPrice
	for _, d := range apiResp.Data {
		dateStr := d.Date
		if len(dateStr) >= 10 {
			dateStr = dateStr[:10]
		}
		prices = append(prices, model.EODPrice{
			Symbol:   d.Symbol,
			Date:     dateStr,
			Open:     d.Open,
			High:     d.High,
			Low:      d.Low,
			Close:    d.Close,
			Volume:   int64(d.Volume),
			AdjClose: d.AdjClose,
			Exchange: d.Exchange,
		})
	}

	return prices, nil
}

func (c *Client) FetchIntraday(symbols []string, from, to time.Time) ([]model.IntradayPrice, error) {
	fromStr := from.Format("2006-01-02")
	toStr := to.Format("2006-01-02")

	params := url.Values{}
	params.Set("access_key", c.apiKey)
	params.Set("symbols", strings.Join(symbols, ","))
	params.Set("interval", "10min")
	params.Set("date_from", fromStr)
	params.Set("date_to", toStr)
	params.Set("limit", "1000")

	reqURL := fmt.Sprintf("%s?%s", intradayURL, params.Encode())

	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("requesting marketstack intraday API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("marketstack intraday API returned status %d", resp.StatusCode)
	}

	var apiResp apiIntradayResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decoding marketstack intraday response: %w", err)
	}

	var prices []model.IntradayPrice
	for _, d := range apiResp.Data {
		prices = append(prices, model.IntradayPrice{
			Symbol:   d.Symbol,
			Date:     d.Date,
			Interval: "10min",
			Open:     d.Open,
			High:     d.High,
			Low:      d.Low,
			Close:    d.Close,
			Volume:   int64(d.Volume),
			Exchange: d.Exchange,
		})
	}

	return prices, nil
}
