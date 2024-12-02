package db

import (
	"crypto-exchange/app/internal/models"
	"crypto-exchange/app/pkg/logger"
	"errors"
	"fmt"
	"gorm.io/gorm"
)

func (db *DB) ExchangeByUserID(userID uint) (res models.Exchange, error error) {
	var found models.Exchange

	if err := db.Model(&models.Exchange{}).Where("user_id = ?", userID).First(&found).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := db.CreateDefaultExchangeIfNotExists(userID); err != nil {
				return found, err
			}
		}
	}

	return found, nil
}

func (db *DB) UpdateExchange(exchange models.Exchange) error {

	var found models.Exchange

	if err := db.CreateDefaultExchangeIfNotExists(exchange.UserID); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			if err := db.Model(&models.Exchange{}).Where("user_id = ?", exchange.UserID).First(&found).Error; err != nil {
				return err
			}

			fmt.Println(found.SoliBalance)
			exchange.BtcBalance = exchange.BtcBalance + found.BtcBalance
			exchange.EthBalance = exchange.EthBalance + found.EthBalance
			exchange.TrxBalance = exchange.TrxBalance + found.TrxBalance
			exchange.SoliBalance = exchange.SoliBalance + found.SoliBalance
			if err := db.Model(&models.Exchange{}).Where("user_id = ?", exchange.UserID).Updates(exchange).Error; err != nil {
				return err
			}
		}
	}
	if err := db.Model(&models.Exchange{}).Where("user_id = ?", exchange.UserID).First(&found).Error; err != nil {
		return err
	}

	exchange.BtcBalance = exchange.BtcBalance + found.BtcBalance
	exchange.EthBalance = exchange.EthBalance + found.EthBalance
	exchange.TrxBalance = exchange.TrxBalance + found.TrxBalance
	exchange.SoliBalance = exchange.SoliBalance + found.SoliBalance
	if err := db.Model(&models.Exchange{}).Where("user_id = ?", exchange.UserID).Updates(exchange).Error; err != nil {
		return err
	}

	return nil
}

func (db *DB) CreateDefaultExchangeIfNotExists(userID uint) error {
	log := logger.GetLogger()

	if err := db.Model(&models.Exchange{}).Create(&models.Exchange{
		UserID:      userID,
		BtcBalance:  0,
		EthBalance:  0,
		TrxBalance:  0,
		SoliBalance: 0,
	}).Error; err != nil {
		log.Error(err)
		return err
	}

	return nil
}
