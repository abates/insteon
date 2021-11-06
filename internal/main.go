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
	"bytes"
	"fmt"
	"go/format"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type autogenTemplate struct {
	input  string
	output string
	data   func() interface{}
}

type autogenCommand struct {
	templates []autogenTemplate
}

var autogenCommands = make(map[string]autogenCommand)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <command>", os.Args[0])
	}

	command, found := autogenCommands[os.Args[1]]
	if !found {
		log.Fatalf("Command %q not found", os.Args[1])
	}

	for _, tmpl := range command.templates {
		cmdsTmpl := template.Must(template.New("").ParseFiles("internal/license.tmpl", tmpl.input))

		buf := bytes.NewBuffer(make([]byte, 0))
		err := cmdsTmpl.ExecuteTemplate(buf, filepath.Base(tmpl.input), struct {
			Copyright string
			Owner     string
			Package   string
			Data      interface{}
		}{
			Copyright: fmt.Sprintf("%4d", time.Now().Year()),
			Owner:     "Andrew Bates",
			Data:      tmpl.data(),
		})

		if err != nil {
			log.Fatalf("%s: Failed to execute template: %v", tmpl.input, err)
		}

		f, err := os.Create(tmpl.output)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		b := buf.Bytes()
		if strings.HasSuffix(tmpl.output, ".go") {
			b, err = format.Source(b)
			if err != nil {
				f.Write(buf.Bytes()) // This is here to debug bad format
				log.Fatalf("%s: error formatting: %s", tmpl.output, err)
			}
		}
		f.Write(b)
	}
}
