package devices

import (
	"bytes"
	"errors"
	"testing"

	"github.com/abates/insteon"
	"github.com/abates/insteon/commands"
)

func TestI1DeviceIsDevice(t *testing.T) {
	var d interface{}
	wantAddress := insteon.Address(0x070605)
	d = &BasicDevice{DeviceInfo: DeviceInfo{Address: wantAddress}}

	if d, ok := d.(Device); ok {
		if d.Address() != wantAddress {
			t.Errorf("Wanted address %v got %v", wantAddress, d.Address())
		}
	} else {
		t.Error("Expected BasicDevice to be Device")
	}

	if _, ok := d.(Linkable); !ok {
		t.Error("Expected BasicDevice to be Linkable")
	}

}

func TestI1DeviceWrite(t *testing.T) {
	tests := []struct {
		desc    string
		version insteon.EngineVersion
		input   *insteon.Message
		want    []byte
	}{
		{"VerI1", insteon.VerI1, &insteon.Message{Payload: []byte{}}, []byte{}},
		{"VerI1 Extended", insteon.VerI1, &insteon.Message{Payload: make([]byte, 14)}, make([]byte, 14)},
		{"VerI2Cs", insteon.VerI2Cs, &insteon.Message{Payload: []byte{}}, []byte{}},
		{"VerI2Cs Extended", insteon.VerI2Cs, &insteon.Message{Payload: []byte{0x2F, 0x00, 0x00, 0x00, 0x0F, 0xFF, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}, []byte{0x2F, 0x00, 0x00, 0x00, 0x0F, 0xFF, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xC2}},
		{"VerI2Cs Extended (truncated)", insteon.VerI2Cs, &insteon.Message{Payload: []byte{0x2e, 0x00, 0x01}}, []byte{0x2E, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xd1}},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			tw := &testWriter{}
			d := New(tw, DeviceInfo{EngineVersion: test.version})
			_, err := d.Write(test.input)
			if err == nil {
				if !bytes.Equal(test.want, tw.written[0].Payload) {
					t.Errorf("Wanted bytes %v got %v", test.want, tw.written[0].Payload)
				}
			} else {
				t.Errorf("unexpected error %v", err)
			}
		})
	}
}

func TestI1DeviceErrLookup(t *testing.T) {
	tests := []struct {
		desc     string
		ver      insteon.EngineVersion
		input    *insteon.Message
		inputErr error
		want     error
	}{
		{"nil error", insteon.VerI1, &insteon.Message{}, nil, nil},
		{"ErrUnknownCommand", insteon.VerI1, &insteon.Message{Command: commands.Command(0x0000fd), Flags: insteon.StandardDirectNak}, ErrNak, insteon.ErrUnknownCommand},
		{"ErrNoLoadDetected", insteon.VerI1, &insteon.Message{Command: commands.Command(0x0000fe), Flags: insteon.StandardDirectNak}, ErrNak, ErrNoLoadDetected},
		{"ErrNotLinked", insteon.VerI1, &insteon.Message{Command: commands.Command(0x0000ff), Flags: insteon.StandardDirectNak}, ErrNak, ErrNotLinked},
		{"ErrUnexpectedResponse", insteon.VerI1, &insteon.Message{Command: commands.Command(0x0000fc), Flags: insteon.StandardDirectNak}, ErrNak, ErrUnexpectedResponse},
		{"nil error", insteon.VerI2Cs, &insteon.Message{}, nil, nil},
		{"ErrIllegalValue", insteon.VerI2Cs, &insteon.Message{Command: commands.Command(0x0000fb), Flags: insteon.StandardDirectNak}, ErrNak, ErrIllegalValue},
		{"ErrPreNak", insteon.VerI2Cs, &insteon.Message{Command: commands.Command(0x0000fc), Flags: insteon.StandardDirectNak}, ErrNak, insteon.ErrPreNak},
		{"ErrIncorrectChecksum", insteon.VerI2Cs, &insteon.Message{Command: commands.Command(0x0000fd), Flags: insteon.StandardDirectNak}, ErrNak, ErrIncorrectChecksum},
		{"ErrNoLoadDetected", insteon.VerI2Cs, &insteon.Message{Command: commands.Command(0x0000fe), Flags: insteon.StandardDirectNak}, ErrNak, ErrNoLoadDetected},
		{"ErrNotLinked", insteon.VerI2Cs, &insteon.Message{Command: commands.Command(0x0000ff), Flags: insteon.StandardDirectNak}, ErrNak, ErrNotLinked},
		{"ErrUnexpectedResponse", insteon.VerI2Cs, &insteon.Message{Command: commands.Command(0x0000fa), Flags: insteon.StandardDirectNak}, ErrNak, ErrUnexpectedResponse},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			d := &BasicDevice{DeviceInfo: DeviceInfo{EngineVersion: test.ver}}
			_, got := d.errLookup(test.input, test.inputErr)
			if !errors.Is(got, test.want) {
				t.Errorf("want error %v got %v", test.want, got)
			}
		})
	}
}

func TestI1DeviceSendCommand(t *testing.T) {
	tests := []struct {
		desc    string
		wantCmd commands.Command
	}{
		{"SD", commands.Command((0xff&int(insteon.StandardDirectMessage))<<16 | 0x0102)},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			tw := &testWriter{}
			device := New(tw, DeviceInfo{})
			device.SendCommand(test.wantCmd, nil)

			gotCmd := tw.written[0].Command

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
		{"Happy Path", &ProductData{insteon.ProductKey{1, 2, 3}, insteon.DevCat{4, 5}}, nil},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			tw := &testWriter{}
			if test.wantErr == nil {
				msg := &insteon.Message{insteon.Address(0x010203), insteon.Address(0x030405), insteon.ExtendedDirectMessage, commands.ProductDataResp, []byte{0, 1, 2, 3, 4, 5, 0xff, 0xff, 0, 0, 0, 0, 0, 0}}
				buf, _ := test.want.MarshalBinary()
				msg.Payload = make([]byte, 14)
				copy(msg.Payload, buf)
				tw.read = append(tw.read, msg)
			}

			device := New(tw, DeviceInfo{})
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

func TestI1DeviceDump(t *testing.T) {
	device := New(nil, DeviceInfo{insteon.Address(0x010203), insteon.DevCat{5, 6}, insteon.FirmwareVersion(42), insteon.EngineVersion(2)})
	want := `
        Device: I2Cs Device (01.02.03)
      Category: 05.06
      Firmware: 42
Engine Version: I2Cs
`[1:]

	got := device.Dump()
	if want != got {
		t.Errorf("Wanted string %q got %q", want, got)
	}
}

func TestI1DeviceCommands(t *testing.T) {
	tests := []struct {
		name        string
		version     insteon.EngineVersion
		run         func(*BasicDevice)
		want        commands.Command
		wantPayload []byte
	}{
		{"EnterLinkingMode", insteon.VerI2, func(d *BasicDevice) { d.EnterLinkingMode(40) }, commands.EnterLinkingMode.SubCommand(40), []byte{}},
		{"EnterLinkingMode Ver2Cs", insteon.VerI2Cs, func(d *BasicDevice) { d.EnterLinkingMode(40) }, commands.EnterLinkingModeExt.SubCommand(40), make([]byte, 14)},
		{"EnterUnlinkingMode", insteon.VerI2, func(d *BasicDevice) { d.EnterUnlinkingMode(41) }, commands.EnterUnlinkingMode.SubCommand(41), []byte{}},
		{"EnterUnlinkingMode Ver2Cs", insteon.VerI2Cs, func(d *BasicDevice) { d.EnterUnlinkingMode(41) }, commands.EnterUnlinkingMode.SubCommand(41), make([]byte, 14)},
		{"ExitLinkingMode", insteon.VerI2, func(d *BasicDevice) { d.ExitLinkingMode() }, commands.ExitLinkingMode, []byte{}},
	}

	oldWait := LinkingModeWaitTime
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			LinkingModeWaitTime = 0
			if len(test.wantPayload) > 0 {
				setChecksum(test.want, test.wantPayload)
			}

			tw := &testWriter{}
			device := &BasicDevice{MessageWriter: tw, DeviceInfo: DeviceInfo{EngineVersion: test.version}}
			test.run(device)
			if test.want != tw.written[0].Command {
				t.Errorf("Wanted command %v got %v", test.want, tw.written[0].Command)
			}

			if !bytes.Equal(test.wantPayload, tw.written[0].Payload) {
				t.Errorf("Wanted payload %v got %v", test.wantPayload, tw.written[0].Payload)
			}
		})
		LinkingModeWaitTime = oldWait
	}
}

func TestI1DeviceExtendedGet(t *testing.T) {
	wantPayload := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14}
	tw := &testWriter{
		read: []*insteon.Message{{Command: commands.ExtendedGetSet, Payload: wantPayload}},
	}
	d := New(tw, DeviceInfo{})
	gotPayload, err := d.ExtendedGet(make([]byte, 14))
	if err == nil {
		if tw.written[0].Command != commands.ExtendedGetSet {
			t.Errorf("Wanted command %v got %v", commands.ExtendedGetSet, tw.written[0].Command)
		}

		if !bytes.Equal(wantPayload, gotPayload) {
			t.Errorf("Wanted bytes %v got %v", wantPayload, gotPayload)
		}
	} else {
		t.Errorf("Unexpected error: %v", err)
	}
}
