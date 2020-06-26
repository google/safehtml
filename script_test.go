// Copyright (c) 2017 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package safehtml

import (
	"strings"
	"testing"
)

func TestScriptFromDataAndConstant(t *testing.T) {
	testStruct := struct {
		ID   int
		Name string
		Data []string
	}{
		ID:   3,
		Name: "Animals",
		Data: []string{"Cats", "Dogs", "Hamsters"},
	}
	for _, test := range [...]struct {
		desc      string
		name      stringConstant
		data      interface{}
		script    stringConstant
		want, err string
	}{
		{
			"string data with HTML special characters",
			`myVar`,
			`</script>`,
			`alert(myVar);`,
			`var myVar = "\u003c/script\u003e";
alert(myVar);`, "",
		},
		{
			"output of custom JSON marshaler escaped",
			`myVar`,
			dataWithUnsafeMarshaler(`"</script>"`),
			`alert(myVar);`,
			`var myVar = "\u003c/script\u003e";
alert(myVar);`, "",
		},
		{
			"invalid output of custom JSON marshaler rejected",
			`myVar`,
			dataWithUnsafeMarshaler(`"hello"; alert(1)`),
			`alert(myVar);`,
			"", "json: error calling MarshalJSON for type safehtml.dataWithUnsafeMarshaler",
		},
		{
			"struct data",
			`myVar`,
			testStruct,
			`alert(myVar);`,
			`var myVar = {"ID":3,"Name":"Animals","Data":["Cats","Dogs","Hamsters"]};
alert(myVar);`, "",
		},
		{
			"multi-line script",
			`myVar`,
			`<foo>`,
			`alert(myVar);
alert("hello world!");`,
			`var myVar = "\u003cfoo\u003e";
alert(myVar);
alert("hello world!");`, "",
		},
		{
			"empty variable name",
			"",
			`<foo>`,
			`alert(myVar);`,
			"", `variable name "" is an invalid Javascript identifier`,
		},
		{
			"invalid variable name",
			"café",
			`<foo>`,
			`alert(myVar);`,
			"", `variable name "café" is an invalid Javascript identifier`,
		},
		{
			"JSON encoding error",
			`myVar`,
			make(chan int),
			`alert(myVar);`,
			"", "json: unsupported type: chan int",
		},
	} {
		s, err := ScriptFromDataAndConstant(test.name, test.data, test.script)
		if test.err != "" && err == nil {
			t.Errorf("%s : expected error", test.desc)
		} else if test.err != "" && !strings.Contains(err.Error(), test.err) {
			t.Errorf("%s : got error:\n\t%s\nwant error:\n\t%s", test.desc, err, test.err)
		} else if test.err == "" && err != nil {
			t.Errorf("%s : unexpected error: %s", test.desc, err)
		} else if got := s.String(); got != test.want {
			t.Errorf("%s : got:\n%s\nwant:\n%s", test.desc, got, test.want)
		}
	}
}

type dataWithUnsafeMarshaler string

func (d dataWithUnsafeMarshaler) MarshalJSON() ([]byte, error) {
	return []byte(string(d)), nil
}

func TestJSIdentifierPattern(t *testing.T) {
	for _, test := range [...]struct {
		in   string
		want bool
	}{
		{`foo`, true},
		{`Foo`, true},
		{`f0o`, true},
		{`_f0o`, true},
		{`$f0o`, true},
		{`f0$_o`, true},
		{`_f0$_o`, true},
		// Starts with digit.
		{`2foo`, false},
		// Contains alphabetic codepoints that are not ASCII letters.
		{`café`, false},
		{`Χαίρετε`, false},
		// Contains non-alphabetic codepoints.
		{`你好`, false},
		// Contains unicode escape sequences.
		{`\u0192oo`, false},
		{`f\u006Fo`, false},
		// Contains zero-width non-joiner.
		{"dea\u200Cly", false},
		// Contains zero-width joiner.
		{"क्\u200D", false},
	} {
		if got := jsIdentifierPattern.MatchString(test.in); got != test.want {
			t.Errorf("jsIdentifierPattern.MatchString(%q) = %t", test.in, got)
		}
	}
}
