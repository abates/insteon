package insteon

import (
	"bytes"
	"reflect"
	"testing"
)

func TestMemDB(t *testing.T) {
	db := NewMemDB()

	addr := Address{5, 6, 7}
	want := DeviceInfo{}
	got, found := db.Get(addr)
	if found {
		t.Fatalf("Expected not found")
	}

	if want != got {
		t.Fatalf("Expected empty info, got %v", got)
	}

	want = DeviceInfo{
		Address:         addr,
		DevCat:          DevCat{1, 2},
		FirmwareVersion: FirmwareVersion(3),
		EngineVersion:   EngineVersion(4),
	}

	db.Put(want)

	got, found = db.Get(addr)
	if !found {
		t.Fatalf("Expected found")
	}

	if want != got {
		t.Fatalf("Expected device info %v got %v", want, got)
	}

	buf := &bytes.Buffer{}
	err := db.Save(buf)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	db1 := NewMemDB()
	err = db1.Load(buf)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !reflect.DeepEqual(db, db1) {
		t.Fatalf("Expected both databases to be the same")
	}
}
