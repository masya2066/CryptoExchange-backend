package db

import (
	"crypto-exchange/app/internal/models"
	"crypto-exchange/app/pkg/crypto"
)

func (db *DB) BtcBalance(userID uint) (models.Balance, error) {
	user, err := db.WalletByUserID(userID)

	if err != nil {
		return models.Balance{}, err
	}

	res, err := crypto.RouterBtcBalance(user.BtcAddress)

	if err != nil {
		return models.Balance{}, err
	}

	return models.Balance{
		Address: res.Address,
		Balance: res.Balance,
	}, nil
}

func (db *DB) EthBalance(userID uint) (models.Balance, error) {
	user, err := db.WalletByUserID(userID)

	if err != nil {
		return models.Balance{}, err
	}

	res, err := crypto.RouterEthBalance(user.EthAddress)

	if err != nil {
		return models.Balance{}, err
	}

	return models.Balance{
		Address: res.Address,
		Balance: res.Balance,
	}, nil
}

func (db *DB) TrxBalance(userID uint) (models.Balance, error) {
	user, err := db.WalletByUserID(userID)

	if err != nil {
		return models.Balance{}, err
	}

	res, err := crypto.RouterTrxBalance(user.TrxAddress)

	if err != nil {
		return models.Balance{}, err
	}

	return models.Balance{
		Address: res.Address,
		Balance: res.Balance,
	}, nil
}

func (db *DB) SoliBalance(userID uint) (models.Balance, error) {
	res, err := db.ExchangeByUserID(userID)

	if err != nil {
		return models.Balance{}, err
	}

	return models.Balance{
		Address: "",
		Balance: res.SoliBalance,
	}, nil
}
