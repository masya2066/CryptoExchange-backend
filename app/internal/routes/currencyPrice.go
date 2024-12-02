package routes

import (
	"crypto-exchange/app/internal/errorCodes"
	"crypto-exchange/app/internal/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (a *App) GetCurrencyPrice(c *gin.Context) {
	var body []string

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, "parse error", errorCodes.ParsingError))
		return
	}

	if len(body) == 0 {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, "empty body", errorCodes.ParsingError))
		return
	}

	currencies, err := a.db.GetCurrencyPrice(body)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, err.Error(), errorCodes.ParsingError))
		return
	}

	c.JSON(http.StatusOK, currencies)
}
