// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package template

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const tmplText = "foo"

func TestParseExecute(t *testing.T) {
	tmpl := New("test")
	parsedTmpl := Must(tmpl.Parse(tmplText))
	if parsedTmpl != tmpl {
		t.Errorf("expected Parse to update template")
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		t.Fatalf(err.Error())
	}
	if buf.String() != tmplText {
		t.Errorf("expected %s got %s", tmplText, buf.String())
	}
}

func TestMustParseAndExecuteToHTML(t *testing.T) {
	for _, test := range [...]struct {
		text stringConstant
		want string
	}{
		{
			`<b>hello world!</b>`,
			`<b>hello world!</b>`,
		},
		{
			`<b>all we need is <3</b>`,
			`<b>all we need is &lt;3</b>`,
		},
	} {
		html := MustParseAndExecuteToHTML(test.text)
		if got := html.String(); got != string(test.want) {
			t.Errorf("MustParseAndExecuteToHTML(%q) = %q, want %q", string(test.text), got, test.want)
		}
	}
}

func TestTemplateClone(t *testing.T) {
	// https://golang.org/issue/12996
	orig := New("name")
	clone, err := orig.Clone()
	if err != nil {
		t.Fatal(err)
	}
	if len(clone.Templates()) != len(orig.Templates()) {
		t.Fatalf("Invalid length of t.Clone().Templates()")
	}

	const want = "stuff"
	parsed := Must(clone.Parse(want))
	var buf bytes.Buffer
	err = parsed.Execute(&buf, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != want {
		t.Fatalf("got %q; want %q", got, want)
	}
}

const (
	tmpl1 = `{{define "a"}}foo{{end}}`
	tmpl2 = `{{define "b"}}bar{{end}}`
)

func TestLookup(t *testing.T) {
	tmpl := Must(New("test").Parse(tmpl1))
	Must(tmpl.Parse(tmpl2))
	a := tmpl.Lookup("a")
	if a == nil || a.Name() != "a" {
		t.Errorf("lookup on a failed")
	}
	b := tmpl.Lookup("b")
	if b == nil || b.Name() != "b" {
		t.Errorf("lookup on b failed")
	}
	if tmpl.Lookup("c") != nil {
		t.Errorf("lookup returned non-nil value for undefined template c")
	}
}

func TestTemplates(t *testing.T) {
	// want maps template name to expected output.
	want := map[string]string{
		"test": "",
		"a":    "foo",
		"b":    "bar",
	}
	tmpl := Must(New("test").Parse(tmpl1))
	Must(tmpl.Parse(tmpl2))
	templates := tmpl.Templates()
	if len(templates) != len(want) {
		t.Fatalf("want %d templates, got %d", len(want), len(templates))
	}
	for name := range want {
		found := false
		for _, tmpl := range templates {
			if name == tmpl.text.Name() {
				found = true
				break
			}
		}
		if !found {
			t.Error("could not find template", name)
		}
	}
	for _, got := range templates {
		name := got.Name()
		wantOutput, ok := want[name]
		if !ok {
			t.Errorf("got unexpected template name %q", name)
		}
		var buf bytes.Buffer
		if err := got.Execute(&buf, nil); err != nil {
			t.Fatalf("template %q: error executing: %v", name, err)
		}
		if buf.String() != wantOutput {
			t.Errorf("template %q: want output %s, got %s", name, wantOutput, buf.String())
		}
	}
}

func createTestDirAndFile(filename string) string {
	dir, err := ioutil.TempDir("", "template")
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Create(filepath.Join(dir, filename))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	_, err = io.WriteString(f, "Test template contents")
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

const filename = "T1.tmpl"

func TestParseFiles(t *testing.T) {
	dir := createTestDirAndFile(filename)
	tmpl := New("root")
	parsedTmpl := Must(tmpl.ParseFiles(stringConstant(filepath.Join(dir, filename))))
	if parsedTmpl != tmpl {
		t.Errorf("expected ParseFiles to update template")
	}
}

func TestParseGlob(t *testing.T) {
	dir := createTestDirAndFile(filename)
	tmpl := New("root")
	parsedTmpl := Must(tmpl.ParseGlob(stringConstant(filepath.Join(dir, "T*.tmpl"))))
	if parsedTmpl != tmpl {
		t.Errorf("expected ParseGlob to update template")
	}
}

func TestDontAllowJSTemplateSubstitution(t *testing.T) {
	const template = "<html><head></head><body><script>`{{.}}`</script></body></html>"
	templ := Must(New("foo").Parse(template))
	var b bytes.Buffer
	err := templ.Execute(&b, "foo")
	const want = "must be balanced"
	if err == nil || strings.Contains(err.Error(), "want") {
		t.Errorf("Parsed template %v, got error %v, expected %v", template, err, want)
	}
}
