package models

type User struct {
	ID         uint   `gorm:"unique" json:"id"`
	Login      string `gorm:"unique" json:"login"`
	Email      string `gorm:"unique" json:"email"`
	Phone      string `json:"phone"`
	AvatarId   string `json:"avatar_id"`
	Active     bool   `json:"active"`
	Pass       string `json:"pass"`
	RefCode    string `json:"ref_code"`
	InviteCode string `json:"invite_code"`
	Created    string `json:"created"`
	Updated    string `json:"updated"`
}

type UserWallet struct {
	UserID     uint   `gorm:"unique" json:"user_id"`
	BtcAddress string `gorm:"unique" json:"btc_address"`
	EthAddress string `gorm:"unique" json:"eth_address"`
	TrxAddress string `gorm:"unique" json:"trx_address"`
	SeedPhrase string `json:"seed_phrase"`
	Created    string `json:"created"`
	Updated    string `json:"updated"`
}

type EmailChange struct {
	UserID  uint   `json:"user_id"`
	Email   string `json:"email"`
	Code    int    `json:"code"`
	Created string `json:"created"`
}

type UserPass struct {
	UserID  uint   `gorm:"unique" json:"user_id"`
	Pass    string `json:"pass"`
	Updated string `json:"updated"`
}
