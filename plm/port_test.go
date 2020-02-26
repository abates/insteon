package plm

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"reflect"
	"strings"
	"testing"

	"github.com/abates/insteon"
)

type ErrWriter struct {
	error
}

func (ew ErrWriter) Write([]byte) (int, error) { return 0, ew.error }

func TestLogWriter(t *testing.T) {
	tests := []struct {
		name    string
		writer  io.Writer
		input   []byte
		want    string
		wantErr error
	}{
		{"Write Error", ErrWriter{io.ErrClosedPipe}, []byte{}, "Failed to write: io: read/write on closed pipe", io.ErrClosedPipe},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			oldLog := insteon.Log
			defer func() { insteon.Log = oldLog }()
			buf := bytes.NewBuffer(nil)
			insteon.Log = &insteon.Logger{Level: insteon.LevelTrace, Logger: log.New(buf, "", 0)}
			lw := LogWriter{test.writer}
			_, gotErr := lw.Write(test.input)
			lines := strings.Split(buf.String(), "\n")
			want := fmt.Sprintf("TRACE TX %s", hexDump("%02x", test.input, " "))
			got := lines[0][strings.Index(lines[0], "TRACE"):]

			if want != got {
				t.Errorf("Wanted log %q got %q", want, got)
			}

			got = strings.TrimPrefix(lines[1], " INFO ")
			if test.want != got {
				t.Errorf("Wanted log %q got %q", test.want, got)
			}

			if test.wantErr != gotErr {
				t.Errorf("Wanted error %v got %v", test.wantErr, gotErr)
			}
		})
	}
}

func TestPacketReaderSync(t *testing.T) {
	tests := []struct {
		name       string
		input      []byte
		wantN      int
		wantPaclen int
		wantErr    error
	}{
		{"partial packet", []byte{0x88, 0x55, 0x02, 0x37, 0x48, 0x02, 0x6a, 0x15}, 2, 1, nil},
		{"start on packet boundary", []byte{0x02, 0x6a, 0x15}, 2, 1, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reader := NewPacketReader(bytes.NewReader(test.input))
			gotN, gotPaclen, gotErr := reader.sync()
			if test.wantN != gotN {
				t.Errorf("Wanted N %d got %d", test.wantN, gotN)
			}

			if test.wantPaclen != gotPaclen {
				t.Errorf("Wanted paclen %d got %d", test.wantPaclen, gotPaclen)
			}

			if test.wantErr != gotErr {
				t.Errorf("Wanted error %v got %v", test.wantErr, gotErr)
			}
		})
	}
}

func TestPacketReaderRead(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		wantN int
	}{
		{"simple packet", []byte{0x02, 0x6a, 0x15}, 3},
		//{"echo'd insteon extended packet", []byte{0x02, 0x6a, 0x15}, 3, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reader := NewPacketReader(bytes.NewReader(test.input))
			gotN, _ := reader.read()
			if test.wantN != gotN {
				t.Errorf("Wanted N %d got %d", test.wantN, gotN)
			}

			if !bytes.Equal(test.input, reader.buf[0:test.wantN]) {
				t.Errorf("Wanted bytes %x got %x", test.input, reader.buf[0:test.wantN])
			}
		})
	}
}

func TestPacketReaderReadPacket(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  Packet
	}{
		{"NAK packet", []byte{0x02, 0x6a, 0x15}, Packet{CmdGetNextAllLink, []byte{}, 0x15}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reader := NewPacketReader(bytes.NewReader(test.input))
			got, _ := reader.ReadPacket()

			if !reflect.DeepEqual(test.want, *got) {
				t.Errorf("Wanted packet %+v got %+v", test.want, *got)
			}
		})
	}
}
