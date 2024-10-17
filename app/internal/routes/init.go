package routes

import (
	"crypto-exchange/app/internal/db"
	"crypto-exchange/app/internal/routes/middlewares"
	"crypto-exchange/app/pkg/broker"
	"crypto-exchange/app/pkg/logger"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type App struct {
	server *gin.Engine
	db     *db.DB
	logger logger.Logger
	broker *broker.Client
}

func New(server *gin.Engine, db *db.DB, logger logger.Logger) (*App, error) {
	client, errInit := broker.RedisInit()
	if errInit != nil {
		return nil, fmt.Errorf("broker was not connected: %v", errInit)
	}
	logger.Info("Redis connected!")

	app := &App{
		server: server,
		db:     db,
		logger: logger,
		broker: client,
	}

	if err := db.RedisSyncAuth(client); err != nil {
		return nil, fmt.Errorf("error while syncing redis: %v", err)
	}
	app.logger.Info("Redis synced!")

	go db.RedisUpdateAuth(client)
	app.logger.Info("Redis update started!")

	app.logger.Info("Env loaded")

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://127.0.0.1:3000", "http://localhost:3000", "http://127.0.0.1:3000/admin", "http://109.71.240.99", "http://127.0.0.1:3001", "http://localhost:3001"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Authorization", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Access-Control-Allow-Origin"}

	app.server.Use(cors.New(config))
	app.routes()

	go db.RedisUpdateCurrencies(client) // update currencies
	app.logger.Info("Redis update currencies started!")

	return app, nil
}

func (a *App) Run() error {
	return a.server.Run(":1102")
}

func (a *App) routes() {
	client := middlewares.Broker{Client: a.broker}
	a.server.GET("/swagger/*any",
		ginSwagger.WrapHandler(swaggerfiles.Handler,
			ginSwagger.DefaultModelsExpandDepth(1),
			ginSwagger.PersistAuthorization(true),
		),
	)
	route := a.server.Group("/api_v1")
	{
		auth := route.Group("/auth")
		{
			auth.POST("/login", a.Login)
			auth.POST("/refresh", a.Refresh)
			auth.POST("/register", a.Register)
			auth.POST("/activate/send", a.Send) // send email
			auth.POST("/activate", a.Activate)
			auth.POST("/logout", client.IsAuthorized, a.Logout)
			auth.POST("/recovery", a.Recovery) // send email
			auth.POST("/recovery/submit", a.RecoverySubmit)
			auth.POST("/register/check", a.CheckRegistrationCode)
			auth.POST("/recovery/check", a.CheckRecoveryCode)
		}
		user := route.Group("/user")
		{
			user.GET("/info", client.IsAuthorized, a.UserInfo)
		}
		crypto := route.Group("/crypto")
		{
			crypto.GET("/trx/usdt_balance", client.IsAuthorized, a.UsdtTrxBalance)
			crypto.GET("/currencies", a.Currencies)
		}
	}
}
