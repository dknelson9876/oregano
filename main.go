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
	"github.com/dknelson9876/oregano/ocli"
	"github.com/dknelson9876/oregano/omoney"
	"github.com/google/shlex"
	"github.com/manifoldco/promptui"
	"github.com/plaid/plaid-go/plaid"
	"github.com/spf13/viper"
	"golang.org/x/exp/maps"
	"golang.org/x/text/language"
)

func main() {
	// disable some of the things that log prints by default
	log.SetFlags(0)
	// TODO: change log level based on command line flags
	olog := ocli.NewOLogger(ocli.DebugDetail)
	oview := ocli.NewOViewPlain(false)

	// Wait for new line to take real action
	reader := bufio.NewReader(os.Stdin)
	// fmt.Println("Press Enter to try launching Link...")
	// reader.ReadString('\n')

	// Establish where to store data as ~/.oregano/
	dirname, _ := os.UserHomeDir()
	viper.SetDefault("oregano.data_dir", filepath.Join(dirname, ".config", "oregano"))

	// Load stored tokens and aliases
	dataDir := viper.GetString("oregano.data_dir")
	model, err := ocli.LoadModel(dataDir)
	if err != nil {
		log.Fatal(err)
	} else {
		olog.Println(ocli.Debug, "Found links to institutions: ")
		for id, acc := range model.Accounts {
			if acc.Alias != "" {
				olog.Printf(ocli.Debug, "-- %s\t(%s)\n", id, acc.Alias)
			} else {
				olog.Printf(ocli.Debug, "-- %s\n", id)
			}
		}
	}

	// Load config.json, preferring the current directory, but if not check ~/.oregano
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(dataDir)
	viper.AddConfigPath(".")
	err = viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// config file not found, not really an error
		} else {
			log.Fatal(err)
		}
	}

	// Allow normal environment variables to be used in place of config.json
	viper.SetEnvPrefix("")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.AutomaticEnv()

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
	workingList := make([]interface{}, 0)

	fmt.Println("Welcome to Oregano, the cli budget program")
	fmt.Println("For help, use 'help' (h). To quit, use 'quit' (q)")
	for {
		fmt.Print("\x1B[32moregano >> \x1B[0m")
		var line string
		// fmt.Scanln(&line)
		line, err = reader.ReadString('\n')
		// tokens, err := shlex.Split(line)
		tokens := strings.Fields(line)
		if err != nil {
			log.Println(err)
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
				case "new":
					log.Println("new - manually create account or transaction\n" +
						"* new account [alias] [type]\t\tcreate a new manual account\n" +
						"* new transaction []...\t\t TODO")
					continue
				case "ls", "list":
					log.Println("ls - list all known accounts")
					log.Println("usage: ls (options)")
					log.Println("\t-l\t(long) Show more details about each account")
				}
			}
			log.Println("oregano-cli - Terminal budgeting app" +
				"Commands:\n" +
				"* help (h)\t\tPrint this menu\n" +
				"* quit (q)\t\tQuit oregano\n" +
				"* link\t\t\tLink a new institution (Opens in a new browser tab)\n" +
				"* list (ls)\t\tList linked institutions\n" +
				"* alias [token] [alias]\tAssign [alias] as the new alias for [token]\n" +
				"* remove (rm) [alias/id...]\tRemove a linked institution\n" +
				"* account (acc) [alias/id...]\tPrint details about specific account(s)\n" +
				"* print (p) [argument index]\tPrint more details about something that was output\n" +
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
			ops := ocli.ShowAccountOptions{ShowType: true}
			if len(tokens) > 1 {
				validFlags := map[string]int {
					"-l": 0, // long (detailed)
				}

				flags, err := ocli.ParseTokensToFlags(tokens[1:], validFlags)
				if err != nil {
					log.Printf("Error: %s\n", err)
				}

				if _, ok := flags["-l"]; ok {
					ops.ShowId = true
					ops.ShowAnchor = true
				}
			}

			oview.ShowAccounts(maps.Values(model.Accounts), ops)

		case "alias":
			if len(tokens) != 3 {
				log.Println("Error: alias requires exactly 2 arguments")
				log.Println("Usage: alias [token] [alias]")
			} else {
				err = model.SetAlias(tokens[1], tokens[2])
				if err != nil {
					log.Printf("Error: %s\n", err)
				} else {
					ocli.Save(model)
				}
			}
		case "remove", "rm":
			for _, input := range tokens[1:] {
				log.Printf("Removing institution %s\n", input)
				err = model.RemoveAcount(input)
				if err != nil {
					log.Println(err)
				} else {
					ocli.Save(model)
				}
			}
		case "account", "acc":
			for _, input := range tokens[1:] {
				acc, err := model.GetAccount(input)
				if err != nil {
					log.Printf("Error: %s\n", err)
					continue
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
					// 	continue
					// }
					// oview.ShowPlaidAccounts(resp.GetAccounts())
				}
			}

		case "transactions", "trs":
			acc, err := model.GetAccount(tokens[1])
			if err != nil {
				log.Fatalln(err)
			}

			var sl []omoney.Transaction
			if len(acc.Transactions) > 10 {
				sl = acc.Transactions[:10]
			} else {
				sl = acc.Transactions
			}

			invert := acc.Type != omoney.CreditCard
			oview.ShowTransactions(sl, invert, len(workingList))

			for i := range sl {
				workingList = append(workingList, &sl[i])
			}

		case "import":
			newTrans := ocli.ReadCsv(tokens[1], maps.Keys(model.Aliases))
			for _, tr := range newTrans {
				model.AddTransaction(*tr)
			}
			err = ocli.Save(model)
			if err != nil {
				log.Fatalln(err)
			}

		case "print", "p":
			// -l	long (detailed)
			if len(tokens) < 2 {
				fmt.Println("Error: 'print' requires an argument")
				continue
			}
			long := false
			arg := ""
			i := 1
			for i < len(tokens) {
				if strings.HasPrefix(tokens[i], "-") {
					switch tokens[i] {
					case "-l":
						long = true
						i++
					}
				} else {
					arg = tokens[i]
					i++
				}
			}
			idx, err := strconv.Atoi(arg)
			if err != nil {
				log.Printf("Error: parsed token %s as arg, but could not parse it into an index\n", arg)
				continue
			}

			if idx < len(workingList) {
				switch v := workingList[idx].(type) {
				default:
					fmt.Printf("%v\n", v)
				case *omoney.Transaction:
					ops := ocli.ShowTransactionOptions{}
					if long {
						ops.ShowId = true
						ops.ShowCategory = true
						ops.ShowInstDesc = true
						ops.ShowDesc = true
					}
					oview.ShowTransaction(*v, ops)
				}
			}

		case "new":
			if len(tokens) < 2 {
				log.Printf("Error: command 'new' requires more arguments\n")
				continue
			}
			switch tokens[1] {
			case "account", "acc":
				acc := ocli.CreateManualAccount(tokens[2:])
				if acc == nil {
					log.Println("Error making new manual account")
					continue
				}
				model.AddAccount(*acc)
				err = ocli.Save(model)
				if err != nil {
					log.Fatalln(err)
				}
			case "transaction", "tr":
				// new tr [acc] [payee] [amount] (date) (cat)
				//      (desc) (-t/--time date) (-c/--category cat) (-d/--description desc)
				if len(tokens) < 3 {
					log.Printf("Error: command 'new transaction' requires more arguments\n")
					log.Printf("Usage: new tr [account] [payee] [amount] (date) (category) (desc)")
					continue
				}
				if !model.IsValidAccount(tokens[2]) {
					log.Printf("Error: %s\n", err)
					continue
				}

				str, _ := shlex.Split(strings.Join(tokens[2:], " "))
				tr := ocli.CreateManualTransaction(str)
				if tr == nil {
					log.Println("Error making new manual transaction")
					continue
				}

				// transaction is purposely handled by model and not
				// account because I intend to later add an always
				// up to date budget model
				model.AddTransaction(*tr)
				log.Printf("Saving new transaction %+v\n", tr)
				err = ocli.Save(model)
				if err != nil {
					log.Fatalln(err)
				}
			default:
				log.Printf("Error: unknown subcommand %s\n", tokens[1])
				log.Println("Valid subcommands are: account, transaction")
			}
		default:
			log.Println("Unrecognized command. Type 'help' for valid commands")
		}

	}

	//-------
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

			if _, ok := model.Aliases[input]; ok {
				return errors.New("that alias is already in use")
			}
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
	err = ocli.Save(model)
	if err != nil {
		log.Fatalln(err)
	}

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
