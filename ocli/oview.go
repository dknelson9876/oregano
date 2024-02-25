package ocli

import (
	"fmt"
	"strconv"

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

func (v *OViewPlain) ShowPlaidAccount(account plaid.AccountBase) {
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

func (v *OViewPlain) ShowPlaidAccounts(accounts []plaid.AccountBase) {
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

func (v *OViewPlain) ShowTransactions(trs []omoney.Transaction, invert bool, startIndex int) {
	var negAmount int
	if invert {
		negAmount = -1
	} else {
		negAmount = 1
	}

	var rows [][]string
	for i, tr := range trs {
		thisRow := []string{
			strconv.Itoa(startIndex + i),
			tr.Date.Format("2006/01/02"),
			tr.Payee,
			tr.Category,
			fmt.Sprintf("$%.2f", tr.Amount*float64(negAmount)),
		}
		rows = append(rows, thisRow)
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("99"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if col == 4 {
				return rightAlignStyle
			} else {
				return lipgloss.NewStyle()
			}
		}).
		Headers("", "DATE", "PAYEE", "CATEGORY", "AMOUNT").
		Rows(rows...)

	fmt.Println(t)
}

type ShowTransactionOptions struct {
	ShowId       bool
	ShowCategory bool
	ShowInstDesc bool
	ShowDesc     bool
}

func (v *OViewPlain) ShowTransaction(tr omoney.Transaction, ops ...ShowTransactionOptions) {
	rows := make([]string, 0)
	var op ShowTransactionOptions
	if len(ops) == 1 {
		op = ops[0]
	} else {
		op = ShowTransactionOptions{}
	}

	if op.ShowId {
		rows = append(rows, fmt.Sprintf("Id: %s", tr.UUID))
	}

	rows = append(rows, fmt.Sprintf("Account: %s", tr.Account))
	rows = append(rows, fmt.Sprintf("Payee: %s", tr.Payee))
	rows = append(rows, fmt.Sprintf("Amount: $%.2f", tr.Amount))
	rows = append(rows, fmt.Sprintf("Date: %s", tr.Date.Format("2006/01/02")))

	if op.ShowCategory {
		rows = append(rows, fmt.Sprintf("Category: %s", tr.Category))
	}

	if op.ShowInstDesc {
		rows = append(rows, fmt.Sprintf("Inst Description: %s", tr.InstDescription))
	}

	if op.ShowDesc {
		rows = append(rows, fmt.Sprintf("Description: %s", tr.Description))
	}

	for _, row := range rows {
		fmt.Println(row)
	}
}

type ShowAccountOptions struct {
	ShowType   bool
	ShowAnchor bool
	ShowId     bool
}

func (v *OViewPlain) ShowAccounts(accounts []omoney.Account, ops ...ShowAccountOptions) {
	rows := make([][]string, len(accounts))
	var op ShowAccountOptions
	var headers []string
	if len(ops) == 1 {
		op = ops[0]
	} else {
		op = ShowAccountOptions{}
	}

	// always show alias
	// TODO: Plaid integration may allow accounts without aliases
	headers = append(headers, "ALIAS")
	for i, acc := range accounts {
		rows[i] = append(rows[i], acc.Alias)
	}

	if op.ShowType {
		headers = append(headers, "TYPE")
		for i, acc := range accounts {
			rows[i] = append(rows[i], string(acc.Type))
		}
	}

	if op.ShowAnchor {
		headers = append(headers, "ANCHOR")
		for i, acc := range accounts {
			rows[i] = append(rows[i], fmt.Sprintf("($%.2f, %s)",
				acc.AnchorBalance,
				acc.AnchorTime.Format("2006/01/02")))
		}
	}

	if op.ShowId {
		headers = append(headers, "ID")
		for i, acc := range accounts {
			rows[i] = append(rows[i], acc.Id)
		}
	}

	t := table.New().Headers(headers...).Rows(rows...)
	fmt.Println(t)
}

func (v *OViewPlain) ShowAccount(acc omoney.Account) {
	fmt.Printf("Id: %s\nAlias: %s\nType: %s\nAnchor: ($%.2f, %s)\n",
		acc.Id,
		acc.Alias,
		acc.Type,
		acc.AnchorBalance,
		acc.AnchorTime)
}
