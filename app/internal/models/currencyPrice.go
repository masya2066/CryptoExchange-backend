package models

type CurrencyPrice struct {
	Currency string  `json:"currency"`
	UsdPrice float64 `json:"usd_price"`
}
