package util

import (
	"strings"

	"github.com/abates/insteon"
)

type Addresses []insteon.Address

func (a *Addresses) Set(str []string) error {
	for _, s := range str {
		addr := insteon.Address(0)
		err := addr.Set(s)
		if err != nil {
			return err
		}
		*a = append(*a, addr)
	}
	return nil
}

func (a *Addresses) Get() interface{} {
	return Addresses(*a)
}

func (a *Addresses) String() string {
	strs := []string{}
	for _, addr := range *a {
		strs = append(strs, addr.String())
	}
	return strings.Join(strs, " ")
}
