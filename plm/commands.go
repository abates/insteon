package plm

import (
	"fmt"

	"github.com/abates/insteon"
)

var commandLens = make(map[byte]int)
var commandNames = make(map[byte]string)
var payloadGenerators = make(map[byte]insteon.PayloadGenerator)

type Command byte

type CommandRegistry struct {
	commands map[byte]Command
}

func NewCommand(name string, cmd byte, length int, generator insteon.PayloadGenerator) Command {
	commandNames[cmd] = name
	commandLens[cmd] = length
	payloadGenerators[cmd] = generator
	return Command(cmd)
}

func (c Command) String() string {
	return fmt.Sprintf("%s", commandNames[byte(c)])
}

var (
	CmdNak                   = NewCommand("NAK", 0x15, 1, nil)
	CmdStdMsgReceived        = NewCommand("Std Msg Received", 0x50, 9, func() insteon.Payload { return &insteon.Message{} })
	CmdExtMsgReceived        = NewCommand("Ext Msg Received", 0x51, 23, func() insteon.Payload { return &insteon.Message{} })
	CmdX10MsgReceived        = NewCommand("X10 Msg Received", 0x52, 2, nil)
	CmdAllLinkComplete       = NewCommand("All Link Complete", 0x53, 8, nil)
	CmdButtonEventReport     = NewCommand("Button Event Report", 0x54, 1, nil)
	CmdUserResetDetected     = NewCommand("User Reset Detected", 0x55, 0, nil)
	CmdAllLinkCleanupFailure = NewCommand("Link Cleanup Report", 0x56, 5, nil)
	CmdAllLinkRecordResp     = NewCommand("Link Record Rsp", 0x57, 8, func() insteon.Payload { return &insteon.Link{} })
	CmdAllLinkCleanupStatus  = NewCommand("Link Cleanup Status", 0x58, 1, nil)
	CmdGetInfo               = NewCommand("Get Info", 0x60, 7, func() insteon.Payload { return &IMInfo{} })
	CmdSendAllLink           = NewCommand("Send All Link", 0x61, 4, nil)
	CmdSendInsteonMsg        = NewCommand("Send INSTEON Msg", 0x62, 7, func() insteon.Payload { return &insteon.Message{} })
	CmdSendX10               = NewCommand("Send X10 Msg", 0x63, 3, nil)
	CmdStartAllLink          = NewCommand("Start All Link", 0x64, 3, nil)
	CmdCancelAllLink         = NewCommand("Cancel All Link", 0x65, 1, nil)
	CmdSetHostCategory       = NewCommand("Set Host Category", 0x66, 4, nil)
	CmdReset                 = NewCommand("Reset", 0x67, 1, nil)
	CmdSetAckMsg             = NewCommand("Set ACK Msg", 0x68, 2, nil)
	CmdGetFirstAllLink       = NewCommand("Get First All Link", 0x69, 1, nil)
	CmdGetNextAllLink        = NewCommand("Get Next All Link", 0x6a, 1, nil)
	CmdSetConfig             = NewCommand("Set Config", 0x6b, 2, nil)
	CmdGetAllLinkForSender   = NewCommand("Get Sender All Link", 0x6c, 1, nil)
	CmdLedOn                 = NewCommand("LED On", 0x6d, 1, nil)
	CmdLedOff                = NewCommand("LED Off", 0x6e, 1, nil)
	CmdManageAllLinkRecord   = NewCommand("Manage All Link Record", 0x6f, 10, func() insteon.Payload { return &manageRecordRequest{} })
	CmdSetNakMsgByte         = NewCommand("Set NAK Msg Byte", 0x70, 2, nil)
	CmdSetNameMsgTwoBytes    = NewCommand("Set NAK Msg Two Bytes", 0x71, 3, nil)
	CmdRfSleep               = NewCommand("RF Slee", 0x72, 1, nil)
	CmdGetConfig             = NewCommand("Get Config", 0x73, 4, nil)
)
