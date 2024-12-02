package db

import (
	"crypto-exchange/app/internal/models"
	"errors"
	"fmt"
	"gorm.io/gorm"
)

func (db *DB) CreateAllWallets(wallet models.UserWallet) error {
	if err := db.Model(&models.UserWallet{}).Create(wallet).Error; err != nil {
		return fmt.Errorf("error create wallet: %v", err)
	}

	return nil
}

func (db *DB) WalletByUserID(userID uint) (models.UserWallet, error) {
	var found models.UserWallet

	if err := db.Model(&models.UserWallet{}).Where("user_id = ?", userID).First(&found).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return found, err
		}
	}
	return found, nil
}
