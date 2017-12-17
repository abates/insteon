package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/abates/insteon"
)

func printLinkDatabase(db insteon.LinkDB) error {
	dbLinks, err := db.Links()
	if err == nil {
		fmt.Printf("Link Database:\n")
		fmt.Printf("Flags Group Address    Data\n")

		links := make(map[string][]*insteon.Link)
		for _, link := range dbLinks {
			links[link.Address.String()] = append(links[link.Address.String()], link)
		}

		linkAddresses := []string{}
		for linkAddress, _ := range links {
			linkAddresses = append(linkAddresses, linkAddress)
		}
		sort.Strings(linkAddresses)

		for _, linkAddress := range linkAddresses {
			for _, link := range links[linkAddress] {
				fmt.Printf("%-5s %5s %8s   %02x %02x %02x\n", link.Flags, link.Group, link.Address, link.Data[0], link.Data[1], link.Data[2])
			}
		}
	}
	return err
}

func dumpLinkDatabase(db insteon.LinkDB) error {
	links, err := db.Links()
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
