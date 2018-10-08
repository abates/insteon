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

type TestDevice struct {
	messages []*Message
}

func (td *TestDevice) SendCommand(cmd Command, payload []byte) (response Command, err error) {
	return
}

func (td *TestDevice) Notify(msg *Message) error {
	td.messages = append(td.messages, msg)
	return nil
}

func (td *TestDevice) Address() Address { return Address{} }
