package ocli

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/araddon/dateparse"
	"github.com/dknelson9876/oregano/omoney"
	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/erikgeiser/promptkit/selection"
	"golang.org/x/exp/maps"
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

// Given the path to a csv file, and the map existingAccounts of alias -> id,
// interactively parse the csv file into a slice of transaction structs
func ReadCsv(filepath string, existingAccounts map[string]string) []*omoney.Transaction {
	f, err := os.Open(filepath)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return nil
	}

	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return nil
	}

	headerPrompt := confirmation.New(
		fmt.Sprintf("->Are these headers for your file? [(%s, %s, %s)", records[0][0], records[0][1], records[0][2]),
		confirmation.No,
	)
	headerPrompt.Template = confirmation.TemplateYN
	headerPrompt.ResultTemplate = confirmation.ResultTemplateYN
	headers, err := headerPrompt.RunPrompt()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return nil
	}

	// loop until columns are accepted as correct
	var colMap map[string]int
	for {
		colMap, err = buildColumnMap(records, headers)
		if err != nil {
			fmt.Printf("Leaving import setup: %s\n", err)
			return nil
		}
		// fmt.Println(colMap)

		var rowIdx int
		if headers {
			rowIdx = 1
		} else {
			rowIdx = 0
		}

		canContinue := false
		tr, err := tryBuildTransaction(records[rowIdx], colMap)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		} else {
			fmt.Printf("The first transaction was parsed into:\n%s\n", tr)
			canContinue = true
		}

		var restartStr string
		if canContinue {
			restartStr = "->Reassign columns? (No will begin processing)"
		} else {
			restartStr = "->Could not assign columns. Retry column assignment?"
		}

		restartPrompt := confirmation.New(restartStr,
			confirmation.NewValue(true))
		restartPrompt.Template = confirmation.TemplateYN
		restartPrompt.ResultTemplate = confirmation.ResultTemplateYN

		response, err := restartPrompt.RunPrompt()
		if err != nil {
			fmt.Printf("Something went wrong with the prompt: %v\n", err)
			return nil
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
				return nil
			}
		}
		// else said yes to reassign, so let the loop restart

	}
	fmt.Println("Processing...")

	accMap := make(map[string]string, 0) // name in csv -> accountId in model
	newTrans := make([]*omoney.Transaction, 0)

	var recordsRange [][]string
	if headers {
		recordsRange = records[1:]
	} else {
		recordsRange = records
	}

	for _, rec := range recordsRange {
		tr, err := tryBuildTransaction(rec, colMap)
		if err != nil {
			fmt.Printf("Failed to import: %s\n", err)
			fmt.Printf("Row: %s\n", rec)
			continue
		}

		// if account for this transaction is not a registered account,
		// if !slices.Contains(existingAccounts, tr.Account) {
		if id, ok := existingAccounts[tr.AccountId]; ok {
			// if account from file is a known account alias, convert it to the ID
			tr.AccountId = id

		} else if !mapContains(existingAccounts, tr.AccountId) {
			// if account from file is a known account ID, no action needs to be taken
			// else, query the user for which account to use
			if matchedAcc, ok := accMap[tr.AccountId]; ok {
				// if a known match exists, change it to that match
				tr.AccountId = matchedAcc

			} else {
				// else ask what the match should be, store it, and set it
				keymap := selection.NewDefaultKeyMap()
				keymap.Up = append(keymap.Up, "k")
				keymap.Down = append(keymap.Down, "j")

				fmt.Printf("Account matching '%s' doesn't match known accounts. Please select the existing account that matches\n", tr.AccountId)
				sel := selection.New("", maps.Keys(existingAccounts))
				sel.Filter = nil
				sel.KeyMap = keymap

				chosenAlias, err := sel.RunPrompt()
				if err != nil {
					fmt.Println("Import canceled")
					return nil
				}

				chosenId := existingAccounts[chosenAlias]
				accMap[tr.AccountId] = chosenId
				tr.AccountId = chosenId
			}
		}

		newTrans = append(newTrans, tr)
	}

	return newTrans
}

func buildColumnMap(records [][]string, headers bool) (map[string]int, error) {
	keymap := selection.NewDefaultKeyMap()
	keymap.Up = append(keymap.Up, "k")
	keymap.Down = append(keymap.Down, "j")

	unmatchedFields := []string{sTransactionID, sAccount, sPayee, sAmount,
		sDate, sCategory, sInstDesc, sDesc, sDir}
	colMap := make(map[string]int, 0)
	colIdx := 0
	for colIdx < len(records[0]) {
		if headers {
			fmt.Printf("->Which field does the column %s belong to?\n\t(%s, %s, %s)\n",
				records[0][colIdx],
				collapseWhitepace(records[1][colIdx]),
				collapseWhitepace(records[2][colIdx]),
				collapseWhitepace(records[3][colIdx]),
			)
		} else {
			fmt.Printf("->Which field does the column with the following values belong to?\n\t(%s, %s, %s)\n",
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
			return nil, err
		}

		if result == "Cancel" {
			fmt.Println()
			return nil, errors.New("column mapping canceled")
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

	return colMap, nil
}

// Given a slice of string tokens from an input file and a map of indicating
// which column belongs to which transaction field, build a new transaction struct
//
// NOTE: Transaction.AccountId inside the returned value is unverified, and may be
// an existing alias, id, or not exist.
func tryBuildTransaction(record []string, colMap map[string]int) (*omoney.Transaction, error) {

	var accStr string
	if accCol, ok := colMap[sAccount]; ok {
		accStr = record[accCol]
	} else {
		return nil, errors.New("missing required field 'Account'")
	}

	var payee string
	if payeeCol, ok := colMap[sPayee]; ok {
		payee = record[payeeCol]
	} else {
		return nil, errors.New("missing required field 'Payee'")
	}

	var amount float64
	if amountCol, ok := colMap[sAmount]; ok {
		var err error
		amount, err = strconv.ParseFloat(record[amountCol], 64)
		if err != nil {
			return nil, errors.New("could not parse number from 'Amount' column")
		}
	} else {
		return nil, errors.New("missing required field 'Amount'")
	}

	ops := make([]omoney.TransactionOption, 0)
	if dateCol, ok := colMap[sDate]; ok {
		date, err := dateparse.ParseLocal(record[dateCol])
		if err != nil {
			return nil, errors.New("could not parse date from 'Date' column")
		}
		ops = append(ops, omoney.WithDate(date))
	}

	if catCol, ok := colMap[sCategory]; ok {
		cat := record[catCol]
		ops = append(ops, omoney.WithCategory(cat))
	}

	if instDescCol, ok := colMap[sInstDesc]; ok {
		instDesc := collapseWhitepace(record[instDescCol])
		ops = append(ops, omoney.WithInstDescription(instDesc))
	}

	if descCol, ok := colMap[sDesc]; ok {
		desc := record[descCol]
		ops = append(ops, omoney.WithDescription(desc))
	}

	var mul float64
	if dirCol, ok := colMap[sDir]; ok {
		dir := record[dirCol]
		if dir == "debit" {
			mul = 1
		} else if dir == "credit" {
			mul = -1
		} else {
			return nil, errors.New("could not parse debit/credit from 'Direction' column")
		}
	} else {
		mul = 1
	}

	return omoney.NewTransaction(accStr, payee, amount*mul, ops...), nil
}

func mapContains(m map[string]string, val string) bool {
	for _, v := range m {
		if v == val {
			return true
		}
	}
	return false
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
