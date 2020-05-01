package accounts

import (
	"log"

	"github.com/jinzhu/gorm"
	"github.com/mislavio/contracter/helpers"
	"golang.org/x/crypto/bcrypt"
)

//Account represents a primary account on Contracter.
type Account struct {
	helpers.BaseModel
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Token     string
	Active    bool
}

// BeforeCreate gorm hook
func (a *Account) BeforeCreate(scope *gorm.Scope) {
	a.BaseModel.BeforeCreate(scope)
	// generate hash
	hash, err := bcrypt.GenerateFromPassword([]byte(a.Password), bcrypt.MinCost)
	if err != nil {
		log.Panic(err)
	}
	// Add verification token
	a.Token = helpers.RandomString(25)
	// Ensure account is not active
	a.Active = false
	// Store hased password.
	a.Password = string(hash)
}

// IsActive indicates if the account email has been verified.
func (a *Account) IsActive() bool {
	return a.Active
}

// ComparePassword returns nil if the provided string matches the Account password.
func (a *Account) ComparePassword(p string) error {
	return bcrypt.CompareHashAndPassword([]byte(a.Password), []byte(p))
}

// FindByEmailOrFalse returns false if record not found.
func (a *Account) FindByEmailOrFalse(e string, db *gorm.DB) bool {
	return db.Where("email = ?", e).Find(a).RecordNotFound()
}
