package models

type FullUserInfo struct {
	ID         uint   `gorm:"unique" json:"id"`
	Login      string `json:"login"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	AvatarId   string `json:"avatar_id"`
	Active     bool   `json:"active"`
	Pass       string `json:"pass"`
	Alive      int    `json:"alive"`
	RefCode    string `json:"ref_code"`
	InviteCode string `json:"invite_code"`
	BtcAddress string `json:"btc_address"`
	EthAddress string `json:"eth_address"`
	TrxAddress string `json:"trx_address"`
	Created    string `json:"created"`
	Updated    string `json:"updated"`
}

type UserLoginInfo struct {
	Info         User   `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Alive        int    `json:"alive"`
}

type TokenData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Activate struct {
	Code     string `json:"code"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegToken struct {
	UserId  int    `json:"user_id"`
	Type    int    `json:"type"`
	Code    string `json:"code"`
	Created string `json:"created"`
}

type RejectedToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
