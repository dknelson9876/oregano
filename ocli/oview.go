package ocli

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/dknelson9876/oregano/omoney"
	"github.com/plaid/plaid-go/plaid"
)

// Define styles
var (
	faintStyle = lipgloss.NewStyle().Faint(true)
	// boldStyle  = lipgloss.NewStyle().Bold(true)
	rightAlignStyle = lipgloss.NewStyle().Align(lipgloss.Right)
)

type OView interface {
	ShowAccount(account plaid.AccountBase)
	ShowAccounts(accounts []plaid.AccountBase)
}

type OViewPlain struct {
	enableColor bool
}

func NewOViewPlain(enableColor bool) *OViewPlain {
	return &OViewPlain{
		enableColor: enableColor,
	}
}

func (v *OViewPlain) ShowAccount(account plaid.AccountBase) {
	// something is up with the way that Nullables are parsed
	//  and causing Ok's and IsSet's to be true, even when the
	//  value is nil, so I'm just gonna directly check for nil
	//  for now

	fmt.Print(account.GetName())
	fmt.Print(" (")
	// if off_name, ok := account.GetOfficialNameOk(); ok {
	if off_name := account.OfficialName.Get(); off_name != nil {
		fmt.Print(faintStyle.Render(*off_name))
	}
	fmt.Print(")\t$")
	if current := account.Balances.Current.Get(); current != nil {
		fmt.Printf("%.2f", *current)
	}
	fmt.Print("  /  $")
	if available := account.Balances.Available.Get(); available != nil {
		fmt.Printf("%.2f", *available)
	}
	fmt.Println()
}

func (v *OViewPlain) ShowAccounts(accounts []plaid.AccountBase) {
	var rows [][]string
	for _, acc := range accounts {
		var thisRow []string
		thisRow = append(thisRow, acc.GetName())

		if off_name := acc.OfficialName.Get(); off_name != nil {
			thisRow = append(thisRow, faintStyle.Render(*off_name))
		} else {
			thisRow = append(thisRow, "")
		}

		if current := acc.Balances.Current.Get(); current != nil {
			thisRow = append(thisRow, fmt.Sprintf("$%.2f", *current))
		} else {
			thisRow = append(thisRow, "")
		}

		if available := acc.Balances.Available.Get(); available != nil {
			thisRow = append(thisRow, fmt.Sprintf("$%.2f", *available))
		} else {
			thisRow = append(thisRow, "")
		}

		rows = append(rows, thisRow)
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("99"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if col > 1 {
				return rightAlignStyle
			} else {
				return lipgloss.NewStyle()
			}
		}).
		Headers("NAME", "FULL NAME", "BALANCE", "AVAILABLE").
		Rows(rows...)

	fmt.Println(t)
}

func (v *OViewPlain) ShowTransactions(acc omoney.Account) {
	var sl []omoney.Transaction
	if len(acc.Transactions) > 10 {
		sl = acc.Transactions[:10]
	} else {
		sl = acc.Transactions
	}

	var negAmount int
	if acc.Type == omoney.CreditCard {
		negAmount = 1
	} else {
		negAmount = -1
	}

	var rows [][]string
	for _, tr := range sl {
		thisRow := []string{
			tr.Date.Format("2006/01/02"),
			tr.Payee,
			tr.Category,
			fmt.Sprintf("%.2f", tr.Amount*float64(negAmount)),
		}
		rows = append(rows, thisRow)
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("99"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if col == 2 {
				return rightAlignStyle
			} else {
				return lipgloss.NewStyle()
			}
		}).
		Headers("DATE", "PAYEE", "CATEGORY", "AMOUNT").
		Rows(rows...)

	fmt.Println(t)
}
