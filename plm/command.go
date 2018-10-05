//go:generate stringer -type=Command -linecomment=true
package plm

var commandLens map[Command]int

type Command byte

const (
	CmdNak                   Command = 0x15 // NAK
	CmdStdMsgReceived        Command = 0x50 // Std Msg Received
	CmdExtMsgReceived        Command = 0x51 // Ext Msg Received
	CmdX10MsgReceived        Command = 0x52 // X10 Msg Received
	CmdAllLinkComplete       Command = 0x53 // All Link Complete
	CmdButtonEventReport     Command = 0x54 // Button Event Report
	CmdUserResetDetected     Command = 0x55 // User Reset Detected
	CmdAllLinkCleanupFailure Command = 0x56 // Link Cleanup Report
	CmdAllLinkRecordResp     Command = 0x57 // Link Record Resp
	CmdAllLinkCleanupStatus  Command = 0x58 // Link Cleanup Status
	CmdGetInfo               Command = 0x60 // Get Info
	CmdSendAllLink           Command = 0x61 // Send All Link
	CmdSendInsteonMsg        Command = 0x62 // Send INSTEON Msg
	CmdSendX10               Command = 0x63 // Send X10 Msg
	CmdStartAllLink          Command = 0x64 // Start All Link
	CmdCancelAllLink         Command = 0x65 // Cancel All Link
	CmdSetHostCategory       Command = 0x66 // Set Host Category
	CmdReset                 Command = 0x67 // Reset
	CmdSetAckMsg             Command = 0x68 // Set ACK Msg
	CmdGetFirstAllLink       Command = 0x69 // Get First All Link
	CmdGetNextAllLink        Command = 0x6a // Get Next All Link
	CmdSetConfig             Command = 0x6b // Set Config
	CmdGetAllLinkForSender   Command = 0x6c // Get Sender All Link
	CmdLedOn                 Command = 0x6d // LED On
	CmdLedOff                Command = 0x6e // LED Off
	CmdManageAllLinkRecord   Command = 0x6f // Manage All Link Record
	CmdSetNakMsgByte         Command = 0x70 // Set NAK Msg Byte
	CmdSetNameMsgTwoBytes    Command = 0x71 // Set NAK Msg Two Bytes
	CmdRfSleep               Command = 0x72 // RF Sleep
	CmdGetConfig             Command = 0x73 // Get Config
)

func init() {
	commandLens = map[Command]int{
		CmdNak:                   1,
		CmdStdMsgReceived:        9,
		CmdExtMsgReceived:        23,
		CmdX10MsgReceived:        2,
		CmdAllLinkComplete:       8,
		CmdButtonEventReport:     1,
		CmdUserResetDetected:     0,
		CmdAllLinkCleanupFailure: 5,
		CmdAllLinkRecordResp:     8,
		CmdAllLinkCleanupStatus:  1,
		CmdGetInfo:               7,
		CmdSendAllLink:           4,
		CmdSendInsteonMsg:        7,
		CmdSendX10:               3,
		CmdStartAllLink:          3,
		CmdCancelAllLink:         1,
		CmdSetHostCategory:       4,
		CmdReset:                 1,
		CmdSetAckMsg:             2,
		CmdGetFirstAllLink:       1,
		CmdGetNextAllLink:        1,
		CmdSetConfig:             2,
		CmdGetAllLinkForSender:   1,
		CmdLedOn:                 1,
		CmdLedOff:                1,
		CmdManageAllLinkRecord:   10,
		CmdSetNakMsgByte:         2,
		CmdSetNameMsgTwoBytes:    3,
		CmdRfSleep:               1,
		CmdGetConfig:             4,
	}
}
