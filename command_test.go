package insteon

import (
	"encoding/hex"
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
}
