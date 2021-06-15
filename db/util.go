package db

import (
	"errors"
	"os"
)

var (
	ErrNotSaveable = errors.New("Database is not saveable")
	ErrNotLoadable = errors.New("Database is not loadable")
)

// Save will attemt to save the database to the named file.  If
// the database is not saveable (does not implement the Saveable
// interface) then ErrNotSaveable is returned
func Save(filename string, db Database) (err error) {
	if saveable, ok := db.(Saveable); ok {
		if saveable.NeedsSaving() {
			var file *os.File
			file, err = os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
			if err == nil {
				err = saveable.Save(file)
			}
		}
	} else {
		err = ErrNotSaveable
	}
	return err
}

// Load will attempt to load db from the named file.  If
// db does not implement the Loadable interface, then nothing
// is done and ErrNotLoadable is returned
func Load(filename string, db Database) (err error) {
	if loadable, ok := db.(Loadable); ok {
		_, err := os.Stat(filename)
		if err == nil {
			file, err := os.Open(filename)
			if err == nil {
				err = loadable.Load(file)
				file.Close()
			}
		}
	} else {
		err = ErrNotLoadable
	}
	return err
}
