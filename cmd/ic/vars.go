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

package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/abates/insteon"
	"github.com/abates/insteon/commands"
)

type addrSlice []insteon.Address

func (a *addrSlice) Set(str []string) error {
	for _, s := range str {
		a := &insteon.Address{}
		err := a.Set(s)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *addrSlice) String() string {
	return fmt.Sprintf("%v", []insteon.Address(*a))
}

type dataVar []byte

func (d *dataVar) Set(str []string) error {
	for _, s := range str {
		var b byte
		_, err := fmt.Sscanf(s, "%x", &b)
		if err == nil {
			*d = append(*d, b)
		} else {
			return err
		}
	}
	return nil
}

func (d *dataVar) String() string { return fmt.Sprintf("%v", *d) }

type cmdVar struct {
	commands.Command
}

// Set satisfies the flag.Value interface
func (cmd *cmdVar) Set(str string) error {
	// Support non-period separated input too.
	index := strings.Index(str, ".")
	if index != -1 {
		str = strings.Join([]string{str[0:index], str[index+1:]}, "")
	}

	if len(str) != 4 {
		return &strconv.NumError{"cmdVar.Set", str, strconv.ErrSyntax}
	}

	c1, err := strconv.ParseInt(str[0:2], 16, 8)
	//c1, err := strconv.Atoi(str[0:2])
	if err != nil {
		return err
	}

	c2, err := strconv.ParseInt(str[2:], 16, 8)
	//c2, err := strconv.Atoi(str[2:])
	if err != nil {
		return err
	}
	cmd.Command = commands.Command((c1&0xff)<<8 | c2&0xff)
	return nil
}
