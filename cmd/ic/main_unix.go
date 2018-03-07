// +build darwin dragonfly freebsd linux netbsd openbsd

package main

import "os"

var EDITOR = ""

func init() {
	EDITOR = os.Getenv("EDITOR")
	if EDITOR == "" {
		EDITOR = "vi"
	}
}
