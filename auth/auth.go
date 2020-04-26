package auth

import (
	"github.com/mislavio/contracter/accounts"

	"github.com/jinzhu/gorm"
)

// OAuthCredentials represents the client id and secret required to
type OAuthCredentials struct {
	OAuthClientID     string `json:"oAuthCredentials"`
	OAuthClientSecret string `json:"oauthSecret"`
}

// Credentials represents the identification and
// authentication variables used to integrate with the Upvest API.
type Credentials struct {
	gorm.Model
	OAuthCredentials *OAuthCredentials `json:"oauthCredentials"`
	UpvestUsername   string            `json:"upvestUsername"`
	DefaultWalletID  string            `json:"defaultWalletId"`
	AccountID        string
	Account          accounts.Account
}
