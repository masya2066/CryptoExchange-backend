package routes

import (
	"crypto-exchange/app/internal/errorCodes"
	"crypto-exchange/app/internal/models"
	"crypto-exchange/app/pkg/jwt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (a *App) BtcBalance(c *gin.Context) {
	token := jwt.GetToken(c)
	userID := jwt.JwtParse(token).ID

	userIDUint := uint(userID.(float64))

	res, err := a.db.BtcBalance(userIDUint)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, err.Error(), errorCodes.ParsingError))
		return
	}

	exchange, err := a.db.ExchangeByUserID(userIDUint)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, err.Error(), errorCodes.ParsingError))
		return
	}

	res.Balance = res.Balance + exchange.BtcBalance
	c.JSON(http.StatusOK, res)
}

func (a *App) EthBalance(c *gin.Context) {
	token := jwt.GetToken(c)
	userID := jwt.JwtParse(token).ID

	userIDUint := uint(userID.(float64))

	res, err := a.db.EthBalance(userIDUint)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, err.Error(), errorCodes.ParsingError))
		return
	}

	exchange, err := a.db.ExchangeByUserID(userIDUint)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, err.Error(), errorCodes.ParsingError))
		return
	}

	res.Balance = res.Balance + exchange.EthBalance
	c.JSON(http.StatusOK, res)
}

func (a *App) TrxBalance(c *gin.Context) {
	token := jwt.GetToken(c)
	userID := jwt.JwtParse(token).ID

	userIDUint := uint(userID.(float64))

	res, err := a.db.TrxBalance(userIDUint)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, err.Error(), errorCodes.ParsingError))
		return
	}

	exchange, err := a.db.ExchangeByUserID(userIDUint)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, err.Error(), errorCodes.ParsingError))
		return
	}

	res.Balance = res.Balance + exchange.TrxBalance
	c.JSON(http.StatusOK, res)
}

func (a *App) SoliBalance(c *gin.Context) {
	token := jwt.GetToken(c)
	userID := jwt.JwtParse(token).ID

	userIDUint := uint(userID.(float64))

	res, err := a.db.SoliBalance(userIDUint)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, err.Error(), errorCodes.ParsingError))
		return
	}

	c.JSON(http.StatusOK, res)
}
