package main

import (
	"bufio"
	"strconv"

	// "context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Xuanwo/go-locale"
	"github.com/araddon/dateparse"
	"github.com/dknelson9876/oregano/ocli"
	"github.com/dknelson9876/oregano/omoney"
	"github.com/google/shlex"
	"github.com/manifoldco/promptui"
	"github.com/plaid/plaid-go/plaid"
	"github.com/spf13/viper"
	"golang.org/x/text/language"
)

type WorkTuple struct {
	typeName string
	id       string
}

var (
	workingList []WorkTuple
	oview       *ocli.OViewPlain
	model       *omoney.Model
)

func main() {
	// disable some of the things that log prints by default
	log.SetFlags(0)
	// TODO: change log level based on command line flags
	olog := ocli.NewOLogger(ocli.Debug)
	oview = ocli.NewOViewPlain(false)

	// Establish default storage folder as ~/.config/oregano/
	dirname, _ := os.UserHomeDir()
	viper.SetDefault("oregano_dir", filepath.Join(dirname, ".config", "oregano"))

	// Allow environment variables to be used for config
	viper.SetEnvPrefix("")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.AutomaticEnv()

	// load the config dir, in case it is set by env var
	configDir := viper.GetString("oregano_dir")

	// Load config.json, either from the current directory or from
	// the directory set above
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(configDir)
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// config file not found, not really an error
		} else {
			log.Fatal(err)
		}
	}

	// Load stored tokens and aliases
	model, err = ocli.LoadModelFromDB(configDir)
	if err != nil {
		log.Fatal(err)
	} else {
		olog.Println(ocli.Debug, "Found links to institutions: ")
		for _, acc := range model.GetAccounts() {
			if acc.Alias != "" {
				olog.Printf(ocli.Debug, "-- %s\t(%s)\n", acc.Id, acc.Alias)
			} else {
				olog.Printf(ocli.Debug, "-- %s\n", acc.Id)
			}
		}
	}

	// Use helper to detect country and lang from env/config
	// countries, lang := DetectRegion()

	// Load the plaid environment from the config
	viper.SetDefault("plaid.environment", "sandbox")
	// plaidEnvStr := strings.ToLower(viper.GetString("plaid.environment"))

	// NOTE: Plaid is manually disabled here until I can get back to integrating it
	plaidDisabled := true
	// var plaidEnv plaid.Environment
	// switch plaidEnvStr {
	// case "sandbox":
	// 	plaidEnv = plaid.Sandbox
	// case "development":
	// 	plaidEnv = plaid.Development
	// default:
	// 	log.Println("Invalid plaid environment. Supported environments are 'sandbox' or 'development'")
	// 	plaidDisabled = true
	// }

	// check that the required plaid api keys are present
	if !viper.IsSet("plaid.client_id") {
		log.Println("⚠️  PLAID_CLIENT_ID not set. Plaid connected features will not work until PLAID_CLIENT_ID is set in as an envvar or in config.json.")
		plaidDisabled = true
	}
	if !viper.IsSet("plaid.secret") {
		log.Println("⚠️ PLAID_SECRET not set. Plaid connected features will not work until PLAID_SECRET is set in as an envvar or in config.json.")
		plaidDisabled = true
	}

	// Build the plaid client using their library
	// var client *plaid.APIClient
	// if !plaidDisabled {
	// 	opts := plaid.NewConfiguration()
	// 	opts.AddDefaultHeader("PLAID-CLIENT-ID", viper.GetString("plaid.client_id"))
	// 	opts.AddDefaultHeader("PLAID-SECRET", viper.GetString("plaid.secret"))
	// 	opts.UseEnvironment(plaidEnv)
	// 	client = plaid.NewAPIClient(opts)
	// }
	// ctx := context.Background()

	// ----- Begin Main Loop -----------------------------------
	reader := bufio.NewReader(os.Stdin)
	workingList = make([]WorkTuple, 0)

	fmt.Println("Welcome to Oregano, the cli budget program")
	fmt.Println("For help, use 'help' (h). To quit, use 'quit' (q)")
	for {
		fmt.Print("\x1B[32moregano >> \x1B[0m")
		var line string
		// fmt.Scanln(&line)
		line, err = reader.ReadString('\n')
		if err != nil {
			log.Println(err)
			continue
		}
		tokens, err := shlex.Split(line)
		if err != nil {
			log.Printf("Error parsing command: %s\n", err)
			continue
		}

		olog.Print(ocli.DebugDetail, tokens)
		if len(tokens) == 0 {
			continue
		}
		switch tokens[0] {
		case "h", "help":
			if len(tokens) == 2 {
				switch tokens[1] {
				case "link":
					log.Println("link - link a new institution (Opens in a new browser tab)")
					log.Println("\tOpens a new browser tab to go through the")
					log.Println("\tPlaid account linking process")
					log.Println("usage: link")
				case "ls", "list":
					log.Println("list - list accounts or transactions")
					log.Println("\tProvide an alias to list transactions under that account,")
					log.Println("\tor don't provide an alias to list all accounts")
					log.Println("usage: ls (alias) (options)")
					log.Println("\t-l\t(long) Show more details")
				case "alias":
					log.Println("alias - Assign a new alias to an account")
					log.Println("\tAssigns the alias [alias] as the alias of ")
					log.Println("\tthe account with id [id]")
					log.Println("\tFind an account's id using `ls -l` or `acc`")
					log.Println("usage: alias [id] [alias]")
				case "rm", "remove":
					log.Println("remove - Remove an account or transaction")
					log.Println("\tSpecify the working id (wid) or actual id")
					log.Println("\tof a transaction or account. If no flag")
					log.Println("\tis provided, the provided id is assumed to")
					log.Println("\tbe a wid")
					log.Println("usage: rm (options) [id]")
					log.Println("\t-w\t(working) Provided id is a working id")
					log.Println("\t-t\t(transaction) Provided id is a transaction id")
					log.Println("\t\t\t TODO not implemented")
					log.Println("\t-c\t(account) Provided id is an account id")
				case "acc", "account":
					log.Println("account - print or edit information about an account")
					log.Println("Usage: account [alias/id] (options)")
					log.Println("\t-a <amount> <date>\t(anchor) set a known amount at a time to base balance off of")
				case "trs", "transactions":
					log.Println("transactions - list transactions from a specific account")
					log.Println("usage: trs [id/alias]")
				case "import":
					log.Println("import - interactively import transactions from a CSV file")
					log.Println("usage: import [filepath]")
				case "p", "print":
					log.Println("print - print more information about a transaction from the working list")
					log.Println("usage: p [wid] (options)")
					log.Println("\t-l\t(long) Show even more details about the transaction")
				case "e", "edit":
					log.Println("edit - edit information about a transaction, by setting with")
					log.Println("\t specific flags which fields to change")
					log.Println("usage: e [wid] (options)")
					log.Println("\t--account <account>")
					log.Println("\t--payee <payee>")
					log.Println("\t--amount <amount>")
					log.Println("\t--date <date>")
					log.Println("\t--category <category>")
					log.Println("\t--desc <desc>")
				case "new":
					log.Println("new - manually create account or transaction")
					log.Println("* new account [alias] [type]\t\tcreate a new manual account")
					log.Println("* new transaction []...\t\t TODO")
				}
				continue
			}
			log.Println("oregano-cli - Terminal budgeting app" +
				"Commands:\n" +
				"* help (h)\t\tPrint this menu\n" +
				"* quit (q)\t\tQuit oregano\n" +
				"* link\t\t\tLink a new institution (Opens in a new browser tab)\n" +
				"* list (ls)\t\tList accounts or transactions\n" +
				"* alias [id] [alias]\tAssign [alias] as the new alias for [id]\n" +
				"* remove (rm) [alias/id...]\tRemove a linked institution\n" +
				"* account (acc) [alias/id...]\tPrint details about specific account(s)\n" +
				"* transactions (trs) [alias/id]\t TODO description\n" +
				"* import [filename]\t TODO description\n" +
				"* print (p) [argument index]\tPrint more details about something that was output\n" +
				"* repair\t\tUsing higher level data as authoritative, correct inconsistencies\n" +
				"* new ...\t\tmanually create account or transaction")
		case "q", "quit":
			return
		case "link":
			if plaidDisabled {
				log.Println("Link has been manually disabled")
				// linkNewInstitution(model, client, countries, lang)
			} else {
				log.Println("link is unavailable while Plaid integration is disabled")
			}
		case "list", "ls":
			listCmd(tokens)
		case "alias":
			aliasCmd(tokens)
		case "remove", "rm":
			removeCmd(tokens)
		case "account", "acc":
			accountCmd(tokens)
		case "transactions", "trs":
			transactionsCmd(tokens)
		case "import":
			importCmd(tokens)
		case "print", "p":
			printCmd(tokens)
		case "edit", "e":
			editCmd(tokens)
		case "repair":
			model.RepairAccounts()
		case "new":
			newCmd(tokens)
		default:
			log.Println("Unrecognized command. Type 'help' for valid commands")
		}

	}

}

//   - if no id is provided, accounts will be listed
//   - if a string id is provided, it will be assumed to be an account alias
//     and transactions in that account will be printed
func listCmd(tokens []string) {
	if len(tokens) == 1 {
		oview.ShowAccounts(model, ocli.ShowAccountOptions{ShowType: true})
		return
	} else if len(tokens) > 1 {
		// Require that if an account is provided, it is the first argument
		//TODO do I prevent account aliases from starting with '-'?
		if strings.HasPrefix(tokens[1], "-") {
			// straight to flags without account name -> list accounts
			// ls (-l)

			long := false
			if tokens[1] == "-l" {
				long = true
			} else {
				log.Println("Error: unknown flag")
				return
			}

			ops := ocli.ShowAccountOptions{ShowType: true}
			if long {
				ops.ShowId = true
				ops.ShowAnchor = true
			}
			oview.ShowAccounts(model, ops)
		} else {
			// has account name -> list transactions
			// ls BOA (-l) (--start/end <date>) (--num <n>)
			// TODO this has a lot of code in common with transactionsCmd

			list := ocli.ListTransactions(tokens, model, len(workingList))

			for i := range list {
				workingList = append(workingList, WorkTuple{"transaction", list[i].Id})
			}
		}

	}
}

func aliasCmd(tokens []string) {
	validFlags := map[string]int{
		"<>": 2,
	}

	flags, err := ocli.ParseTokensToFlags(tokens, validFlags)
	if err != nil {
		log.Println("Fail to parse 'alias' command")
		log.Println("Usage: alias [id] [alias]")
		log.Println("Use 'help alias' for details")
		return
	}

	err = model.SetAlias(flags["<>"][0], flags["<>"][1])
	if err != nil {
		log.Printf("Error: %s\n", err)
	}

}

func removeCmd(tokens []string) {
	validFlags := map[string]int{
		"<>": 1,
		"-t": 0,
		"-w": 0,
		"-c": 0,
	}

	flags, err := ocli.ParseTokensToFlags(tokens, validFlags)
	if err != nil {
		log.Println("Fail to parse 'rm' command")
		log.Println("Usage: rm [alias/id]")
		log.Println("Use 'help remove' for details")
		return
	}

	flag := "working"
	var tr *omoney.Transaction
	var acc *omoney.Account

	if _, ok := flags["-c"]; ok {
		flag = "account"
	} else if _, ok := flags["-t"]; ok {
		flag = "transaction"
		log.Println("ERROR: Removing by transaction id not yet implemented")
		return
	}

	input := flags["<>"][0]
	if flag == "working" {
		item, err := fromWorkingList(input)
		if err != nil {
			log.Printf("Could not find item in working list.\nError: %s\n", err)
			return
		}

		if transaction, ok := item.(omoney.Transaction); ok {
			log.Printf("Pulled transaction of amount %.2f from working list\n", transaction.Amount)
			tr = &transaction
		} else if account, ok := item.(omoney.Account); ok {
			log.Printf("Pulled account %s from working list\n", account.Alias)
			acc = &account
		} else {
			log.Println("Failed to recognize type of item from working list")
			return
		}
	}

	if flag == "working" && tr != nil {
		log.Printf("Removing transaction %s\n", input)
		err = model.RemoveTransaction(tr)

	} else if flag == "account" || (flag == "working" && acc != nil) {
		log.Printf("Removing account %s\n", input)
		if acc != nil {
			input = acc.Id
		}
		err = model.RemoveAccount(input)
	}

	if err != nil {
		log.Println(err)
	}

}

func accountCmd(tokens []string) {
	validFlags := map[string]int{
		"<>": 1,
		"-a": 2,
	}

	flags, err := ocli.ParseTokensToFlags(tokens, validFlags)
	if err != nil {
		log.Println("Fail to parse 'acc' command")
		log.Println("Usage: acc [alias/id]")
		log.Println("Use 'help account' for details")
		return
	}

	input := flags["<>"][0]
	if anchor, ok := flags["-a"]; ok {
		err = model.SetAnchor(input, anchor)
		if err != nil {
			log.Printf("Error: %s\n", err)
		} else {
			acc, _ := model.GetAccount(input)
			log.Printf("Updated anchor to $%.2f on %s", acc.AnchorBalance, acc.AnchorTime.Format("2006/01/02"))
		}
		return
	}

	acc, err := model.GetAccount(input)
	if err != nil {
		log.Printf("Error: %s\n", err)
		return
	}
	if acc.PlaidToken == "" {
		// account was manually created
		oview.ShowAccount(acc)
	} else {
		log.Println("Details about Plaid accounts has been manually disabled")
		// Not sure where to move this code for now
		// I guess I might end up making some kind of
		// Plaid-compatibility layer for snippets like this
		// resp, _, err := client.PlaidApi.AccountsGet(ctx).AccountsGetRequest(
		// 	*plaid.NewAccountsGetRequest(token),
		// ).Execute()
		// if err != nil {
		// 	log.Println(err)
		// 	return
		// }
		// oview.ShowPlaidAccounts(resp.GetAccounts())
	}

}

func transactionsCmd(tokens []string) {
	list := ocli.ListTransactions(tokens, model, len(workingList))

	for i := range list {
		workingList = append(workingList, WorkTuple{"transaction", list[i].Id})
	}

}

func importCmd(tokens []string) {
	validFlags := map[string]int{
		"<>": 1,
	}

	flags, err := ocli.ParseTokensToFlags(tokens, validFlags)
	if err != nil {
		log.Println("Fail to parse 'import' command")
		log.Println("Usage: import [filename]")
		log.Println("Use 'help import' for details")
		return
	}

	input := flags["<>"][0]
	newTrans := ocli.ReadCsv(input, model.GetAliases())
	for _, tr := range newTrans {
		model.AddTransaction(tr)
	}

}

func printCmd(tokens []string) {
	// -l	long (detailed)
	validFlags := map[string]int{
		"<>": 1,
		"-l": 0,
	}

	flags, err := ocli.ParseTokensToFlags(tokens, validFlags)
	if err != nil {
		log.Println("Fail to parse 'import' command")
		log.Println("Usage: import [filename]")
		log.Println("Use 'help import' for details")
		return
	}

	_, long := flags["-l"]
	arg := flags["<>"][0]

	v, err := fromWorkingList(arg)
	if err != nil {
		log.Printf("Error: %s\n", err)
		return
	}

	switch t := v.(type) {
	default:
		fmt.Printf("%v\n", v)
	case omoney.Transaction:
		ops := ocli.ShowTransactionOptions{}
		if long {
			ops.ShowId = true
			ops.ShowCategory = true
			ops.ShowInstDesc = true
			ops.ShowDesc = true
		}
		oview.ShowTransaction(t, ops)
	}

}

// e <wid> (--account/--payee/--amount/--date/--category/--desc)
func editCmd(tokens []string) {
	if len(tokens) < 2 {
		log.Println("Error: not enough arguments")
		return
	}

	v, err := fromWorkingList(tokens[1])
	if err != nil {
		log.Printf("Error: %s\n", err)
		return
	}
	tr, ok := v.(omoney.Transaction)
	if !ok {
		log.Println("Error: wid does not point to a transaction")
		return
	}

	ops := make([]omoney.UpdateTransactionOptions, 0)
	i := 2
	for i < len(tokens)-1 {
		switch tokens[i] {
		case "--account":
			acc, err := model.GetAccount(tokens[i+1])
			if err != nil {
				log.Printf("Error: %s\n", err)
				return
			}
			ops = append(ops, omoney.WithAccountUpdate(acc.Id))
			i += 2
		case "--payee":
			ops = append(ops, omoney.WithPayeeUpdate(tokens[i+1]))
			i += 2
		case "--amount":
			amount, err := strconv.ParseFloat(tokens[i+1], 64)
			if err != nil {
				log.Println("Error: failed to parse amount")
				return
			}
			ops = append(ops, omoney.WithAmountUpdate(amount))
			i += 2
		case "--date":
			date, err := dateparse.ParseLocal(tokens[i+1])
			if err != nil {
				log.Println("Error: failed to parse date")
				return
			}
			ops = append(ops, omoney.WithDateUpdate(date))
			i += 2
		case "--category":
			ops = append(ops, omoney.WithCategoryUpdate(tokens[i+1]))
			i += 2
		case "--desc":
			ops = append(ops, omoney.WithDescUpdate(tokens[i+1]))
			i += 2
		}
	}

	err = model.UpdateTransaction(tr.Id, ops...)
	if err != nil {
		log.Printf("Error: %s\n", err)
	}
}

func newCmd(tokens []string) {
	if len(tokens) < 2 {
		log.Printf("Error: command 'new' requires more arguments\n")
		return
	}

	// trim 'new' off front of cmd
	tokens = tokens[1:]

	switch tokens[0] {
	case "account", "acc":
		// new acc [alias] [type]
		validFlags := map[string]int{
			"<>": 2,
		}

		flags, err := ocli.ParseTokensToFlags(tokens, validFlags)
		if err != nil {
			log.Println("Fail to parse 'new acc' command")
			log.Println("Usage: new acc [alias] [type]")
			log.Println("Use 'help new' for details")
			return
		}

		acc := ocli.CreateManualAccount(flags["<>"])
		if acc == nil {
			log.Println("Error making new manual account")
			return
		}
		model.AddAccount(*acc)
	case "transaction", "tr":
		// new tr [acc] [payee] [amount] (date) (cat)
		//      (desc) (-t/--time date) (-c/--category cat) (-d/--description desc)
		if len(tokens) < 3 {
			log.Printf("Error: command 'new transaction' requires more arguments\n")
			log.Printf("Usage: new tr [account] [payee] [amount] (date) (category) (desc)")
			return
		}

		// if account is valid alias, replace with ID
		// else if account is not valid id, error
		if model.IsValidAccountAlias(tokens[1]) {
			alias := tokens[1]
			tokens[1] = model.GetAccountId(tokens[1])
			log.Printf("converted alias %s to accid %s\n", alias, tokens[1])
		} else if !model.IsValidAccountId(tokens[1]) {
			log.Printf("Error: %s is not a valid account alias or id\n", tokens[1])
			return
		} else {
			log.Printf("Continuing with provided account id %s\n", tokens[1])
		}

		// trim 'tr' off front of cmd
		tokens = tokens[1:]

		tr := ocli.CreateManualTransaction(tokens)
		if tr == nil {
			log.Println("Error making new manual transaction")
			return
		}

		// transaction is purposely handled by model and not
		// account because I intend to later add an always
		// up to date budget model
		model.AddTransaction(tr)
		log.Printf("Saving new transaction %+v\n", tr)
	default:
		log.Printf("Error: unknown subcommand %s\n", tokens[1])
		log.Println("Valid subcommands are: account, transaction")
	}
}

func linkNewInstitution(model *omoney.Model, client *plaid.APIClient, countries []string, lang string) {
	// Build a linker struct to run Plaid Link
	linker := ocli.NewLinker(client, countries, lang)
	port := viper.GetString("link.port")

	// Attempt to run link in a browser window
	var tokenPair *ocli.TokenPair
	tokenPair, err := linker.Link(port)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Institution linked!")
	log.Printf("Item ID: %s\n", tokenPair.ItemID)

	acc := omoney.NewAccount(
		omoney.WithPlaidIds(tokenPair.ItemID, tokenPair.AccessToken),
	)
	//TODO: Pull AccountType from Plaid

	log.Printf("Default alias in struct: %s\n", acc.Alias)

	prompt := promptui.Prompt{
		Label: "Provide an alias to use for this institution: (default: none)",
		Validate: func(input string) error {
			matched, err := regexp.Match(`^\w+$`, []byte(input))
			if err != nil {
				return err
			}

			if !matched && input != "" {
				return errors.New("alias must contain only letters, numbers, or underscore")
			}

			// if _, ok := model.Aliases[input]; ok {
			// 	return errors.New("that alias is already in use")
			// }
			return nil
		},
	}

	input, err := prompt.Run()
	if err != nil {
		log.Fatalln(err)
	}
	if input != "" {
		acc.Alias = input
	}

	// Store the long term access token from plaid
	model.AddAccount(*acc)
}

func fromWorkingList(input string) (interface{}, error) {
	i, err := strconv.Atoi(input)
	if err != nil || i >= len(workingList) {
		return nil, fmt.Errorf("%s is not a valid wid", input)
	}

	pair := workingList[i]
	if pair.typeName == "transaction" {
		return model.GetTransactionById(pair.id)
	} else if pair.typeName == "account" {
		return model.GetAccount(pair.id)
	}

	return nil, fmt.Errorf("typename from working list %s not recognized", pair.typeName)

}

func DetectRegion() ([]string, string) {
	// A whole bunch of code to safely detect country and language
	tag, err := locale.Detect()
	if err != nil {
		tag = language.AmericanEnglish
	}

	region, _ := tag.Region()
	base, _ := tag.Base()

	var country string
	if region.IsCountry() {
		country = region.String()
	} else {
		country = "US"
	}

	lang := base.String()

	viper.SetDefault("plaid.countries", []string{country})
	countriesOpt := viper.GetStringSlice("plaid.countries")
	var countries []string
	for _, c := range countriesOpt {
		countries = append(countries, strings.ToUpper(c))
	}

	viper.SetDefault("plaid.language", lang)
	lang = viper.GetString("plaid.language")

	if !AreValidCountries(countries) {
		log.Fatalln("⚠️  Invalid countries. Please configure `plaid.countries` (using an envvar, PLAID_COUNTRIES, or in oregano's config file) to a subset of countries that Plaid supports. Plaid supports the following countries: ", plaidSupportedCountries)
	}

	if !IsValidLanguageCode(lang) {
		log.Fatalln("⚠️  Invalid language code. Please configure `plaid.language` (using an envvar, PLAID_LANGUAGE, or in oregano's config file) to a language that Plaid supports. Plaid supports the following languages: ", plaidSupportedLanguages)
	}
	return countries, lang
}

// See https://plaid.com/docs/link/customization/#language-and-country
var plaidSupportedCountries = []string{"US", "CA", "GB", "IE", "ES", "FR", "NL"}
var plaidSupportedLanguages = []string{"en", "fr", "es", "nl"}

func AreValidCountries(countries []string) bool {
	supportedCountries := sliceToMap(plaidSupportedCountries)
	for _, c := range countries {
		if !supportedCountries[c] {
			return false
		}
	}

	return true
}

func IsValidLanguageCode(lang string) bool {
	supportedLanguages := sliceToMap(plaidSupportedLanguages)
	return supportedLanguages[lang]
}

func sliceToMap(slice []string) map[string]bool {
	set := make(map[string]bool, len(slice))
	for _, s := range slice {
		set[s] = true
	}
	return set
}
