package insteon

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"sync"
)

var DB Database

type Database interface {
	Get(addr Address) (DeviceInfo, bool)
	Put(info DeviceInfo)
	Open(bus Bus, dst Address) (device Device, err error)
}

func init() {
	//DB = MemDB{values: make(map[Address]DeviceInfo)}
	DB = DummyDB{}
}

type DummyDB struct{}

func (DummyDB) Get(addr Address) (DeviceInfo, bool) {
	return DeviceInfo{}, false
}

func (DummyDB) Put(info DeviceInfo) {
	return
}

func (db DummyDB) Open(bus Bus, dst Address) (device Device, err error) {
	return open(db, bus, dst)
}

func NewMemDB() *MemDB {
	return &MemDB{
		values: make(map[Address]DeviceInfo),
	}
}

type MemDB struct {
	lock   sync.Mutex
	values map[Address]DeviceInfo
}

func (db *MemDB) Get(addr Address) (DeviceInfo, bool) {
	db.lock.Lock()
	defer db.lock.Unlock()
	info, found := db.values[addr]
	return info, found
}

func (db *MemDB) Put(info DeviceInfo) {
	db.lock.Lock()
	defer db.lock.Unlock()
	db.values[info.Address] = info
}

func (db *MemDB) Open(bus Bus, dst Address) (device Device, err error) {
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
	buf, err := db.MarshalJSON()
	if err == nil {
		_, err = writer.Write(buf)
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
	db.values = make(map[Address]DeviceInfo)
	return json.Unmarshal(data, &db.values)
}
