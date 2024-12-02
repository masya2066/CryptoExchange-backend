package routes

import (
	"crypto-exchange/app/internal/errorCodes"
	"crypto-exchange/app/internal/models"
	"crypto-exchange/app/internal/models/requestData"
	"crypto-exchange/app/pkg/jwt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (a *App) Exchange(c *gin.Context) {

	var body requestData.Exchange

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, "parse error", errorCodes.ParsingError))
		return
	}

	token := jwt.GetToken(c)
	userID := jwt.JwtParse(token).ID

	userIDUint := uint(userID.(float64))

	if err := a.db.UpdateExchange(models.Exchange{
		UserID:      userIDUint,
		BtcBalance:  body.BtcBalance,
		EthBalance:  body.EthBalance,
		TrxBalance:  body.TrxBalance,
		SoliBalance: body.SoliBalance,
	}); err != nil {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, err.Error(), errorCodes.ParsingError))
		return
	}

	c.JSON(http.StatusOK, models.ResponseMsg(true, "Success", 0))
}
