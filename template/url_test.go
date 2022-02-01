// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package template

import (
	"strings"
	"testing"
)

func TestValidateURLPrefix(t *testing.T) {
	for _, test := range [...]struct {
		in    string
		valid bool
	}{
		// Allowed schemes or MIME types.
		{`http:`, true},
		{`http://www.foo.com/`, true},
		{`https://www.foo.com/`, true},
		{`mailto://foo@foo.com.com/`, true},
		{`ftp://foo.com/`, true},
		{`data:image/png;base64,abc`, true},
		{`data:video/mpeg;base64,abc`, true},
		{`data:audio/ogg;base64,abc`, true},
		// Leading and trailing newlines.
		{"\nhttp:", false},
		{"http:\n", false},
		// Disallowed schemes or MIME types.
		{`tel:+1-234-567-8901`, false},
		{`javascript:foo()`, false},
		{`data:image/png,abc`, false},
		{`data:text/html;base64,abc`, false},
		// No scheme, but not a scheme prefix.
		{`//www.foo.com/`, true},
		{`/path`, true},
		{`/path/x`, true},
		{`/path#x`, true},
		{`/path?x`, true},
		{`?q=`, true},
		// Scheme prefix.
		{`j`, false},
		{`java`, false},
		{`on`, false},
		{`data-`, false},
		// Unsafe scheme
		{`javascript:`, false},
		{`javascript:alert`, false},
		// Ends with incomplete HTML character reference.
		{`https&colon`, false},
		// Ends with valid HTML character reference, which forms a safe prefix
		// after HTML-unescaping.
		{`https&colon;`, true},
		{`?q&equals;`, true},
		// Ends with valid HTML character reference, but forms an
		// unsafe scheme after HTML-unescaping.
		{`javascript&#58`, false},
	} {
		err := validateURLPrefix(test.in)
		if err != nil && test.valid {
			t.Errorf("validateURLPrefix(%q) failed: %s", test.in, err)
		} else if err == nil && !test.valid {
			t.Errorf("validateURLPrefix(%q) succeeded unexpectedly", test.in)
		}
	}
}

func TestValidateTrustedResourceURLPrefix(t *testing.T) {
	for _, test := range [...]struct {
		in    string
		valid bool
	}{
		// Basic test cases for clearly safe and unsafe prefixes. Comprehensive test cases can
		// be found in TestIsSafeTrustedResourceURLPrefix in package safehtml/internal/safehtmlutil.
		{`https://www.foo.com/`, true},
		{`javascript:alert`, false},
		// Leading and trailing newlines.
		{"\n/path", false},
		{"/path\n", false},
		// Ends with incomplete HTML character reference.
		{`https://www.foo.com&sol`, false},
		// Ends with valid HTML character reference, which forms a safe prefix
		// after HTML-unescaping.
		{`https://www.foo.com&sol;`, true},
		{`//www.foo.com&sol;`, true},
		// Ends with valid HTML character reference, but forms an
		// unsafe scheme after HTML-unescaping.
		{`javascript&#58`, false},
	} {
		err := validateTrustedResourceURLPrefix(test.in)
		if err != nil && test.valid {
			t.Errorf("validateTrustedResourceURLPrefix(%q) failed: %s", test.in, err)
		} else if err == nil && !test.valid {
			t.Errorf("validateTrustedResourceURLPrefix(%q) succeeded unexpectedly", test.in)
		}
	}
}

func TestDecodeURLPrefix(t *testing.T) {
	const containsWhitespaceMsg = ` contains whitespace or control characters`
	const incompleteCharRefMsg = ` ends with an incomplete HTML character reference; did you mean "&amp;" instead of "&"?`
	const incopletePercentEncodingMsg = ` ends with an incomplete percent-encoding character triplet`
	for _, test := range [...]struct {
		in, want, err string
	}{
		// Contains whitespace or control characters.
		{" ", ``, containsWhitespaceMsg},
		{"\0000", ``, containsWhitespaceMsg},
		{" javascript", ``, containsWhitespaceMsg},
		{"javascript&#5\t8", ``, containsWhitespaceMsg},
		{"java\nscript&#58", ``, containsWhitespaceMsg},
		{"java script&#58;", ``, containsWhitespaceMsg},
		{"javascript&#5\n8", ``, containsWhitespaceMsg},
		{"javascript&#5\f8", ``, containsWhitespaceMsg},
		{"javascript&#5\r8", ``, containsWhitespaceMsg},
		{"javascript&#5 8", ``, containsWhitespaceMsg},
		{"javascript&#5&NewLine;8", ``, containsWhitespaceMsg},
		{"javascript&#5\00008", ``, containsWhitespaceMsg},
		// Incomplete HTML character escape sequences.
		{`https://www.foo.com?q=bar&`, ``, incompleteCharRefMsg},
		{`javascript&colon`, ``, incompleteCharRefMsg},
		// Complete HTML character references.
		{`javascript&colon;`, `javascript:`, ``},
		{`javascript&#58;`, `javascript:`, ``},
		// Invalid URL-encoding sequences.
		{`/fo%`, ``, incopletePercentEncodingMsg},
		{`/fo%6`, ``, incopletePercentEncodingMsg},
		{`/fo%6f`, `/fo%6f`, ``},
		{`/fo%6F`, `/fo%6F`, ``},
		// Only HTML-unescaping, not URL-unescaping, takes place.
		{`foo&#37;3a`, `foo%3a`, ``},
		{`foo&#37;3A`, `foo%3A`, ``},
	} {
		decoded, err := decodeURLPrefix(test.in)
		switch {
		case test.err != "":
			if err == nil {
				t.Errorf("url prefix %s : expected error", test.in)
			} else if got := err.Error(); !strings.Contains(got, test.err) {
				t.Errorf("url prefix %s : error\n\t%s\ndoes not contain expected string\n\t%s", test.in, got, test.err)
			}
		case test.want != "":
			if got := decoded; got != test.want {
				t.Errorf("url prefix %s : got decoded = %s, want %s", test.in, got, test.want)
			}
		}
	}
}
