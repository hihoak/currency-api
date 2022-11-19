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
	ID int64
	UserID int64
	Currency Currencies
	Value int64
}
