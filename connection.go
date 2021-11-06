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

import (
	"time"

	"github.com/abates/insteon/commands"
)

var (
	// See "INSTEON Message Retrying" section in Insteon Developer's Guide
	retryTimes = []struct {
		standard time.Duration
		extended time.Duration
	}{
		{1400 * time.Millisecond, 2220 * time.Millisecond},
		{1700 * time.Millisecond, 2690 * time.Millisecond},
		{1900 * time.Millisecond, 3010 * time.Millisecond},
		{2000 * time.Millisecond, 3170 * time.Millisecond},
	}
)

type messageReader interface {
	Read() (*Message, error)
}

type MessageWriter interface {
	Read() (*Message, error)
	Write(*Message) (ack *Message, err error)
}

func retry(retries int, cb func() error) (err error) {
	tries := retries
	for {
		err = cb()
		if err == nil || err != ErrReadTimeout {
			break
		}
		if retries > 1 {
			retries--
			Log.Printf("Read Timeout, retrying")
		} else {
			Log.Printf("Retry count exceeded (%d)", tries)
			break
		}
	}
	return err
}

func IDRequest(mw MessageWriter, dst Address) (version FirmwareVersion, devCat DevCat, err error) {
	msg, err := mw.Write(&Message{Dst: dst, Flags: StandardDirectMessage, Command: commands.IDRequest})
	if err == nil {
		msg, err = Read(mw, Or(CmdMatcher(commands.SetButtonPressedResponder), CmdMatcher(commands.SetButtonPressedController)))
		if err == nil {
			version = FirmwareVersion(msg.Dst[2])
			devCat = DevCat{msg.Dst[0], msg.Dst[1]}
		}
	}
	return
}

func GetEngineVersion(mw MessageWriter, dst Address) (version EngineVersion, err error) {
	ack, err := mw.Write(&Message{Dst: dst, Flags: StandardDirectMessage, Command: commands.GetEngineVersion})
	if err == nil {
		LogDebug.Printf("Device %v responded with an engine version %d", dst, ack.Command.Command2())
		version = EngineVersion(ack.Command.Command2())
	} else if err == ErrNak {
		// This only happens if the device is an I2Cs device and
		// is not linked to the queryier
		if ack.Command.Command2() == 0xff {
			LogDebug.Printf("Device %v is an unlinked I2Cs device", dst)
			version = VerI2Cs
			err = ErrNotLinked
		} else {
			err = ErrNak
		}
	}
	return
}

func Read(reader messageReader, matcher Matcher) (*Message, error) {
	msg, err := reader.Read()
	for ; err == nil; msg, err = reader.Read() {
		if matcher.Matches(msg) {
			break
		}
	}
	return msg, err
}
