// Copyright (c) 2017 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package safehtmlutil

import (
	"testing"
)

func TestIsSafeTrustedResourceURLPrefix(t *testing.T) {
	for _, test := range [...]struct {
		in   string
		want bool
	}{
		// With scheme.
		{`httpS://www.foO.com/`, true},
		// Scheme-relative.
		{`//www.foo.com/`, true},
		// Origin with hypen and port.
		{`//ww-w.foo.com:1000/path`, true},
		// IPv6 origin.
		{`//[::1]/path`, true},
		// Path-absolute.
		{`/path`, true},
		{`/path/x`, true},
		{`/path#x`, true},
		{`/path?x`, true},
		// Mixed case.
		{`httpS://www.foo.cOm/pAth`, true},
		{`about:blank#`, true},
		{`about:blank#x`, true},
		// Scheme prefix.
		{`j`, false},
		{`java`, false},
		{`on`, false},
		{`data-`, false},
		// Unsafe scheme
		{`javascript:`, false},
		{`javascript:alert`, false},
		// Invalid scheme.
		{`ftp://`, false},
		// Missing origin.
		{`https://`, false},
		{`https:///`, false}, // NOTYPO
		{`//`, false},
		{`///`, false},
		// Missing / after origin.
		{`https://foo.com`, false},
		// Invalid char in origin.
		{`https://www.foo%.com/`, false},
		{`https://www.foo\\.com/`, false},
		{`https://user:password@www.foo.com/`, false},
		// Two slashes, would allow origin to be set dynamically.
		{`//`, false},
		// Two slashes. IE allowed (allows?) '\' instead of '/'.
		{`/\\`, false},
		// Relative path.
		{`abc`, false},
		{`about:blank`, false},
		{`about:blankX`, false},
	} {
		if got := IsSafeTrustedResourceURLPrefix(test.in); got != test.want {
			t.Errorf("IsSafeTrustedResourceURLPrefix(%q) = %t", test.in, got)
		}
	}
}

func TestURLContainsDoubleDotSegment(t *testing.T) {
	for _, test := range [...]struct {
		in   string
		want bool
	}{
		// Permutations of double dot-segment URL substrings.
		{`..`, true},
		{`%2e%2e`, true},
		{`%2E%2e`, true},
		{`%2e%2E`, true},
		{`%2E%2E`, true},
		{`.%2e`, true},
		{`.%2E`, true},
		{`%2e.`, true},
		{`%2E.`, true},
		// Permutations of single dot-segments URL substrings.
		{`.`, false},
		{`%2e`, false},
		{`%2E`, false},
		// These URL substrings do not technically denote dot-segments, but are very
		// unlikely to be part of a legitimate URL.
		{`foo..`, true},
		{`..foo`, true},
		// Non-contiguous dots
		{`.foo.`, false},
		// Full URLs with a sampling of the double and single dot-segment substrings
		// from the previous test cases.
		{`http://www.test.com/../bar`, true},
		{`http://www.test.com/foo../bar`, true},
		{`http://www.test.com/bar/%2E%2e`, true},
		{`http://www.test.com/./bar`, false},
		{`http://www.test.com/.foo./bar`, false},
		{`http://www.test.com/bar/%2E`, false},
	} {
		if got := URLContainsDoubleDotSegment(test.in); got != test.want {
			t.Errorf("URLContainsDoubleDotSegment(%q) = %t", test.in, got)
		}
	}
}

func TestNormalizeURL(t *testing.T) {
	for _, test := range [...]struct {
		url, want string
	}{
		{"", ""},
		{
			"http://example.com:80/foo/bar?q=foo%20&bar=x+y#frag",
			"http://example.com:80/foo/bar?q=foo%20&bar=x+y#frag",
		},
		{" ", "%20"},
		{"%7c", "%7c"},
		{"%7C", "%7C"},
		{"%2", "%252"},
		{"%", "%25"},
		{"%z", "%25z"},
		{"/foo|bar/%5c\u1234", "/foo%7cbar/%5c%e1%88%b4"},
	} {
		if got := NormalizeURL(test.url); test.want != got {
			t.Errorf("%q: got\n\t%q\n, want\n\t%q", test.url, got, test.want)
		}
		if test.want != NormalizeURL(test.want) {
			t.Errorf("not idempotent: %q", test.want)
		}
	}
}

func TestQueryEscapeURL(t *testing.T) {
	const input = "\x00\x01\x02\x03\x04\x05\x06\x07\x08\t\n\x0b\x0c\r\x0e\x0f" +
		"\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f" +
		` !"#$%&'()*+,-./` +
		`0123456789:;<=>?` +
		`@ABCDEFGHIJKLMNO` +
		`PQRSTUVWXYZ[\]^_` +
		"`abcdefghijklmno" +
		"pqrstuvwxyz{|}~\x7f" +
		"\u00A0\u0100\u2028\u2029\ufeff\U0001D11E"
	const want = "%00%01%02%03%04%05%06%07%08%09%0a%0b%0c%0d%0e%0f" +
		"%10%11%12%13%14%15%16%17%18%19%1a%1b%1c%1d%1e%1f" +
		"%20%21%22%23%24%25%26%27%28%29%2a%2b%2c-.%2f" +
		"0123456789%3a%3b%3c%3d%3e%3f" +
		"%40ABCDEFGHIJKLMNO" +
		"PQRSTUVWXYZ%5b%5c%5d%5e_" +
		"%60abcdefghijklmno" +
		"pqrstuvwxyz%7b%7c%7d~%7f" +
		"%c2%a0%c4%80%e2%80%a8%e2%80%a9%ef%bb%bf%f0%9d%84%9e"
	if got := QueryEscapeURL(input); want != got {
		t.Fatalf("got\n\t%q\nwant\n\t%q", got, want)
	}
}

func BenchmarkQueryEscapeURL(b *testing.B) {
	for i := 0; i < b.N; i++ {
		QueryEscapeURL("http://example.com:80/foo?q=bar%20&baz=x+y#frag")
	}
}

func BenchmarkQueryEscapeURLNoSpecials(b *testing.B) {
	for i := 0; i < b.N; i++ {
		QueryEscapeURL("TheQuickBrownFoxJumpsOverTheLazyDog.")
	}
}

func BenchmarkNormalizeURL(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NormalizeURL("The quick brown fox jumps over the lazy dog.\n")
	}
}

func BenchmarkNormalizeURLNoSpecials(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NormalizeURL("http://example.com:80/foo?q=bar%20&baz=x+y#frag")
	}
}
