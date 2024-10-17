package models

type Currency struct {
	Name                string `json:"name"`
	PriceUsd            string `json:"price_usd"`
	MarketCap           string `json:"market_cap"`
	Change24hPercentage string `json:"change_24h_percentage"`
	Change7dPercentage  string `json:"change_7d_percentage"`
}
