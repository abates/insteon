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
	"fmt"
	"strings"

	"github.com/abates/insteon"
)

func dumpLinkDatabase(linkable insteon.Linkable) error {
	links, err := linkable.Links()
	if err == nil {
		fmt.Printf("links:\n")
		for _, link := range links {
			buf, _ := link.MarshalBinary()
			s := make([]string, len(buf))
			for i, b := range buf {
				s[i] = fmt.Sprintf("0x%02x", b)
			}
			fmt.Printf("- [ %s ]\n", strings.Join(s, ", "))
		}
	}
	return err
}
