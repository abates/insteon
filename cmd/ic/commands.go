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
	"strings"

	"github.com/abates/cli"
	"github.com/abates/insteon"
)

type cmdVar struct {
	insteon.Command
}

// Set satisfies the flag.Value interface
func (cmd *cmdVar) Set(str string) error {
	// Support non-period separated input too.
	if len(str) == 4 {
		str = strings.Join([]string{str[0:2], str[2:4]}, ".")
	}

	if len(str) != 5 {
		return fmt.Errorf("Bad command format need xx.xx or xxxx where xx represents a valid hex value.  Got: %v", str)
	}
	var c1, c2 int
	_, err := fmt.Sscanf(str, "%2x.%2x", &c1, &c2)
	if err != nil {
		return fmt.Errorf("Bad command format need xx.xx or xxxx where xx represents a valid hex value.  Got: %v", str)
	}
	cmd.Command = insteon.Command((c1&0xff)<<8 | c2&0xff)
	return nil
}

type Command interface {
	Name() string
	Desc() string
	Usage() string
	Setup(*cli.Arguments)
	Command() (insteon.Command, []byte)
}

type command struct {
	name  string
	desc  string
	usage string
}

func (c *command) Name() string         { return c.name }
func (c *command) Desc() string         { return c.desc }
func (c *command) Usage() string        { return c.usage }
func (c *command) Setup(*cli.Arguments) {}

type voidCmd struct {
	*command
	cmd insteon.Command
}

func Cmd(name, desc string, cmd insteon.Command) Command {
	return &voidCmd{&command{name, desc, ""}, cmd}
}

func (c *voidCmd) Command() (insteon.Command, []byte) { return c.cmd, nil }

type intCmd struct {
	*command
	arg int
	cmd func(int) (insteon.Command, []byte)
}

func IntCmd(name, desc, usage string, cmd func(int) (insteon.Command, []byte)) Command {
	return &intCmd{&command{name, desc, usage}, 0, cmd}
}

func (ic *intCmd) Setup(args *cli.Arguments) {
	args.Int(&ic.arg, "")
}

func (ic *intCmd) Command() (insteon.Command, []byte) {
	return ic.cmd(ic.arg)
}

type twintCmd struct {
	*command
	arg1 int
	arg2 int
	cmd  func(int, int) (insteon.Command, []byte)
}

func TwintCmd(name, desc, usage string, cmd func(int, int) (insteon.Command, []byte)) Command {
	return &twintCmd{&command{name, desc, usage}, 0, 0, cmd}
}

func (tc *twintCmd) Setup(args *cli.Arguments) {
	args.Int(&tc.arg1, "")
	args.Int(&tc.arg2, "")
}

func (tc *twintCmd) Command() (insteon.Command, []byte) {
	return tc.cmd(tc.arg1, tc.arg2)
}

type boolCmd struct {
	*command
	arg bool
	cmd func(bool) (insteon.Command, []byte)
}

func BoolCmd(name, desc, usage string, cmd func(bool) (insteon.Command, []byte)) Command {
	return &boolCmd{&command{name, desc, usage}, false, cmd}
}

func (bc *boolCmd) Setup(args *cli.Arguments) {
	args.Bool(&bc.arg, "")
}

func (bc *boolCmd) Command() (insteon.Command, []byte) {
	return bc.cmd(bc.arg)
}
