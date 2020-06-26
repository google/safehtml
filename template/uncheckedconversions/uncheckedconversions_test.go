// Copyright (c) 2017 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package uncheckedconversions

import (
	"testing"
)

func TestTrustedSourceFromStringKnownToSatisfyTypeContract(t *testing.T) {
	src := `some src`
	if out := TrustedSourceFromStringKnownToSatisfyTypeContract(src).String(); src != out {
		t.Errorf("uncheckedconversions.HTMLFromStringKnownToSatisfyTypeContract(%q).String() = %q, want %q",
			src, out, src)
	}
}

func TestTrustedTemplateFromStringKnownToSatisfyTypeContract(t *testing.T) {
	tmpl := `some tmpl`
	if out := TrustedTemplateFromStringKnownToSatisfyTypeContract(tmpl).String(); tmpl != out {
		t.Errorf("uncheckedconversions.HTMLFromStringKnownToSatisfyTypeContract(%q).String() = %q, want %q",
			tmpl, out, tmpl)
	}
}
