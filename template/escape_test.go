// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package template

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
	"text/template"
	"text/template/parse"
)

// TODO: consider merging this file with sanitize_test.go or other test files.

type badMarshaler struct{}

func (x *badMarshaler) MarshalJSON() ([]byte, error) {
	// Keys in valid JSON must be double quoted as must all strings.
	return []byte("{ foo: 'not quite valid JSON' }"), nil
}

type goodMarshaler struct{}

func (x *goodMarshaler) MarshalJSON() ([]byte, error) {
	return []byte(`{ "<foo>": "O'Reilly" }`), nil
}

func TestEscapeMap(t *testing.T) {
	data := map[string]string{
		"html":     `<h1>Hi!</h1>`,
		"urlquery": `http://www.foo.com/index.html?title=main`,
	}
	for _, test := range [...]struct {
		desc   string
		input  stringConstant
		output string
	}{
		// covering issue 20323
		{
			"field with predefined escaper name 1",
			`{{.html | print}}`,
			`&lt;h1&gt;Hi!&lt;/h1&gt;`,
		},
		// covering issue 20323
		{
			"field with predefined escaper name 2",
			`{{.urlquery | print}}`,
			`http://www.foo.com/index.html?title=main`,
		},
	} {
		tmpl := Must(New("").Parse(test.input))
		b := new(bytes.Buffer)
		if err := tmpl.Execute(b, data); err != nil {
			t.Errorf("%s: template execution failed: %s", test.desc, err)
			continue
		}
		if w, g := test.output, b.String(); w != g {
			t.Errorf("%s: escaped output: want\n\t%q\ngot\n\t%q", test.desc, w, g)
			continue
		}
	}
}

func TestEscapeSet(t *testing.T) {
	type dataItem struct {
		Children []*dataItem
		X        string
	}

	data := dataItem{
		Children: []*dataItem{
			{X: "foo"},
			{X: "<bar>"},
			{
				Children: []*dataItem{
					{X: "baz"},
				},
			},
		},
	}

	tests := []struct {
		inputs map[string]string
		want   string
	}{
		// The trivial set.
		{
			map[string]string{
				"main": ``,
			},
			``,
		},
		// A template called in the start context.
		{
			map[string]string{
				"main":   `Hello, {{template "helper"}}!`,
				"helper": `{{"<World>"}}`,
			},
			`Hello, &lt;World&gt;!`,
		},
		// A template called in a context other than the start.
		{
			map[string]string{
				"main": `<a href="/foo?q={{template "helper"}}">Link</a>`,
				// Not a valid top level HTML template.
				// "<b" is not a full tag.
				"helper": `{{"bar&x=baz"}}<b`,
			},
			`<a href="/foo?q=bar%26x%3dbaz<b">Link</a>`,
		},
		// A recursive template that ends in its start context.
		{
			map[string]string{
				"main": `{{range .Children}}{{template "main" .}}{{else}}{{.X}} {{end}}`,
			},
			`foo &lt;bar&gt; baz `,
		},
		// A recursive helper template that ends in its start context.
		{
			map[string]string{
				"main":   `{{template "helper" .}}`,
				"helper": `{{if .Children}}<ul>{{range .Children}}<li>{{template "main" .}}</li>{{end}}</ul>{{else}}{{.X}}{{end}}`,
			},
			`<ul><li>foo</li><li>&lt;bar&gt;</li><li><ul><li>baz</li></ul></li></ul>`,
		},
		// Co-recursive templates that end in its start context.
		{
			map[string]string{
				"main":   `<blockquote>{{range .Children}}{{template "helper" .}}{{end}}</blockquote>`,
				"helper": `{{if .Children}}{{template "main" .}}{{else}}{{.X}}<br>{{end}}`,
			},
			`<blockquote>foo<br>&lt;bar&gt;<br><blockquote>baz<br></blockquote></blockquote>`,
		},
		// A template that is called in two different contexts.
		{
			map[string]string{
				"main":   `<a href="/foo?q={{template "helper"}}">{{template "helper"}}</a>`,
				"helper": `{{"bar&x=baz"}}`,
			},
			`<a href="/foo?q=bar%26x%3dbaz">bar&amp;x=baz</a>`,
		},
		// A non-recursive template that ends in a different context.
		// helper starts in stateTag and ends in stateAttr.
		{
			map[string]string{
				"main":   `<a {{template "helper"}}">Link</a>`,
				"helper": `href="{{"https://www.foo.com"}}`,
			},
			`<a href="https://www.foo.com">Link</a>`,
		},
		// A recursive template that ends in a similar context.
		{
			map[string]string{
				"main":        `<a href="/foo?{{template "queryParams" 4}}">Link</a>`,
				"queryParams": `key{{.}}=val{{.}}{{if .}}&{{template "queryParams" . | pred}}{{end}}`,
			},
			`<a href="/foo?key4=val4&key3=val3&key2=val2&key1=val1&key0=val0">Link</a>`,
		},
	}

	// pred is a template function that returns the predecessor of a
	// natural number for testing recursive templates.
	fns := FuncMap{"pred": func(a ...interface{}) (interface{}, error) {
		if len(a) == 1 {
			if i, _ := a[0].(int); i > 0 {
				return i - 1, nil
			}
		}
		return nil, fmt.Errorf("undefined pred(%v)", a)
	}}

	for _, test := range tests {
		source := ""
		for name, body := range test.inputs {
			source += fmt.Sprintf("{{define %q}}%s{{end}} ", name, body)
		}
		tmpl, err := New("root").Funcs(fns).Parse(stringConstant(source))
		if err != nil {
			t.Errorf("error parsing %q: %v", source, err)
			continue
		}
		var b bytes.Buffer

		if err := tmpl.ExecuteTemplate(&b, "main", data); err != nil {
			t.Errorf("%q executing %v", err.Error(), tmpl.Lookup("main"))
			continue
		}
		if got := b.String(); test.want != got {
			t.Errorf("want\n\t%q\ngot\n\t%q", test.want, got)
		}
	}
}

func TestNestedRuntimeError(t *testing.T) {
	buf := new(bytes.Buffer)
	tmpl := New("")
	Must(tmpl.Parse(`<style>{{template "inner" .}}</style>`))
	Must(tmpl.Parse(`{{define "inner"}}{{"foo"}}{{end}}`))
	err := tmpl.Execute(buf, nil)
	if err == nil {
		t.Fatalf("expected error")
	}
	if got, want := err.Error(), `expected a safehtml.StyleSheet value`; !strings.Contains(got, want) {
		t.Fatalf("got: %s, does not contain: %s", got, want)
	}
}

func TestEscapeText(t *testing.T) {
	tests := []struct {
		input  string
		output context
	}{
		{
			``,
			context{},
		},
		{
			`Hello, World!`,
			context{},
		},
		{
			// An orphaned "<" is OK.
			`I <3 Ponies!`,
			context{},
		},
		{
			`<a`,
			context{state: stateTag, element: element{name: "a"}},
		},
		{
			`<a `,
			context{state: stateTag, element: element{name: "a"}},
		},
		{
			`<a>`,
			context{state: stateText, element: element{name: "a"}},
		},
		{
			`<a href`,
			context{state: stateAttrName, element: element{name: "a"}, attr: attr{name: "href"}},
		},
		{
			`<a on`,
			context{state: stateAttrName, element: element{name: "a"}, attr: attr{name: "on"}},
		},
		{
			`<a href `,
			context{state: stateAfterName, element: element{name: "a"}, attr: attr{name: "href"}},
		},
		{
			`<a style  =  `,
			context{state: stateBeforeValue, element: element{name: "a"}, attr: attr{name: "style"}},
		},
		{
			`<a href=`,
			context{state: stateBeforeValue, element: element{name: "a"}, attr: attr{name: "href"}},
		},
		{
			`<a href=x `,
			context{state: stateTag, element: element{name: "a"}},
		},
		{
			`<a href=>`,
			context{state: stateText, element: element{name: "a"}},
		},
		{
			`<a href=x>`,
			context{state: stateText, element: element{name: "a"}},
		},
		{
			`<a href=''`,
			context{state: stateTag, element: element{name: "a"}},
		},
		{
			`<a href=""`,
			context{state: stateTag, element: element{name: "a"}},
		},
		{
			`<a title="`,
			context{state: stateAttr, delim: delimDoubleQuote, element: element{name: "a"}, attr: attr{name: "title"}},
		},
		{
			`<img alt="1">`,
			context{state: stateText},
		},
		{
			`<img alt="1>"`,
			context{state: stateTag, element: element{name: "img"}},
		},
		{
			`<img alt="1>">`,
			context{state: stateText},
		},
		{
			`<input checked type="checkbox"`,
			context{state: stateTag, element: element{name: "input"}},
		},
		{
			`<!-- foo`,
			context{state: stateHTMLCmt},
		},
		{
			`<!-->`,
			context{state: stateHTMLCmt},
		},
		{
			`<!--->`,
			context{state: stateHTMLCmt},
		},
		{
			`<!-- foo -->`,
			context{state: stateText},
		},
		{
			`<script`,
			context{state: stateTag, element: element{name: "script"}},
		},
		{
			`<script `,
			context{state: stateTag, element: element{name: "script"}},
		},
		{
			`<script src="foo.js" `,
			context{state: stateTag, element: element{name: "script"}},
		},
		{
			`<script src='foo.js' `,
			context{state: stateTag, element: element{name: "script"}},
		},
		{
			`<script type=text/javascript `,
			context{state: stateTag, element: element{name: "script"}, scriptType: "text/javascript"},
		},
		{
			`<script>`,
			context{state: stateSpecialElementBody, element: element{name: "script"}},
		},
		{
			`<script>foo`,
			context{state: stateSpecialElementBody, element: element{name: "script"}},
		},
		{
			`<script>foo</script>`,
			context{state: stateText},
		},
		{
			`<script>foo</script><!--`,
			context{state: stateHTMLCmt},
		},
		{
			`<script>document.write("<p>foo</p>");`,
			context{state: stateSpecialElementBody, element: element{name: "script"}},
		},
		{
			`<script>document.write("<p>foo<\/script>");`,
			context{state: stateSpecialElementBody, element: element{name: "script"}},
		},
		{
			`<script>document.write("<script>alert(1)</script>");`,
			context{state: stateText},
		},
		{
			`<script type="text/template">`,
			context{state: stateSpecialElementBody, element: element{name: "script"}, scriptType: "text/template"},
		},
		// covering issue 19968
		{
			`<script type="TEXT/JAVASCRIPT">`,
			context{state: stateSpecialElementBody, element: element{name: "script"}, scriptType: "text/javascript"},
		},
		// covering issue 19965
		{
			`<script TYPE="text/template">`,
			context{state: stateSpecialElementBody, element: element{name: "script"}, scriptType: "text/template"},
		},
		{
			`<script type="notjs">`,
			context{state: stateSpecialElementBody, element: element{name: "script"}, scriptType: "notjs"},
		},
		{
			`<script type="notjs">foo`,
			context{state: stateSpecialElementBody, element: element{name: "script"}, scriptType: "notjs"},
		},
		{
			`<script type="notjs">foo</script>`,
			context{state: stateText},
		},
		{
			`<Script>`,
			context{state: stateSpecialElementBody, element: element{name: "script"}},
		},
		{
			`<SCRIPT>foo`,
			context{state: stateSpecialElementBody, element: element{name: "script"}},
		},
		{
			`<textarea>value`,
			context{state: stateSpecialElementBody, element: element{name: "textarea"}},
		},
		{
			`<textarea>value</TEXTAREA>`,
			context{state: stateText},
		},
		{
			`<textarea name=html><b`,
			context{state: stateSpecialElementBody, element: element{name: "textarea"}},
		},
		{
			`<title>value`,
			context{state: stateSpecialElementBody, element: element{name: "title"}},
		},
		{
			`<style>value`,
			context{state: stateSpecialElementBody, element: element{name: "style"}},
		},
		{
			`<style>/* comment </b`,
			context{state: stateSpecialElementBody, element: element{name: "style"}},
		},
		{
			`<style>a[href="a</b"] {}`,
			context{state: stateSpecialElementBody, element: element{name: "style"}},
		},
		{
			// The solidus ("/") after "</style" causes "/bar)" to be consumed as attribute names.
			// See https://html.spec.whatwg.org/multipage/parsing.html#rawtext-end-tag-name-state.
			`<style>.foo { background-image: url(/</style/bar)`,
			context{state: stateAttrName, attr: attr{name: "/bar)"}},
		},
		{
			`<a xlink:href`,
			context{state: stateAttrName, element: element{name: "a"}, attr: attr{name: "xlink:href"}},
		},
		{
			`<a xmlns`,
			context{state: stateAttrName, element: element{name: "a"}, attr: attr{name: "xmlns"}},
		},
		{
			`<a xmlns:foo`,
			context{state: stateAttrName, element: element{name: "a"}, attr: attr{name: "xmlns:foo"}},
		},
		{
			`<a xmlnsxyz`,
			context{state: stateAttrName, element: element{name: "a"}, attr: attr{name: "xmlnsxyz"}},
		},
		{
			`<a data-url`,
			context{state: stateAttrName, element: element{name: "a"}, attr: attr{name: "data-url"}},
		},
		{
			`<a data-iconUri`,
			context{state: stateAttrName, element: element{name: "a"}, attr: attr{name: "data-iconuri"}},
		},
		{
			`<a data-urlItem`,
			context{state: stateAttrName, element: element{name: "a"}, attr: attr{name: "data-urlitem"}},
		},
		{
			`<a g:`,
			context{state: stateAttrName, element: element{name: "a"}, attr: attr{name: "g:"}},
		},
		{
			`<a g:url`,
			context{state: stateAttrName, element: element{name: "a"}, attr: attr{name: "g:url"}},
		},
		{
			`<a g:iconUri`,
			context{state: stateAttrName, element: element{name: "a"}, attr: attr{name: "g:iconuri"}},
		},
		{
			`<a g:urlItem`,
			context{state: stateAttrName, element: element{name: "a"}, attr: attr{name: "g:urlitem"}},
		},
		{
			`<a g:value`,
			context{state: stateAttrName, element: element{name: "a"}, attr: attr{name: "g:value"}},
		},
		{
			`<svg:font-face`,
			context{state: stateTag, element: element{name: "svg:font-face"}},
		},
		{
			`<svg:a svg:onclick="x()">`,
			context{element: element{name: "svg:a"}},
		},
		{
			`<link rel="bookmark" href=`,
			context{state: stateBeforeValue, element: element{name: "link"}, attr: attr{name: "href"}, linkRel: " bookmark "},
		},
		{
			`<link rel="   AuThOr cite    LICENSE   " href=`,
			context{state: stateBeforeValue, element: element{name: "link"}, attr: attr{name: "href"}, linkRel: " author cite license "},
		},
		{
			`<link rel="bookmark" href="www.foo.com">`,
			context{state: stateText},
		},
	}

	for _, test := range tests {
		b, e := []byte(test.input), makeEscaper(&nameSpace{})
		c := e.escapeText(context{}, &parse.TextNode{NodeType: parse.NodeText, Text: b})
		if !test.output.eq(c) {
			t.Errorf("input %s: want context\n\t%+v\ngot\n\t%+v", test.input, test.output, c)
			continue
		}
		if test.input != string(b) {
			t.Errorf("input %s: text node was modified: want %q got %q", test.input, test.input, b)
			continue
		}
	}
}

func TestEnsurePipelineContains(t *testing.T) {
	tests := []struct {
		input, output string
		ids           []string
	}{
		{
			"{{.X}}",
			".X",
			[]string{},
		},
		{
			"{{.X | html}}",
			".X | html",
			[]string{},
		},
		{
			"{{.X}}",
			".X | html",
			[]string{"html"},
		},
		{
			"{{html .X}}",
			"_evalArgs .X | html | urlquery",
			[]string{"html", "urlquery"},
		},
		{
			"{{html .X .Y .Z}}",
			"_evalArgs .X .Y .Z | html | urlquery",
			[]string{"html", "urlquery"},
		},
		{
			"{{.X | print}}",
			".X | print | urlquery",
			[]string{"urlquery"},
		},
		{
			"{{.X | print | urlquery}}",
			".X | print | urlquery",
			[]string{"urlquery"},
		},
		{
			"{{.X | urlquery}}",
			".X | html | urlquery",
			[]string{"html", "urlquery"},
		},
		{
			"{{.X | print 2 | .f 3}}",
			".X | print 2 | .f 3 | urlquery | html",
			[]string{"urlquery", "html"},
		},
		{
			// covering issue 10801
			"{{.X | println.x }}",
			".X | println.x | urlquery | html",
			[]string{"urlquery", "html"},
		},
		{
			// covering issue 10801
			"{{.X | (print 12 | println).x }}",
			".X | (print 12 | println).x | urlquery | html",
			[]string{"urlquery", "html"},
		},
		// The following test cases ensure that the merging of internal escapers
		// with the predefined "html" and "urlquery" escapers is correct.
		{
			"{{.X | urlquery}}",
			".X | _sanitizeURL | urlquery",
			[]string{"_sanitizeURL", "_normalizeURL"},
		},
		{
			"{{.X | urlquery}}",
			".X | urlquery | _sanitizeURL",
			[]string{"_sanitizeURL"},
		},
		{
			"{{.X | urlquery}}",
			".X | urlquery",
			[]string{"_normalizeURL"},
		},
		{
			"{{.X | urlquery}}",
			".X | urlquery",
			[]string{"_queryEscapeURL"},
		},
		{
			"{{.X | html}}",
			".X | html",
			[]string{"_sanitizeHTML"},
		},
		{
			"{{.X | html}}",
			".X | html",
			[]string{"_sanitizeRCDATA"},
		},
	}
	for i, test := range tests {
		tmpl := template.Must(template.New("test").Parse(test.input))
		action, ok := (tmpl.Tree.Root.Nodes[0].(*parse.ActionNode))
		if !ok {
			t.Errorf("First node is not an action: %s", test.input)
			continue
		}
		pipe := action.Pipe
		originalIDs := make([]string, len(test.ids))
		copy(originalIDs, test.ids)
		ensurePipelineContains(pipe, test.ids)
		got := pipe.String()
		if got != test.output {
			t.Errorf("#%d: %s, %v: want\n\t%s\ngot\n\t%s", i, test.input, originalIDs, test.output, got)
		}
	}
}

func TestPredefinedEscaperMerging(t *testing.T) {
	for _, test := range [...]struct {
		desc string
		in   stringConstant
		want string
	}{
		{
			`html merged with HTML escaper and URL query escaper`,
			`<a href="http://www.foo.com/main.html?a={{html "b&c=d" "></a>bar"}}">Link</a>`,
			`<a href="http://www.foo.com/main.html?a=b%26c%3dd%3e%3c%2fa%3ebar">Link</a>`,
		},
		{
			`urlquery merged with HTML escaper and URL query escaper`,
			`<a href="http://www.foo.com/main.html?a={{urlquery "b&c=d" "></a>bar"}}">Link</a>`,
			// Note: percent encoding sequences are uppercase; this is expected when the built-in
			// urlquery escaper is used instead of the internal one.
			`<a href="http://www.foo.com/main.html?a=b%26c%3Dd%3E%3C%2Fa%3Ebar">Link</a>`,
		},
		{
			`urlquery merged with HTML escaper and URL normalizer`,
			`<a href="http://www.foo.com/{{urlquery "a=b" "></a>bar"}}">Link</a>`,
			// Note: the internal URL normalizer, which would normally be applied to this action,
			// is replaced by urlquery, which is stricter. URL query reserved character '=' is
			// therefore escaped.
			`<a href="http://www.foo.com/a%3Db%3E%3C%2Fa%3Ebar">Link</a>`,
		},
	} {
		tmpl := Must(New("").Parse(test.in))
		var b bytes.Buffer
		tmpl.Execute(&b, nil)
		if got := b.String(); got != test.want {
			t.Errorf("%s: got: %s, want: %s", test.desc, got, test.want)
		}
	}
}

func TestEscapeMalformedPipelines(t *testing.T) {
	tests := []stringConstant{
		"{{ 0 | $ }}",
		"{{ 0 | $ | urlquery }}",
		"{{ 0 | (nil) }}",
		"{{ 0 | (nil) | html }}",
	}
	for _, test := range tests {
		var b bytes.Buffer
		tmpl, err := New("test").Parse(test)
		if err != nil {
			t.Errorf("failed to parse set: %q", err)
		}
		err = tmpl.Execute(&b, nil)
		if err == nil {
			t.Errorf("Expected error for %q", test)
		}
	}
}

func TestEscapeErrorsNotIgnorable(t *testing.T) {
	var b bytes.Buffer
	tmpl, _ := New("dangerous").Parse("<a")
	err := tmpl.Execute(&b, nil)
	if err == nil {
		t.Errorf("Expected error")
	} else if b.Len() != 0 {
		t.Errorf("Emitted output despite escaping failure")
	}
}

func TestEscapeSetErrorsNotIgnorable(t *testing.T) {
	var b bytes.Buffer
	tmpl, err := New("root").Parse(`{{define "t"}}<a{{end}}`)
	if err != nil {
		t.Errorf("failed to parse set: %q", err)
	}
	err = tmpl.ExecuteTemplate(&b, "t", nil)
	if err == nil {
		t.Errorf("Expected error")
	} else if b.Len() != 0 {
		t.Errorf("Emitted output despite escaping failure")
	}
}

func TestIndirectPrint(t *testing.T) {
	a := 3
	ap := &a
	b := "hello"
	bp := &b
	bpp := &bp
	tmpl := Must(New("t").Parse(`{{.}}`))
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, ap)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if buf.String() != "3" {
		t.Errorf(`Expected "3"; got %q`, buf.String())
	}
	buf.Reset()
	err = tmpl.Execute(&buf, bpp)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if buf.String() != "hello" {
		t.Errorf(`Expected "hello"; got %q`, buf.String())
	}
}

// This is a test for issue 3272.
func TestEmptyTemplate(t *testing.T) {
	page := Must(New("page").ParseFiles(os.DevNull))
	if err := page.ExecuteTemplate(os.Stdout, "page", "nothing"); err == nil {
		t.Fatal("expected error")
	}
}

type Issue7379 int

func (Issue7379) SomeMethod(x int) string {
	return fmt.Sprintf("<%d>", x)
}

// This is a test for issue 7379: type assertion error caused panic, and then
// the code to handle the panic breaks escaping. It's hard to see the second
// problem once the first is fixed, but its fix is trivial so we let that go. See
// the discussion for issue 7379.
func TestPipeToMethodIsEscaped(t *testing.T) {
	tmpl := Must(New("x").Parse("<html>{{0 | .SomeMethod}}</html>\n"))
	tryExec := func() string {
		defer func() {
			panicValue := recover()
			if panicValue != nil {
				t.Errorf("panicked: %v\n", panicValue)
			}
		}()
		var b bytes.Buffer
		tmpl.Execute(&b, Issue7379(0))
		return b.String()
	}
	for i := 0; i < 3; i++ {
		str := tryExec()
		const expect = "<html>&lt;0&gt;</html>\n"
		if str != expect {
			t.Errorf("expected %q got %q", expect, str)
		}
	}
}

// Unlike text/template, html/template crashed if given an incomplete
// template, that is, a template that had been named but not given any content.
// This is issue #10204.
func TestErrorOnUndefined(t *testing.T) {
	tmpl := New("undefined")

	err := tmpl.Execute(nil, nil)
	if err == nil {
		t.Error("expected error")
	}
	if !strings.Contains(err.Error(), "incomplete") {
		t.Errorf("expected error about incomplete template; got %s", err)
	}
}

// This covers issue #20842.
func TestIdempotentExecute(t *testing.T) {
	tmpl := Must(New("").
		Parse(`{{define "main"}}<body>{{template "hello"}}</body>{{end}}`))
	Must(tmpl.
		Parse(`{{define "hello"}}Hello, {{"Ladies & Gentlemen!"}}{{end}}`))
	got := new(bytes.Buffer)
	var err error
	// Ensure that "hello" produces the same output when executed twice.
	want := "Hello, Ladies &amp; Gentlemen!"
	for i := 0; i < 2; i++ {
		err = tmpl.ExecuteTemplate(got, "hello", nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if got.String() != want {
			t.Errorf("after executing template \"hello\", got:\n\t%q\nwant:\n\t%q\n", got.String(), want)
		}
		got.Reset()
	}
	// Ensure that the implicit re-execution of "hello" during the execution of
	// "main" does not cause the output of "hello" to change.
	err = tmpl.ExecuteTemplate(got, "main", nil)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	// If the HTML escaper is added again to the action {{"Ladies & Gentlemen!"}},
	// we would expected to see the ampersand overescaped to "&amp;amp;".
	want = "<body>Hello, Ladies &amp; Gentlemen!</body>"
	if got.String() != want {
		t.Errorf("after executing template \"main\", got:\n\t%q\nwant:\n\t%q\n", got.String(), want)
	}
}

func TestCSPCompatibilityError(t *testing.T) {
	var b bytes.Buffer
	for _, test := range [...]struct {
		in  stringConstant
		err string
	}{
		{`<a href="javascript:alert(1)">foo</a>`, `"javascript:" URI disallowed for CSP compatibility`},
		{`<a href='javascript:alert(1)'>foo</a>`, `"javascript:" URI disallowed for CSP compatibility`},
		{`<a href=javascript:alert(1)>foo</a>`, `"javascript:" URI disallowed for CSP compatibility`},
		{`<a href=javascript:alert(1)>foo</a>`, `"javascript:" URI disallowed for CSP compatibility`},
		{`<a href="javascript:alert({{ "10" }})">foo</a>`, `"javascript:" URI disallowed for CSP compatibility`},
		{`<span onclick="handle();">foo</span>`, `inline event handler "onclick" is disallowed for CSP compatibility`},
		{`<span onchange="handle();">foo</span>`, `inline event handler "onchange" is disallowed for CSP compatibility`},
		{`<span onmouseover="handle();">foo</span>`, `inline event handler "onmouseover" is disallowed for CSP compatibility`},
		{`<span onmouseout="handle();">foo</span>`, `inline event handler "onmouseout" is disallowed for CSP compatibility`},
		{`<span onkeydown="handle();">foo</span>`, `inline event handler "onkeydown" is disallowed for CSP compatibility`},
		{`<span onload="handle();">foo</span>`, `inline event handler "onload" is disallowed for CSP compatibility`},
		{`<span title="foo" onclick="handle();" id="foo">foo</span>`, `inline event handler "onclick" is disallowed for CSP compatibility`},
		{`<img src=foo.png Onerror="handle();">`, `inline event handler "onerror" is disallowed for CSP compatibility`},
	} {
		tmpl := Must(New("").CSPCompatible().Parse(test.in))
		err := tmpl.Execute(&b, nil)
		if err == nil {
			t.Errorf("template %s : expected error", test.in)
			continue
		}
		parseErr, ok := err.(*Error)
		if !ok {
			t.Errorf("template %s : expected error of type Error", test.in)
			continue
		}
		if parseErr.ErrorCode != ErrCSPCompatibility {
			t.Errorf("template %s : parseErr.ErrorCode == %d, want %d (ErrCSPCompatibility)", test.in, parseErr.ErrorCode, ErrCSPCompatibility)
			continue
		}
		if !strings.Contains(err.Error(), test.err) {
			t.Errorf("template %s : got error:\n\t%s\ndoes not contain:\n\t%s", test.in, err, test.err)
		}
	}
}
func TestScriptUnbalancedError(t *testing.T) {
	tests := [...]struct {
		in  stringConstant
		err string
	}{
		{"<script>alert(``)</script>", ""},
		{"<script>alert(`{{.}}`)</script>", "Mixing template systems"},
		{"<script>alert(`)</script>", "Missing closing `"},
		{"<script>alert(`${``})</script>", "Mixing template systems"},
		{"<script>alert(`${``}`)</script>", ""},
		{"<script>alert(`${````}`)</script>", ""},
		{"<script>alert(`${``${``}`)</script>", ""},
		{"<script>alert(`${`}`)</script>", "Mixing template systems"},
		{"<script>alert(`{{.}}`)</script>", "Missing closing `"},
		{"<script>alert(`${{.}}`)</script>", "Missing closing `"},
		{"<script>alert(`${\"`\"}`)</script>", "Mixing template systems"},
	}
	var b bytes.Buffer
	for _, test := range tests {
		tmpl := Must(New("").Parse(test.in))
		err := tmpl.Execute(&b, nil)
		if err == nil && test.err != "" {
			t.Errorf("template %s : expected error", test.in)
			continue
		}
		if err == nil && test.err == "" {
			continue
		}
		parseErr, ok := err.(*Error)
		if !ok {
			t.Errorf("template %s : expected error of type Error", test.in)
			continue
		}
		if parseErr.ErrorCode != ErrUnbalancedJsTemplate {
			t.Errorf("template %s : parseErr.ErrorCode == %d, want %d (ErrUnbalancedJsTemplate)", test.in, parseErr.ErrorCode, ErrUnbalancedJsTemplate)
			continue
		}
		if !strings.Contains(err.Error(), test.err) {
			t.Errorf("template %s : got error:\n\t%s\ndoes not contain:\n\t%s", test.in, err, test.err)
		}
	}
}

func BenchmarkEscapedExecute(b *testing.B) {
	tmpl := Must(New("t").Parse(`<a onclick="alert('{{.}}')">{{.}}</a>`))
	var buf bytes.Buffer
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmpl.Execute(&buf, "foo & 'bar' & baz")
		buf.Reset()
	}
}

// Covers issue 22780.
func TestOrphanedTemplate(t *testing.T) {
	t1 := Must(New("foo").Parse(`<a href="{{.}}">link1</a>`))
	t2 := Must(t1.New("foo").Parse(`bar`))

	var b bytes.Buffer
	const wantError = `template: "foo" is an incomplete or empty template`
	if err := t1.Execute(&b, "javascript:alert(1)"); err == nil {
		t.Fatal("expected error executing t1")
	} else if gotError := err.Error(); gotError != wantError {
		t.Fatalf("got t1 execution error:\n\t%s\nwant:\n\t%s", gotError, wantError)
	}
	b.Reset()
	if err := t2.Execute(&b, nil); err != nil {
		t.Fatalf("error executing t2: %s", err)
	}
	const want = "bar"
	if got := b.String(); got != want {
		t.Fatalf("t2 rendered %q, want %q", got, want)
	}
}
