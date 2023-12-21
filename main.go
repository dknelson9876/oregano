package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/Xuanwo/go-locale"
	"github.com/dknelson9876/oregano/ocli"
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

	usr, _ := user.Current()
	dir := usr.HomeDir
	viper.SetDefault("oregano.data_dir", filepath.Join(dir, ".oregano"))

	dataDir := viper.GetString("oregano.data_dir")
	data, err := ocli.LoadData(dataDir)

	if err != nil {
		log.Fatal(err)
	}

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

	viper.SetEnvPrefix("")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.AutomaticEnv()

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

	opts := plaid.NewConfiguration()
	opts.AddDefaultHeader("PLAID-CLIENT-ID", viper.GetString("plaid.client_id"))
	opts.AddDefaultHeader("PLAID-SECRET", viper.GetString("plaid.secret"))
	opts.UseEnvironment(plaidEnv)

	client := plaid.NewAPIClient(opts)

	linker := ocli.NewLinker(data, client, countries, lang)
	port := viper.GetString("link.port")

	var tokenPair *ocli.TokenPair

	tokenPair, err = linker.Link(port)
	if err != nil {
		log.Fatalln(err)
	}
	data.Tokens[tokenPair.ItemID] = tokenPair.AccessToken
	// err = data.Save()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Institution linked!")
	log.Println(fmt.Sprintf("Item ID: %s", tokenPair.ItemID))

	if alias, ok := data.BackAliases[tokenPair.ItemID]; ok {
		log.Println(fmt.Sprintf("Alias: %s", alias))
		return
	}
	//-------
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
