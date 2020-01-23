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
	cmd := Command{0x00, 0x01, 0x02}
	cmd = cmd.SubCommand(3)
	if cmd[2] != 3 {
		t.Errorf("Expected 3 got %v", cmd[2])
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
				cmd := Command{value[0], value[1], value[2]}
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

	cmd = Command{0xff, 0xff, 0xff}
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
		{"Ping", Ping, CmdPing, nil},
		{"ExitLinkingMode", ExitLinkingMode, CmdExitLinkingMode, nil},
		{"EnterLinkingMode", func() (Command, []byte) { return EnterLinkingMode(10) }, CmdEnterLinkingMode.SubCommand(10), nil},
		{"EnterUnlinkingMode", func() (Command, []byte) { return EnterUnlinkingMode(10) }, CmdEnterUnlinkingMode.SubCommand(10), nil},
		{"Turn Light On", func() (Command, []byte) { return TurnLightOn(27) }, CmdLightOn.SubCommand(27), nil},
		{"Turn Light On Fast", func() (Command, []byte) { return TurnLightOnFast(42) }, CmdLightOnFast.SubCommand(42), nil},
		{"Brighten Light", Brighten, CmdLightBrighten, nil},
		{"Dim Light", Dim, CmdLightDim, nil},
		{"Start Brighten Light", StartBrighten, CmdLightStartManual.SubCommand(1), nil},
		{"Start Dim Light", StartDim, CmdLightStartManual.SubCommand(0), nil},
		{"Stop Light Change", StopChange, CmdLightStopManual, nil},
		{"Light Instant Change", func() (Command, []byte) { return InstantChange(40) }, CmdLightInstantChange.SubCommand(40), nil},
		{"Set Light Status", func() (Command, []byte) { return SetLightStatus(2) }, CmdLightSetStatus.SubCommand(2), nil},
		{"Light On At Ramp", func() (Command, []byte) { return OnAtRamp(0x42, 0x54) }, CmdLightOnAtRamp.SubCommand(0x24), nil},
		{"Light Off At Ramp", func() (Command, []byte) { return OffAtRamp(0x37) }, CmdLightOffAtRamp.SubCommand(0x7), nil},
		{"Light Default Ramp", func() (Command, []byte) { return SetDefaultRamp(0x37) }, CmdExtendedGetSet, []byte{0x01, 0x05, 0x37}},
		{"Light Default On Level", func() (Command, []byte) { return SetDefaultOnLevel(0x37) }, CmdExtendedGetSet, []byte{0x01, 0x06, 0x37}},
		{"Set Program Lock", func() (Command, []byte) { return SetProgramLock(true) }, CmdSetOperatingFlags.SubCommand(0), nil},
		{"Clear Program Lock", func() (Command, []byte) { return SetProgramLock(false) }, CmdSetOperatingFlags.SubCommand(1), nil},
		{"Set Tx LED", func() (Command, []byte) { return SetTxLED(true) }, CmdSetOperatingFlags.SubCommand(2), nil},
		{"Clear Tx LED", func() (Command, []byte) { return SetTxLED(false) }, CmdSetOperatingFlags.SubCommand(3), nil},
		{"Set Resume Dim", func() (Command, []byte) { return SetResumeDim(true) }, CmdSetOperatingFlags.SubCommand(4), nil},
		{"Clear Resume Dim", func() (Command, []byte) { return SetResumeDim(false) }, CmdSetOperatingFlags.SubCommand(5), nil},
		{"Set Load Sense", func() (Command, []byte) { return SetLoadSense(true) }, CmdSetOperatingFlags.SubCommand(7), nil},
		{"Clear Load Sense", func() (Command, []byte) { return SetLoadSense(false) }, CmdSetOperatingFlags.SubCommand(6), nil},
		{"Set LED", func() (Command, []byte) { return SetLED(true) }, CmdSetOperatingFlags.SubCommand(9), nil},
		{"Clear LED", func() (Command, []byte) { return SetLED(false) }, CmdSetOperatingFlags.SubCommand(8), nil},
		{"SetX10Address", func() (Command, []byte) { return SetX10Address(7, 8, 9) }, CmdExtendedGetSet, []byte{7, 4, 8, 9}},
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
