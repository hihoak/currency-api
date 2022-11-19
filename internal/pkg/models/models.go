package models

import "time"

type User struct {
	ID int64 `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
	MiddleName string `json:"middle_name" db:"middle_name"`
	Surname string `json:"surname" db:"surname"`
	Mail string `json:"mail" db:"mail"`
	PhoneNumber string `json:"phone_number" db:"phone_number"`
	Blocked bool `json:"blocked" db:"blocked"`
	Registered bool `json:"registered" db:"registered"`
	Admin bool `json:"admin" db:"admin"`
	Password string `json:"password" db:"password"`
}

type Currencies string
const (
	RUB = "RUB"
	USD = "USD"
	EUR = "EUR"
)

func (c Currencies) String() string {
	return string(c)
}

type Wallet struct {
	ID int64 `json:"id" db:"id"`
	UserID int64 `json:"user_id" db:"user_id"`
	Currency Currencies `json:"currency" db:"currency"`
	Value int64 `json:"value" db:"value"`
}

type Course struct {
	Timestamp time.Time `json:"timestamp"`
	From Currencies `json:"from"`
	To Currencies `json:"to"`
	Value float64 `json:"value"`
}
