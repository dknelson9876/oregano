package ocli

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

type Data struct {
	DataDir     string
	Tokens      map[string]string // item id -> access token
	Aliases     map[string]string // alias -> item id
	BackAliases map[string]string // item id -> alias
}

func LoadData(dataDir string) (*Data, error) {
	os.MkdirAll(filepath.Join(dataDir, "data"), os.ModePerm)

	data := &Data{
		DataDir:     dataDir,
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
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0755)

	if err != nil {
		return err
	} else {
		defer f.Close()
		b, err := io.ReadAll(f)
		if err != nil {
			return err
		}

		return json.Unmarshal(b, v)
	}
}

func (d *Data) Save() error {
	err := d.SaveTokens()
	if err != nil {
		return err
	}

	err = d.SaveAliases()
	if err != nil {
		return err
	}

	return nil
}

func (d *Data) SaveTokens() error {
	return save(d.Tokens, d.tokensPath())
}

func (d *Data) SaveAliases() error {
	return save(d.Aliases, d.aliasesPath())
}

func save(v interface{}, filePath string) error {
	// O_TRUNC to truncate to 0 bytes on open, in other words deleting
	// the old file contents
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return nil
	}
	defer f.Close()

	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	_, err = f.Write(b)
	return err
}

func (d *Data) GetAccessToken(input string) (string, error) {
	if id, ok := d.Aliases[input]; ok {
		return d.Tokens[id], nil
	} else if token, ok := d.Tokens[input]; ok {
		return token, nil
	} else {
		return "", fmt.Errorf("input not recognized as valid item ID or alias: %s", input)
	}
}

func (d *Data) RemoveToken(input string) error {
	// Check if input was an alias
	if id, ok := d.Aliases[input]; ok {
		delete(d.Aliases, input)
		delete(d.BackAliases, id)
		delete(d.Tokens, id)
		d.Save()
		return nil
	} else if alias, ok := d.BackAliases[input]; ok {
		delete(d.Aliases, alias)
		delete(d.BackAliases, input)
		delete(d.Tokens, input)
		d.Save()
		return nil
	}
	return fmt.Errorf("input not recognized as valid item ID or alias: %s", input)
}

func (d *Data) SetAlias(itemID string, alias string) error {
	if _, ok := d.Tokens[itemID]; !ok {
		return fmt.Errorf("item ID `%s` not recognized", itemID)
	}

	d.Aliases[alias] = itemID
	d.BackAliases[itemID] = alias
	err := d.Save()
	if err != nil {
		return err
	}

	log.Printf("Aliased %s to %s\n", itemID, alias)

	return nil
}
