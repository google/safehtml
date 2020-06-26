// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package template

import (
	"bytes"
	"fmt"
	"testing"
)

// Test that we print using the String method. Was issue 3073.
type stringer struct {
	v int
}

func (s *stringer) String() string {
	return fmt.Sprintf("string=%d", s.v)
}

type errorer struct {
	v int
}

func (s *errorer) Error() string {
	return fmt.Sprintf("error=%d", s.v)
}

func TestStringer(t *testing.T) {
	s := &stringer{3}
	b := new(bytes.Buffer)
	tmpl := Must(New("x").Parse("{{.}}"))
	if err := tmpl.Execute(b, s); err != nil {
		t.Fatal(err)
	}
	var expect = "string=3"
	if b.String() != expect {
		t.Errorf("expected %q got %q", expect, b.String())
	}
	e := &errorer{7}
	b.Reset()
	if err := tmpl.Execute(b, e); err != nil {
		t.Fatal(err)
	}
	expect = "error=7"
	if b.String() != expect {
		t.Errorf("expected %q got %q", expect, b.String())
	}
}

// https://golang.org/issue/5982
func TestEscapingNilNonemptyInterfaces(t *testing.T) {
	tmpl := Must(New("x").Parse("{{.E}}"))

	got := new(bytes.Buffer)
	testData := struct{ E error }{} // any non-empty interface here will do; error is just ready at hand
	tmpl.Execute(got, testData)

	// Use this data instead of just hard-coding "&lt;nil&gt;" to avoid
	// dependencies on the html escaper and the behavior of fmt w.r.t. nil.
	want := new(bytes.Buffer)
	data := struct{ E string }{E: fmt.Sprint(nil)}
	tmpl.Execute(want, data)

	if !bytes.Equal(want.Bytes(), got.Bytes()) {
		t.Errorf("expected %q got %q", string(want.Bytes()), string(got.Bytes()))
	}
}
