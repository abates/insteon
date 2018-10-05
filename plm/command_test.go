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
								t.Errorf("Expected %v string to be %q but got %q", name, str, cmd.String())
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
