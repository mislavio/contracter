package accounts

import "github.com/jinzhu/gorm"

//Account represents a primary account on Contracter.
type Account struct {
	gorm.Model
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}
