package util

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"

	"github.com/abates/insteon"
)

// Database is the interface representing a collection of
// Insteon devices.  This provides a way to collect and store
// data about linked devices thus reducing the need to perform
// common first time queries (namely EngineVersion request and
// ID Request) every time you want to interact with an Insteon
// network.  EngineVersion requests and ID Requests can actually
// take longer than the intended direct message (such as turning
// on a light) therefore using a database for a long running
// process that interacts with many devices can significantly
// reduce Insteon network load
type Database interface {
	// Get will look up the Address in the database and return the
	// matching DeviceInfo object.  If no entry is found, then
	// found returns false
	Get(addr insteon.Address) (info insteon.DeviceInfo, found bool)

	// Put will store the DeviceInfo object in the Database overwriting
	// any existing object
	Put(info insteon.DeviceInfo)

	// Filter will return a list of addresses that match the
	// given device categories
	Filter(domains ...insteon.Domain) []insteon.Address
}

// Saveable is any database that can be written to an io.Writer
type Saveable interface {
	Database

	// Save will write the current database state to the given writer
	Save(io.Writer) error

	// NeedsSaving indicates whether the database has changed since the
	// last save
	NeedsSaving() bool
}

// Loadable is any database that can be loaded from an io.Reader
type Loadable interface {
	// Load will replace the current database content with that
	// provided from the io.Reader
	Load(io.Reader) error
}

// NewMemDB returns a memory-backed database
func NewMemDB() Saveable {
	return &memDB{
		values: make(map[insteon.Address]insteon.DeviceInfo),
	}
}

type memDB struct {
	values map[insteon.Address]insteon.DeviceInfo
	dirty  bool
}

func (db *memDB) Filter(domains ...insteon.Domain) (matches []insteon.Address) {
	for addr, info := range db.values {
		if info.DevCat.In(domains...) {
			matches = append(matches, addr)
		}
	}
	return matches
}

func (db *memDB) Get(addr insteon.Address) (insteon.DeviceInfo, bool) {
	info, found := db.values[addr]
	return info, found
}

func (db *memDB) Put(info insteon.DeviceInfo) {
	if old := db.values[info.Address]; old != info {
		db.dirty = true
		db.values[info.Address] = info
	}
}

func (db *memDB) Load(reader io.Reader) error {
	buf, err := ioutil.ReadAll(reader)
	if err == nil {
		err = db.UnmarshalJSON(buf)
	}
	return err
}

func (db *memDB) NeedsSaving() bool {
	return db.dirty
}

func (db *memDB) Save(writer io.Writer) error {
	buf, err := db.MarshalJSON()
	if err == nil {
		_, err = writer.Write(buf)
		if err == nil {
			db.dirty = false
		}
	}
	return err
}

func (db *memDB) MarshalJSON() ([]byte, error) {
	return json.Marshal(db.values)
}

func (db *memDB) UnmarshalJSON(data []byte) error {
	db.values = make(map[insteon.Address]insteon.DeviceInfo)
	return json.Unmarshal(data, &db.values)
}

var (
	ErrNotSaveable = errors.New("Database is not saveable")
	ErrNotLoadable = errors.New("Database is not loadable")
)

// Open will look for the device info in the database and return
// an initialized device if found.  If not found, Open will call
// insteon.Open and store the info upon success.  If dbfile is
// not an empty string, SaveDB will be called at the end
func Open(mw insteon.MessageWriter, addr insteon.Address, db Database, dbfile string) (*insteon.BasicDevice, error) {
	info, found := db.Get(addr)
	if found {
		return insteon.NewDevice(mw, info), nil
	}

	device, info, err := insteon.Open(mw, addr)
	if err == nil {
		db.Put(info)
		if dbfile != "" {
			err = SaveDB(dbfile, db)
		}
	}
	return device, err
}

// SaveDB will attemt to save the database to the named file.  If
// the database is not saveable (does not implement the Saveable
// interface) then ErrNotSaveable is returned
func SaveDB(filename string, db Database) (err error) {
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

// LoadDB will attempt to load db from the named file.  If
// db does not implement the Loadable interface, then nothing
// is done and ErrNotLoadable is returned
func LoadDB(filename string, db Database) (err error) {
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