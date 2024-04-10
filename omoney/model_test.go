package omoney

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/extra/bundebug"
)

func CreateEmptyDB() *bun.DB {
	sqldb, err := sql.Open(sqliteshim.ShimName, "file::memory:?cache=shared")
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

func TestRetrieveSingleAccount(t *testing.T) {
	m := &Model{db: CreateEmptyDB()}

	acc := *NewAccount(WithAlias("dummy"),
		WithAnchor(20, time.Date(2001, 01, 01, 0, 0, 0, 0, time.Local)),
		WithAccountType(Checking),
	)

	m.AddAccount(acc)

	retrieved, err := m.GetAccount("dummy")
	if err != nil {
		t.Error(err)
	}

	if !acc.LooseEquals(&retrieved) {
		t.Fatalf("Account insert and retrieval failed"+
			"\ngot: %+v"+
			"\nneed: %+v",
			retrieved, acc)
	}
}
