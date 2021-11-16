package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/abates/insteon/util"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <file to anonymize>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	inputFile := os.Args[1]
	inputStr, err := ioutil.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read file %v: %v\n", inputFile, err)
		os.Exit(1)
	}

	comments := []string{}
	for _, line := range strings.Split(string(inputStr), "\n") {
		if strings.HasPrefix(line, "#") {
			comments = append(comments, line)
		}
	}
	if len(comments) > 0 {
		comments = append(comments, "")
	}

	links, err := util.TextToLinks(string(inputStr))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to convert input file to links %v\n", err)
		os.Exit(1)
	}

	links = util.Anonymize(links)
	output := fmt.Sprintf("%s%s", strings.Join(comments, "\n"), util.LinksToText(links, false))

	err = ioutil.WriteFile(inputFile, []byte(output), 0640)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write file %v\n", err)
		os.Exit(1)
	}
}
