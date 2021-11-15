// Copyright 2018 Andrew Bates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package insteon

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
)

// Address is a 3 byte insteon address
type Address uint32

// String will format the Address object into a form
// common to Insteon devices: 00.00.00 where each byte
// is represented in hexadecimal form (e.g. 01.b4.a5) the
// string will always be 8 characters long, bytes are zero
// padded
func (a Address) String() string {
	return fmt.Sprintf("%02x.%02x.%02x", byte(a>>16), byte(a>>8), byte(a))
}

// UnmarshalText converts a human readable string into an
// Insteon address. If the address cannot be parsed then
// UnmarshalText returns an ErrAddressFormat error
func (a *Address) UnmarshalText(text []byte) error {
	// Support non-period separated input too.
	if len(text) == 6 {
		text = bytes.Join([][]byte{text[0:2], text[2:4], text[4:6]}, []byte("."))
	}

	var b [4]byte
	_, err := fmt.Sscanf(string(text), "%2x.%2x.%2x", &b[1], &b[2], &b[3])
	if err != nil {
		return ErrAddrFormat
	}
	*a = Address(binary.BigEndian.Uint32(b[:]))
	return nil
}

// Set satisfies the flag.Value interface
func (a *Address) Set(str string) error {
	return a.UnmarshalText([]byte(str))
}

// Set satisfies the flag.Getter interface
func (a *Address) Get() interface{} {
	return Address(*a)
}

func (a *Address) Put(buf []byte) {
	b := make([]byte, 4)
	copy(b[1:], buf)
	*a = Address(binary.BigEndian.Uint32(b))
}

func (a *Address) Bytes() []byte {
	b := [4]byte{}
	binary.BigEndian.PutUint32(b[:], uint32(*a))
	return b[1:]
}

// MarshalText fulfills the requiresments of encoding.TextMarshaler so that
// Address can be used as a map key in other encoding
func (a Address) MarshalText() ([]byte, error) {
	return []byte(a.String()), nil
}

// MarshalJSON will convert the address to a JSON string
func (a Address) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

// UnmarshalJSON will populate the address from the input JSON string
func (a *Address) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err = json.Unmarshal(data, &s); err == nil {
		err = a.Set(s)
	}
	return err
}
