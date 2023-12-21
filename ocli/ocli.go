package ocli

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
)

type Data struct {
	DataDir     string
	Tokens      map[string]string
	Aliases     map[string]string
	BackAliases map[string]string
}

func LoadData(dataDir string) (*Data, error) {
	os.MkdirAll(filepath.Join(dataDir, "data"), os.ModePerm)

	data := &Data {
		DataDir: dataDir,
		BackAliases: make(map[string]string),
	}

	data.loadTokens()
	data.loadAliases()

	return data, nil
}

func (d *Data) loadTokens() {
	var tokens map[string]string = make(map[string]string)
	filePath := d.tokensPath()
	err := load(filePath, &tokens)
	if err != nil {
		log.Printf("Error loading tokens from %s. Assuming empty tokens. Error: %s", d.tokensPath(), err)
	}

	d.Tokens = tokens
}

func (d *Data) tokensPath() string {
	return filepath.Join(d.DataDir, "data", "tokens.json")
}

func (d *Data) loadAliases() {
	var aliases map[string]string = make(map[string]string)
	filePath := d.aliasesPath()
	err := load(filePath, &aliases)
	if err != nil {
		log.Printf("Error loading aliases from %s. Assuming empty tokens. Error: %s", d.aliasesPath(), err)
	}

	d.Aliases = aliases

	for alias, itemID := range aliases {
		d.BackAliases[itemID] = alias
	}
}

func (d *Data) aliasesPath() string {
	return filepath.Join(d.DataDir, "data", "aliases.json")
}

func load(filePath string, v interface{}) error {
	f, err := os.OpenFile(filePath, os.O_RDWR | os.O_CREATE, 0755)
	defer f.Close()

	if err != nil {
		return err
	} else {
		b, err := io.ReadAll(f)
		if err != nil {
			return err
		}

		return json.Unmarshal(b, v)
	}
}