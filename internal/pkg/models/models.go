package models

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
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
}

type Currencies string
const (
	RUB = "RUB"
	USD = "USD"
)

type Wallet struct {
	ID int64
	UserID int64
	Currency Currencies
	Value int64
}
