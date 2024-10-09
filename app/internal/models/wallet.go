package models

type EthWallet struct {
	Login      string `json:"login"`
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	Address    string `json:"address"`
}

type BtcWallet struct {
	Login      string `json:"login"`
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	Address    string `json:"address"`
}

type UsdtWallet struct {
	Login      string `json:"login"`
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	Address    string `json:"address"`
}

// Define a structure to unmarshal the JSON response
type BlockchairResponse struct {
	Data struct {
		Address struct {
			Balance int64 `json:"balance"`
		} `json:"address"`
	} `json:"data"`
}
