package auth

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"
	"github.com/mislavio/contracter/accounts"
	"github.com/mislavio/contracter/helpers"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/net/context"
)

// Request Response payloads.

// SignUpPayload represents a sign up request body.
type SignUpPayload struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// SignUpResponse represents a sign up response.
type SignUpResponse struct{}

// SignInPayload represents a sign in request body.
type SignInPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SignInResponse represents a sign in response.
type SignInResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Token     string    `json:"token"`
}

// VerifyResponse represents the email verification response
type VerifyResponse struct{}

// WhoAmIResponse represents a sign in response.
type WhoAmIResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Token     string    `json:"token"`
}

// NewSignInResponse returns the signup response
func NewSignInResponse(a *accounts.Account, t string) *SignInResponse {
	return &SignInResponse{ID: a.ID, Email: a.Email, FirstName: a.FirstName, LastName: a.LastName, Token: t}
}

// Bind implements the binder interface.
func (a *SignInPayload) Bind(r *http.Request) error {
	return validateEmailAndPassword(a.Email, a.Password)
}

// Bind implements the binder interface.
func (a *SignUpPayload) Bind(r *http.Request) error {
	return validateEmailAndPassword(a.Email, a.Password)
}

// Render implements the renderer interface.
func (a *SignUpResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, 201)
	return nil
}

// Render implements the renderer interface.
func (v *VerifyResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, 200)
	return nil
}

// Render implements the renderer interface.
func (s *SignInResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, 200)
	return nil
}

// Render implements the renderer interface.
func (s *WhoAmIResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, 200)
	return nil
}

// Request Handlers

// SignUp creates a new unverified account
func SignUp(db *gorm.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := &SignUpPayload{}

		if err := render.Bind(r, data); err != nil {
			render.Render(w, r, helpers.ErrBadRequest(err))
			return
		}

		a := &accounts.Account{}

		if !a.FindByEmailOrFalse(data.Email, db) {
			render.Render(w, r, helpers.ErrConflict(errors.New("email already exists")))
			return
		}

		a = &accounts.Account{
			Email:     data.Email,
			Password:  data.Password,
			FirstName: data.FirstName,
			LastName:  data.LastName,
		}

		if err := db.Create(a).Error; err != nil {
			log.Panic(err)
		}

		log.Printf("Created: account %v (%v)", a.Email, a.ID)
		// Temp token print for local testing
		log.Printf("http://localhost:8000/auth/verify?token=%v", a.Token)

		render.Render(w, r, &SignUpResponse{})
	})
}

// SignIn creates a new valid jwt corresponding to an Account
func SignIn(db *gorm.DB, j *ContracterJWT) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := &SignInPayload{}

		if err := render.Bind(r, data); err != nil {
			render.Render(w, r, helpers.ErrBadRequest(err))
			return
		}

		a := &accounts.Account{}

		if a.FindByEmailOrFalse(data.Email, db) {
			// 401 response to not leak email info
			render.Render(w, r, helpers.ErrUnauthorized(errors.New("incorrect email or password")))
			return
		}

		log.Println(a.ID)

		if !a.IsActive() {
			render.Render(w, r, helpers.ErrUnauthorized(errors.New("email not verified")))
		}

		if err := a.ComparePassword(data.Password); err != nil {
			render.Render(w, r, helpers.ErrUnauthorized(errors.New("incorrect email or password")))
			return
		}

		_, tokenString, err := j.NewJWTFromAccount(a)
		if err != nil {
			log.Panic(err)
		}

		render.Render(w, r, NewSignInResponse(a, tokenString))

	})
}

// VerifyEmail activates an account if .
func VerifyEmail(db *gorm.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		a := &accounts.Account{}
		log.Print(token)
		// TODO: add token validation
		if db.Where("token = ?", token).Find(a).RecordNotFound() {
			render.Render(w, r, helpers.ErrNotFound("account", ""))
			return
		}
		if a.IsActive() {
			render.Render(w, r, helpers.ErrBadRequest(errors.New("account already active")))
			return
		}

		a.Active = true
		a.Token = ""

		if err := db.Model(a).Updates(map[string]interface{}{
			"active": true,
			"token":  gorm.Expr("NULL"),
		},
		).Error; err != nil {
			log.Panic(err)
		}

		render.Render(w, r, &VerifyResponse{})
	})
}

// WhoAmI returns current account information
func WhoAmI() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		a, _ := accountFromContext(ctx)
		token, _, _ := tokenFromContext(ctx)

		render.Render(w, r, &WhoAmIResponse{
			ID:        a.ID,
			Email:     a.Email,
			FirstName: a.FirstName,
			LastName:  a.LastName,
			Token:     token.Raw,
		})
	})
}

// Middlewares

// Verifier is Contracters JWT verification middleware
func Verifier(j *ContracterJWT) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			if tokenString := findTokenFromRequest(r); tokenString != "" {
				var claims = &ContracterClaims{}

				token, err := jwt.ParseWithClaims(tokenString, claims, j.Keyfunc)
				if err != nil {
					http.Error(w, http.StatusText(400), 400)
					return
				}
				ctx = context.WithValue(ctx, TokenCtxKey, token)
				ctx = context.WithValue(ctx, ErrorCtxKey, err)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AccountAuthenticator is the Contracter authentication middleware
func AccountAuthenticator(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			t, claims, err := tokenFromContext(ctx)

			if err != nil {
				log.Print(err)
				http.Error(w, http.StatusText(401), 401)
				return
			}

			if t == nil || !t.Valid {
				http.Error(w, http.StatusText(401), 401)
				return
			}

			a, err := claims.GetAccountFromClaims(db)
			if err != nil {
				fmt.Print(err)
				http.Error(w, "we couldn't find your account", 401)
				return
			}

			ctx = newContextWithAccount(ctx, a)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Helpers

func validateEmailAndPassword(e string, p string) error {
	p = strings.TrimSpace(p)
	if len(p) < 3 {
		return errors.New("password must be longer than 5 characters")
	}
	// Taken from https://www.alexedwards.net/blog/validation-snippets-for-go#email-validation
	regEmail := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if len(e) > 254 || !regEmail.MatchString(e) {
		return errors.New("invalid email address")
	}
	return nil
}
