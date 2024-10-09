package db

import (
	"crypto-exchange/app/internal/models"
	"crypto-exchange/app/pkg/broker"
	"crypto-exchange/app/pkg/jwt"
	"crypto-exchange/app/pkg/logger"
	"encoding/json"
	"time"
)

const (
	RedisAuthTokens = "auth_tokens"
)

func (db *DB) RedisSyncAuth(client *broker.Client) error {
	var tokens []models.RejectedToken
	if err := db.DB.Model(models.RejectedToken{}).Find(&tokens).Error; err != nil {
		return err
	}

	marshalled, err := json.Marshal(tokens)
	if err != nil {
		return err
	}

	if err := client.Client.Set(RedisAuthTokens, marshalled, 0).Err(); err != nil {
		return err
	}

	return nil
}

func (db *DB) RedisUpdateAuth(client *broker.Client) {
	// Не совсем понятен смысл. Ты забираешь всю таблицу отозванных токенов и проверяешь каждый в
	// цикле протух ли он. А зачем его проверять если судя по названию это отозванные токены?
	log := logger.GetLogger().Logger
	for {
		var tokens []models.RejectedToken
		if err := db.Model(&models.RejectedToken{}).Find(&tokens).Error; err != nil {
			log.Printf("Error retrieving tokens from database: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		var tokensArray []models.RejectedToken
		for _, token := range tokens {
			if jwt.CheckTokenExpiration(token.AccessToken) {
				if err := db.Model(&models.RejectedToken{}).Where("access_token = ?", token.AccessToken).Delete(&token).Error; err != nil {
					log.Errorf("Error deleting token from database: %v", err)
				}
			}
			tokensArray = append(tokensArray, token)
		}

		marshalled, err := json.Marshal(tokensArray)
		if err != nil {
			log.Errorf("Error marshalling tokens: %v", err)
			continue
		}

		if err := client.Client.Set(RedisAuthTokens, marshalled, 0).Err(); err != nil {
			log.Errorf("Error setting Redis key: %v", err)
			continue
		}

		time.Sleep(5 * time.Minute)
	}
}
