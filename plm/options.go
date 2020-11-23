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

package plm

import (
	"time"

	"github.com/abates/insteon"
	"github.com/abates/insteon/db"
)

// The Option mechanism is based on the method described at https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
type Option func(p *PLM) error

// WriteDelay can be passed as a parameter to New to change the delay used after writing a command before reading the response.
func WriteDelay(d time.Duration) Option {
	return func(p *PLM) error {
		p.writeDelay = d
		return nil
	}
}

// Database sets the insteoen database that the PLM will use for device category lookups
func Database(db db.Database) Option {
	return func(p *PLM) error {
		p.db = db
		return nil
	}
}

// ConnectionOptions specifies options to be set for each new device connection
func ConnectionOptions(options ...insteon.ConnectionOption) Option {
	return func(p *PLM) error {
		p.connOptions = options
		return nil
	}
}
