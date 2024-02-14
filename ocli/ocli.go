package ocli

import (
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
