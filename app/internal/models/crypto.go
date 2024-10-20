package models

type Currency struct {
	Name                string `json:"name"`
	PriceUsd            string `json:"price_usd"`
	MarketCap           string `json:"market_cap"`
	Change24hPercentage string `json:"change_24h_percentage"`
	Change7dPercentage  string `json:"change_7d_percentage"`
}

type Withdraw struct {
	UserID     uint    `json:"user_id"`
	WithdrawID int64   `json:"withdraw_id" gorm:"unique"`
	Coin       string  `json:"coin"`
	Address    string  `json:"address"`
	Amount     float64 `json:"amount"`
	Status     int     `json:"status"`
	Created    string  `json:"created"`
	Updated    string  `json:"updated"`
}
