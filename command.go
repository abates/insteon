//go:generate go run internal/autogen_commands.go

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
		return str
	}
	return fmt.Sprintf("Command(0x%02x, 0x%02x, 0x%02x)", cmd[0], cmd[1], cmd[2])
}
