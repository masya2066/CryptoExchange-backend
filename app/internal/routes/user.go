package routes

import (
	"crypto-exchange/app/internal/models/responses"
	"crypto-exchange/app/pkg/jwt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *App) UserInfo(c *gin.Context) {

	token := jwt.GetToken(c)
	parsedToken := jwt.JwtParse(token)

	user, err := a.db.UserInfoById(parsedToken.ID)
	if err != nil {
		c.JSON(
			http.StatusUnauthorized,
			"error user",
		)
	}

	c.JSON(http.StatusOK, responses.UserInfo{
		Email:    user.Email,
		Login:    user.Login,
		Active:   user.Active,
		AvatarId: user.AvatarId,
		Phone:    user.Phone,
		RefCode:  user.RefCode,
		Invite:   user.InviteCode,

		TrxAddress: user.TrxAddress,
		EthAddress: user.EthAddress,
		BtcAddress: user.BtcAddress,

		Created: user.Created,
		Updated: user.Updated,
	})
}
