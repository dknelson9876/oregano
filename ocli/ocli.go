package ocli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/dknelson9876/oregano/omoney"
)

func CreateManualAccount(input []string) *omoney.Account {
	if len(input) < 2 {
		fmt.Println("Error: new account requires exactly 2 arguments")
		fmt.Println("Usage: new account [alias] [type]")
		return nil
	}
	alias := input[0]
	accType, err := omoney.ParseAccountType(input[1])
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return nil
	}

	return omoney.NewAccount(
		omoney.WithAlias(alias),
		omoney.WithAccountType(accType),
	)
}

func CreateManualTransaction(input []string) *omoney.Transaction {
	// new tr [acc] [payee] [amount] (date) (cat)
	//      (desc) (-t/--time date) (-c/--category cat) (-d/--description desc)

	acc := input[0]

	payee := input[1]

	amount, err := strconv.ParseFloat(input[2], 64)
	if err != nil {
		fmt.Printf("Error: Unable to parse amount %s\n", input[1])
		return nil
	}

	var date time.Time
	dateFound := false
	cat := ""
	desc := ""
	endPositional := false
	i := 3
	for i < len(input) {
		if strings.HasPrefix(input[i], "-") {
			endPositional = true
			if i+1 == len(input) {
				fmt.Println("Error: Found flag with no value at end of command")
				return nil
			}
			switch input[i] {
			case "-t", "--time":
				date, err = dateparse.ParseLocal(input[i+1])
				if err != nil {
					fmt.Printf("Error: Unable to parse datetime %s\n", input[i+1])
					return nil
				}
				dateFound = true
			case "-c", "--category":
				cat = input[i+1]
			case "-d", "--description":
				desc = input[i+1]
			default:
				fmt.Printf("Error: Unrecognized flag %s\n", input[i])
				return nil
			}
			i += 2
		} else {
			if endPositional {
				// found positional arg, after using flag. Ambiguous, so fail
				fmt.Printf("Error: Positional argument after using flag. Cannot parse new transaction\n")
				return nil
			} else {
				// still on positional args
				switch i {
				case 3:
					date, err = dateparse.ParseLocal(input[3])
					if err != nil {
						fmt.Printf("Error: Unable to parse datetime %s\n", input[3])
						return nil
					}
					dateFound = true
				case 4:
					cat = input[4]
				case 5:
					desc = input[5]
				default:
					fmt.Println("ERROR: this should be unreachable, unless a missed an earlier check")
					return nil
				}
				i++
			}
		}
	}

	if !dateFound {
		date = time.Now()
	}

	return omoney.NewTransaction(acc, payee, amount,
		omoney.WithDate(date),
		omoney.WithCategory(cat),
		omoney.WithDescription(desc))
}

// format of flagMap is {<flag>: <number of arguments for flag>}
func ParseTokensToFlags(tokens []string, flagMap map[string]int) (map[string][]string, error) {
	toreturn := make(map[string][]string, len(flagMap))

	tokens = tokens[1:] // trim off command name every time
	i := 0
	for i < len(tokens) {
		if argCount, ok := flagMap[tokens[i]]; ok {
			if i+argCount > len(tokens) {
				return nil, fmt.Errorf("missing arguments for flag %s", tokens[i])
			} else {
				//TODO: check that tokens captured as args here are not
				// other flags
				toreturn[tokens[i]] = tokens[i+1 : i+1+argCount]
				i += argCount + 1
			}

		} else if argCount, ok := flagMap["<>"]; ok {
			// use "<>" as special flag for plain arguments
			if i+argCount > len(tokens) {
				return nil, fmt.Errorf("missing arguments for flag %s", tokens[i])
			} else {
				//TODO: check that tokens captured as args here are not
				// other flags
				toreturn["<>"] = tokens[i : i+argCount]
				i += argCount
			}

		} else {
			return nil, fmt.Errorf("invalid flag %s", tokens[i])
		}
	}

	if _, ok := flagMap["<>"]; ok {
		if _, ok := toreturn["<>"]; !ok {
			// if we found no plain arguments, but they are needed
			return nil, errors.New("missing required positional arguments")
		}
	}

	return toreturn, nil
}

// example input: [ls <account> -n 20 -l]
//
// return the list of transactions printed, for the working list
func ListTransactions(input []string, model *omoney.Model, workingIndex int) []omoney.Transaction {
	// -l	long: show all possible details about each transaction
	// --num	number: an integer number of transactions to print (default: 10)
	// --start	a time for the oldest transaction cutoff
	// --end    a time for the newest transaction cutoff

	acc, err := model.GetAccount(input[1])
	if err != nil {
		fmt.Printf("Alias %s not recognized\n", input[0])
		return nil
	}

	filterOps := omoney.GetTransactionsOptions{Count: 10}
	showOps := ShowTransactionOptions{}

	i := 2
	for i < len(input) {
		if strings.HasPrefix(input[i], "-") {
			switch input[i] {
			case "-l":
				showOps.ShowId = true
				showOps.ShowCategory = true
				showOps.ShowInstDesc = true
				showOps.ShowDesc = true
				i++
			case "--num":
				n, err := strconv.Atoi(input[i+1])
				if err != nil {
					fmt.Printf("Failed to parse number %s\n", input[i+1])
					return nil
				}
				filterOps.Count = n
				i += 2
			case "--start":
				date, err := dateparse.ParseLocal(input[i+1])
				if err != nil {
					fmt.Printf("Failed to parse date %s\n", input[i+1])
					return nil
				}
				filterOps.StartDate = &date
				i += 2
			case "--end":
				date, err := dateparse.ParseLocal(input[i+1])
				if err != nil {
					fmt.Printf("Failed to parse date %s\n", input[i+1])
					return nil
				}
				filterOps.EndDate = &date
				i += 2
			default:
				fmt.Printf("Unknown flag: %s\n", input[i])
				return nil
			}
		} else {
			fmt.Println("Failed to parse command")
			return nil
		}
	}

	list, err := model.GetTransactionsByAccount(acc.Id, filterOps)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return nil
	}

	invert := acc.Type != omoney.CreditCard
	ShowTransactions(list, invert, workingIndex)

	return list

}
