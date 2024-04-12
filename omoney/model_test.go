package omoney

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/extra/bundebug"
)

func CreateEmptyDB() *bun.DB {
	sqldb, err := sql.Open(sqliteshim.ShimName, "file::memory:")
	if err != nil {
		panic(err)
	}

	db := bun.NewDB(sqldb, sqlitedialect.New())

	_, err = db.NewCreateTable().
		Model((*Account)(nil)).
		IfNotExists().
		Exec(context.TODO())
	if err != nil {
		panic(err)
	}

	_, err = db.NewCreateTable().
		Model((*Transaction)(nil)).
		IfNotExists().
		Exec(context.TODO())
	if err != nil {
		panic(err)
	}

	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true),
		bundebug.WithEnabled(true),
	))

	return db
}

func TestRetrieveOnlyAccountByAlias(t *testing.T) {
	m := &Model{db: CreateEmptyDB()}

	acc := *NewAccount(WithAlias("dummy"),
		WithAnchor(20, time.Date(2001, 01, 01, 0, 0, 0, 0, time.Local)),
		WithAccountType(Checking),
	)

	m.AddAccount(acc)

	retrieved, err := m.GetAccount(acc.Alias)
	if err != nil {
		t.Error(err)
	}

	if !acc.LooseEquals(&retrieved) {
		t.Fatalf("Account insert and retrieval failed"+
			"\nhave: %+v"+
			"\nneed: %+v",
			retrieved, acc)
	}
}

func TestRetrieveSingleAccount(t *testing.T) {
	m := &Model{db: CreateEmptyDB()}

	acc := *NewAccount(WithAlias("acc10"))
	m.AddAccount(acc)

	for i := 0; i < 9; i++ {
		acc := *NewAccount(WithAlias(fmt.Sprintf("acc%d", i)))
		m.AddAccount(acc)
	}

	retrieved, err := m.GetAccount(acc.Alias)
	if err != nil {
		t.Error(err)
	}

	if !acc.LooseEquals(&retrieved) {
		t.Fatalf("Account retrieval from group failed"+
			"\nhave: %+v"+
			"\nneed: %+v",
			retrieved, acc)
	}

	retrieved, err = m.GetAccount(acc.Id)
	if err != nil {
		t.Error(err)
	}

	if !acc.LooseEquals(&retrieved) {
		t.Fatalf("Account retrieval from group failed"+
			"\nhave: %+v"+
			"\nneed: %+v",
			retrieved, acc)
	}
}

func TestRetrieveEachAccount(t *testing.T) {
	m := &Model{db: CreateEmptyDB()}

	accs := make([]Account, 10)
	for i := 0; i < 10; i++ {
		acc := *NewAccount(WithAlias(fmt.Sprintf("acc%d", i)))
		accs[i] = acc
		m.AddAccount(acc)
	}

	for _, acc := range accs {
		retrieved, err := m.GetAccount(acc.Id)
		if err != nil {
			t.Error(err)
		}

		if !acc.LooseEquals(&retrieved) {
			t.Fatalf("Account retrieval from group failed"+
				"\nhave: %+v"+
				"\nneed: %+v",
				retrieved, acc)
		}
	}
}

func TestRetrieveAllAccounts(t *testing.T) {
	m := &Model{db: CreateEmptyDB()}

	accs := make([]Account, 10)
	for i := 0; i < 10; i++ {
		acc := *NewAccount(WithAlias(fmt.Sprintf("acc%d", i)))
		accs[i] = acc
		m.AddAccount(acc)
	}

	retrieved := m.GetAccounts()
	if len(retrieved) != len(accs) {
		t.Fatalf("All account retrieval failed"+
			"\nneed length: %d"+
			"\nhave length: %d",
			len(accs), len(retrieved))
	}

	sort.Slice(retrieved, func(i, j int) bool {
		return retrieved[i].Alias < retrieved[j].Alias
	})

	for i, acc := range accs {
		r := retrieved[i]

		if !acc.LooseEquals(&r) {
			t.Fatalf("Account retrieval from group failed"+
				"\nhave: %+v"+
				"\nneed: %+v",
				r, acc)
		}
	}
}

func TestNewAccountFailNonuniqueAlias(t *testing.T) {
	m := &Model{db: CreateEmptyDB()}

	acc1 := *NewAccount(WithAlias("dummy"))
	m.AddAccount(acc1)

	acc2 := *NewAccount(WithAlias("dummy"))
	m.AddAccount(acc2)

	if len(m.GetAccounts()) != 1 {
		t.Fatalf("Reject duplicate alias failed")
	}

	retrieved, err := m.GetAccount(acc1.Id)
	if err != nil {
		t.Error(err)
	}

	if !acc1.LooseEquals(&retrieved) {
		t.Fatalf("Reject duplicate alias failed"+
			"\nhave: %+v"+
			"\nneed: %+v",
			retrieved, acc1)
	}
}

func TestIsValidAccountId(t *testing.T) {
	m := &Model{db: CreateEmptyDB()}

	acc := *NewAccount(WithAlias("acc1"))
	m.AddAccount(acc)

	if !m.IsValidAccountId(acc.Id) {
		t.Fatalf("Validate account id: False negative")
	}

	if m.IsValidAccountId("1234-5678-123095") {
		t.Fatalf("Validate account id: False positive")
	}
}

func TestIsValidAccountAlias(t *testing.T) {
	m := &Model{db: CreateEmptyDB()}

	acc := *NewAccount(WithAlias("dummy"))
	m.AddAccount(acc)

	if !m.IsValidAccountAlias(acc.Alias) {
		t.Fatalf("Validate account alias: False negative")
	}

	if m.IsValidAccountId("1234-5678-123095") {
		t.Fatalf("Validate account alias: False positive")
	}
}

func TestSetAlias(t *testing.T) {
	m := &Model{db: CreateEmptyDB()}

	acc := *NewAccount(WithAlias("dummy"))
	m.AddAccount(acc)

	m.SetAlias(acc.Id, "other")

	retrieved, err := m.GetAccount("other")
	if err != nil {
		t.Error(err)
	}

	if !acc.LooseEquals(&retrieved) {
		t.Fatal("Change alias failed")
	}

	retrieved, err = m.GetAccount("dummy")
	if err == nil {
		t.Fatal("Change alias failed")
	}

}