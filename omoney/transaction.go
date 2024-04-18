package omoney

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// A representation of a single transaction that took place between two
// accounts at a specific point in time
type Transaction struct {
	// Unique identifier for this transaction within
	// this application. Required field.
	Id string
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
		Id:        uuid.New().String(),
		AccountId: accountId,
		Payee:     payee,
		Amount:    amount,
		Date:      time.Now().Truncate(time.Second),
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
		t.Id,
		t.AccountId,
		t.Payee,
		t.Amount,
		t.Date.Format("2006/01/02"),
		t.Category,
		t.InstDescription,
		t.Description,
	)
}
func (m *Model) AddTransaction(tr *Transaction) error {
	_, err := m.db.NewInsert().
		Model(tr).
		Exec(context.TODO())
	return err
}

func (m *Model) GetTransactionById(id string) (Transaction, error) {
	tr := &Transaction{}

	err := m.db.NewSelect().
		Model(tr).
		Where("id = ?", id).
		Limit(1).
		Scan(context.TODO())
	return *tr, err
}

type GetTransactionsOptions struct {
	Count     int
	StartDate *time.Time
	EndDate   *time.Time
}

func (m *Model) GetTransactionsByAccount(accId string, ops ...GetTransactionsOptions) ([]Transaction, error) {
	var trs []Transaction

	var op GetTransactionsOptions
	if len(ops) == 1 {
		op = ops[0]
	} else {
		op = GetTransactionsOptions{Count: 10}
	}

	var query strings.Builder
	var args []interface{}

	query.WriteString("SELECT * FROM transactions WHERE account_id = ?")
	args = append(args, bun.Ident(accId))

	if op.StartDate != nil {
		query.WriteString(" AND date > ?")
		args = append(args, op.StartDate.Format("2006-01-02 15:04:05-07:00"))
	}

	if op.EndDate != nil {
		query.WriteString(" AND date < ?")
		args = append(args, op.EndDate.Format("2006-01-02 15:04:05-07:00"))
	}

	query.WriteString(" ORDER BY date DESC LIMIT ?")
	args = append(args, op.Count)

	err := m.db.NewRaw(query.String(), args...).Scan(context.TODO(), &trs)

	return trs, err
}

func (m *Model) RemoveTransaction(tr *Transaction) error {
	return m.RemoveTransactionById(tr.Id)
}

func (m *Model) RemoveTransactionById(id string) error {
	fmt.Printf("Removing tr %s\n", id)
	_, err := m.db.NewDelete().
		Model((*Transaction)(nil)).
		Where("id = ?", id).
		Exec(context.TODO())
	return err
}
