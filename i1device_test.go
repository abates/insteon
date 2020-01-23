package insteon

import (
	"testing"
	"time"
)

func TestI1DeviceIsDevice(t *testing.T) {
	var device interface{}
	device = &i1Device{}

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
			if !IsError(got, test.want) {
				t.Errorf("want %v got %v", test.want, got)
			}
		})
	}
}

func TestI1DeviceSendCommand(t *testing.T) {
	tests := []struct {
		desc      string
		wantCmd   Command
		payload   []byte
		wantFlags Flags
	}{
		{"SD", Command{byte(StandardDirectMessage), 1, 2}, nil, StandardDirectMessage},
		{"ED", Command{byte(ExtendedDirectMessage), 2, 3}, []byte{1, 2, 3, 4}, ExtendedDirectMessage},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ackFlags := StandardDirectAck
			if len(test.payload) > 0 {
				ackFlags = ExtendedDirectAck
			}
			conn := &testConnection{acks: []*Message{{Flags: ackFlags}}}
			device := newI1Device(conn, time.Millisecond)
			device.SendCommand(test.wantCmd, test.payload)

			gotCmd := conn.sent[0].Command
			gotFlags := conn.sent[0].Flags

			if test.wantCmd != gotCmd {
				t.Errorf("want command %v got %v", test.wantCmd, gotCmd)
			}

			if test.wantFlags != gotFlags {
				t.Errorf("want flags %v got %v", test.wantFlags, gotFlags)
			}
		})
	}
}

func TestI1DeviceProductData(t *testing.T) {
	tests := []struct {
		desc    string
		want    *ProductData
		wantErr error
	}{
		{"Happy Path", &ProductData{ProductKey{1, 2, 3}, DevCat{4, 5}}, nil},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			conn := &testConnection{acks: []*Message{TestAck}}
			if test.wantErr == nil {
				msg := *TestProductDataResponse
				buf, _ := test.want.MarshalBinary()
				msg.Payload = make([]byte, 14)
				copy(msg.Payload, buf)
				conn.recv = []*Message{&msg}
			} else {
				conn.recv = []*Message{TestAck, TestAck}
			}

			device := newI1Device(conn, time.Millisecond)
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
			conn := &testConnection{recv: []*Message{test.input}}
			device := newI1Device(conn, time.Millisecond)
			_, err := device.Receive()
			if !IsError(err, test.wantErr) {
				t.Errorf("want error %v got %v", test.wantErr, err)
			}
		})
	}
}
