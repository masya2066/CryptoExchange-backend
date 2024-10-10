package crypto

import (
	"bytes"
	"crypto-exchange/app/pkg/generator"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type AllWallets struct {
	Trx Wallet `json:"trx_wallet"`
	Eth Wallet `json:"eth_wallet"`
	Btc Wallet `json:"btx_wallet"`
}

type Wallet struct {
	Address    string `json:"address"`
	SeedPhrase string `json:"seed_phrase"`
}

type createWalletResponse struct {
	PrivateKey string `json:"private_key"`
	Address    string `json:"address"`
}

func createWallet(phrase string, typeWallet string) (Wallet, error) {
	var url string
	var response createWalletResponse
	if typeWallet == "eth" {
		url = os.Getenv("CRYPTO_ROUTER_URL") + "/api/create_eth_wallet"
	}
	if typeWallet == "trx" {
		url = os.Getenv("CRYPTO_ROUTER_URL") + "/api/create_trx_wallet"
	}
	if typeWallet == "btc" {
		url = os.Getenv("CRYPTO_ROUTER_URL") + "/api/create_btc_wallet"
	}

	data := []byte(`{
			"phrase": "` + phrase + `"
		}`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return Wallet{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return Wallet{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Wallet{}, err
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return Wallet{}, err
	}

	seedPhrase := generator.SeedPhraseGenerator()

	return Wallet{
		Address:    response.Address,
		SeedPhrase: seedPhrase,
	}, nil
}

func CreateAllWallets() (AllWallets, error) {
	seedPhrase := generator.SeedPhraseGenerator()

	trxWallet, err := createWallet(seedPhrase, "trx")
	if err != nil {
		return AllWallets{}, err
	}
	ethWallet, err := createWallet(seedPhrase, "eth")
	if err != nil {
		return AllWallets{}, err
	}
	btcWallet, err := createWallet(seedPhrase, "btc")
	if err != nil {
		return AllWallets{}, err
	}

	return AllWallets{
		Btc: Wallet{
			Address:    btcWallet.Address,
			SeedPhrase: seedPhrase,
		},
		Trx: Wallet{
			Address:    trxWallet.Address,
			SeedPhrase: seedPhrase,
		},
		Eth: Wallet{
			Address:    ethWallet.Address,
			SeedPhrase: seedPhrase,
		},
	}, nil
}
