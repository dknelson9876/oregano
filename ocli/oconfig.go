package ocli

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/dknelson9876/oregano/omoney"
)

// type Data struct {
// 	DataDir     string
// 	Tokens      map[string]string // item id -> access token
// 	Aliases     map[string]string // alias -> item id
// 	BackAliases map[string]string // item id -> alias
// }

func LoadModelFromJson(dataDir string) (*omoney.Model, error) {
	os.MkdirAll(filepath.Join(dataDir, "data"), os.ModePerm)

	model := &omoney.Model{
		FilePath: filepath.Join(dataDir, "data", "accounts.json"),
		Accounts: make(map[string]omoney.Account),
		Aliases:  make(map[string]string),
	}

	err := load(model.FilePath, &model.Accounts)
	if err != nil {
		log.Printf("Error loading data from %s. Proceeding with no account data. Error: %s", model.FilePath, err)
	} else {
		for id, account := range model.Accounts {
			model.Aliases[account.Alias] = id
		}
	}

	return model, nil
}

func LoadModelFromDB(dataDir string) (*omoney.Model, error) {
	os.MkdirAll(filepath.Join(dataDir, "data"), os.ModePerm)

	return omoney.NewModelFromDB(filepath.Join(dataDir, omoney.DbFilename))
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

func Save(model *omoney.Model) error {
	return save(model.Accounts, model.FilePath)
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
