package db

import (
	"crypto-exchange/app/internal/models"
	"crypto-exchange/app/pkg/logger"
	"crypto-exchange/app/pkg/utils"
	"fmt"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type DB struct {
	*gorm.DB
}

func New() (*DB, error) {
	cfg := config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&models.Config{},
		&models.User{},
		&models.RegToken{},
		&models.EmailChange{},
		&models.RejectedToken{},
		&models.ActionLogs{},
		&models.UserWallet{},
	); err != nil {
		return nil, err
	}

	if err := createConfig(db); err != nil {
		return nil, err
	}
	errUser := createDefaultUserIfNotExists(db)
	if errUser != nil {
		return nil, errUser
	}

	logger.GetLogger().Info("Database migrated successfully")

	return &DB{
		db,
	}, nil
}

func createConfig(db *gorm.DB) error {
	log := logger.GetLogger()

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	var count int64
	if err := tx.Model(&models.Config{}).Where("param = ?", "smtp_host").Count(&count); err.Error != nil {
		log.Error(err.Error)
		return tx.Rollback().Error
	}
	if count == 0 {
		if err := tx.Model(&models.Config{}).Create(&models.Config{
			Param:    "smtp_host",
			Value:    "",
			Activate: true,
			Updated:  TimeNow(),
		}); err.Error != nil {
			log.Error(err.Error)
			return tx.Rollback().Error
		}
	}
	if err := tx.Model(&models.Config{}).Where("param = ?", "smtp_port").Count(&count); err.Error != nil {
		log.Error(err.Error)
		return tx.Rollback().Error
	}
	if count == 0 {
		if err := tx.Model(&models.Config{}).Create(&models.Config{
			Param:    "smtp_port",
			Value:    "",
			Activate: true,
			Updated:  TimeNow(),
		}); err.Error != nil {
			log.Error(err.Error)
			return tx.Rollback().Error
		}
	}
	if err := tx.Model(&models.Config{}).Where("param = ?", "smtp_email").Count(&count); err.Error != nil {
		log.Error(err.Error)
		return tx.Rollback().Error
	}
	if count == 0 {
		if err := tx.Model(&models.Config{}).Create(&models.Config{
			Param:    "smtp_email",
			Value:    "",
			Activate: true,
			Updated:  TimeNow(),
		}); err.Error != nil {
			log.Error(err.Error)
			return tx.Rollback().Error
		}
	}
	if err := tx.Model(&models.Config{}).Where("param = ?", "smtp_pass").Count(&count); err.Error != nil {
		log.Error(err.Error)
		return tx.Rollback().Error
	}
	if count == 0 {
		if err := tx.Model(&models.Config{}).Create(&models.Config{
			Param:    "smtp_pass",
			Value:    "",
			Activate: true,
			Updated:  TimeNow(),
		}); err.Error != nil {
			log.Error(err.Error)
			return tx.Rollback().Error
		}
	}

	if err := tx.Commit(); err.Error != nil {
		log.Error(err.Error)
		return tx.Rollback().Error
	}

	return nil
}

func createDefaultUserIfNotExists(db *gorm.DB) error {
	log := logger.GetLogger()

	var count int64
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&models.User{}).Where("login = ? OR email = ?", "universal", "uni@example.com").Count(&count).Error; err != nil {
		log.Error(err)
		return tx.Rollback().Error
	}
	if count == 0 {
		defaultUser := models.User{
			ID:      uint(0),
			Login:   "universal",
			Email:   "uni@example.com",
			Phone:   "00000000000",
			Active:  true,
			Pass:    utils.Hash("admin"),
			Created: TimeNow(),
			Updated: TimeNow(),
		}
		if err := tx.Create(&defaultUser).Error; err != nil {
			log.Error(err)
			return tx.Rollback().Error
		}
	}

	if err := tx.Commit(); err.Error != nil {
		log.Error(err.Error)
		return tx.Rollback().Error
	}
	return nil
}

func TimeNow() string {
	return time.Now().UTC().Format(os.Getenv("DATE_FORMAT"))
}
