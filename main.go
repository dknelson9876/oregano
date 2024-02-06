package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Xuanwo/go-locale"
	"github.com/dknelson9876/oregano/ocli"
	"github.com/manifoldco/promptui"
	"github.com/plaid/plaid-go/plaid"
	"github.com/spf13/viper"
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
	data, err := ocli.LoadData(dataDir)
	if err != nil {
		log.Fatal(err)
	} else {
		olog.Println(ocli.Debug, "Found links to institutions: ")
		for itemID := range data.Tokens {
			if alias, ok := data.BackAliases[itemID]; ok {
				olog.Printf(ocli.Debug, "-- %s\t(%s)\n", itemID, alias)
			} else {
				olog.Printf(ocli.Debug, "-- %s\n", itemID)
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
	countries, lang := DetectRegion()

	// Load the plaid environment from the config
	viper.SetDefault("plaid.environment", "sandbox")
	plaidEnvStr := strings.ToLower(viper.GetString("plaid.environment"))

	var plaidEnv plaid.Environment
	switch plaidEnvStr {
	case "sandbox":
		plaidEnv = plaid.Sandbox
	case "development":
		plaidEnv = plaid.Development
	default:
		log.Fatalln("Invalid plaid environment. Supported environments are 'sandbox' or 'development'")
	}

	// check that the required plaid api keys are present
	if !viper.IsSet("plaid.client_id") {
		log.Println("⚠️  PLAID_CLIENT_ID not set. Please set in as an envvar or in config.json.")
		os.Exit(1)
	}
	if !viper.IsSet("plaid.secret") {
		log.Println("⚠️ PLAID_SECRET not set. Please set in as an envvar or in config.json.")
		os.Exit(1)
	}

	// Build the plaid client using their library
	opts := plaid.NewConfiguration()
	opts.AddDefaultHeader("PLAID-CLIENT-ID", viper.GetString("plaid.client_id"))
	opts.AddDefaultHeader("PLAID-SECRET", viper.GetString("plaid.secret"))
	opts.UseEnvironment(plaidEnv)
	client := plaid.NewAPIClient(opts)
	ctx := context.Background()

	// ----- Begin Main Loop -----------------------------------
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
		switch tokens[0] {
		case "h", "help":
			log.Println("oregano-cli - Terminal budgeting app" +
				"Commands:\n" +
				"* help (h)\t\tPrint this menu\n" +
				"* quit (q)\t\tQuit oregano\n" +
				"* link\t\t\tLink a new institution (Opens in a new browser tab)\n" +
				"* list (ls)\t\tList linked institutions\n" +
				"* remove (rm) [alias/id...]\tRemove a linked institution\n" +
				"* alias [token] [alias]\tAssign [alias] as the new alias for [token]" +
				"* accounts (acc) [alias/id...]\tShow account information for the institution under [alias]" +
				"* transactions (trsn) [args]\tFetch transactions for [account]")
		case "q", "quit":
			return
		case "link":
			linkNewInstitution(data, client, countries, lang)
		case "list", "ls":
			log.Println("Institutions:")
			for itemID := range data.Tokens {
				if alias, ok := data.BackAliases[itemID]; ok {
					log.Printf("-- %s\t(%s)\n", itemID, alias)
				} else {
					log.Printf("-- %s\n", itemID)
				}
			}
		case "alias":
			if len(tokens) != 3 {
				log.Println("Error: alias requires exactly 2 arguments")
				log.Println("Usage: alias [token] [alias]")
			} else {
				err = data.SetAlias(tokens[1], tokens[2])
				if err != nil {
					log.Printf("Error: %s\n", err)
				}
			}
		case "remove", "rm":
			for _, input := range tokens[1:] {
				log.Printf("Removing institution %s\n", input)
				err = data.RemoveToken(input)
				if err != nil {
					log.Println(err)
				}
			}
		case "accounts", "acc":
			for _, input := range tokens[1:] {
				token, err := data.GetAccessToken(input)
				if err != nil {
					log.Printf("Error: %s\n", err)
					continue
				}
				resp, _, err := client.PlaidApi.AccountsGet(ctx).AccountsGetRequest(
					*plaid.NewAccountsGetRequest(token),
				).Execute()
				if err != nil {
					log.Println(err)
					continue
				}
				oview.ShowAccounts(resp.GetAccounts())
			}
		case "transactions", "trsn":
			token, err := data.GetAccessToken(tokens[1])
			if err != nil {
				log.Printf("Error: %s\n", err)
				continue
			}
			showTransactions(client, token)
		default:
			log.Println("Unrecognized command. Type 'help' for valid commands")
		}

	}

	//-------
}

func linkNewInstitution(data *ocli.Data, client *plaid.APIClient, countries []string, lang string) {
	// Build a linker struct to run Plaid Link
	linker := ocli.NewLinker(data, client, countries, lang)
	port := viper.GetString("link.port")

	// Attempt to run link in a browser window
	var tokenPair *ocli.TokenPair
	tokenPair, err := linker.Link(port)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Institution linked!")
	log.Printf("Item ID: %s\n", tokenPair.ItemID)

	// Store the long term access token from plaid
	data.Tokens[tokenPair.ItemID] = tokenPair.AccessToken
	err = data.Save()
	if err != nil {
		log.Fatalln(err)
	}

	// If an alias already exists, print it
	if alias, ok := data.BackAliases[tokenPair.ItemID]; ok {
		log.Printf("Existing Alias: %s\n", alias)
		return
	}

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

			return nil
		},
	}

	input, err := prompt.Run()
	if err != nil {
		log.Fatalln(err)
	}
	if input != "" {
		err = data.SetAlias(tokenPair.ItemID, input)
		if err != nil {
			log.Fatalln(err)
		}
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

func showTransactions(client *plaid.APIClient, token string) error {
	ctx := context.Background()

	// Set cursor to empty to receive all historical updates
	var cursor *string

	// New transaction updates since "cursor"
	var added []plaid.Transaction
	var modified []plaid.Transaction
	var removed []plaid.RemovedTransaction
	hasMore := true
	// Iterate through each page until no mroe pages
	for hasMore {
		request := plaid.NewTransactionsSyncRequest(token)
		if cursor != nil {
			request.SetCursor(*cursor)
		}
		resp, _, err := client.PlaidApi.TransactionsSync(
			ctx,
		).TransactionsSyncRequest(*request).Execute()
		if err != nil {
			return err
		}

		added = append(added, resp.GetAdded()...)
		modified = append(modified, resp.GetModified()...)
		removed = append(removed, resp.GetRemoved()...)
		hasMore = resp.GetHasMore()
		nextCursor := resp.GetNextCursor()
		cursor = &nextCursor

		// data, _ := added[len(added)-1].MarshalJSON()
		// fmt.Println(string(data))
	}

	sort.Slice(added, func(i, j int) bool {
		return added[i].GetDate() < added[j].GetDate()
	})
	latestTransactions := added[len(added)-9:]

	// fmt.Println(latestTransactions.)
	for _, t := range latestTransactions {
		// fmt.Println(t.MarshalJSON())
		showTransaction(t)
	}

	return nil
}

func showTransaction(t plaid.Transaction) {
	fmt.Print("-- ")

	fmt.Printf("(%s) ", t.Date)
	fmt.Printf("%s - ", t.AccountId)
	fmt.Printf("$%.2f / ", t.Amount)
	fmt.Printf("%s - ", t.Name)
	fmt.Println()


	// fmt.Printf("\t\t%s\n", t.Location.)
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
