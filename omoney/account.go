package omoney

import (
	"fmt"
)

type AccountType string

const (
	UnknownAccount = "unknown"
	Checking       = "checking"
	Savings        = "savings"
	CreditCard     = "creditCard"
	Investment     = "investment"
	PersonalLoan   = "personalLoan"
)

type Account struct {
	Id         string // main unique identifier for account
	Alias      string // english nickname
	PlaidToken string // plaid generated key for getting data on this account
	// empty string if manually created account
	Type         AccountType
	Transactions []Transaction
}

func ParseAccountType(input string) (AccountType, error) {
	switch input {
	case "ch", "checking":
		return Checking, nil
	case "sa", "savings":
		return Savings, nil
	case "cc", "credit":
		return CreditCard, nil
	case "in", "investment":
		return Investment, nil
	case "pl", "personalLoan":
		return PersonalLoan, nil
	default:
		return UnknownAccount, fmt.Errorf("account type %s not recognized", input)
	}
}
