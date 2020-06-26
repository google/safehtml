// Copyright (c) 2017 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package template

import (
	"testing"
)

func TestMakeTrustedTemplate(t *testing.T) {
	const want = `foo`
	if got := MakeTrustedTemplate(want).String(); got != want {
		t.Errorf("got: %q, want: %q", got, want)
	}
}
