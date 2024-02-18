package ocli

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/araddon/dateparse"
	"github.com/dknelson9876/oregano/omoney"
	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/erikgeiser/promptkit/selection"
)

const (
	// string constants, to be printed and used as keys
	sTransactionID = "Transaction ID"
	sAccount       = "Account"
	sPayee         = "Payee"
	sAmount        = "Amount"
	sDate          = "Date"
	sCategory      = "Category"
	sInstDesc      = "Institution Description"
	sDesc          = "Description"
	sDir           = "Direction (Debit/Credit)"
)

func ReadCsv(filepath string) {
	f, err := os.Open(filepath)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	headerPrompt := confirmation.New(
		fmt.Sprintf("Are these headers for your file? [(%s, %s, %s)\n", records[0][0], records[0][1], records[0][2]),
		confirmation.No,
	)
	headers, err := headerPrompt.RunPrompt()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	keymap := selection.NewDefaultKeyMap()
	keymap.Up = append(keymap.Up, "k")
	keymap.Down = append(keymap.Down, "j")

	var unmatchedFields []string
	var colMap map[string]int
	for {
		unmatchedFields = []string{sTransactionID, sAccount, sPayee, sAmount,
			sDate, sCategory, sInstDesc, sDesc, sDir}
		colMap = make(map[string]int, 0)
		colIdx := 0
		for colIdx < len(records[0]) {
			if headers {
				fmt.Printf("Which field does the column %s belong to?\n\t(%s, %s, %s)\n",
					records[0][colIdx],
					collapseWhitepace(records[1][colIdx]),
					collapseWhitepace(records[2][colIdx]),
					collapseWhitepace(records[3][colIdx]),
				)
			} else {
				fmt.Printf("Which field does the column with the following values belong to?\n\t(%s, %s, %s)\n",
					collapseWhitepace(records[0][colIdx]),
					collapseWhitepace(records[1][colIdx]),
					collapseWhitepace(records[2][colIdx]),
				)
			}

			sel := selection.New("", append(unmatchedFields, "Ignore", "Cancel"))
			sel.PageSize = 5
			sel.Filter = nil // disable filtering
			sel.KeyMap = keymap

			result, err := sel.RunPrompt()
			if err != nil {
				fmt.Println(err)
				return
			}

			if result == "Cancel" {
				fmt.Println()
				return
			} else if result != "Ignore" {
				if headers {
					fmt.Printf("Mapping %s column to %s field\n", records[0][colIdx], result)
				} else {
					fmt.Printf("Mapping column %d to %s field\n", colIdx, result)
				}
				unmatchedFields = removeElement(unmatchedFields, result)
				colMap[result] = colIdx
			}
			colIdx++
		}
		// fmt.Println(colMap)

		canContinue := false
		// Also go back if a required field is missing
		if _, ok := colMap[sAccount]; !ok {
			fmt.Println("Error: Missing required field 'Account'")
		} else if _, ok := colMap[sPayee]; !ok {
			fmt.Println("Error: Missing required field 'Payee'")
		} else if _, ok := colMap[sAmount]; !ok {
			fmt.Println("Error: Missing required field 'Amount'")
		} else {
			var rowIdx int
			if headers {
				rowIdx = 1
			} else {
				rowIdx = 0
			}

			accName := records[rowIdx][colMap[sAccount]]
			payee := records[rowIdx][colMap[sPayee]]
			amount, err := strconv.ParseFloat(records[rowIdx][colMap[sAmount]], 64)
			if err != nil {
				fmt.Println("Error: Could not parse number from 'Amount' column")
				goto ShowRestartPrompt
			}

			ops := make([]omoney.TransactionOption, 0)
			if dateCol, ok := colMap[sDate]; ok {
				date, err := dateparse.ParseLocal(records[rowIdx][dateCol])
				if err != nil {
					fmt.Println("Error: Could not parse date from 'Date' column")
					goto ShowRestartPrompt
				}
				ops = append(ops, omoney.WithDate(date))
			}

			if catCol, ok := colMap[sCategory]; ok {
				cat := records[rowIdx][catCol]
				ops = append(ops, omoney.WithCategory(cat))
			}

			if instDescCol, ok := colMap[sInstDesc]; ok {
				instDesc := collapseWhitepace(records[rowIdx][instDescCol])
				ops = append(ops, omoney.WithInstDescription(instDesc))
			}

			if descCol, ok := colMap[sDesc]; ok {
				desc := records[rowIdx][descCol]
				ops = append(ops, omoney.WithDescription(desc))
			}

			var mul float64
			if dirCol, ok := colMap[sDir]; ok {
				dir := records[rowIdx][dirCol]
				if dir == "debit" {
					mul = 1
				} else if dir == "credit" {
					mul = -1
				} else {
					fmt.Println("Error: Could not parse debit/credit from 'Direction' column")
					goto ShowRestartPrompt
				}
			} else {
				mul = 1
			}

			tr := omoney.NewTransaction(accName, payee, amount*mul, ops...)
			fmt.Printf("The first transaction was parsed into:\n%s", tr)

			canContinue = true

		}

	ShowRestartPrompt:
		var restartStr string
		if canContinue {
			restartStr = "Reassign columns? (No will begin processing)"
		} else {
			restartStr = "Could not assign columns. Retry column assignment?"
		}
		restartPrompt := confirmation.New(restartStr,
			confirmation.NewValue(true))
		restartPrompt.Template = confirmation.TemplateYN
		restartPrompt.ResultTemplate = confirmation.ResultTemplateYN

		response, err := restartPrompt.RunPrompt()
		if err != nil {
			fmt.Printf("Something went wrong with the prompt: %v\n", err)
			return
		}
		if !response {
			if canContinue {
				// said no to reassign when columns could be parsed
				// so break the loop to start actually parsing
				break
			} else {
				// said no to reassign when columns could not be parsed
				// so cancel the import
				fmt.Println("Canceling import...")
				return
			}
		}

	}
	fmt.Println("Processing...")

	// accMap := make(map[string]string, 0) // name in csv -> alias in model
	// newTrans := make([]omoney.Transaction, 0)

	// for _, rec := range records {

	// }
}

func removeElement(s []string, element string) []string {
	for i, v := range s {
		if v == element {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func collapseWhitepace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
