package db

import (
	"crypto-exchange/app/internal/models"
	"fmt"
)

func (db *DB) Withdraw(withdraw models.Withdraw) error {

	withdraw.Created = TimeNow()
	withdraw.Updated = TimeNow()

	if err := db.Model(&models.Withdraw{}).Create(withdraw).Error; err != nil {
		return fmt.Errorf("error create wallet: %v", err)
	}
	return nil
}
