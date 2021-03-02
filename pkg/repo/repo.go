// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package repo

import (
	"net/url"
	"strings"
)

// ParseURL returns the organization and project for a URL or partial path
func ParseURL(rawURL string) (string, string) {
	u, err := url.Parse(rawURL)
	if err == nil {
		p := strings.Split(u.Path, "/")
		if u.Hostname() != "" {
			return p[1], p[2]
		}
		return p[0], p[1]
	}
	// Not a URL
	p := strings.Split(rawURL, "/")
	return p[0], p[1]
}
