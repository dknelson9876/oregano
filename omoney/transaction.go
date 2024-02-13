package omoney

import (
	"github.com/google/uuid"
	"time"
)

type Transaction struct {
	UUID        string
	Date        time.Time
	Amount      float64
	Account     string
	Payee       string
	Category    string
	Description string
}

func NewTransaction(date time.Time, amount float64, account string,
	payee string, category string, description string) Transaction {
	return Transaction{
		UUID:        uuid.New().String(),
		Date:        date,
		Amount:      amount,
		Account:     account,
		Payee:       payee,
		Category:    category,
		Description: description,
	}
}

// Returns whether or not all fields, excepting uuid, match
func (t *Transaction) LooseEquals(other *Transaction) bool {
	//TODO: do more research on comparing dates
	return t.Date.Equal(other.Date) &&
		t.Amount == other.Amount &&
		t.Account == other.Account &&
		t.Payee == other.Payee &&
		t.Category == other.Category &&
		t.Description == other.Description
}
