package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/mislavio/contracter/accounts"
	"github.com/mislavio/contracter/helpers"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/net/context"
)

// Context keys
var (
	AccountCtxKey = contextKey("account")
	TokenCtxKey   = contextKey("token")
	ErrorCtxKey   = contextKey("error")
)

// ContracterClaims represent jwt claims used in Contracter jwts
type ContracterClaims struct {
	jwt.StandardClaims
	Email string `json:"email"`
}

// ContracterJWT represents the contracter jwt standard auth.
type ContracterJWT struct {
	SigningKey []byte
	Signer     jwt.SigningMethod
}

// NewJWTFromAccount returns a JWT token will all claims relating to the accounts.Account
func (j *ContracterJWT) NewJWTFromAccount(a *accounts.Account) (*jwt.Token, string, error) {
	iat := time.Now().UTC().Unix()
	exp := iat + int64(time.Hour.Seconds())
	sub := a.BaseModel.ID.String()

	claims := &ContracterClaims{Email: a.Email, StandardClaims: jwt.StandardClaims{
		IssuedAt:  iat,
		ExpiresAt: exp,
		Subject:   sub,
	}}

	t := jwt.New(j.Signer)
	t.Claims = claims
	tokenString, err := t.SignedString(j.SigningKey)
	t.Raw = tokenString
	return t, tokenString, err
}

// Valid function implements the Claims interface
func (c ContracterClaims) Valid() error {
	return c.StandardClaims.Valid()
}

// GetAccountFromClaims retrieves the accounts.Account from the JWT claims
func (c ContracterClaims) GetAccountFromClaims() (*accounts.Account, error) {
	if c.Subject == strconv.FormatUint(1, 10) {
		return &accounts.Account{BaseModel: helpers.BaseModel{ID: uuid.NewV4()}}, nil
	}
	return &accounts.Account{}, errors.New("auth: account does not exist")
}

// Keyfunc returns the key used to sign the JWTs
func (j *ContracterJWT) Keyfunc(t *jwt.Token) (interface{}, error) {
	return j.SigningKey, nil
}

func newContextWithAccount(ctx context.Context, a *accounts.Account) context.Context {
	ctx = context.WithValue(ctx, AccountCtxKey, a)
	return ctx
}

func tokenFromContext(ctx context.Context) (*jwt.Token, ContracterClaims, error) {
	token, _ := ctx.Value(TokenCtxKey).(*jwt.Token)

	var claims ContracterClaims
	if token != nil {
		claims = *token.Claims.(*ContracterClaims)
	} else {
		claims = ContracterClaims{}
	}

	err, _ := ctx.Value(ErrorCtxKey).(error)

	return token, claims, err
}

func findTokenFromRequest(r *http.Request) string {
	fmt.Println(r.Cookie("jwt"))
	if cookie, err := r.Cookie("jwt"); err != http.ErrNoCookie {
		return cookie.Value
	}
	if bearer := r.Header.Get("Authorization"); len(bearer) > 7 && strings.ToUpper(bearer[0:6]) == "BEARER" {
		return bearer[7:]
	}
	return r.URL.Query().Get("jwt")
}
