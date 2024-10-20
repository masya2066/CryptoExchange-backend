package routes

import (
	"crypto-exchange/app/consumer"
	"crypto-exchange/app/internal/errorCodes"
	"crypto-exchange/app/internal/models"
	"crypto-exchange/app/internal/models/requestData"
	"crypto-exchange/app/internal/models/responses"
	"crypto-exchange/app/pkg/jwt"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"

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

func (a *App) Withdraw(c *gin.Context) {
	log := a.logger.Logger

	var body requestData.Withdraw

	token := jwt.GetToken(c)
	parsedToken := jwt.JwtParse(token)

	if err := c.ShouldBindJSON(&body); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, "parse error", errorCodes.ParsingError))
		return
	}

	if body.Coin == "" {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, "Incorrect value of coin", errorCodes.IncorrectCoin))
		return
	}

	if body.Address == "" {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, "Incorrect address", errorCodes.IncorrectAddress))
		return
	}

	if body.Amount == 0 {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, "Incorrect value of amount", errorCodes.IncorrectAmount))
		return
	}

	id, ok := parsedToken.ID.(float64)
	if !ok {
		fmt.Println(parsedToken.ID)
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, "parse error", errorCodes.ParsingError))
		return
	}

	withdrawID := rand.Int63n(100000000)

	withdraw := models.Withdraw{
		UserID:     uint(id),
		WithdrawID: withdrawID,
		Coin:       body.Coin,
		Address:    body.Address,
		Amount:     body.Amount,
		Status:     0,
	}

	if err := a.db.Withdraw(withdraw); err != nil {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, err.Error(), errorCodes.ParsingError))
		return
	}

	c.JSON(http.StatusOK, responses.Withdraw{
		WithdrawID: withdrawID,
		Status:     withdraw.Status,
		Success:    true,
	})

	email, ok := parsedToken.Email.(string)
	if !ok {
		log.Errorf("Parse error: %v", ok)
		return
	}

	go consumer.Send(email, "Finchain: New withdraw", "New withdraw \n\nID: "+
		strconv.FormatInt(withdrawID, 10)+
		"\nCoin: "+body.Coin+"\nAddress: "+
		body.Address+"\nAmount: "+
		strconv.FormatFloat(body.Amount, 'f', 2, 64)+
		"\n\nTransfer processing may take up to 48 hours\n\nThank you for using our service!",
		a.db.DB)
}
