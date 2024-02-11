package omoney

import (
	"time"
	"github.com/google/uuid"
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

func (t *Transaction) New(date time.Time, amount float64, account string,
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
