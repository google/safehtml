// Copyright (c) 2017 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package template

import (
	"fmt"
	"os"
	"testing"
)

func TestTrustedSourceFromConstant(t *testing.T) {
	const want = `foo`
	if got := TrustedSourceFromConstant(want).String(); got != want {
		t.Errorf("got: %q, want: %q", got, want)
	}
}

func TestTrustedSourceFromConstantDir(t *testing.T) {
	// Use a short alias to make test cases more readable.
	c := TrustedSourceFromConstant
	for _, test := range [...]struct {
		dir                 string
		src                 TrustedSource
		filename, want, err string
	}{
		{"foo/", c(""), "file", "foo/file", ""},
		{"foo/", TrustedSource{}, "file", "foo/file", ""},
		{"", c("foo/"), "file", "foo/file", ""},
		{"foo", c("bar"), "file", "foo/bar/file", ""},
		{"foo/bar", c("baz"), "file", "foo/bar/baz/file", ""},
		{"foo/bar", c("baz"), "file.html", "foo/bar/baz/file.html", ""},
		{"foo", c("bar"), "dir:otherPath", "", `filename "dir:otherPath" must not contain the separator ':'`},
		{"foo", c("bar"), "dir/file.html", "", `filename "dir/file.html" must not contain the separator '/'`},
		{"foo", c("bar"), "../file.html", "", `filename "../file.html" must not contain the separator '/'`},
		{"foo/bar", c("baz"), "..", "", `filename must not be the special name ".."`},
	} {
		ts, err := TrustedSourceFromConstantDir(stringConstant(test.dir), test.src, test.filename)
		prefix := fmt.Sprintf("dir %q src %q filename %q", test.dir, test.src, test.filename)
		if test.err == "" && err != nil {
			t.Errorf("%s : unexpected error: %s", prefix, err)
		} else if test.err != "" && err == nil {
			t.Errorf("%s : expected error", prefix)
		} else if test.err != "" && err.Error() != test.err {
			t.Errorf("%s : got error:\n\t%s\nwant:\n\t%s", prefix, err, test.err)
		} else if ts.String() != test.want {
			t.Errorf("%s : got %q, want %q", prefix, ts.String(), test.want)
		}
	}
}

func TestTrustedSourceJoin(t *testing.T) {
	// Use a short alias to make test cases more readable.
	c := TrustedSourceFromConstant
	for _, test := range [...]struct {
		desc string
		in   []TrustedSource
		want string
	}{
		{"Path separators added if necessary",
			[]TrustedSource{c("foo"), c("bar/"), c("/baz"), c("/far")},
			"foo/bar/baz/far",
		},

		{".. path segments formed by concatenating multiple TrustedSource values do not affect the resultant path",
			[]TrustedSource{c("foo"), c("bar/."), c("./baz")},
			"foo/bar/baz",
		},
		{
			".. path segments that are explicitly specified in individual TrustedSource values will take effect",
			[]TrustedSource{c("foo"), c("bar"), c("baz/.."), c("../far")},
			"foo/far",
		},
	} {
		if got := TrustedSourceJoin(test.in...).String(); got != test.want {
			t.Errorf("%s : got: %q, want: %q", test.desc, got, test.want)
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

func TestTrustedSourceFromFlag(t *testing.T) {
	const want = `foo`
	value := testFlagValue(want)
	if got := TrustedSourceFromFlag(&value).String(); got != want {
		t.Errorf("got: %q, want: %q", got, want)
	}
}

func TestTrustedSourceFromEnvVar(t *testing.T) {
	const tmpDirEnvVar = `TMPDIR`
	const want = `/my/tmp`
	os.Setenv(tmpDirEnvVar, want)
	defer os.Unsetenv(tmpDirEnvVar)
	if got := TrustedSourceFromEnvVar(tmpDirEnvVar).String(); got != want {
		t.Errorf("got: %q, want: %q", got, want)
	}
}
