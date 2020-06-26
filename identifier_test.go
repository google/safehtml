// Copyright (c) 2017 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package safehtml

import (
	"fmt"
	"strings"
	"testing"
)

func TestIdentifierFromConstant(t *testing.T) {
	tryIdentifierFromConstant := func(value string) (id Identifier, panicMsg string) {
		defer func() {
			r := recover()
			if r == nil {
				panicMsg = ""
				return
			}
			panicMsg = fmt.Sprint(r)
		}()
		return IdentifierFromConstant(stringConstant(value)), ""
	}

	for _, test := range [...]struct {
		value, panicMsg string
	}{
		{"foo", ""},
		{"F0ob4r", ""},
		{"foo-bar", ""},
		{"foo--bar", ""},
		{"foo-bar-baz", ""},
		{"foo-bar_baz", ""},
		{"foo!", "invalid identifier"},
		{"foo ", "invalid identifier"},
		{"fo o", "invalid identifier"},
		{" foo", "invalid identifier"},
		{"foo\t", "invalid identifier"},
		{"4wesome", "invalid identifier"},
	} {
		id, panicMsg := tryIdentifierFromConstant(test.value)
		if test.panicMsg != "" {
			if !strings.Contains(panicMsg, test.panicMsg) {
				t.Errorf("value %q: got panic message:\n\t%q\nwant:\n\t%q", test.value, panicMsg, test.panicMsg)
			}
			continue
		}
		if panicMsg != "" {
			t.Errorf("value %q: unexpected panic: %q", test.value, panicMsg)
			continue
		}
		if got := id.String(); got != test.value {
			t.Errorf("value %q: got id: %q\twant: %q", test.value, got, test.value)
		}
	}
}

func TestIdentifierFromConstantPrefix(t *testing.T) {
	tryIdentifierFromConstantPrefix := func(prefix, value string) (id Identifier, panicMsg string) {
		defer func() {
			r := recover()
			if r == nil {
				panicMsg = ""
				return
			}
			panicMsg = fmt.Sprint(r)
		}()
		return IdentifierFromConstantPrefix(stringConstant(prefix), value), ""
	}

	for _, test := range [...]struct {
		prefix, value, want, panicMsg string
	}{
		{"foo", "bar", "foo-bar", ""},
		{"foo", "-bar", "foo--bar", ""},
		{"foo", "bar-baz", "foo-bar-baz", ""},
		{"foo", "bar-baz-", "foo-bar-baz-", ""},
		{"foo", "bar_baz-", "foo-bar_baz-", ""},
		{"foo", "", "foo-", ""},
		{"", "bar", "", "invalid prefix"},
		{"foo!", "bar", "", "invalid prefix"},
		{"4wesome", "bar", "", "invalid prefix"},
		{" foo", "bar", "", "invalid prefix"},
		{"fo o", "bar", "", "invalid prefix"},
		{"foo ", "bar", "", "invalid prefix"},
		{"foo\t", "bar", "", "invalid prefix"},
		{"foo\n", "bar", "", "invalid prefix"},
		{"\nfoo", "bar", "", "invalid prefix"},
		{"foo", "bar!", "", "contains non-alphanumeric runes"},
		{"foo", " bar", "", "contains non-alphanumeric runes"},
		{"foo", "b ar", "", "contains non-alphanumeric runes"},
		{"foo", "bar ", "", "contains non-alphanumeric runes"},
		{"foo", "bar\t", "", "contains non-alphanumeric runes"},
	} {
		id, panicMsg := tryIdentifierFromConstantPrefix(test.prefix, test.value)
		inputs := fmt.Sprintf("prefix %q, value %q", test.prefix, test.value)
		if test.panicMsg != "" {
			if !strings.Contains(panicMsg, test.panicMsg) {
				t.Errorf("%s: got panic message:\n\t%q\nwant:\n\t%q", inputs, panicMsg, test.panicMsg)
			}
			continue
		}
		if panicMsg != "" {
			t.Errorf("%s: unexpected panic: %q", inputs, panicMsg)
			continue
		}
		if got := id.String(); got != test.want {
			t.Errorf("%s: got id: %q\twant: %q", inputs, got, test.want)
		}
	}
}
