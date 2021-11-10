package plm

import (
	"bytes"
	"io"
	"log"
	"reflect"
	"strings"
	"testing"
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
			buf := bytes.NewBuffer(nil)
			Log = log.New(buf, "", 0)
			LogDebug = log.New(buf, "DEBUG ", 0)
			lw := logWriter{Writer: test.writer}
			_, gotErr := lw.Write(test.input)
			lines := strings.Split(buf.String(), "\n")

			got := strings.TrimPrefix(lines[1], " INFO ")
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
			reader := newPacketReader(bytes.NewReader(test.input), false)
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
		want  Packet
	}{
		{
			name:  "NAK packet",
			input: []byte{0x02, 0x6a, 0x15},
			want:  Packet{CmdGetNextAllLink, nil, 0x15},
		},
		{
			name:  "insteon send extended packet ack",
			input: []byte{0x02, 0x62, 0x54, 0x88, 0x55, 0x1f, 0x2f, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xd1, 0x06},
			want:  Packet{CmdSendInsteonMsg, []byte{0, 0, 0, 0x54, 0x88, 0x55, 0x1f, 0x2f, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xd1}, 0x06},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reader := newPacketReader(bytes.NewReader(test.input), false)
			got, err := reader.ReadPacket()
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			} else {
				if !reflect.DeepEqual(test.want, *got) {
					t.Errorf("Wanted packet %+v got %+v", test.want, *got)
				}
			}
		})
	}
}
