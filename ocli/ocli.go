package ocli

import (
	"fmt"

	"github.com/dknelson9876/oregano/omoney"
	"github.com/google/uuid"
)

func CreateManualAccount(input []string) *omoney.Account {
	if len(input) < 2 {
		fmt.Println("Error: new account requires exactly 2 arguments")
		fmt.Println("Usage: new account [alias] [type]")
		return nil
	}
	id := uuid.New().String()
	alias := input[0]
	accType, err := omoney.ParseAccountType(input[1])
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return nil
	}

	return &omoney.Account{
		Id:           id,
		Alias:        alias,
		Type:         accType,
		Transactions: make([]omoney.Transaction, 0),
	}
}
