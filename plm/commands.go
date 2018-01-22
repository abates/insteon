package plm

import (
	"fmt"
)

var commandLens = make(map[byte]int)
var commandNames = make(map[byte]string)

type Command byte

type CommandRegistry struct {
	commands map[byte]Command
}

func NewCommand(name string, cmd byte, length int) Command {
	commandNames[cmd] = name
	commandLens[cmd] = length
	return Command(cmd)
}

func (c Command) String() string {
	return fmt.Sprintf("%s", commandNames[byte(c)])
}

var (
	CmdNak                   = NewCommand("NAK", 0x15, 1)
	CmdStdMsgReceived        = NewCommand("Std Msg Received", 0x50, 9)
	CmdExtMsgReceived        = NewCommand("Ext Msg Received", 0x51, 23)
	CmdX10MsgReceived        = NewCommand("X10 Msg Received", 0x52, 2)
	CmdAllLinkComplete       = NewCommand("All Link Complete", 0x53, 8)
	CmdButtonEventReport     = NewCommand("Button Event Report", 0x54, 1)
	CmdUserResetDetected     = NewCommand("User Reset Detected", 0x55, 0)
	CmdAllLinkCleanupFailure = NewCommand("Link Cleanup Report", 0x56, 5)
	CmdAllLinkRecordResp     = NewCommand("Link Record Resp", 0x57, 8)
	CmdAllLinkCleanupStatus  = NewCommand("Link Cleanup Status", 0x58, 1)
	CmdGetInfo               = NewCommand("Get Info", 0x60, 7)
	CmdSendAllLink           = NewCommand("Send All Link", 0x61, 4)
	CmdSendInsteonMsg        = NewCommand("Send INSTEON Msg", 0x62, 7)
	CmdSendX10               = NewCommand("Send X10 Msg", 0x63, 3)
	CmdStartAllLink          = NewCommand("Start All Link", 0x64, 3)
	CmdCancelAllLink         = NewCommand("Cancel All Link", 0x65, 1)
	CmdSetHostCategory       = NewCommand("Set Host Category", 0x66, 4)
	CmdReset                 = NewCommand("Reset", 0x67, 1)
	CmdSetAckMsg             = NewCommand("Set ACK Msg", 0x68, 2)
	CmdGetFirstAllLink       = NewCommand("Get First All Link", 0x69, 1)
	CmdGetNextAllLink        = NewCommand("Get Next All Link", 0x6a, 1)
	CmdSetConfig             = NewCommand("Set Config", 0x6b, 2)
	CmdGetAllLinkForSender   = NewCommand("Get Sender All Link", 0x6c, 1)
	CmdLedOn                 = NewCommand("LED On", 0x6d, 1)
	CmdLedOff                = NewCommand("LED Off", 0x6e, 1)
	CmdManageAllLinkRecord   = NewCommand("Manage All Link Record", 0x6f, 10)
	CmdSetNakMsgByte         = NewCommand("Set NAK Msg Byte", 0x70, 2)
	CmdSetNameMsgTwoBytes    = NewCommand("Set NAK Msg Two Bytes", 0x71, 3)
	CmdRfSleep               = NewCommand("RF Slee", 0x72, 1)
	CmdGetConfig             = NewCommand("Get Config", 0x73, 4)
)
