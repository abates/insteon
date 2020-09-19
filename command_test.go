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
	"encoding/hex"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestCommandSubCommand(t *testing.T) {
	cmd := Command(0x000102)
	cmd = cmd.SubCommand(3)
	if cmd.Command2() != 3 {
		t.Errorf("Expected 3 got %v", cmd.Command2())
	}
}

func TestCommandSet(t *testing.T) {
	cmd := Command(0x000102)
	cmd.Set("42")
	if cmd.Command2() != 42 {
		t.Errorf("Expected 42 got %v", cmd.Command2())
	}
}

func TestCommandString(t *testing.T) {
	fs := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fs, "commands.go", nil, parser.ParseComments)
	if err == nil {
		for _, decl := range parsedFile.Decls {
			decl, ok := decl.(*ast.GenDecl)

			if !ok || decl.Tok != token.VAR {
				continue
			}

			for _, spec := range decl.Specs {
				spec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}

				name := spec.Names[0]
				str := strings.TrimSpace(spec.Comment.Text())
				i, ok := spec.Values[0].(*ast.CompositeLit)

				if !ok {
					continue
				}

				if t, ok := i.Type.(*ast.Ident); !ok || t.Name != "Command" {
					continue
				}

				valueStr := ""
				for _, v := range i.Elts {
					valueStr += strings.TrimPrefix(v.(*ast.BasicLit).Value, "0x")
				}
				value, _ := hex.DecodeString(valueStr)
				cmd := Command((value[0] << 16) | (value[1] << 8) | value[2])
				if cmd.String() != str {
					t.Errorf("Expected %v string to be %q but got %q", name, str, cmd.String())
				}
			}
		}
	} else {
		t.Errorf("Failed to parse file: %v", err)
	}

	cmd := CmdAssignToAllLinkGroup.SubCommand(1)
	expected := fmt.Sprintf("%s(1)", CmdAssignToAllLinkGroup.String())
	if cmd.String() != expected {
		t.Errorf("expected %q got %q", expected, cmd.String())
	}

	cmd = Command(0xffffff)
	expected = "Command(0xff, 0xff, 0xff)"
	if cmd.String() != expected {
		t.Errorf("expected %q got %q", expected, cmd.String())
	}
}

func TestCommandGenerators(t *testing.T) {
	tests := []struct {
		name        string
		input       func() (Command, []byte)
		wantCmd     Command
		wantPayload []byte
	}{
		{"AssignToAllLinkGroup", func() (Command, []byte) { return AssignToAllLinkGroup(10) }, CmdAssignToAllLinkGroup.SubCommand(10), nil},
		{"DeleteFromAllLinkGroup", func() (Command, []byte) { return DeleteFromAllLinkGroup(10) }, CmdDeleteFromAllLinkGroup.SubCommand(10), nil},
		{"EnterLinkingMode", func() (Command, []byte) { return EnterLinkingMode(10) }, CmdEnterLinkingMode.SubCommand(10), nil},
		{"EnterUnlinkingMode", func() (Command, []byte) { return EnterUnlinkingMode(10) }, CmdEnterUnlinkingMode.SubCommand(10), nil},
		{"Turn Light On", func() (Command, []byte) { return TurnLightOn(27) }, CmdLightOn.SubCommand(27), nil},
		{"Turn Light On Fast", func() (Command, []byte) { return TurnLightOnFast(42) }, CmdLightOnFast.SubCommand(42), nil},
		{"Light Instant Change", func() (Command, []byte) { return InstantChange(40) }, CmdLightInstantChange.SubCommand(40), nil},
		{"Set Light Status", func() (Command, []byte) { return SetLightStatus(2) }, CmdLightSetStatus.SubCommand(2), nil},
		{"Light On At Ramp", func() (Command, []byte) { return LightOnAtRamp(0x42, 0x54) }, CmdLightOnAtRamp.SubCommand(0x24), nil},
		{"Light Off At Ramp", func() (Command, []byte) { return LightOffAtRamp(0x37) }, CmdLightOffAtRamp.SubCommand(0x7), nil},
		{"Light Default Ramp", func() (Command, []byte) { return SetDefaultRamp(0x37) }, CmdExtendedGetSet, []byte{0x01, 0x05, 0x37}},
		{"Light Default On Level", func() (Command, []byte) { return SetDefaultOnLevel(0x37) }, CmdExtendedGetSet, []byte{0x01, 0x06, 0x37}},
		{"SetX10Address", func() (Command, []byte) { return SetX10Address(7, 8, 9) }, CmdExtendedGetSet, []byte{7, 4, 8, 9}},
		{"Backlight On", func() (Command, []byte) { return Backlight(true) }, CmdEnableLED, make([]byte, 14)},
		{"Backlight Off", func() (Command, []byte) { return Backlight(false) }, CmdDisableLED, make([]byte, 14)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotCmd, gotPayload := test.input()
			if test.wantCmd != gotCmd {
				t.Errorf("Wanted command %v got %v", test.wantCmd, gotCmd)
			}

			if !bytes.Equal(test.wantPayload, gotPayload) {
				t.Errorf("Wanted payload %v got %v", test.wantPayload, gotPayload)
			}
		})
	}
}
