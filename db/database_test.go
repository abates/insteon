package db

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/abates/insteon"
)

func TestMemDB(t *testing.T) {
	db := NewMemDB()

	addr := insteon.Address{5, 6, 7}
	want := insteon.DeviceInfo{}
	got, found := db.Get(addr)
	if found {
		t.Fatalf("Expected not found")
	}

	if want != got {
		t.Fatalf("Expected empty info, got %v", got)
	}

	want = insteon.DeviceInfo{
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

/*func TestOpen(t *testing.T) {
	tests := []struct {
		desc       string
		input      []*insteon.Message
		messages   []*insteon.Message
		publishErr error
		wantType   reflect.Type
		wantErr    error
	}{
		{"I1Device", []*insteon.Message{TestMessageEngineVersion1, TestAck}, []*insteon.Message{SetButtonPressed(false, 0, 0, 0)}, nil, reflect.TypeOf(&i1Device{}), nil},
		{"I2Device", []*insteon.Message{TestMessageEngineVersion2, TestAck}, []*insteon.Message{SetButtonPressed(false, 0, 0, 0)}, nil, reflect.TypeOf(&i2Device{}), nil},
		{"I2CsDevice", []*insteon.Message{TestMessageEngineVersion2cs, TestAck}, []*insteon.Message{SetButtonPressed(false, 0, 0, 0)}, nil, reflect.TypeOf(&i2CsDevice{}), nil},
		{"Dimmer", []*insteon.Message{TestMessageEngineVersion2cs, TestAck}, []*insteon.Message{SetButtonPressed(false, 1, 0, 0)}, nil, reflect.TypeOf(&Dimmer{}), nil},
		{"Switch", []*insteon.Message{TestMessageEngineVersion2cs, TestAck}, []*insteon.Message{SetButtonPressed(false, 2, 0, 0)}, nil, reflect.TypeOf(&Switch{}), nil},
		{"ErrVersion", []*insteon.Message{TestMessageEngineVersion3, TestAck}, []*insteon.Message{SetButtonPressed(false, 0, 0, 0)}, nil, reflect.TypeOf(nil), ErrVersion},
		{"Not Linked", []*insteon.Message{Ack(false, 0, 255), SetButtonPressed(false, 0, 0, 0)}, nil, ErrNak, reflect.TypeOf(&i2CsDevice{}), ErrNotLinked},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			db := &DummyDB{}
			ch := make(chan *Message, len(test.messages))
			for _, msg := range test.messages {
				ch <- msg
			}
			tb := &testBus{subscribeCh: ch, publishResp: test.input, publishErr: test.publishErr}
			device, gotErr := open(db, tb, Address{})
			gotType := reflect.TypeOf(device)

			if test.wantErr != gotErr {
				t.Errorf("want err %v got %v", test.wantErr, gotErr)
			}
			if test.wantType != gotType {
				t.Errorf("want type %v got %v", test.wantType, gotType)
			}
		})
	}
}*/
