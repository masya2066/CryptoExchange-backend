package db

import (
	"crypto-exchange/app/internal/models"
	"errors"
	"gorm.io/gorm"
)

func (db *DB) GetCurrencyPrice(currency []string) ([]models.CurrencyPrice, error) {
	var found []models.CurrencyPrice

	if err := db.Model(&models.CurrencyPrice{}).Where("currency IN (?)", currency).Find(&found).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	return found, nil
}

func (db *DB) UpdateCurrencyPrice(currency models.CurrencyPrice) error {
	if err := db.Model(&models.CurrencyPrice{}).Where("currency = ?", currency.Currency).Updates(currency).Error; err != nil {
		return err
	}

	return nil
}
