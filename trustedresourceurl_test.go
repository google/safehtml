// Copyright (c) 2017 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package safehtml

import (
	"testing"
)

func TestTrustedResourceURLWithParams(t *testing.T) {
	for _, test := range [...]struct {
		tru    TrustedResourceURL
		params map[string]string
		want   string
	}{
		{
			TrustedResourceURLFromConstant(`https://example.com/`),
			nil,
			`https://example.com/`,
		},
		{
			TrustedResourceURLFromConstant(`https://example.com/`),
			map[string]string{``: ``},
			`https://example.com/`,
		},
		{
			TrustedResourceURLFromConstant(`https://example.com/`),
			map[string]string{`b`: `1`, `c`: ``, ``: `d`},
			`https://example.com/?b=1`,
		},
		{
			TrustedResourceURLFromConstant(`https://example.com/`),
			map[string]string{`b`: `1`, `a`: `2`, `c`: `3`},
			`https://example.com/?a=2&b=1&c=3`,
		},
		{
			TrustedResourceURLFromConstant(`https://example.com/`),
			map[string]string{`a`: `&`},
			`https://example.com/?a=%26`,
		},
		{
			TrustedResourceURLFromConstant(`https://example.com/?a=x`),
			map[string]string{`b`: `y`},
			`https://example.com/?a=x&b=y`,
		},
		{
			TrustedResourceURLFromConstant(`https://example.com/?`),
			map[string]string{`b`: `y`},
			`https://example.com/?b=y`,
		},
		{
			TrustedResourceURLFromConstant(`https://example.com/??`),
			map[string]string{`b`: `y`},
			`https://example.com/??&b=y`,
		},
		{
			TrustedResourceURLFromConstant(`https://example.com/?a=x#foo`),
			map[string]string{`b`: `y`},
			`https://example.com/?a=x&b=y#foo`,
		},
	} {
		if got := TrustedResourceURLWithParams(test.tru, test.params).String(); got != test.want {
			t.Errorf("TrustedResourceURLWithParams(%#v, %v) = %q, want %q", test.tru, test.params, got, test.want)
		}
	}
}

type testFlagValue string

func (t *testFlagValue) String() string { return string(*t) }

func (t *testFlagValue) Get() interface{} { return *t }

func (t *testFlagValue) Set(s string) error {
	*t = testFlagValue(s)
	return nil
}

func TestTrustedResourceURLFromFlag(t *testing.T) {
	const want = `foo`
	value := testFlagValue(want)
	if got := TrustedResourceURLFromFlag(&value).String(); got != want {
		t.Errorf("got: %q, want: %q", got, want)
	}
}

func TestTrustedResourceURLFormat(t *testing.T) {
	for _, test := range [...]struct {
		desc, format string
		args         map[string]string
		want, err    string
	}{
		{
			"single arg with reserved characters",
			`/path/%{path}/`,
			map[string]string{"path": `d%/?#=`},
			`/path/d%25%2f%3f%23%3d/`,
			``,
		},
		{
			"multiple args",
			`/path/%{path1}/%{path2}?n1=v1`,
			map[string]string{"path1": `d%/?#=`, "path2": `2`},
			`/path/d%25%2f%3f%23%3d/2?n1=v1`,
			``,
		},
		{
			"extra arg ignored",
			`/path/%{path1}/%{path2}?n1=v1`,
			map[string]string{"path1": `d%/?#=`, "path2": `2`, "path3": "foo"},
			`/path/d%25%2f%3f%23%3d/2?n1=v1`,
			``,
		},
		{
			"missing arg 1",
			`/path/%{path1}/%{path2}?n1=v1`,
			map[string]string{"path2": `d%/?#=`},
			``,
			`expected argument named "path1"`,
		},
		{
			"missing arg 2",
			`/path/%{path1}/%{path2}?n1=v1`,
			map[string]string{"path1": `d%/?#=`},
			``,
			`expected argument named "path2"`,
		},
		{
			"nil args",
			`/path/%{path1}/%{path2}?n1=v1`,
			nil,
			``,
			`expected argument named "path1"`,
		},
		{
			"invalid arg name",
			`/path/%{path!name}/`,
			map[string]string{"path": `d%/?#=`},
			`/path/%{path!name}/`,
			``,
		},
		// Basic test cases for format string validation. Comprehensive test cases can
		// be found in TestIsSafeTrustedResourceURLPrefix in package safehtml/internal/safehtmlutil.
		{
			"path ambiguity",
			`/%{path}/`,
			map[string]string{"path": `/example.com/`},
			`/%2fexample.com%2f/`,
			``,
		},
		{
			"unsafe format string",
			`javascript:%{data}`,
			map[string]string{"data": `alert(1)`},
			``,
			`"javascript:%{data}" is a disallowed TrustedResourceURL format string`,
		},
		{
			"authority substitution",
			`https://%{authority}/%{path}`,
			map[string]string{"authority": `example.com`, "path": `foo`},
			``,
			`"https://%{authority}/%{path}" is a disallowed TrustedResourceURL format string`,
		},
		{
			"double dot segment disallowed",
			`/path/%{doubleDot}/%{path}?n1=v1`,
			map[string]string{"doubleDot": `..`, "path": "foo"},
			``,
			`argument "doubleDot" with value ".." must not contain ".."`,
		},
	} {
		got, err := trustedResourceURLFormat(test.format, test.args)
		if test.err != "" && err == nil {
			t.Errorf("%s: expected error, unexpectedly got output: %s", test.desc, got)
		} else if test.err == "" && err != nil {
			t.Errorf("%s: unexpected error: %s", test.desc, err)
		} else if test.err != "" && err.Error() != test.err {
			t.Errorf("%s: got error:\n\t%s\nwant:\n\t%s", test.desc, err, test.err)
		} else if test.err == "" && got.String() != test.want {
			t.Errorf("%s: got:\n\t%s\nwant:\n\t%s", test.desc, got, test.want)
		}
	}
}

func TestTrustedResourceURLAppend(t *testing.T) {
	for _, tc := range []struct {
		name, toAppend, wantURL string
		baseURL                 TrustedResourceURL
		err                     bool
	}{
		{
			name:    "empty append",
			baseURL: TrustedResourceURLFromConstant("//base.url/"),
			wantURL: "//base.url/",
		}, {
			name:     "alphanumerics appended",
			baseURL:  TrustedResourceURLFromConstant("//base.url/"),
			toAppend: "12-34-56-abc",
			wantURL:  "//base.url/12-34-56-abc",
		}, {
			name:     "path appended",
			baseURL:  TrustedResourceURLFromConstant("//base.url/"),
			toAppend: "sub/path/1",
			wantURL:  "//base.url/sub%2fpath%2f1",
		}, {
			name:     "path and query appended",
			baseURL:  TrustedResourceURLFromConstant("//base.url/"),
			toAppend: "sub/path/1?a=1&b=2",
			wantURL:  "//base.url/sub%2fpath%2f1%3fa%3d1%26b%3d2",
		}, {
			name:     "allowed chars appended",
			baseURL:  TrustedResourceURLFromConstant("//base.url/"),
			toAppend: "-_.*",
			wantURL:  "//base.url/-_.%2a",
		}, {
			name:     "star appended",
			baseURL:  TrustedResourceURLFromConstant("//base.url/"),
			toAppend: "*",
			wantURL:  "//base.url/%2a",
		}, {
			name:     "insecure prefix",
			baseURL:  TrustedResourceURLFromConstant("http://not.good"),
			toAppend: "foo",
			err:      true,
		}, {
			name:     "insecure prefix",
			baseURL:  TrustedResourceURLFromConstant("not.good"),
			toAppend: "foo",
			err:      true,
		},
	} {
		u, err := TrustedResourceURLAppend(tc.baseURL, tc.toAppend)
		if (err != nil) != tc.err {
			t.Fatalf("%s: err: %v, want err ? %t", tc.name, err, tc.err)
		}
		if err != nil {
			return
		}
		if got, want := u.String(), tc.wantURL; got != want {
			t.Errorf("%s: got: %q, want: %q", tc.name, got, want)
		}
	}
}
