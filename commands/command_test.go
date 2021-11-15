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

package commands

import (
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

	cmd := AssignToAllLinkGroup.SubCommand(1)
	expected := fmt.Sprintf("%s(1)", AssignToAllLinkGroup.String())
	if cmd.String() != expected {
		t.Errorf("expected %q got %q", expected, cmd.String())
	}

	cmd = Command(0xffffff)
	expected = "Command(0xff, 0xff, 0xff)"
	if cmd.String() != expected {
		t.Errorf("expected %q got %q", expected, cmd.String())
	}
}

func TestFrom(t *testing.T) {
	tests := []struct {
		name  string
		input [3]int
		want  Command
	}{
		{"standard direct", [3]int{0x0a, AssignToAllLinkGroup.Command1(), AssignToAllLinkGroup.Command2()}, AssignToAllLinkGroup},
		{"standard direct ACK", [3]int{0x2a, AssignToAllLinkGroup.Command1(), AssignToAllLinkGroup.Command2()}, AssignToAllLinkGroup},
		{"standard direct NAK", [3]int{0xaa, AssignToAllLinkGroup.Command1(), AssignToAllLinkGroup.Command2()}, AssignToAllLinkGroup},
		{"extended direct", [3]int{0x1a, ProductDataResp.Command1(), ProductDataResp.Command2()}, ProductDataResp},
		{"extended direct ACK", [3]int{0x3a, ProductDataResp.Command1(), ProductDataResp.Command2()}, ProductDataResp},
		{"extended direct NAK", [3]int{0xfa, ProductDataResp.Command1(), ProductDataResp.Command2()}, ProductDataResp},
		{"all-link cleanup", [3]int{0x4a, AllLinkRecall.Command1(), AllLinkRecall.Command2()}, AllLinkRecall},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := From(byte(test.input[0]), byte(test.input[1]), byte(test.input[2]))
			if test.want != got {
				t.Errorf("Wanted command %v got %v", test.want, got)
			}
		})
	}
}
