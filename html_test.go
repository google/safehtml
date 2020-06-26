// Copyright (c) 2017 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package safehtml

import (
	"testing"
)

const (
	rawHTML     = `<>'"&`
	escapedHTML = "&lt;&gt;&#39;&#34;&amp;"
)

func TestHTMLEscaped(t *testing.T) {
	if got := HTMLEscaped(rawHTML); got.String() != escapedHTML {
		t.Errorf("HTMLEscaped(%q) == %q, want %q", rawHTML, got.String(), escapedHTML)
	}
}

func TestHTMLConcat(t *testing.T) {
	for _, test := range [...]struct {
		in   []string
		want string
	}{
		{
			nil,
			"",
		},
		{
			[]string{""},
			"",
		},
		{
			[]string{"Hello world!"},
			"Hello world!",
		},
		{
			[]string{"Hello", " ", "world!"},
			"Hello world!",
		},
	} {
		var htmls []HTML
		for _, str := range test.in {
			htmls = append(htmls, HTML{str})
		}
		if got := HTMLConcat(htmls...); got.String() != test.want {
			t.Errorf("HTMLConcat with args %q returns %q, want %q", test.in, got.String(), test.want)
		}
	}
}

func TestCoerceToInterchangeValid(t *testing.T) {
	// Single character tests
	for _, tt := range [...]struct {
		in       string
		replaced bool
	}{
		// Control characters.
		{"\x00", true},
		{"\x04", true},
		{"\x08", true},
		{"\t", false}, // x09
		{"\n", false}, // x0A
		{"\v", true},  // x0B
		{"\f", false}, // x0C
		{"\r", false}, // x0D
		{"\x0E", true},
		{"\x0F", true},
		// Non-character codepoints. See
		// http://www.w3.org/TR/html5/syntax.html#preprocessing-the-input-stream.
		{"\uFDCF", false}, // Begin border of \uFDD0 to \uFDEF range.
		{"\uFDD0", true},  // Begin range.
		{"\uFDD0", true},  // Mid range.
		{"\uFDEF", true},  // End range.
		{"\uFDF0", false}, // End border.
		{"\uFFFE", true},
		{"\uFFFF", true},
		{"\U0001FFFE", true},
		{"\U0001FFFF", true},
		{"\U0002FFFE", true},
		{"\U0002FFFF", true},
		{"\U0003FFFE", true},
		{"\U0003FFFF", true},
		{"\U0004FFFE", true},
		{"\U0004FFFF", true},
		{"\U0005FFFE", true},
		{"\U0005FFFF", true},
		{"\U0006FFFE", true},
		{"\U0006FFFF", true},
		{"\U0007FFFE", true},
		{"\U0007FFFF", true},
		{"\U0008FFFE", true},
		{"\U0008FFFF", true},
		{"\U0009FFFE", true},
		{"\U0009FFFF", true},
		{"\U000AFFFE", true},
		{"\U000AFFFF", true},
		{"\U000BFFFE", true},
		{"\U000BFFFF", true},
		{"\U000CFFFE", true},
		{"\U000CFFFF", true},
		{"\U000DFFFE", true},
		{"\U000DFFFF", true},
		{"\U000EFFFE", true},
		{"\U000EFFFF", true},
		{"\U000FFFFE", true},
		{"\U000FFFFF", true},
		{"\U0010FFFE", true},
		{"\U0010FFFF", true},
		// Invalid UTF8.
		{"\xed", true},
		// Valid UTF8.
		{" ", false},
		// Replacement character.
		{"\uFFFD", false},
	} {
		coerced := coerceToUTF8InterchangeValid(tt.in)
		if tt.replaced && coerced != "\uFFFD" {
			t.Errorf(`coerceToInterchangeValid(%q) == %q, want "0xFFFD"`, tt.in, coerced)
		} else if !tt.replaced && coerced != tt.in {
			t.Errorf("coerceToInterchangeValid(%q) == %q, want %q", tt.in, coerced, tt.in)
		}
	}

	// String tests
	for _, tt := range [...]struct {
		in       string
		expected string
	}{
		{"abcd", "abcd"},
		{"丄ê𐒖t", "丄ê𐒖t"},
		// String with all classes of codepoints above.
		{"\n丄\xed \x00\U0001FFFEa\uFFFD", "\n丄\uFFFD \uFFFD\uFFFDa\uFFFD"},
		// Invalid byte sequence.
		{"\xff\x7e", "\uFFFD\x7e"},
		// Single surrogate.
		// Go range clause advances one byte at a time on invalid UTF-8,
		// including codepoints that represent surrogates.
		{"\xED\xA0\x80", "\uFFFD\uFFFD\uFFFD"},
		// Supplementary point U+233B4 encoded as a surrogate pair.
		{"\xED\xA1\x8C\xED\xBE\xB4", "\uFFFD\uFFFD\uFFFD\uFFFD\uFFFD\uFFFD"},
		// Overlong sequence.
		{"\xC0\x80", "\uFFFD\uFFFD"},
	} {
		if coerced := coerceToUTF8InterchangeValid(tt.in); tt.expected != coerced {
			t.Errorf(`coerceToInterchangeValid(%q) == %q, want %q`, tt.in, coerced, tt.expected)
		}
	}
}
