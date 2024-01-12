package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Xuanwo/go-locale"
	"github.com/dknelson9876/oregano/ocli"
	"github.com/manifoldco/promptui"
	"github.com/plaid/plaid-go/plaid"
	"github.com/spf13/viper"
	"golang.org/x/text/language"
	// ui "github.com/gizak/termui/v3"
	// "github.com/gizak/termui/v3/widgets"
)

func main() {
	// disable some of the things that log prints by default
	log.SetFlags(0)

	// Wait for new line to take real action
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Press Enter to try launching Link...")
	reader.ReadString('\n')

	// Establish where to store data as ~/.oregano/
	usr, _ := user.Current()
	dir := usr.HomeDir
	viper.SetDefault("oregano.data_dir", filepath.Join(dir, ".config", "oregano"))

	// Load stored tokens and aliases
	dataDir := viper.GetString("oregano.data_dir")
	data, err := ocli.LoadData(dataDir)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Found links to institutions: ")
		for itemID := range data.Tokens {
			if alias, ok := data.Aliases[itemID]; ok {
				log.Printf("-- %s\t(%s)\n", itemID, alias)
			} else {
				log.Printf("-- %s\n", itemID)
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

	// ----- Begin Main Loop -----------------------------------
	fmt.Println("Welcome to Oregano, the cli budget program")
	fmt.Println("For help, use 'help' (h). To quit, use 'quit' (q)")
	for {
		fmt.Print("oregano >>")
		var input string
		fmt.Scanln(&input)

		switch input {
		case "h", "help":
			fmt.Println("oregano-cli - Terminal budgeting app" +
				"Commands:\n" +
				"* help\t[h]\tPrint this menu\n" +
				"* quit\t[q]\tQuit oregano\n" +
				"* link\t\tLink a new institution\n" +
				"* list\t[ls]\tList linked institutions")
		case "q", "quit":
			fmt.Println("goodbye")
			return
		case "link":
			linkNewInstitution(data, client, countries, lang)
		case "list", "ls":
			fmt.Println("Institutions:")
			for itemID := range data.Tokens {
				if alias, ok := data.Aliases[itemID]; ok {
					log.Printf("-- %s\t(%s)\n", itemID, alias)
				} else {
					log.Printf("-- %s\n", itemID)
				}
			}
		default:
			fmt.Println("Unrecognized command. Type 'help' for valid commands")
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
		err = SetAlias(data, tokenPair.ItemID, input)
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

func SetAlias(data *ocli.Data, itemID string, alias string) error {
	if _, ok := data.Tokens[itemID]; !ok {
		return errors.New(fmt.Sprintf("No access token found for item ID `%s`. Try linking again.", itemID))
	}

	data.Aliases[alias] = itemID
	data.BackAliases[itemID] = alias
	err := data.Save()
	if err != nil {
		return err
	}

	log.Println(fmt.Sprintf("Aliased %s to %s", itemID, alias))

	return nil
}
