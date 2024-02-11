package omoney

type AccountType string

const (
	Checking   = "checking"
	Savings    = "savings"
	CreditCard = "creditCard"
	Investment = "investment"
)

type Account struct {
	Id         string // main unique identifier for account
	Alias      string // english nickname
	PlaidToken string // plaid generated key for getting data on this account
	// null if manually created account
	Type         AccountType
	Transactions []Transaction
}
