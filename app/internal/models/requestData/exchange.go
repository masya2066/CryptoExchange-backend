package requestData

type Exchange struct {
	BtcBalance  float64 `json:"btc_balance"`
	EthBalance  float64 `json:"eth_balance"`
	TrxBalance  float64 `json:"trx_balance"`
	SoliBalance float64 `json:"soli_balance"`
}
