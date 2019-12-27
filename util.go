package insteon

import (
	"time"
)

// PropagationDelay attempts to calculate the amount of time an Insteon
// message needs in order to completely propogate (including relays) throughout
// the network.  This is based on the number of AC zero crossings, the
// number of times a message can be repeated (ttl) and whether or not
// the message is an extended message
func PropagationDelay(ttl int, extended bool) (pd time.Duration) {
	// wait 2 * ttl * message length zero crossings
	if extended {
		pd = time.Second * 26 * time.Duration(ttl) / 60
	} else {
		pd = time.Second * 12 * time.Duration(ttl) / 60
	}
	return
}
