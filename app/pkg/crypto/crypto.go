package crypto

import (
	"crypto-exchange/app/internal/models/responses"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

func GetBtcInfo() responses.Currency {
	res, err := http.Get("https://api.coingecko.com/api/v3/coins/bitcoin")
	if err != nil {
		log.Println("Error getting btc info: ", err)
		return responses.Currency{}
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(res.Body)

	if res.StatusCode != 200 {
		log.Println("Error getting info: ", err)
		return responses.Currency{}
	}

	var btcInfo responses.Currency

	if err := json.NewDecoder(res.Body).Decode(&btcInfo); err != nil {
		log.Println("Error decoding btc info: ", err)
		return responses.Currency{}
	}

	return btcInfo
}

func GetEthInfo() responses.Currency {
	res, err := http.Get("https://api.coingecko.com/api/v3/coins/ethereum")
	if err != nil {
		log.Println("Error getting eth info: ", err)
		return responses.Currency{}
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(res.Body)

	if res.StatusCode != 200 {
		log.Println("Error getting info: ", err)
		return responses.Currency{}
	}

	var ethInfo responses.Currency

	if err := json.NewDecoder(res.Body).Decode(&ethInfo); err != nil {
		log.Println("Error decoding eth info: ", err)
		return responses.Currency{}
	}

	return ethInfo
}

func GetUsdtInfo() responses.Currency {
	res, err := http.Get("https://api.coingecko.com/api/v3/coins/tether")
	if err != nil {
		log.Println("Error getting usdt info: ", err)
		return responses.Currency{}
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(res.Body)

	if res.StatusCode != 200 {
		log.Println("Error getting info: ", err)
		return responses.Currency{}
	}

	var usdtInfo responses.Currency

	if err := json.NewDecoder(res.Body).Decode(&usdtInfo); err != nil {
		log.Println("Error decoding usdt info: ", err)
		return responses.Currency{}
	}

	return usdtInfo
}

func GetSolanaInfo() responses.Currency {
	res, err := http.Get("https://api.coingecko.com/api/v3/coins/solana")
	if err != nil {
		log.Println("Error getting solana info: ", err)
		return responses.Currency{}
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(res.Body)

	if res.StatusCode != 200 {
		log.Println("Error getting info: ", err)
		return responses.Currency{}
	}

	var solanaInfo responses.Currency

	if err := json.NewDecoder(res.Body).Decode(&solanaInfo); err != nil {
		log.Println("Error decoding solana info: ", err)
		return responses.Currency{}
	}

	return solanaInfo
}

func GetBnbInfo() responses.Currency {
	res, err := http.Get("https://api.coingecko.com/api/v3/coins/binancecoin")
	if err != nil {
		log.Println("Error getting bnb info: ", err)
		return responses.Currency{}
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(res.Body)

	if res.StatusCode != 200 {
		log.Println("Error getting info: ", err)
		return responses.Currency{}
	}

	var bnbInfo responses.Currency

	if err := json.NewDecoder(res.Body).Decode(&bnbInfo); err != nil {
		log.Println("Error decoding bnb info: ", err)
		return responses.Currency{}
	}

	return bnbInfo
}

func GetRippleInfo() responses.Currency {
	res, err := http.Get("https://api.coingecko.com/api/v3/coins/ripple")
	if err != nil {
		log.Println("Error getting ripple info: ", err)
		return responses.Currency{}
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(res.Body)

	if res.StatusCode != 200 {
		log.Println("Error getting info: ", err)
		return responses.Currency{}
	}

	var rippleInfo responses.Currency

	if err := json.NewDecoder(res.Body).Decode(&rippleInfo); err != nil {
		log.Println("Error decoding ripple info: ", err)
		return responses.Currency{}
	}

	return rippleInfo
}

func GetCardanoInfo() responses.Currency {
	res, err := http.Get("https://api.coingecko.com/api/v3/coins/cardano")
	if err != nil {
		log.Println("Error getting cardano info: ", err)
		return responses.Currency{}
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(res.Body)

	if res.StatusCode != 200 {
		log.Println("Error getting info: ", err)
		return responses.Currency{}
	}

	var cardanoInfo responses.Currency

	if err := json.NewDecoder(res.Body).Decode(&cardanoInfo); err != nil {
		log.Println("Error decoding cardano info: ", err)
		return responses.Currency{}
	}

	return cardanoInfo
}

func GetAvalancheInfo() responses.Currency {
	res, err := http.Get("https://api.coingecko.com/api/v3/coins/avalanche-2")
	if err != nil {
		log.Println("Error getting avalanche info: ", err)
		return responses.Currency{}
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(res.Body)

	if res.StatusCode != 200 {
		log.Println("Error getting info: ", err)
		return responses.Currency{}
	}

	var avalancheInfo responses.Currency

	if err := json.NewDecoder(res.Body).Decode(&avalancheInfo); err != nil {
		log.Println("Error decoding avalanche info: ", err)
		return responses.Currency{}
	}

	return avalancheInfo
}

func GetAllCurrencies() []responses.Currency {
	var currencies []responses.Currency
	currencyFuncs := []func() responses.Currency{
		GetBtcInfo,
		GetEthInfo,
		GetUsdtInfo,
		GetSolanaInfo,
		GetBnbInfo,
		GetRippleInfo,
		GetCardanoInfo,
		GetAvalancheInfo,
	}

	// Канал для сбора данных
	currencyChan := make(chan responses.Currency, len(currencyFuncs))

	// Вызов каждой функции с задержкой в 10 секунд
	for _, getCurrency := range currencyFuncs {
		go func(fn func() responses.Currency) {
			// Задержка в 10 секунд
			currencyChan <- fn() // Отправка результата в канал
		}(getCurrency)
		time.Sleep(30 * time.Second)
	}

	// Чтение из канала и добавление в массив
	for i := 0; i < len(currencyFuncs); i++ {
		currencies = append(currencies, <-currencyChan)
	}

	return currencies
}
