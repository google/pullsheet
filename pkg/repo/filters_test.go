// Copyright 2026 Google LLC
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

import "testing"

func TestSetIgnorePathReAndTruncRe(t *testing.T) {
	t.Cleanup(func() {
		if err := SetIgnorePathRe(DefaultIgnorePathRe); err != nil {
			t.Fatalf("reset ignore-re: %v", err)
		}
		if err := SetTruncRe(DefaultTruncRe); err != nil {
			t.Fatalf("reset trunc-re: %v", err)
		}
	})

	if err := SetIgnorePathRe(`(?i)vendor/`); err != nil {
		t.Fatalf("SetIgnorePathRe: %v", err)
	}
	if !ignorePathRe.MatchString("Vendor/pkg/foo.go") {
		t.Fatalf("expected ignorePathRe to match Vendor/pkg/foo.go")
	}

	if err := SetTruncRe(`(?i)changelog`); err != nil {
		t.Fatalf("SetTruncRe: %v", err)
	}
	if !truncRe.MatchString("docs/CHANGELOG.md") {
		t.Fatalf("expected truncRe to match docs/CHANGELOG.md")
	}

	if err := SetIgnorePathRe(`(`); err == nil {
		t.Fatalf("expected invalid ignore-re to fail")
	}
	if err := SetTruncRe(`(`); err == nil {
		t.Fatalf("expected invalid trunc-re to fail")
	}
}
