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

package plm

import (
	"encoding/hex"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestCommand(t *testing.T) {
	fs := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fs, "command.go", nil, parser.ParseComments)
	if err == nil {
		for _, decl := range parsedFile.Decls {
			if decl, ok := decl.(*ast.GenDecl); ok && decl.Tok == token.CONST {
				for _, spec := range decl.Specs {
					if spec, ok := spec.(*ast.ValueSpec); ok {
						name := spec.Names[0]
						str := strings.TrimSpace(spec.Comment.Text())
						if i, ok := spec.Values[0].(*ast.BasicLit); ok {
							value, _ := hex.DecodeString(strings.TrimPrefix(i.Value, "0x"))
							cmd := Command(value[0])
							if cmd.String() != str {
								t.Errorf("%v: got String %q, want %q", name, cmd.String(), str)
							}

							if _, found := commandLens[cmd]; !found {
								t.Errorf("No command length found for %v", name)
							}
						}
					}
				}
			}
		}

		// check for default string
		cmd := Command(255)
		if cmd.String() != "Command(255)" {
			t.Errorf("Expected default string to be %q got %q", "Command(255)", cmd.String())
		}
	} else {
		t.Errorf("Failed to parse file: %v", err)
	}
}
