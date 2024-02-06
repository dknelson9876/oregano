// Based on plaid/plaid-go/model_account_base.go
//  but simplified to only include information relevant to me
//  and to have conversion methods that fix nil checks

package omoney

type Account struct {
	// Unique identifier from Plaid
	AccountId string `json:"account_id"`
	Balances AccountBalance `json:"balances"`
	// Name of acount, either as given by user or institution
	Name string `json:"name"`
	// Official name of the account given by the institution
	OfficialName *string
	Type AccountType `json:"type"`
	Subtype *AccountSubtype `json:"subtype"`
}
