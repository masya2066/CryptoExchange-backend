package template

import (
	"crypto-exchange/app/internal/models"
	"crypto-exchange/app/internal/models/language"
	"os"
)

func UserRegister(lang string, user models.User, code string) (subj string, msg string) {
	return language.Language(lang, "welcome_admin_panel"),
		language.Language(lang, "link_to_register") + os.Getenv("DOMAIN") + "/registration/submit/" + code +
			"\n\nEmail: " + user.Email +
			"\nLogin: " + user.Login +
			"\nCreated: " + user.Created
}
