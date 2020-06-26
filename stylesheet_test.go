// Copyright (c) 2017 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package safehtml

import (
	"fmt"
	"testing"
)

func TestCSSRule(t *testing.T) {
	for _, test := range [...]struct {
		selector  string
		style     Style
		want, err string
	}{
		{
			`#id`, StyleFromConstant(`top:0;left:0;`),
			`#id{top:0;left:0;}`, ``,
		},
		{
			`.class`, StyleFromConstant(`margin-left:5px;`),
			`.class{margin-left:5px;}`, ``,
		},
		{
			`tag #id, .class`, StyleFromConstant(`color:black !important;`),
			`tag #id, .class{color:black !important;}`, ``,
		},
		{
			`[title='son\'s']`, Style{},
			`[title='son\'s']{}`, ``,
		},
		{
			`[title="{"]`, Style{},
			`[title="{"]{}`, ``,
		},
		{
			`:nth-child(1)`, Style{},
			`:nth-child(1){}`, ``,
		},
		{
			`tag{color:black;}`, Style{},
			``, `selector "tag{color:black;}" contains "{", which is disallowed outside of CSS strings`,
		},
		{
			`]`, Style{},
			``, `selector "]" contains unbalanced () or [] brackets`,
		},
		{
			`[title`, Style{},
			``, `selector "[title" contains unbalanced () or [] brackets`,
		},
		{
			`[foo)bar]`, Style{},
			``, `selector "[foo)bar]" contains unbalanced () or [] brackets`,
		},
		{
			`[foo[bar]`, Style{},
			``, `selector "[foo[bar]" contains unbalanced () or [] brackets`,
		},
		{
			`foo(bar(baz)`, Style{},
			``, `selector "foo(bar(baz)" contains unbalanced () or [] brackets`,
		},
		{
			`:nth-child(1`, Style{},
			``, `selector ":nth-child(1" contains unbalanced () or [] brackets`,
		},
		{
			`[type="a]`, Style{},
			``, `selector "[type=\"a]" contains "\"", which is disallowed outside of CSS strings`,
		},
		{
			`[type=\'a]`, Style{},
			``, `selector "[type=\\'a]" contains "\\", which is disallowed outside of CSS strings`,
		},
		{
			`<`, Style{},
			``, `selector "<" contains '<'`,
		},
		{
			`@import "foo";#id`, Style{},
			``, `selector "@import \"foo\";#id" contains "@", which is disallowed outside of CSS strings`,
		},
		{
			`/* `, Style{},
			``, `selector "/* " contains "/", which is disallowed outside of CSS strings`,
		},
	} {
		errPrefix := fmt.Sprintf("CSSRule(%q, %#v)", test.selector, test.style)
		ss, err := CSSRule(test.selector, test.style)
		if test.want != "" && err != nil {
			t.Errorf("%s returned unexpected error: %s", errPrefix, err)
		} else if test.want != "" && ss.String() != test.want {
			t.Errorf("%s = %q, want: %q", errPrefix, ss.String(), test.want)
		} else if test.want == "" && err == nil {
			t.Errorf("%s expected error", errPrefix)
		} else if test.want == "" && err.Error() != test.err {
			t.Errorf("%s returned error:\n\t%s,\nwant:\n\t%q", errPrefix, err, test.want)
		}
	}
}
