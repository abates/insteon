package devices

import "github.com/abates/insteon/commands"

func setChecksum(cmd commands.Command, buf []byte) {
	buf[len(buf)-1] = checksum(cmd, buf)
}

func checksum(cmd commands.Command, buf []byte) byte {
	sum := byte(cmd.Command1() + cmd.Command2())
	for _, b := range buf {
		sum += b
	}
	return ^sum + 1
}
