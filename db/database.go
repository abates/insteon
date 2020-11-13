package db

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"sync"

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
		} else if err == insteon.ErrNotLinked {
			device, _ = insteon.Create(bus, insteon.DeviceInfo{Address: dst, EngineVersion: insteon.VerI2Cs})
		}
	}
	return device, err
}

type Database interface {
	Get(addr insteon.Address) (insteon.DeviceInfo, bool)
	Put(info insteon.DeviceInfo)
	Open(bus insteon.Bus, dst insteon.Address) (device insteon.Device, err error)
}

type DummyDB struct{}

func (DummyDB) Get(addr insteon.Address) (insteon.DeviceInfo, bool) {
	return insteon.DeviceInfo{}, false
}

func (DummyDB) Put(info insteon.DeviceInfo) {
	return
}

func (db DummyDB) Open(bus insteon.Bus, dst insteon.Address) (device insteon.Device, err error) {
	return open(db, bus, dst)
}

func NewMemDB() *MemDB {
	return &MemDB{
		values: make(map[insteon.Address]insteon.DeviceInfo),
	}
}

type MemDB struct {
	lock   sync.Mutex
	values map[insteon.Address]insteon.DeviceInfo
	dirty  bool
}

func (db *MemDB) Get(addr insteon.Address) (insteon.DeviceInfo, bool) {
	db.lock.Lock()
	defer db.lock.Unlock()
	info, found := db.values[addr]
	return info, found
}

func (db *MemDB) Put(info insteon.DeviceInfo) {
	db.lock.Lock()
	defer db.lock.Unlock()
	if old := db.values[info.Address]; old != info {
		db.dirty = true
		db.values[info.Address] = info
	}
}

func (db *MemDB) Open(bus insteon.Bus, dst insteon.Address) (device insteon.Device, err error) {
	return open(db, bus, dst)
}

func (db *MemDB) Load(reader io.Reader) error {
	buf, err := ioutil.ReadAll(reader)
	if err == nil {
		err = db.UnmarshalJSON(buf)
	}
	return err
}

func (db *MemDB) Save(writer io.Writer) error {
	db.lock.Lock()
	dirty := db.dirty
	db.lock.Unlock()
	if !dirty {
		return nil
	}

	buf, err := db.MarshalJSON()
	if err == nil {
		_, err = writer.Write(buf)
		if err == nil {
			db.lock.Lock()
			db.dirty = false
			db.lock.Unlock()
		}
	}
	return err
}

func (db *MemDB) MarshalJSON() ([]byte, error) {
	db.lock.Lock()
	defer db.lock.Unlock()
	return json.Marshal(db.values)
}

func (db *MemDB) UnmarshalJSON(data []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()
	db.values = make(map[insteon.Address]insteon.DeviceInfo)
	return json.Unmarshal(data, &db.values)
}
