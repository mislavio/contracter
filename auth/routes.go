package auth

import (
	"github.com/go-chi/chi"
	"github.com/jinzhu/gorm"
)

// Router compiles all auth routes
func Router(db *gorm.DB, j *ContracterJWT) chi.Router {
	r := chi.NewRouter()

	r.Post("/signup", SignUp(db))
	r.Post("/signin", SignIn(db, j))
	r.Get("/verify", VerifyEmail(db))
	return r
}
