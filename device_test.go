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
	"testing"
	"time"
)

func TestDeviceRegistry(t *testing.T) {
	dr := &DeviceRegistry{}

	if _, found := dr.Find(Category(1)); found {
		t.Error("Expected nothing found for Category(1)")
	}

	dr.Register(Category(1), func(DeviceInfo, Device, time.Duration) (Device, error) {
		return nil, nil
	})

	if _, found := dr.Find(Category(1)); !found {
		t.Error("Expected to find Category(1)")
	}

	dr.Delete(Category(1))
	if _, found := dr.Find(Category(1)); found {
		t.Error("Expected nothing found for Category(1)")
	}
}

func mkPayload(buf ...byte) []byte {
	return append(buf, make([]byte, 14-len(buf))...)
}
