package util

import (
	"errors"
	"io/fs"
	"os"
	"testing"

	"github.com/abates/insteon"
	"github.com/abates/insteon/devices"
)

func TestDatabase(t *testing.T) {
	f, err := os.CreateTemp("", "insteon_db_test_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	f.Close()
	// start with no file.  A missing file should not
	// cause NewFileDB to error our
	err = os.Remove(f.Name())

	db, err := NewFileDB(f.Name())
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Wanted %v got %v", fs.ErrNotExist, err)
	}
	defer os.Remove(f.Name())

	addr := insteon.Address(0x050607)
	want := devices.DeviceInfo{}
	got, found := db.Get(addr)
	if found {
		t.Fatalf("Expected not found")
	}

	if want != got {
		t.Fatalf("Expected empty info, got %v", got)
	}

	want = devices.DeviceInfo{
		Address:         addr,
		DevCat:          insteon.DevCat{1, 2},
		FirmwareVersion: insteon.FirmwareVersion(3),
		EngineVersion:   insteon.EngineVersion(4),
	}

	db.Put(want)

	got, found = db.Get(addr)
	if !found {
		t.Fatalf("Expected found")
	}

	if want != got {
		t.Fatalf("Expected device info %v got %v", want, got)
	}

	db1, err := NewFileDB(f.Name())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	got, found = db1.Get(addr)
	if !found {
		t.Fatalf("Expected found")
	}

	if want != got {
		t.Fatalf("Expected device info %v got %v", want, got)
	}

}
