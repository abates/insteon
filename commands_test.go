package insteon

import (
	"reflect"
	"testing"
)

func TestCommandBytesSubCommand(t *testing.T) {
	tests := []struct {
		input      CommandBytes
		subCommand int
		expected   CommandBytes
	}{
		{CommandBytes{Command1: 0x01, Command2: 0x02}, 3, CommandBytes{Command1: 0x01, Command2: 0x03}},
	}

	for i, test := range tests {
		subCommand := test.input.SubCommand(test.subCommand)
		if test.expected != subCommand {
			t.Errorf("tests[%d] expected %v got %v", i, test.expected, subCommand)
		}
	}
}

func TestFirmwareIndexSwap(t *testing.T) {
	tests := []struct {
		input    []int
		i        int
		j        int
		expected []int
	}{
		{[]int{0, 1, 2, 3, 4}, 1, 2, []int{0, 2, 1, 3, 4}},
	}

	for i, test := range tests {
		fi := make(FirmwareIndex, len(test.input))
		for x, value := range test.input {
			fi[x] = &CommandBytes{FirmwareVersion(value), 0x00, 0x00}
		}
		fi.Swap(test.i, test.j)
		values := make([]int, len(test.input))
		for x, value := range fi {
			values[x] = int(value.version)
		}

		if !reflect.DeepEqual(test.expected, values) {
			t.Errorf("tests[%d] expected %v got %v", i, test.expected, values)
		}
	}
}

func TestFirmwareIndexFind(t *testing.T) {
	tests := []struct {
		input    []byte
		find     int
		expected byte
	}{
		{[]byte{0xf0}, 0, 0xf0},
		{[]byte{0xf0, 0xf1, 0xf2, 0xf3, 0xf4}, 2, 0xf2},
		{[]byte{0xf0, 0xf1, 0xf2, 0xf3, 0xf4}, 5, 0xf4},
	}

	for i, test := range tests {
		command := &Command{}
		for x, value := range test.input {
			command.Register(FirmwareVersion(x), value, 0x00)
		}

		value := command.Version(FirmwareVersion(test.find))
		if value.Command1 != test.expected {
			t.Errorf("tests[%d] expected 0x%x got 0x%x", i, test.expected, value.Command1)
		}
	}
}
