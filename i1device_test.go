package insteon

import (
	"testing"
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
		desc     string
		input    *Message
		inputErr error
		want     error
	}{
		{"nil error", &Message{}, nil, nil},
		{"ErrUnknownCommand", &Message{Command: Command(0x0000fd), Flags: StandardDirectNak}, ErrNak, ErrUnknownCommand},
		{"ErrNoLoadDetected", &Message{Command: Command(0x0000fe), Flags: StandardDirectNak}, ErrNak, ErrNoLoadDetected},
		{"ErrNotLinked", &Message{Command: Command(0x0000ff), Flags: StandardDirectNak}, ErrNak, ErrNotLinked},
		{"ErrUnexpectedResponse", &Message{Command: Command(0x0000fc), Flags: StandardDirectNak}, ErrNak, ErrUnexpectedResponse},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			_, got := errLookup(test.input, test.inputErr)
			if !IsError(got, test.want) {
				t.Errorf("want error %v got %v", test.want, got)
			}
		})
	}
}

func TestI1DeviceSendCommand(t *testing.T) {
	tests := []struct {
		desc    string
		wantCmd Command
	}{
		{"SD", Command((0xff&int(StandardDirectMessage))<<16 | 0x0102)},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			conn := &testConnection{acks: []*Message{TestAck}}
			device := newI1Device(testDialer{conn}, DeviceInfo{})
			device.SendCommand(test.wantCmd, nil)

			gotCmd := conn.sent[0].Command

			if test.wantCmd != gotCmd {
				t.Errorf("want command %v got %v", test.wantCmd, gotCmd)
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

			device := newI1Device(testDialer{conn}, DeviceInfo{})
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

func TestI1DeviceLinkDatabase(t *testing.T) {
	device := &i1Device{}
	want := ErrNotSupported
	_, got := device.LinkDatabase()
	if want != got {
		t.Errorf("Expected error %v got %v", want, got)
	}
}
