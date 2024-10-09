package middlewares

import (
	"crypto-exchange/app/internal/db"
	"crypto-exchange/app/internal/errorCodes"
	"crypto-exchange/app/internal/models"
	"crypto-exchange/app/internal/models/language"
	"crypto-exchange/app/pkg/broker"
	"crypto-exchange/app/pkg/jwt"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Broker struct {
	*broker.Client
}

func (br *Broker) IsAuthorized(c *gin.Context) {
	lang := language.LangValue(c)
	token := jwt.GetToken(c)
	if token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, models.ResponseMsg(false, language.Language(lang, "incorrect_email_or_password"), errorCodes.Unauthorized))
		return
	}
	if jwt.CheckTokenExpiration(token) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, models.ResponseMsg(false, language.Language(lang, "incorrect_email_or_password"), errorCodes.Unauthorized))
		return
	}

	array, err := br.RedisGetArray(db.RedisAuthTokens)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ResponseMsg(false, "db_error", errorCodes.DBError))
		return
	}

	for _, item := range array {
		var tok models.RejectedToken
		er, errMarshal := json.Marshal(item)
		if errMarshal != nil {
			continue
		}
		errUnmarshal := json.Unmarshal(er, &tok)
		if errUnmarshal != nil {
			continue
		}
		if tok.AccessToken == token {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ResponseMsg(false, language.Language(lang, "incorrect_email_or_password"), errorCodes.Unauthorized))
			return
		}
	}
}
