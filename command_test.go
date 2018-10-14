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
