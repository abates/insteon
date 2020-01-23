//go:generate go run internal/autogen_commands.go
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

package insteon

import "fmt"

// Command is a 3 byte sequence that indicates command flags (direct, all-link or broadcast, standard/extended)
// and two byte commands
type Command [3]byte

// SubCommand will return a new command where the subcommand byte is updated
// to reflect command2 from the arguments
func (cmd Command) SubCommand(command2 int) Command {
	return Command{cmd[0], cmd[1], byte(command2)}
}

func (cmd Command) String() string {
	if str, found := cmdStrings[cmd]; found {
		return str
	} else if str, found := cmdStrings[Command{cmd[0], cmd[1], 0x00}]; found {
		return fmt.Sprintf("%s(%d)", str, cmd[2])
	}
	return fmt.Sprintf("Command(0x%02x, 0x%02x, 0x%02x)", cmd[0], cmd[1], cmd[2])
}

func AssignToAllLinkGroup(group Group) (Command, []byte) {
	return CmdAssignToAllLinkGroup.SubCommand(int(group)), nil
}

func DeleteFromAllLinkGroup(group Group) (Command, []byte) {
	return CmdDeleteFromAllLinkGroup.SubCommand(int(group)), nil
}

func Ping() (Command, []byte) {
	return CmdPing, nil
}

// EnterLinkingMode is the programmatic equivalent of holding down
// the set button for two seconds. If the device is the first
// to enter linking mode, then it is the controller. The next
// device to enter linking mode is the responder.  LinkingMode
// is usually indicated by a flashing GREEN LED on the device
func EnterLinkingMode(group Group) (Command, []byte) {
	return CmdEnterLinkingMode.SubCommand(int(group)), nil
}

// EnterUnlinkingMode puts a controller device into unlinking mode
// when the set button is then pushed (EnterLinkingMode) on a linked
// device the corresponding links in both the controller and responder
// are deleted.  EnterUnlinkingMode is the programmatic equivalent
// to pressing the set button until the device beeps, releasing, then
// pressing the set button again until the device beeps again. UnlinkingMode
// is usually indicated by a flashing RED LED on the device
func EnterUnlinkingMode(group Group) (Command, []byte) {
	return CmdEnterUnlinkingMode.SubCommand(int(group)), nil
}

// ExitLinkingMode takes a controller out of linking/unlinking mode.
func ExitLinkingMode() (Command, []byte) {
	return CmdExitLinkingMode, nil
}

func TurnLightOn(level int) (Command, []byte) {
	return CmdLightOn.SubCommand(level), nil
}

func TurnLightOnFast(level int) (Command, []byte) {
	return CmdLightOnFast.SubCommand(level), nil
}

func Brighten() (Command, []byte) {
	return CmdLightBrighten, nil
}

func Dim() (Command, []byte) {
	return CmdLightDim, nil
}

func StartBrighten() (Command, []byte) {
	return CmdLightStartManual.SubCommand(1), nil
}

func StartDim() (Command, []byte) {
	return CmdLightStartManual.SubCommand(0), nil
}

func StopChange() (Command, []byte) {
	return CmdLightStopManual, nil
}

func InstantChange(level int) (Command, []byte) {
	return CmdLightInstantChange.SubCommand(level), nil
}

func SetLightStatus(level int) (Command, []byte) {
	return CmdLightSetStatus.SubCommand(level), nil
}

func OnAtRamp(level, ramp int) (Command, []byte) {
	levelRamp := byte(level) << 4
	levelRamp |= byte(ramp) & 0x0f
	return CmdLightOnAtRamp.SubCommand(int(levelRamp)), nil
}

func OffAtRamp(ramp int) (Command, []byte) {
	return CmdLightOffAtRamp.SubCommand(0x0f & ramp), nil
}

func SetDefaultRamp(rate int) (Command, []byte) {
	// See notes on DimmerConfig() for information about D1 (payload[0])
	return CmdExtendedGetSet, []byte{0x01, 0x05, byte(rate)}
}

func SetDefaultOnLevel(level int) (Command, []byte) {
	// See notes on DimmerConfig() for information about D1 (payload[0])
	return CmdExtendedGetSet, []byte{0x01, 0x06, byte(level)}
}

func SetX10Address(button int, houseCode, unitCode byte) (Command, []byte) {
	return CmdExtendedGetSet, []byte{byte(button), 0x04, houseCode, unitCode}
}

func setOperatingFlags(flags byte, conditional bool) (Command, []byte) {
	if conditional {
		return CmdSetOperatingFlags.SubCommand(int(flags)), nil
	}
	return CmdSetOperatingFlags.SubCommand(int(flags) + 1), nil
}

func SetProgramLock(flag bool) (Command, []byte) { return setOperatingFlags(0, flag) }
func SetTxLED(flag bool) (Command, []byte)       { return setOperatingFlags(2, flag) }
func SetResumeDim(flag bool) (Command, []byte)   { return setOperatingFlags(4, flag) }
func SetLoadSense(flag bool) (Command, []byte)   { return setOperatingFlags(6, !flag) }
func SetLED(flag bool) (Command, []byte)         { return setOperatingFlags(8, !flag) }
