package plm

import (
	"bytes"
	"encoding"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/abates/insteon"
)

type marshalUnmarshal interface {
	MarshalBinary() (data []byte, err error)
	UnmarshalBinary(data []byte) error
}

func TestMarshalUnmarshalBinary(t *testing.T) {
	tests := []struct {
		desc    string
		input   []byte
		got     marshalUnmarshal
		want    marshalUnmarshal
		wantErr error
	}{
		{"manageRecordRequest", []byte{byte(LinkCmdFindFirst), byte(insteon.UnavailableController) | 0x02, 0x42, 4, 5, 6, 0, 0, 0}, &manageRecordRequest{}, &manageRecordRequest{LinkCmdFindFirst, insteon.ControllerLink(0x42, insteon.Address{4, 5, 6})}, nil},
		{"allLinkReq", []byte{0x42, 0x75}, &allLinkReq{}, &allLinkReq{0x42, 0x75}, nil},
		{"allLinkReq", []byte{}, &allLinkReq{}, &allLinkReq{}, insteon.ErrBufferTooShort},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			err := test.got.UnmarshalBinary(test.input)
			if !insteon.IsError(err, test.wantErr) {
				t.Errorf("Want error %v got %v", test.wantErr, err)
			} else if err == nil {
				buf, err := test.want.MarshalBinary()
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else {
					if !bytes.Equal(test.input, buf) {
						t.Errorf("want %x got %x", test.input, buf)
					}
				}
			}
		})
	}
}

func TestLinkDBOld(t *testing.T) {
	ldb := linkdb{
		timeout: time.Hour,
	}
	if !ldb.old() {
		t.Errorf("Expected database to be marked old")
	}
	ldb.age = time.Now()
	if ldb.old() {
		t.Errorf("Expected database to not be marked old")
	}
}

type testModem struct {
	rx    []*Packet
	rxErr error
	tx    []*Packet
	txErr error
	ack   []*Packet
}

func (tlplm *testModem) ReadPacket() (p *Packet, err error) {
	err = tlplm.rxErr
	if len(tlplm.rx) > 0 {
		p = tlplm.rx[0]
		tlplm.rx = tlplm.rx[1:]
	} else if err == nil {
		err = errors.New("No RX error and no rx packets")
	}
	return
}

func (tlplm *testModem) WritePacket(packet *Packet) (ack *Packet, err error) {
	tlplm.tx = append(tlplm.tx, packet)
	err = tlplm.txErr
	if len(tlplm.ack) > 0 {
		ack = tlplm.ack[0]
		tlplm.ack = tlplm.ack[1:]
		if ack.Ack == 0x15 {
			err = ErrNak
		}
	} else if err == nil {
		err = errors.New("No TX error and no ack packets")
	}
	return
}

func TestLinkDBRefresh(t *testing.T) {
	pkt := func(cmd Command, marshaler encoding.BinaryMarshaler) *Packet {
		packet := &Packet{Command: cmd}
		packet.Payload, _ = marshaler.MarshalBinary()
		return packet
	}

	tests := []struct {
		name      string
		age       time.Time
		timeout   time.Duration
		rx        []*Packet
		rxErr     error
		ack       []*Packet
		txErr     error
		wantTx    []*Packet
		wantLinks []*insteon.LinkRecord
		wantErr   error
	}{
		{
			name:      "Happy Path",
			rx:        []*Packet{pkt(CmdAllLinkRecordResp, insteon.ControllerLink(42, insteon.Address{1, 2, 3}))},
			ack:       []*Packet{{Command: CmdGetFirstAllLink, Ack: 0x06}, {Command: CmdGetNextAllLink, Ack: 0x15}},
			wantTx:    []*Packet{{Command: CmdGetFirstAllLink}, {Command: CmdGetNextAllLink}},
			wantLinks: []*insteon.LinkRecord{insteon.ControllerLink(42, insteon.Address{1, 2, 3})},
		},
		{
			name:      "Happy Path 1",
			rx:        []*Packet{pkt(CmdAllLinkRecordResp, insteon.ControllerLink(42, insteon.Address{1, 2, 3})), pkt(CmdAllLinkRecordResp, insteon.ControllerLink(35, insteon.Address{4, 5, 6}))},
			ack:       []*Packet{{Command: CmdGetFirstAllLink, Ack: 0x06}, {Command: CmdGetNextAllLink, Ack: 0x06}, {Command: CmdGetNextAllLink, Ack: 0x15}},
			wantTx:    []*Packet{{Command: CmdGetFirstAllLink}, {Command: CmdGetNextAllLink}, {Command: CmdGetNextAllLink}},
			wantLinks: []*insteon.LinkRecord{insteon.ControllerLink(42, insteon.Address{1, 2, 3}), insteon.ControllerLink(35, insteon.Address{4, 5, 6})},
		},
		{
			name: "New",
			age:  time.Now().Add(42 * time.Hour),
		},
		{
			name:    "Read Timeout",
			timeout: -time.Hour,
			ack:     []*Packet{{Command: CmdGetFirstAllLink}},
			rxErr:   ErrReadTimeout,
			wantErr: ErrReadTimeout,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plm := &testModem{rx: test.rx, rxErr: test.rxErr, ack: test.ack, txErr: test.txErr}
			ldb := linkdb{plm: plm, age: test.age, timeout: test.timeout}
			gotLinks, gotErr := ldb.Links()
			if test.wantErr == gotErr {
				if gotErr == nil {
					if !reflect.DeepEqual(test.wantTx, plm.tx) {
						t.Errorf("Wanted %v packets got %v", test.wantTx, plm.tx)
					}

					if !reflect.DeepEqual(test.wantLinks, gotLinks) {
						t.Errorf("Wanted %v links got %v", test.wantLinks, ldb.links)
					}
				}
			} else {
				t.Errorf("Wanted error %v got %v", test.wantErr, gotErr)
			}
		})
	}
}

func TestLinkdbCommand(t *testing.T) {
	pkt := func(cmd Command, marshaler encoding.BinaryMarshaler) *Packet {
		packet := &Packet{Command: cmd}
		packet.Payload, _ = marshaler.MarshalBinary()
		return packet
	}

	tests := []struct {
		name string
		test func(ldb *linkdb) error
		want *Packet
	}{
		{"EnterLinkingMode", func(ldb *linkdb) error { return ldb.EnterLinkingMode(42) }, pkt(CmdStartAllLink, &allLinkReq{Mode: linkingMode(0x03), Group: 42})},
		{"ExitLinkingMode", func(ldb *linkdb) error { return ldb.ExitLinkingMode() }, &Packet{Command: CmdCancelAllLink}},
		{"EnterUnlinkingMode", func(ldb *linkdb) error { return ldb.EnterUnlinkingMode(42) }, pkt(CmdStartAllLink, &allLinkReq{Mode: linkingMode(0xff), Group: 42})},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plm := &testModem{ack: []*Packet{{Ack: 0x06}}}
			ldb := &linkdb{plm: plm}
			gotErr := test.test(ldb)
			if gotErr == nil {
				if !reflect.DeepEqual(plm.tx, []*Packet{test.want}) {
					t.Errorf("Wanted packet %v got %v", test.want, plm.tx[0])
				}
			} else {
				t.Errorf("Unexpected error: %v", gotErr)
			}
		})
	}
}
