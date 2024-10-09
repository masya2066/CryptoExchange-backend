package db

import (
	"crypto-exchange/app/internal/models"
	"crypto-exchange/app/pkg/logger"
)

func (db *DB) AttachAction(logs models.ActionLogs) {
	log := logger.GetLogger()
	tx := db.Begin()
	if err := tx.Create(&logs).Error; err != nil {
		log.Errorf("error create action log: %v", err)
	}
	//ошибка и откат
	tx.Commit()
}

func (db *DB) GetActionLogs() []models.ActionLogs {
	var logs []models.ActionLogs
	db.Find(&logs)
	return logs
}
