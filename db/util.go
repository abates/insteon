package db

import (
	"io"
	"os"
	"path/filepath"

	"github.com/kirsle/configdir"
)

func ConfigPath(elem ...string) (string, error) {
	dir := configdir.LocalConfig("go-insteon")
	configPath := filepath.Join(append([]string{dir}, elem...)...)
	return configPath, configdir.MakePath(dir) // Ensure it exists.
}

type Saveable interface {
	Save(io.Writer) error
}

func Save(filename string, s Saveable) error {
	configFile, err := ConfigPath(filename)
	if err == nil {
		var file *os.File
		file, err = os.OpenFile(configFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
		if err == nil {
			err = s.Save(file)
		}
	}
	return err
}

type Loadable interface {
	Load(io.Reader) error
}

func LoadMemDB(filename string) (db *MemDB, err error) {
	configFile, err := ConfigPath(filename)

	if err == nil {
		db = NewMemDB()
		err = Load(configFile, db)
		if os.IsNotExist(err) {
			err = nil
		}
	}
	return db, err
}

func Load(filename string, l Loadable) error {
	_, err := os.Stat(filename)
	if err == nil {
		file, err := os.Open(filename)
		if err == nil {
			err = l.Load(file)
			file.Close()
		}
	}
	return err
}
