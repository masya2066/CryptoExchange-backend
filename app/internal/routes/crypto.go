package routes

import (
	"crypto-exchange/app/pkg/jwt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *App) UsdtTrxBalance(c *gin.Context) {
	token := jwt.GetToken(c)
	parsedToken := jwt.JwtParse(token)

	user, err := a.db.UserInfoById(parsedToken.ID)
	if err != nil {
		c.JSON(
			http.StatusUnauthorized,
			"error user",
		)
	}
	if user.TrxAddress == "" {
		c.JSON(
			http.StatusUnauthorized,
			"error user",
		)
	}

}

func (a *App) Currencies(c *gin.Context) {

	currencies, err := a.db.RedisGetCurrencies(a.broker)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
	}

	c.JSON(http.StatusOK, currencies)
}
