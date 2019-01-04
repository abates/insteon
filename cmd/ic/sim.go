package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

type simReaderWriter struct {
}

func (srw *simReaderWriter) Read(p []byte) (n int, err error) {
	var buf []byte
	reader := bufio.NewReader(os.Stdin)
	fmt.Fprintf(os.Stderr, " Read %d bytes: ", len(p))
	line, _, err := reader.ReadLine()
	if err == nil {
		str := strings.TrimSpace(strings.Replace(strings.Replace(string(line), "0x", "", -1), " ", "", -1))
		buf, err = hex.DecodeString(str)
		copy(p, buf)
	}
	return len(buf), err
}

func (srw *simReaderWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
