package ocli

import (
	"encoding/json"
	"io"
	// "log"
	"os"
	"path/filepath"

	"github.com/dknelson9876/oregano/omoney"
)

func LoadModelFromDB(dataDir string) (*omoney.Model, error) {
	os.MkdirAll(dataDir, os.ModePerm)

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

// func Save(model *omoney.Model) error {
// 	return save(model.Accounts, model.FilePath)
// }

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
