package insteon

import (
	"time"
)

// PropagationDelay attempts to calculate the amount of time an Insteon
// message needs in order to completely propagate (including relays) throughout
// the network.  This is based on the number of AC zero crossings, the
// number of times a message can be repeated (ttl) and whether or not
// the message is an extended message
func PropagationDelay(ttl uint8, l int) (pd time.Duration) {
	// wait 2 * ttl * message length zero crossings
	return time.Second * time.Duration(l) * time.Duration(ttl+1) / 60
}

// ReadWithTimeout will attempt to read a message from a channel and will
// return the read message.  If no message is received after the timeout
// duration, then ErrReadTimeout is returned
func ReadWithTimeout(ch <-chan *Message, timeout time.Duration) (msg *Message, err error) {
	select {
	case msg = <-ch:
	case <-time.After(timeout):
		err = ErrReadTimeout
	}
	return
}

/*func setChecksum(cmd commands.Command, buf []byte) {
	buf[len(buf)-1] = checksum(cmd, buf)
}

func checksum(cmd commands.Command, buf []byte) byte {
	sum := byte(cmd.Command1() + cmd.Command2())
	for _, b := range buf {
		sum += b
	}
	return ^sum + 1
}*/
