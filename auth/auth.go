package auth

import (
	"github.com/mislavio/contracter/accounts"
	"github.com/mislavio/contracter/helpers"
)

// OAuthCredentials represents the client id and secret required to
type OAuthCredentials struct {
	OAuthClientID     string `json:"oAuthCredentials"`
	OAuthClientSecret string `json:"oauthSecret"`
}

// Credentials represents the identification and
// authentication variables used to integrate with the Upvest API.
type Credentials struct {
	helpers.BaseModel
	OAuthCredentials *OAuthCredentials `json:"oauthCredentials"`
	UpvestUsername   string            `json:"upvestUsername"`
	DefaultWalletID  string            `json:"defaultWalletId"`
	AccountID        string
	Account          accounts.Account
}

type contextKey string

func (c contextKey) String() string {
	return "contracter/auth context key " + string(c)
}
