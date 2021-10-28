// Copyright (c) 2021 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

//go:build go1.16
// +build go1.16

package template

import (
	"embed"
	"testing"
)

//go:embed testdata
var testFS embed.FS

func TestParseFS(t *testing.T) {
	tmpl := New("root")
	parsedTmpl := Must(tmpl.ParseFS(TrustedFSFromEmbed(testFS), "testdata/glob_*.tmpl"))
	if parsedTmpl != tmpl {
		t.Errorf("expected ParseFS to update template")
	}
}

func TestSub(t *testing.T) {
	tfs := TrustedFSFromEmbed(testFS)
	sub, err := tfs.Sub(TrustedSourceFromConstant("testdata"))
	if err != nil {
		t.Fatal(err)
	}
	tmpl := New("t1")
	parsedTmpl := Must(tmpl.ParseFS(sub, "dir1/parsefiles_t1.tmpl"))
	if parsedTmpl != tmpl {
		t.Errorf("expected ParseFS to update template")
	}
}
