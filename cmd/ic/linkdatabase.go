package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/abates/insteon"
)

func printLinkDatabase(linkable insteon.LinkableDevice) error {
	dbLinks, err := linkable.Links()
	fmt.Printf("Link Database:\n")
	if len(dbLinks) > 0 {
		fmt.Printf("    Flags Group Address    Data\n")

		links := make(map[string][]*insteon.LinkRecord)
		for _, link := range dbLinks {
			links[link.Address.String()] = append(links[link.Address.String()], link)
		}

		linkAddresses := []string{}
		for linkAddress := range links {
			linkAddresses = append(linkAddresses, linkAddress)
		}
		sort.Strings(linkAddresses)

		for _, linkAddress := range linkAddresses {
			for _, link := range links[linkAddress] {
				fmt.Printf("    %-5s %5s %8s   %02x %02x %02x\n", link.Flags, link.Group, link.Address, link.Data[0], link.Data[1], link.Data[2])
			}
		}
	} else {
		fmt.Printf("    No links defined\n")
	}
	return err
}

func dumpLinkDatabase(linkable insteon.LinkableDevice) error {
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
