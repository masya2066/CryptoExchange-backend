package main

import (
	"crypto-exchange/app/internal/config"
	"crypto-exchange/app/internal/db"
	"crypto-exchange/app/internal/routes"
	"crypto-exchange/app/pkg/logger"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag/example/basic/docs"
)

type server interface {
	Run() error
}

func main() {
	log := logger.GetLogger()

	err := config.Get()
	if err != nil {
		log.Fatalln("error config:", err)
	}

	db, err := db.New()
	if err != nil {
		log.Fatalln("error init database:", err)
	}
	log.Logger.Info("DB connected")

	var srv server
	srv, err = routes.New(gin.Default(), db, log)
	if err != nil {
		log.Fatalln("error init api:", err)
	}
	log.Logger.Info("API created")

	log.Logger.Info("Server starting on: ")
	log.Logger.Info("PORT: " + os.Getenv("APP_PORT"))
	log.Logger.Info("DB_HOST: " + os.Getenv("DB_HOST"))
	log.Logger.Info("DB_PORT: " + os.Getenv("DB_PORT"))
	log.Logger.Info("ACCESS_ALIVE: " + os.Getenv("ACCESS_ALIVE"))
	log.Logger.Info("REFRESH_ALIVE: " + os.Getenv("REFRESH_ALIVE"))

	log.Logger.Info("Starting server...")
	docs.SwaggerInfo.BasePath = "/api_v1"

	runErr := srv.Run()
	if runErr != nil {
		log.Fatalln("error run server:", runErr)
	}

}
