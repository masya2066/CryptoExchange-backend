package requestData

type Withdraw struct {
	Coin    string  `json:"coin"`
	Address string  `json:"address"`
	Amount  float64 `json:"amount"`
}
