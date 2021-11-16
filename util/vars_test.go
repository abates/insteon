package util

import (
	"reflect"
	"testing"

	"github.com/abates/insteon"
)

func TestAddresses(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		want    Addresses
		wantStr string
	}{
		{"one address", []string{"01.02.03"}, Addresses{insteon.Address(0x010203)}, "01.02.03"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := Addresses{}
			err := got.Set(test.input)
			if err != nil {
				t.Errorf("Unexpected error %v", err)
			}

			if !reflect.DeepEqual(test.want, got.Get()) {
				t.Errorf("Wanted addresses %v got %v", test.want, got)
			}

			if test.wantStr != got.String() {
				t.Errorf("Wanted string %q got %q", test.wantStr, got.String())
			}
		})
	}
}
