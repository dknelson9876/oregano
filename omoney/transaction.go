package omoney

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// A representation of a single transaction that took place between two
// accounts at a specific point in time
type Transaction struct {
	// Unique identifier for this transaction within
	// this application. Required field.
	UUID string
	// The account belonging to the user which the money
	// is being pulled from. Required field. Assumed to be a valid alias
	AccountId string
	// The 'account' which money is being sent to, typically
	// a business. Required field.
	Payee string
	// The value of this transaction. A negative value indicates
	// money moving into the account. Note that this follows
	// the convention of a credit card, which is the opposite
	// convention of a savings account. Required field.
	Amount float64
	// The date and time at which this transaction
	// took place. Required field which defaults to
	// current time in local time zone.
	Date time.Time
	// The category to which this transaction should be sorted
	// by. Optional field which defaults to empty string.
	// TODO: Details remain to be figured out for parent/sub-
	// categories and how they are parsed and stored
	Category string
	// The description assigned by the institution
	// typically will actually match the payee, with
	// some formatting that is not preferrable.
	// Optional field which defaults to empty string.
	InstDescription string
	// The description assigned by the user, to provide
	// specifics about this individual transaction.
	// Optional field which defaults to empty string.
	Description string
}

type TransactionOption func(*Transaction)

func NewTransaction(accountId string, payee string, amount float64,
	options ...TransactionOption) *Transaction {
	tr := &Transaction{
		UUID:    uuid.New().String(),
		AccountId: accountId,
		Payee:   payee,
		Amount:  amount,
		Date:    time.Now(),
	}

	for _, op := range options {
		op(tr)
	}

	return tr
}

func WithDate(date time.Time) TransactionOption {
	return func(t *Transaction) {
		t.Date = date
	}
}
func WithCategory(category string) TransactionOption {
	return func(t *Transaction) {
		t.Category = category
	}
}
func WithDescription(description string) TransactionOption {
	return func(t *Transaction) {
		t.Description = description
	}
}
func WithInstDescription(instDescription string) TransactionOption {
	return func(t *Transaction) {
		t.InstDescription = instDescription
	}
}

// Returns whether or not all fields, excepting uuid, match
func (t *Transaction) LooseEquals(other *Transaction) bool {
	//TODO: do more research on comparing dates
	return t.Date.Equal(other.Date) &&
		t.Amount == other.Amount &&
		t.AccountId == other.AccountId &&
		t.Payee == other.Payee &&
		t.Category == other.Category &&
		t.Description == other.Description
}

func (t *Transaction) String() string {
	return fmt.Sprintf("ID: %s\nAcc: %s\nPayee: %s\nAmount: %.2f\nDate: %s\nCategory: %s\nInstDescription: %s\nDescription: %s",
		t.UUID,
		t.AccountId,
		t.Payee,
		t.Amount,
		t.Date.Format("2006/01/02"),
		t.Category,
		t.InstDescription,
		t.Description,
	)
}
