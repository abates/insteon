package db

import (
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/abates/insteon"
)

func open(db Database, bus insteon.Bus, dst insteon.Address) (device insteon.Device, err error) {
	if info, found := db.Get(dst); found {
		device, err = insteon.New(bus, info)
	} else {
		var version insteon.EngineVersion
		version, err = insteon.GetEngineVersion(bus, dst)
		if err == nil {
			info := insteon.DeviceInfo{
				Address:       dst,
				EngineVersion: version,
			}
			info.FirmwareVersion, info.DevCat, err = insteon.IDRequest(bus, dst)
			if err == nil {
				device, err = insteon.New(bus, info)
			}

			if err == nil {
				db.Put(info)
			}
		}
	}
	return device, err
}

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

	// Open will return an initialized Insteon device object.  If the
	// DeviceInfo object does not exist in the database, then the database
	// will query the device for its engine version and device category
	// and store the responses.  Once the information has been gathered (either
	// from an existing entry in the database, or by querying the device)
	// Open will open a connection to the device and return the
	// correct device type (Light, Thermostat, etc).  See insteon.New and
	// insteon.Create for additional details
	Open(bus insteon.Bus, dst insteon.Address) (device insteon.Device, err error)
}

// Saveable is any database that can be written to an io.Writer
type Saveable interface {
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
func NewMemDB() Database {
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

func (db *memDB) Open(bus insteon.Bus, dst insteon.Address) (device insteon.Device, err error) {
	return open(db, bus, dst)
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
