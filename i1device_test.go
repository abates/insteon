package insteon

import (
	"bytes"
	"testing"
	"time"
)

func TestI1DeviceIsDevice(t *testing.T) {
	var device interface{}
	device = &I1Device{}

	if _, ok := device.(Device); !ok {
		t.Error("Expected I1Device to be Device")
	}
}

func TestI1DeviceErrLookup(t *testing.T) {
	tests := []struct {
		desc  string
		input *Message
		err   error
		want  error
	}{
		{"nil error", &Message{}, nil, nil},
		{"ErrUnknownCommand", &Message{Command: Command{0, 0, 0xfd}, Flags: StandardDirectNak}, nil, ErrUnknownCommand},
		{"ErrNoLoadDetected", &Message{Command: Command{0, 0, 0xfe}, Flags: StandardDirectNak}, nil, ErrNoLoadDetected},
		{"ErrNotLinked", &Message{Command: Command{0, 0, 0xff}, Flags: StandardDirectNak}, nil, ErrNotLinked},
		{"ErrUnexpectedResponse", &Message{Command: Command{0, 0, 0xfc}, Flags: StandardDirectNak}, nil, ErrUnexpectedResponse},
		{"ErrReadTimeout", &Message{Command: Command{0, 0, 0xfc}, Flags: StandardDirectNak}, ErrReadTimeout, ErrReadTimeout},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			_, got := errLookup(test.input, test.err)
			if !isError(got, test.want) {
				t.Errorf("want %v got %v", test.want, got)
			}
		})
	}
}

func TestI1DeviceSendCommand(t *testing.T) {
	tests := []struct {
		desc    string
		command Command
		payload []byte
		flags   Flags
	}{
		{"SD", Command{byte(StandardDirectMessage), 1, 2}, nil, StandardDirectMessage},
		{"ED", Command{byte(ExtendedDirectMessage), 2, 3}, []byte{1, 2, 3, 4}, ExtendedDirectMessage},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			conn := &testConnection{sendCh: make(chan *Message, 1), ackCh: make(chan *Message, 1)}
			device := NewI1Device(conn, time.Millisecond)
			ackFlags := StandardDirectAck
			if len(test.payload) > 0 {
				ackFlags = ExtendedDirectAck
			}
			conn.ackCh <- &Message{Flags: ackFlags}

			device.SendCommand(test.command, test.payload)

			msg := <-conn.sendCh
			if msg.Command != test.command {
				t.Errorf("want %v got %v", test.command, msg.Command)
			}

			if msg.Flags != test.flags {
				t.Errorf("want %v got %v", test.flags, msg.Flags)
			}
		})
	}
}

type commandTest struct {
	desc        string
	callback    func(Device) error
	wantCmd     Command
	wantErr     error
	wantPayload []byte
}

func testDeviceCommands(t *testing.T, constructor func(*testConnection) Device, tests []*commandTest) {
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			conn := &testConnection{sendCh: make(chan *Message, 1), ackCh: make(chan *Message, 1)}
			device := constructor(conn)

			conn.ackCh <- TestAck
			err := test.callback(device)

			if err != test.wantErr {
				t.Errorf("got error %v, want %v", err, test.wantErr)
			}

			if test.wantErr == nil {
				msg := <-conn.sendCh
				if msg.Command != test.wantCmd {
					t.Errorf("want Command %v got %v", test.wantCmd, msg.Command)
				}

				if len(test.wantPayload) > 0 {
					wantPayload := make([]byte, len(test.wantPayload))
					copy(wantPayload, test.wantPayload)
					if !bytes.Equal(wantPayload, msg.Payload) {
						t.Errorf("want payload %x got %x", wantPayload, msg.Payload)
					}
				}
			}
		})
	}
}

func TestI1DeviceCommands(t *testing.T) {
	tests := []*commandTest{
		{"AssignToAllLinkGroup", func(d Device) error { return d.(*I1Device).AssignToAllLinkGroup(10) }, CmdAssignToAllLinkGroup.SubCommand(10), nil, nil},
		{"DeleteFromAllLinkGroup", func(d Device) error { return d.(*I1Device).DeleteFromAllLinkGroup(10) }, CmdDeleteFromAllLinkGroup.SubCommand(10), nil, nil},
		{"CmdPing", func(d Device) error { return d.(*I1Device).Ping() }, CmdPing, nil, nil},
		{"SetAllLinkCommandAlias", func(d Device) error { return d.(*I1Device).SetAllLinkCommandAlias(Command{}, Command{}) }, Command{}, ErrNotImplemented, nil},
		{"SetAllLinkCommandAliasData", func(d Device) error { return d.(*I1Device).SetAllLinkCommandAliasData(nil) }, Command{}, ErrNotImplemented, nil},
		{"BlockDataTransfer", func(d Device) error { _, err := d.(*I1Device).BlockDataTransfer(0, 0, 0); return err }, Command{}, ErrNotImplemented, nil},
	}

	testDeviceCommands(t, func(conn *testConnection) Device { return NewI1Device(conn, time.Millisecond) }, tests)
}

func TestI1DeviceProductData(t *testing.T) {
	tests := []struct {
		desc    string
		want    *ProductData
		wantErr error
	}{
		{"Happy Path", &ProductData{ProductKey{1, 2, 3}, DevCat{4, 5}}, nil},
		{"Sad Path", nil, ErrReadTimeout},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			conn := &testConnection{sendCh: make(chan *Message, 1), ackCh: make(chan *Message, 1), recvCh: make(chan *Message, 1)}
			conn.ackCh <- TestAck
			if test.wantErr == nil {
				msg := *TestProductDataResponse
				buf, _ := test.want.MarshalBinary()
				msg.Payload = make([]byte, 14)
				copy(msg.Payload, buf)
				conn.recvCh <- &msg
			} else {
				go func() {
					conn.recvCh <- TestAck
					time.Sleep(time.Nanosecond)
					conn.recvCh <- TestAck
				}()
			}

			device := NewI1Device(conn, time.Nanosecond)
			pd, err := device.ProductData()
			if err != test.wantErr {
				t.Errorf("want error %v got %v", test.wantErr, err)
			} else if err == nil {
				if *pd != *test.want {
					t.Errorf("want %v got %v", *test.want, *pd)
				}
			}
		})
	}
}

func TestI1DeviceReceive(t *testing.T) {
	tests := []struct {
		desc    string
		input   *Message
		wantErr error
	}{
		{"nil error", &Message{}, nil},
		{"ErrUnknownCommand", &Message{Command: Command{0, 0, 0xfd}, Flags: StandardDirectNak}, ErrUnknownCommand},
		{"ErrNoLoadDetected", &Message{Command: Command{0, 0, 0xfe}, Flags: StandardDirectNak}, ErrNoLoadDetected},
		{"ErrNotLinked", &Message{Command: Command{0, 0, 0xff}, Flags: StandardDirectNak}, ErrNotLinked},
		{"ErrUnexpectedResponse", &Message{Command: Command{0, 0, 0xfc}, Flags: StandardDirectNak}, ErrUnexpectedResponse},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			conn := &testConnection{recvCh: make(chan *Message, 1)}
			conn.recvCh <- test.input
			device := NewI1Device(conn, time.Millisecond)
			_, err := device.Receive()
			if !isError(err, test.wantErr) {
				t.Errorf("want error %v got %v", test.wantErr, err)
			}
		})
	}
}
