package db

import (
	"crypto-exchange/app/internal/models"
	"fmt"
)

func (db *DB) CreateAllWallets(wallet models.UserWallet) error {
	if err := db.Model(&models.UserWallet{}).Create(wallet).Error; err != nil {
		return fmt.Errorf("error create wallet: %v", err)
	}

	return nil
}
