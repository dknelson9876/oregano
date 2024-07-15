package ocli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/dknelson9876/oregano/omoney"
	"github.com/dknelson9876/oregano/outil"
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

// example input: r --by payee --start 2023/06 --range month
func ListReport(input []string, model *omoney.Model) {
	// --by {category, payee}
	// --start <date>
	// --end <date>
	// --range {day, month, year}

	var startDate, endDate outil.Option[time.Time]
	var date_range, grouping string

	i := 1
	for i < len(input) {
		if strings.HasPrefix(input[i], "-") {
			switch input[i] {
			case "--by":
				if input[i+1] == "category" || input[i+1] == "payee" {
					grouping = input[i+1]
					i += 2
				} else {
					fmt.Printf("Invalid --by option: %s\n", input[i+1])
					return
				}
			case "--start":
				date, err := dateparse.ParseLocal(input[i+1])
				if err != nil {
					fmt.Printf("Failed to parse date %s\n", input[i+1])
					return
				}
				startDate.Set(date)
				i += 2
			case "--end":
				if date_range != "" {
					fmt.Println("Error: Cannot use --range and --end at the same time")
					return
				}
				date, err := dateparse.ParseLocal(input[i+1])
				if err != nil {
					fmt.Printf("Failed to parse date %s\n", input[i+1])
					return
				}
				endDate.Set(date)
				i += 2
			case "--range":
				if endDate.IsSet() {
					fmt.Println("Error: Cannot use --range and --end at the same time")
					return
				}
				if input[i+1] == "day" || input[i+1] == "month" || input[i+1] == "year" {
					date_range = input[i+1]
					i += 2
				} else {
					fmt.Printf("Invalid --range option: %s\n", input[i+1])
					return
				}
			default:
				fmt.Println("Failed to parse report command")
				return
			}
		} else {
			fmt.Println("Failed to parse report command")
			return
		}
	}
	getOps := omoney.GetSumsOptions{Grouping: grouping}

	// previous code should enforce that only one between {date_range, endDate} are set
	now := time.Now()
	var err error
	getOps.StartDate, getOps.EndDate, err = ParseDateInput(startDate, endDate, date_range, now)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	if getOps.Grouping == "" {
		getOps.Grouping = "category"
	}

	fmt.Printf("Fetching transactions from %v to %v\n", getOps.StartDate, getOps.EndDate)
	list, err := model.GetTransactionSums(getOps)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	ShowReport(list)
}

func ParseDateInput(startInput, endInput outil.Option[time.Time], date_range string, now time.Time) (time.Time, time.Time, error) {
	var startDate, endDate time.Time
	if date_range != "" {
		switch date_range {
		case "day":
			if startInput.IsSet() {
				// r --start 2023/05/02 --range day
				// -> 2023/05/02 - 2023/05/02
				startDate = startInput.StrongGet()
				endDate = startInput.StrongGet().AddDate(0, 0, 1)
			} else {
				// r --range day
				// -> 2023/05/02 - 2023/05/02
				startDate = BeginningOfDay(now)
				endDate = EndOfDay(now)
			}
		case "month":
			if startInput.IsSet() {
				// r --start 2023/02 --range month
				// -> 2023/02/01 - 2023/02/28
				startDate = startInput.StrongGet()
				endDate = startInput.StrongGet().AddDate(0, 1, 0)
			} else {
				// r --range month
				// -> 2023/05/01 - 2023/05/31
				startDate = BeginningOfMonth(now)
				endDate = EndOfMonth(now)
			}
		case "year":
			if startInput.IsSet() {
				// r --start 2023/01 --range year
				startDate = startInput.StrongGet()
				endDate = startInput.StrongGet().AddDate(1, 0, 0)
			} else {
				// r --range year
				startDate = BeginningOfYear(now)
				endDate = EndOfYear(now)
			}
		}
	} else if endInput.IsSet() {
		if startInput.IsSet() {
			startDate = startInput.StrongGet()
			endDate = endInput.StrongGet()
		} else {
			return now, now, errors.New("cannot use --end without --start")
		}
	} else {
		if startInput.IsSet() {
			startDate = startInput.StrongGet()
			endDate = EndOfMonth(startInput.StrongGet())
		} else {
			startDate = BeginningOfMonth(time.Now())
			endDate = EndOfMonth(time.Now())
		}
	}
	return startDate, endDate, nil
}

func BeginningOfDay(date time.Time) time.Time {
	// time.Truncate assumes UTC, so to recognize local time
	// must manually extract the day value
	// see https://stackoverflow.com/a/25255294/19302239
	year, month, day := date.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, date.Location())
}

func EndOfDay(date time.Time) time.Time {
	year, month, day := date.Date()
	return time.Date(year, month, day, 23, 59, 59, 0, date.Location())
}

func BeginningOfMonth(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
}

func EndOfMonth(date time.Time) time.Time {
	// add one month, truncate to the first of that month,
	// then subtract one second
	date = date.AddDate(0, 1, 0)
	date = time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	date = date.Add(time.Second * -1)
	return date
}

func BeginningOfYear(date time.Time) time.Time {
	return time.Date(date.Year(), 1, 1, 0, 0, 0, 0, date.Location())
}

func EndOfYear(date time.Time) time.Time {
	date = date.AddDate(1, 0, 0)
	date = time.Date(date.Year(), 1, 1, 0, 0, 0, 0, date.Location())
	date = date.Add(time.Second * -1)
	return date
}

