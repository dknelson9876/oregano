package omoney

import (
	"context"
	"database/sql"
	"fmt"
	// "os"
	"strconv"
	"time"

	"github.com/araddon/dateparse"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	// "github.com/uptrace/bun/extra/bundebug"
)

const (
	DbFilename = "oregano_data.db"
)

type Model struct {
	// FilePath string
	// Accounts map[string]Account
	// Aliases  map[string]string // alias -> uuid

	db *bun.DB
}

func NewModelFromDB(filepath string) (*Model, error) {
	sqldb, err := sql.Open(sqliteshim.ShimName, filepath)
	if err != nil {
		return nil, err
	}

	db := bun.NewDB(sqldb, sqlitedialect.New())

	// db.AddQueryHook(bundebug.NewQueryHook(
	// 	bundebug.WithEnabled(true),
	// 	bundebug.WithVerbose(true),
	// 	bundebug.WithWriter(os.Stdout)))

	_, err = db.NewCreateTable().
		Model((*Account)(nil)).
		IfNotExists().
		Exec(context.TODO())
	if err != nil {
		return nil, err
	}

	_, err = db.NewCreateTable().
		Model((*Transaction)(nil)).
		IfNotExists().
		Exec(context.TODO())
	if err != nil {
		return nil, err
	}

	return &Model{db: db}, nil
}

func (m *Model) GetAccount(input string) (Account, error) {
	acc := &Account{}

	// WARNING: This will likely break if an alias that looks like
	//  an id is ever assigned. So for now, just don't
	err := m.db.NewSelect().
		Model(acc).
		Where("id = ?", input).
		WhereOr("alias = ?", input).
		Limit(1).
		Scan(context.TODO())
	if err != sql.ErrNoRows {
		if err != nil {
			return Account{}, err
		}
		if acc.Id != "" {
			return *acc, nil
		}
	}

	return Account{}, fmt.Errorf("input not recognized as valid account ID or alias: %s", input)

}

func (m *Model) GetAccounts() []Account {
	var accs []Account
	err := m.db.NewSelect().
		Model(&accs).
		Scan(context.TODO())
	if err != nil {
		return make([]Account, 0)
	}
	return accs
}

func (m *Model) GetAliases() map[string]string {
	var ids []string
	var aliases []string
	err := m.db.NewSelect().
		Model((*Account)(nil)).
		Column("id", "alias").
		Scan(context.TODO(), &ids, &aliases)
	if err != nil {
		return make(map[string]string, 0)
	}

	toreturn := make(map[string]string, len(ids))
	for i := range ids {
		toreturn[aliases[i]] = ids[i]
	}

	return toreturn
}

func (m *Model) IsValidAccountId(input string) bool {
	exists, _ := m.db.NewSelect().
		Model((*Account)(nil)).
		Where("id = ?", input).
		Exists(context.TODO())
	return exists
}

func (m *Model) IsValidAccountAlias(input string) bool {
	exists, _ := m.db.NewSelect().
		Model((*Account)(nil)).
		Where("alias = ?", input).
		Exists(context.TODO())
	return exists
}

// Given the alias for an account, return the id of that account.
// If alias does not exist, returns empty string
func (m *Model) GetAccountId(alias string) string {
	id := ""
	m.db.NewSelect().
		Model((*Account)(nil)).
		Column("id").
		Where("alias = ?", alias).
		Scan(context.TODO(), &id)
	return id
}

// Given an id, returns that id.
// Given an alias, returns the matching alias.
// So that, given an alias or an id as input, reliably
// change it to an id
func (m *Model) resolveToId(input string) (string, error) {
	acc := &Account{}

	err := m.db.NewSelect().
		Model(acc).
		Where("id = ?", input).
		WhereOr("alias = ?", input).
		Limit(1).
		Scan(context.TODO())
	return acc.Id, err
}

// given a string that is an id or an alias, return the matching
// Account's PlaidToken
func (m *Model) GetAccessToken(input string) (string, error) {
	acc, err := m.GetAccount(input)
	if err != nil {
		return "", err
	} else {
		return acc.PlaidToken, nil
	}
}

func (m *Model) AddAccount(acc Account) {
	m.db.NewInsert().
		Model(&acc).
		Exec(context.TODO())
}

func (m *Model) RemoveAccount(input string) error {
	id, err := m.resolveToId(input)
	if err != nil {
		return err
	}

	_, err = m.db.NewDelete().
		Model((*Account)(nil)).
		Where("id = ?", id).
		Exec(context.TODO())

	return err
}

// iterate over accounts, ensuring consistency in data
func (m *Model) RepairAccounts() {
	panic("Repair Accounts has been disabled. It may not be needed anymore?")
	// Repair List:
	// - Set Account ID within transactions to match account it's stored in
	// - Recalculate current balance
	// for _, acc := range m.Accounts {
	// 	acc.RepairTransactions()
	// 	acc.UpdateCurrentBalance()
	// }
}

func (m *Model) SetAlias(id string, alias string) error {
	err := m.db.NewUpdate().
		Model((*Account)(nil)).
		Set("alias = ?", alias).
		Where("id = ?", id).
		Scan(context.TODO())
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	return nil
}

func (m *Model) SetAnchor(account string, anchor []string) error {
	id, err := m.resolveToId(account)
	if err != nil {
		return err
	}

	amount, err := strconv.ParseFloat(anchor[0], 64)
	if err != nil {
		return err
	}

	date, err := dateparse.ParseLocal(anchor[1])
	if err != nil {
		return err
	}

	acc := &Account{}

	err = m.db.NewSelect().
		Model(acc).
		Where("id = ?", id).
		Scan(context.TODO())
	if err != nil {
		fmt.Printf("set anchor failed: %s\n", err)
		return err
	}

	acc.AnchorBalance = amount
	acc.AnchorTime = date

	_, err = m.db.NewUpdate().
		Model(acc).
		Column("anchor_balance").
		Column("anchor_time").
		WherePK().
		Exec(context.TODO())

	return err
}

func (m *Model) GetCurrentBalance(accId string) (float64, error) {
	sum := 0.0
	err := m.db.NewRaw(
		"SELECT sum(amount) FROM transactions WHERE account_id = ? AND date > (SELECT anchor_time FROM accounts WHERE id = ?)",
		bun.Ident(accId), bun.Ident(accId),
	).Scan(context.TODO(), &sum)
	if err != nil {
		return 0, err
	}

	anchor := 0.0
	err = m.db.NewSelect().
		Model((*Account)(nil)).
		Column("anchor_balance").
		Where("id = ?", accId).
		Scan(context.TODO(), &anchor)
	if err != nil {
		return 0, err
	}

	return sum + anchor, nil
}

type GetSumsOptions struct {
	StartDate *time.Time
	EndDate   *time.Time
	Range     string
	Grouping  string // equivalent to a field of transactions
}

func (m *Model) GetTransactionSums(ops ...GetSumsOptions) (map[string]float64, error) {
	var op GetSumsOptions
	if len(ops) == 1 {
		op = ops[0]
	} else {
		panic("Failed to pass options to getTransactionSums")
	}

	var labels []string
	var values []float64

	err := m.db.NewSelect().
		Model((*Transaction)(nil)).
		ColumnExpr(fmt.Sprintf("%s, sum(amount)", op.Grouping)).
		Where("date > ?", op.StartDate.Format(dateFormatStr)).
		Where("date < ?", op.EndDate.Format(dateFormatStr)).
		Group(op.Grouping).
		Order("amount DESC").
		Scan(context.TODO(), &labels, &values)
	if err != nil {
		return nil, err
	}

	return ToMap(labels, values), nil

}

func ToMap[K comparable, V any](keys []K, values []V) map[K]V {
	resultMap := make(map[K]V)
	for i, key := range keys {
		resultMap[key] = values[i]
	}
	return resultMap
}
