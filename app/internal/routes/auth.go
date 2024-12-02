package routes

import (
	"crypto-exchange/app/consumer"
	"crypto-exchange/app/internal/db"
	"crypto-exchange/app/internal/errorCodes"
	"crypto-exchange/app/internal/models"
	"crypto-exchange/app/internal/models/language"
	"crypto-exchange/app/internal/models/requestData"
	"fmt"

	"crypto-exchange/app/internal/models/responses"
	"crypto-exchange/app/pkg/client"
	"crypto-exchange/app/pkg/crypto"
	"crypto-exchange/app/pkg/jwt"
	"crypto-exchange/app/pkg/mail"
	"crypto-exchange/app/pkg/utils"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// @Summary Login into account
// @Description Endpoint to login into account
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body requestData.Login true "request requestData"
// @Success 200 object responses.AuthResponse
// @Failure 400 object models.ErrorResponse
// @Failure 401 object models.ErrorResponse
// @Router /auth/login [post]
func (a *App) Login(c *gin.Context) {
	var user requestData.Login

	lang := language.LangValue(c)
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			models.ResponseMsg(false, language.Language(lang, "parse_error"), errorCodes.ParsingError),
		)
		return
	}

	userInfo, err := a.db.UserInfo(user.Login, user.Login)
	if err != nil {
		c.JSON(
			http.StatusUnauthorized,
			models.ResponseMsg(false, language.Language(lang, "incorrect_email_or_password"), errorCodes.Unauthorized),
		)
		return
	}

	if !userInfo.Active {
		c.JSON(
			http.StatusUnauthorized,
			models.ResponseMsg(
				false,
				language.Language(lang, "user")+" "+user.Login+" "+language.Language(lang, "is_not_active"),
				errorCodes.UserIsNotActive,
			),
		)
		return
	}

	userPass := utils.Hash(user.Password)
	if userPass != userInfo.Pass {
		c.JSON(
			http.StatusUnauthorized,
			models.ResponseMsg(false, language.Language(lang, "incorrect_email_or_password"), errorCodes.Unauthorized),
		)
		a.db.AttachAction(models.ActionLogs{
			Action:  "Try to login with incorrect password",
			Login:   user.Login,
			Ip:      client.GetIP(c),
			Created: db.TimeNow(),
		})
		return
	}

	authResponse := responses.AuthResponse{
		User: responses.UserInfo{
			Login:   userInfo.Login,
			Email:   userInfo.Email,
			Phone:   userInfo.Phone,
			RefCode: userInfo.RefCode,
			Created: userInfo.Created,
			Updated: userInfo.Updated,
		},
	}

	accessToken, refreshToken, err := jwt.GenerateJWT(jwt.TokenData{
		ID:         int(userInfo.ID),
		Authorized: true,
		Email:      userInfo.Email,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	alive, err := jwt.CheckTokenRemaining(accessToken)
	if err != nil {
		a.logger.Error(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	authResponse.AccessToken = accessToken
	authResponse.RefreshToken = refreshToken
	authResponse.Alive = alive

	c.JSON(http.StatusOK, authResponse)

	a.db.AttachAction(models.ActionLogs{
		Action:  "Login in to account",
		Login:   user.Login,
		Ip:      client.GetIP(c),
		Created: db.TimeNow(),
	})
}

// @Summary Register account
// @Description Endpoint to register a new user account
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body requestData.Register true "request requestData"
// @Success 200 object responses.RegisterResponse
// @Failure 400 object models.ErrorResponse
// @Failure 403 object models.ErrorResponse
// @Failure 500
// @Router /auth/register [post]
func (a *App) Register(c *gin.Context) {
	lang := language.LangValue(c)
	var user requestData.Register
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, language.Language(lang, "parse_error"), errorCodes.ParsingError))
		return
	}
	if !mail.MailValidator(user.Email) {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, language.Language(lang, "incorrect_email"), errorCodes.IncorrectEmail))
		return
	}
	if user.Login == "" {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, language.Language(lang, "login_empty"), errorCodes.LoginCanBeEmpty))
		return
	}

	// To do Validation of password
	if user.Pass == "" {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, language.Language(lang, "password_empty"), errorCodes.PasswordCantBeEmpty))
		return
	}
	if ok := utils.ValidateLogin(user.Login); !ok {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, language.Language(lang, "login_can_be_include_letters_digits"), errorCodes.IncorrectLogin))
		return
	}

	var ifExist models.User
	var foundLogin models.User

	user.Login = strings.ToLower(user.Login)
	user.Email = strings.ToLower(user.Email)
	if err := a.db.Where("email = ?", user.Email).First(&ifExist); err.Error != nil {
		a.logger.Errorf("error get user: %v", err.Error)
	}
	if err := a.db.Model(&models.User{}).Where("login = ?", user.Login).First(&foundLogin); err.Error != nil {
		a.logger.Errorf("error get user: %v", err.Error)
	}

	if ifExist.Email != "" {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, language.Language(lang, "user_already_exist"), errorCodes.UserAlreadyExist))
		return
	}
	if foundLogin.Login != "" {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, language.Language(lang, "login_already_exist"), errorCodes.LoginAlreadyExist))
		return
	}

	code, err := utils.GenerateReferralCode()
	if err != nil {
		a.logger.Error(err)
		c.JSON(http.StatusInternalServerError, models.ResponseMsg(false, language.Language(lang, "internal_error"), errorCodes.DBError))
		return
	}

	timeNow := db.TimeNow()

	completeUser := models.User{
		Login:      strings.ToLower(user.Login),
		Email:      strings.ToLower(user.Email),
		RefCode:    code,
		InviteCode: user.InviteCode,
		Pass:       utils.Hash(user.Pass),
		Active:     true,
		Created:    timeNow,
		Updated:    timeNow,
	}

	errCreate := a.db.CreateUser(completeUser)
	if errCreate != nil {
		c.JSON(http.StatusInternalServerError, models.ResponseMsg(false, language.Language(lang, "db_error"), errorCodes.DBError))
		return
	}

	c.JSON(http.StatusOK, responses.RegisterResponse{
		Error: false,
		User: responses.UserInfo{
			Login:   strings.ToLower(completeUser.Login),
			Email:   strings.ToLower(completeUser.Email),
			Phone:   strings.ToLower(completeUser.Phone),
			RefCode: completeUser.RefCode,
			Created: completeUser.Created,
			Updated: completeUser.Updated,
		},
	})

	allWallets, err := crypto.CreateAllWallets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ResponseMsg(false, language.Language(lang, "db_error"), errorCodes.DBError))
		return
	}

	var cuser models.User

	if err := a.db.Model(&models.User{}).Where("login = ?", strings.ToLower(completeUser.Login)).First(&cuser).Error; err != nil {
		fmt.Println("Fatal error to create user, when user tried to find:", err)
		return
	}

	if createWalletsErr := a.db.CreateAllWallets(models.UserWallet{
		UserID:     cuser.ID,
		BtcAddress: allWallets.Btc.Address,
		EthAddress: allWallets.Eth.Address,
		TrxAddress: allWallets.Trx.Address,
		SeedPhrase: allWallets.Btc.SeedPhrase,
		Created:    timeNow,
		Updated:    timeNow,
	}); createWalletsErr != nil {
		c.JSON(http.StatusInternalServerError, models.ResponseMsg(false, language.Language(lang, "db_error"), errorCodes.DBError))
		return
	}

	if errCreateExchangeModel := a.db.CreateDefaultExchangeIfNotExists(cuser.ID); errCreateExchangeModel != nil {
		a.logger.Logger.Errorln("Error create default exchange model:", errCreateExchangeModel)
		return
	}
}

// @Summary Send register email
// @Description Endpoint to send register email to submit registration
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body requestData.Send true "request requestData"
// @Success 200 object models.SuccessResponse
// @Failure 400 object models.ErrorResponse
// @Failure 403 object models.ErrorResponse
// @Failure 404 object models.ErrorResponse
// @Failure 500
// @Router /auth/activate/send [post]
func (a *App) Send(c *gin.Context) {
	lang := language.LangValue(c)
	var user requestData.Send

	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, language.Language(lang, "parse_error"), errorCodes.ParsingError))
		return
	}

	if user.Email == "" {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, language.Language(lang, "user_not_registered"), errorCodes.UserNotFound))
		return
	}
	var foundUser models.User
	a.db.Model(&models.User{}).Where("email = ?", user.Email).First(&foundUser)
	if foundUser.Email == "" {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, language.Language(lang, "user_not_found"), errorCodes.UserNotFound))
		return
	}

	var checkUser models.RegToken

	if err := a.db.Model(&models.RegToken{}).Where("user_id = ? AND type = ?", foundUser.ID, 0).First(&checkUser); err.Error != nil {
		a.logger.Errorf("error get user: %v", err.Error)
	}

	if checkUser.Created > time.Now().UTC().Add(-2*time.Minute).Format(os.Getenv("DATE_FORMAT")) {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, language.Language(lang, "email_already_sent")+user.Email, errorCodes.EmailAlreadySent))
		return
	} else {
		del := a.db.Model(&models.RegToken{}).Delete("user_id = ?", checkUser.UserId)
		if del.Error != nil {
			c.JSON(http.StatusInternalServerError, models.ResponseMsg(false, language.Language(lang, "db_error"), errorCodes.DBError))
			return
		}
	}
	code, errGen := utils.CodeGen()
	if errGen != nil {
		a.logger.Errorf("error generate code: %v", errGen)
		c.JSON(http.StatusInternalServerError, models.ResponseMsg(false, language.Language(lang, "error"), errorCodes.ServerError))
		return
	}

	c.JSON(http.StatusOK, models.ResponseMsg(true, language.Language(lang, "email_sent")+foundUser.Email, 0))

	go func(code string) {
		create := a.db.Model(&models.RegToken{}).Create(models.RegToken{
			UserId:  int(foundUser.ID),
			Type:    0,
			Code:    code,
			Created: db.TimeNow(),
		})
		if create.Error != nil {
			a.logger.Error("Create mail in table error: " + create.Error.Error())
			return
		}

		if !consumer.SendRegisterMail(foundUser.Email, lang, foundUser, code, a.db.DB) {
			a.logger.Error("Email send error to address: " + user.Email)
			return
		}

		a.logger.Info("Email sent to address: " + user.Email)
	}(code)
}

// @Summary Activate account
// @Description Endpoint to activate account by registration code
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body requestData.Activate true "request requestData"
// @Success 200 object models.SuccessResponse
// @Failure 400 object models.ErrorResponse
// @Failure 403 object models.ErrorResponse
// @Failure 404 object models.ErrorResponse
// @Failure 500
// @Router /auth/activate [post]
func (a *App) Activate(c *gin.Context) {
	lang := language.LangValue(c)
	var user requestData.Activate
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, language.Language(lang, "parse_error"), errorCodes.ParsingError))
		return
	}

	if user.Code == "" {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, language.Language(lang, "incorrect_activation_code"), errorCodes.IncorrectActivationCode))
		return
	}
	if user.Password == "" {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, language.Language(lang, "password_null"), errorCodes.NameOfSurnameIncorrect))
		return
	}
	digit, symbols := utils.PasswordChecker(user.Password)
	if !digit || !symbols {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, language.Language(lang, "password_should_be_include_digits"), errorCodes.PasswordShouldByIncludeSymbols))
		return
	}

	tx := a.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	activate, err := a.db.CheckActivationCode(models.RegToken{
		Code: user.Code,
	})
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, language.Language(lang, "incorrect_activation_code"), errorCodes.IncorrectActivationCode))
		return
	}

	var foundUsers models.User
	// Check if user exist
	tx.Model(models.User{}).Where("id = ?", uint(activate.UserId)).First(&foundUsers)
	if foundUsers.ID <= 0 {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, language.Language(lang, "user_not_found"), errorCodes.UserNotFound))
		return
	}

	// Delete activation code
	if deleteCode := tx.Model(&models.RegToken{}).Where("code = ?", activate.Code).Delete(activate); deleteCode.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, models.ResponseMsg(false, language.Language(lang, "db_error"), errorCodes.DBError))
		return
	}

	// Create (UPDATE) new password
	if updatePass := tx.Model(&models.User{}).Where("id = ?", foundUsers.ID).Update("pass", utils.Hash(user.Password)); updatePass.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, models.ResponseMsg(false, language.Language(lang, "db_error"), errorCodes.DBError))
		return
	}
	// Activate user
	if update := tx.Model(&models.User{}).Where("id = ?", activate.UserId).Updates(models.User{
		Active:  true,
		Updated: db.TimeNow(),
	}); update.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, models.ResponseMsg(false, language.Language(lang, "db_error"), errorCodes.DBError))
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ResponseMsg(false, language.Language(lang, "db_error"), errorCodes.DBError))
		return
	}

	c.JSON(http.StatusOK, models.ResponseMsg(true, language.Language(lang, "account")+foundUsers.Email+" "+language.Language(lang, "success_activate"), 0))

	a.db.AttachAction(models.ActionLogs{
		Action:  "Activate account by registration code",
		Login:   foundUsers.Login,
		Ip:      client.GetIP(c),
		Created: db.TimeNow(),
	})
}

// @Summary Get new access token
// @Description Endpoint to get a new access token by refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body requestData.Refresh true "request requestData"
// @Success 200 object responses.Refresh
// @Failure 400 object models.ErrorResponse
// @Failure 401 object models.ErrorResponse
// @Failure 500
// @Router /auth/refresh [post]
func (a *App) Refresh(c *gin.Context) {
	lang := language.LangValue(c)
	token := jwt.GetToken(c)
	var dataToken requestData.Refresh
	if err := c.ShouldBindJSON(&dataToken); err != nil {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, language.Language(lang, "parse_error"), errorCodes.ParsingError))
		return
	}

	//Мне не нравится история что ты будешь вечно весь массив тянуть
	array, errGet := a.broker.RedisGetArray(db.RedisAuthTokens)
	if errGet != nil {
		return
	}
	var exist bool
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
		if tok.RefreshToken == dataToken.Token {
			exist = true
			break
		}
	}
	if exist {
		c.JSON(http.StatusUnauthorized, models.ResponseMsg(false, language.Language(lang, "incorrect_email_or_password"), errorCodes.Unauthorized))
		return
	}
	parsedRefresh := jwt.JwtParse(dataToken.Token)
	if jwt.CheckTokenExpiration(dataToken.Token) {
		c.JSON(http.StatusUnauthorized, models.ResponseMsg(false, language.Language(lang, "incorrect_email_or_password"), errorCodes.Unauthorized))
		return
	}
	if parsedRefresh.Email == nil {
		c.JSON(http.StatusUnauthorized, models.ResponseMsg(false, language.Language(lang, "incorrect_email_or_password"), errorCodes.Unauthorized))
		return
	}
	if parsedRefresh.Email != jwt.JwtParse(token).Email {
		c.JSON(http.StatusUnauthorized, models.ResponseMsg(false, language.Language(lang, "incorrect_email_or_password"), errorCodes.Unauthorized))
		return
	}

	var user models.User
	a.db.Model(models.User{}).Where("email = ?", jwt.JwtParse(token).Email).First(&user)
	if user.ID == 0 {
		c.JSON(http.StatusUnauthorized, models.ResponseMsg(false, language.Language(lang, "incorrect_email_or_password"), errorCodes.Unauthorized))
		return
	}
	access, refresh, err := jwt.GenerateJWT(jwt.TokenData{
		Authorized: true,
		Email:      user.Email,
	})
	if err != nil {
		a.logger.Error(err)
	}

	rejectedTokens := models.RejectedToken{
		AccessToken:  token,
		RefreshToken: dataToken.Token,
	}

	err = a.db.Model(models.RejectedToken{}).Create(rejectedTokens).Error
	if err != nil {
		a.logger.Error(err)
		c.JSON(http.StatusInternalServerError, models.ResponseMsg(false, language.Language(lang, "db_error"), errorCodes.DBError))
		return
	}

	if err := a.broker.RedisAddToArray(db.RedisAuthTokens, rejectedTokens); err != nil {
		a.logger.Error(err)
		c.JSON(http.StatusInternalServerError, models.ResponseMsg(false, language.Language(lang, "db_error"), errorCodes.DBError))
		return
	}

	c.JSON(http.StatusOK, responses.Refresh{
		AccessToken:  access,
		RefreshToken: refresh,
		UserId:       int(user.ID),
	})
}

// @Summary Logout from account
// @Description Endpoint to logout from account
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 object models.SuccessResponse
// @Failure 401 object models.ErrorResponse
// @Failure 403 object models.ErrorResponse
// @Failure 500
// @Router /auth/logout [post]
func (a *App) Logout(c *gin.Context) {
	lang := language.LangValue(c)
	token := jwt.GetAuth(c)
	if token == "" {
		c.JSON(http.StatusOK, models.ResponseMsg(true, language.Language(lang, "successfuly_logout"), 0))
		return
	}
	if jwt.JwtParse(token).Email == nil {
		c.JSON(http.StatusOK, models.ResponseMsg(true, language.Language(lang, "successfuly_logout"), 0))
		return
	}

	rejected := models.RejectedToken{
		AccessToken:  token,
		RefreshToken: "",
	}

	if err := a.broker.RedisAddToArray(db.RedisAuthTokens, rejected); err != nil {
		a.logger.Error(err)
		c.JSON(http.StatusInternalServerError, models.ResponseMsg(false, language.Language(lang, "db_error"), errorCodes.DBError))
		return
	}

	if err := a.db.Model(models.RejectedToken{}).Create(rejected).Error; err != nil {
		a.logger.Error(err)
		c.JSON(http.StatusInternalServerError, models.ResponseMsg(false, language.Language(lang, "db_error"), errorCodes.DBError))
		return
	}

	c.JSON(http.StatusOK, models.ResponseMsg(true, language.Language(lang, "successfuly_logout"), 0))
}

// @Summary Check registration code if exist
// @Description Endpoint to check registration code if exist
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body requestData.RegistrationCode true "request requestData"
// @Success 200 object responses.CodeCheck
// @Failure 400 object models.ErrorResponse
// @Failure 401 object models.ErrorResponse
// @Failure 403 object models.ErrorResponse
// @Failure 404 object models.ErrorResponse
// @Failure 500
// @Router /auth/register/check [post]
func (a *App) CheckRegistrationCode(c *gin.Context) {
	lang := language.LangValue(c)
	var code requestData.RegistrationCode

	if err := c.ShouldBindJSON(&code); err != nil {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, language.Language(lang, "parse_error"), errorCodes.ParsingError))
		return
	}

	var foundCodes models.RegToken
	if err := a.db.Model(models.RegToken{}).Where("code = ? AND type = ?", code.Code, 0).First(&foundCodes).Error; err != nil {
		a.logger.Info(err)
		if foundCodes.Code == "" {
			c.JSON(http.StatusBadRequest, models.ResponseMsg(false, language.Language(lang, "register_code_not_found"), errorCodes.NotFoundRegistrationCode))
			return
		}
	}

	var user models.User
	if err := a.db.Model(models.User{}).Where("id = ?", uint(foundCodes.UserId)).First(&user); err.Error != nil {
		a.logger.Error(err)
		c.JSON(http.StatusInternalServerError, models.ResponseMsg(false, language.Language(lang, "db_error"), errorCodes.DBError))
		return
	}
	if user.Login == "" {
		c.JSON(http.StatusNotFound, models.ResponseMsg(false, language.Language(lang, "register_code_not_found"), errorCodes.NotFoundRegistrationCode))
		return
	}

	if foundCodes.Created < time.Now().UTC().Add(-24*time.Hour).Format(os.Getenv("DATE_FORMAT")) {
		if err := a.db.Model(&models.RegToken{}).Where("code = ?", foundCodes).Delete(foundCodes); err.Error != nil {
			a.logger.Error(err)
			c.JSON(http.StatusInternalServerError, models.ResponseMsg(false, language.Language(lang, "db_error"), errorCodes.DBError))
			return
		}
		c.JSON(http.StatusUnauthorized, models.ResponseMsg(false, language.Language(lang, "activation_code_expired"), errorCodes.ActivationCodeExpired))
		return
	}

	c.JSON(http.StatusOK, responses.CodeCheck{
		ID:    user.ID,
		Email: user.Email,
	})
}

// @Summary Check recovery code if exist
// @Description Endpoint to check recovery code if exist
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body requestData.CheckRecoveryCode true "request requestData"
// @Success 200 object responses.CheckRecoveryCode
// @Failure 400 object models.ErrorResponse
// @Failure 401 object models.ErrorResponse
// @Failure 403 object models.ErrorResponse
// @Failure 404 object models.ErrorResponse
// @Failure 500
// @Router /auth/recovery/check [post]
func (a *App) CheckRecoveryCode(c *gin.Context) {
	lang := language.LangValue(c)
	var code requestData.CheckRecoveryCode

	if err := c.ShouldBindJSON(&code); err != nil {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, language.Language(lang, "parse_error"), errorCodes.ParsingError))
		return
	}

	var foundCodes models.RegToken
	if err := a.db.Model(models.RegToken{}).Where("code = ? AND type = ?", code.Code, 1).First(&foundCodes).Error; err != nil {
		a.logger.Error(err)
	}
	if foundCodes.Code == "" {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, language.Language(lang, "register_code_not_found"), errorCodes.NotFoundRegistrationCode))
		return
	}

	var user models.User
	if err := a.db.Model(models.User{}).Where("id = ?", uint(foundCodes.UserId)).First(&user); err.Error != nil {
		a.logger.Error(err)
	}
	if user.Login == "" {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, language.Language(lang, "register_code_not_found"), errorCodes.NotFoundRegistrationCode))
		return
	}

	if foundCodes.Created < time.Now().UTC().Add(-24*time.Hour).Format(os.Getenv("DATE_FORMAT")) {
		if err := a.db.Model(&models.RegToken{}).Where("code = ?", foundCodes).Delete(foundCodes); err.Error != nil {
			a.logger.Error(err)
			c.JSON(http.StatusInternalServerError, models.ResponseMsg(false, language.Language(lang, "db_error"), errorCodes.DBError))
			return
		}
		c.JSON(http.StatusUnauthorized, models.ResponseMsg(false, language.Language(lang, "activation_code_expired"), errorCodes.ActivationCodeExpired))
		return
	}

	c.JSON(http.StatusOK, responses.CheckRecoveryCode{
		ID:    user.ID,
		Email: user.Email,
	})
}

// @Summary Recovery user account
// @Description Endpoint to recovery user account by email
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body requestData.SendMail true "request requestData"
// @Success 200 object models.SuccessResponse
// @Failure 400 object models.ErrorResponse
// @Failure 403 object models.ErrorResponse
// @Failure 500
// @Router /auth/recovery [post]
func (a *App) Recovery(c *gin.Context) {
	lang := language.LangValue(c)
	var user requestData.SendMail

	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, language.Language(lang, "parse_error"), errorCodes.ParsingError))
		return
	}

	if user.Email == "" {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, language.Language(lang, "email_empty"), errorCodes.EmptyEmail))
		return
	}

	// Check if user exist
	var foundUser models.User
	a.db.Model(&models.User{}).Where("email = ?", user.Email).First(&foundUser)
	if foundUser.Email == "" {
		c.JSON(http.StatusOK, models.ResponseMsg(true, language.Language(lang, "email_sent")+user.Email, 0))
		return
	}

	// Check if user already sent email
	var checkUser models.RegToken
	a.db.Model(&models.RegToken{}).Where("user_id = ? AND type = ?", foundUser.ID, 1).First(&checkUser)
	if checkUser.Created < time.Now().UTC().Add(-2*time.Minute).Format(os.Getenv("DATE_FORMAT")) {
		a.db.Model(&models.RegToken{}).Where("user_id = ?", checkUser.UserId).Delete(models.RegToken{UserId: checkUser.UserId, Type: 0})
	} else {
		c.JSON(http.StatusForbidden, models.ResponseMsg(false, language.Language(lang, "email_already_sent")+user.Email, errorCodes.EmailAlreadySent))
		return
	}

	code, errGen := utils.CodeGen()
	if errGen != nil {
		a.logger.Error(errGen)
		c.JSON(http.StatusInternalServerError, models.ResponseMsg(false, language.Language(lang, "error"), errorCodes.ServerError))
		return
	}

	// Create new code in database
	if err := a.db.Model(&models.RegToken{}).Create(models.RegToken{
		UserId:  int(foundUser.ID),
		Type:    1,
		Code:    code,
		Created: db.TimeNow(),
	}); err.Error != nil {
		a.logger.Error(err)
		c.JSON(http.StatusInternalServerError, models.ResponseMsg(false, language.Language(lang, "db_error"), errorCodes.DBError))
		return
	}

	c.JSON(http.StatusOK, models.ResponseMsg(true, "Email sent to "+foundUser.Email, 0))

	a.db.AttachAction(models.ActionLogs{
		Action:  "Recovery password",
		Login:   foundUser.Login,
		Ip:      client.GetIP(c),
		Created: db.TimeNow(),
	})

	// Send email
	go func(code string) {
		consumer.Send(
			foundUser.Email,
			"Admin Panel password recovery!", "Your link for continue is:  "+os.Getenv("DOMAIN")+"/recovery/submit/"+code+
				"\n\nEmail: "+user.Email+
				"\nLogin: "+foundUser.Login+
				"\nCreated: "+foundUser.Created,
			a.db.DB)
	}(code)
}

// @Summary Recovery submit
// @Description Endpoint to submit recovery account and create a new password for account
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body requestData.RecoverySubmit true "request requestData"
// @Success 200 object models.SuccessResponse
// @Failure 400 object models.ErrorResponse
// @Failure 401 object models.ErrorResponse
// @Failure 404 object models.ErrorResponse
// @Failure 500
// @Router /auth/recovery/submit [post]
func (a *App) RecoverySubmit(c *gin.Context) {
	lang := language.LangValue(c)
	var recoveryBody requestData.RecoverySubmit

	err := c.ShouldBindJSON(&recoveryBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, language.Language(lang, "parse_error"), errorCodes.ParsingError))
		return
	}
	if recoveryBody.Code == "" || recoveryBody.Password == "" {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, language.Language(lang, "code_password_empty"), errorCodes.CodeOrPasswordEmpty))
		return
	}
	digit, symbols := utils.PasswordChecker(recoveryBody.Password)
	if !digit || !symbols {
		c.JSON(http.StatusBadRequest, models.ResponseMsg(false, language.Language(lang, "password_should_be_include_digits"), errorCodes.PasswordShouldByIncludeSymbols))
		return
	}

	tx := a.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Find of code
	var foundCodes models.RegToken
	if err := tx.Model(models.RegToken{}).Where("code = ?", recoveryBody.Code).First(&foundCodes).Error; err != nil {
		tx.Rollback()
		a.logger.Error(err)
	}
	if foundCodes.Code == "" {
		c.JSON(http.StatusNotFound, models.ResponseMsg(false, language.Language(lang, "recovery_code_not_found"), errorCodes.RecoveryCodeNotFound))
		return
	}

	// Find user by code
	var foundUser models.User
	if err := tx.Model(models.User{}).Where("id = ?", uint(foundCodes.UserId)).First(&foundUser).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, models.ResponseMsg(false, language.Language(lang, "recovery_code_not_found"), errorCodes.RecoveryCodeNotFound))
		return
	}

	// Check if code is expired
	if foundCodes.Created < time.Now().UTC().Add(-24*time.Hour).Format(os.Getenv("DATE_FORMAT")) {
		tx.Model(&models.RegToken{}).Where("code = ?", foundCodes.Code).Delete(foundCodes)
		c.JSON(http.StatusUnauthorized, models.ResponseMsg(false, language.Language(lang, "recovery_code_expired"), errorCodes.RecoveryCodeExpired))
		return
	}

	// Hash password
	hashPassword := utils.Hash(recoveryBody.Password)

	// Delete code
	err = tx.Model(&models.RegToken{}).Where("code = ?", foundCodes.Code).Delete(foundCodes).Error
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, models.ResponseMsg(false, language.Language(lang, "db_error"), errorCodes.DBError))
		return
	}

	// Update password
	err = tx.Model(&models.User{}).Where("id = ?", foundUser.ID).Update("pass", hashPassword).Error
	if err != nil {
		tx.Rollback()
		a.logger.Error(err)
		c.JSON(http.StatusInternalServerError, models.ResponseMsg(false, language.Language(lang, "db_error"), errorCodes.DBError))
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, models.ResponseMsg(true, language.Language(lang, "password_reseted"), 0))

	a.db.AttachAction(models.ActionLogs{
		Action:  "Submit recovery password",
		Login:   foundUser.Login,
		Ip:      client.GetIP(c),
		Created: db.TimeNow(),
	})
}
