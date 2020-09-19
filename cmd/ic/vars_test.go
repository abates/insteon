package main

import (
	"bytes"
	"errors"
	"strconv"
	"testing"

	"github.com/abates/insteon"
)

func TestDataVar(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		want    []byte
		wantStr string
	}{
		{"one value", []string{"01"}, []byte{1}, "[1]"},
		{"two values", []string{"01", "02"}, []byte{1, 2}, "[1 2]"},
		{"hex values", []string{"a", "b"}, []byte{10, 11}, "[10 11]"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := dataVar{}
			for _, s := range test.input {
				err := got.Set(s)
				if err != nil {
					t.Errorf("Unexpected error %v", err)
				}
			}

			if !bytes.Equal(test.want, []byte(got)) {
				t.Errorf("Wanted %v got %v", test.want, got)
			}

			if test.wantStr != got.String() {
				t.Errorf("Wanted string %q got %q", test.wantStr, got.String())
			}
		})
	}
}

func TestCmdVar(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    insteon.Command
		wantErr error
	}{
		{"common format", "01.02", insteon.Command(0x0102), nil},
		{"no dots", "0102", insteon.Command(0x0102), nil},
		{"short", "123", insteon.Command(0), strconv.ErrSyntax},
		{"first byte syntax error", "xx.01", insteon.Command(0), strconv.ErrSyntax},
		{"second byte syntax error", "01.xx", insteon.Command(0), strconv.ErrSyntax},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := cmdVar{}
			err := got.Set(test.input)
			if err == nil {
				if test.want != got.Command {
					t.Errorf("Wanted command %v got %v", test.want, got.Command)
				}
			} else if !errors.Is(err, test.wantErr) {
				t.Errorf("Wanted error %v got %v", test.wantErr, err)
			}
		})
	}
}
