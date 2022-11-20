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
	GBP = "GBP"
	JPY = "JPY"
	CHF = "CHF"
	CNY = "INR"
)
var AllSupportedCurrencies = []Currencies{RUB, EUR, USD, GBP, JPY, CHF, CNY}

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
	ID int64 `json:"id"`
	Timestamp int64 `json:"timestamp"`
	From Currencies `json:"from" db:"from_currency"`
	To Currencies `json:"to" db:"to_currency"`
	Value float64 `json:"value" db:"course"`
}

type Transaction struct {
	ID int64 `json:"id"`
	UserID int64 `json:"user_id"`
	Date time.Time `json:"date"`
	OperationName string `json:"operation_name"`
	IncomeAmount int64 `json:"income_amount"`
	OutcomeAmount int64 `json:"outcome_amount"`
	IncomeWalletID int64 `json:"income_wallet_id"`
	OutcomeWalletID int64 `json:"outcome_wallet_id"`
	IncomeWalletCurrency string `json:"income_wallet_currency"`
	CourseValue float64 `json:"course_value"`
}
