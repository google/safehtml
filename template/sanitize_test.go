// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package template

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/google/safehtml"
	"github.com/google/safehtml/testconversions"
)

func TestDynamicElementNamePrefixEscaped(t *testing.T) {
	for _, test := range [...]struct {
		input, want string
	}{
		{
			`<{{if 1}}area{{else}}link{{end}} title="bar">`,
			`&lt;area title="bar">`,
		},
		{
			`<{{ "FOO" }} title="bar">`,
			`&lt;FOO title="bar">`,
		},
		{
			`<{{ "FOO" }}a title="bar">`,
			`&lt;FOOa title="bar">`,
		},
		{
			`<{{"script"}}>{{"doEvil()"}}</{{"script"}}>`,
			`&lt;script>doEvil()&lt;/script>`,
		},
	} {
		tmpl := Must(New(test.input).Parse(stringConstant(test.input)))
		var b bytes.Buffer
		err := tmpl.Execute(&b, nil)
		if err != nil {
			t.Errorf("%s : template execution failed:\n%s", test.input, err)
			continue
		}
		if got := b.String(); got != test.want {
			t.Errorf("%s : escaped output: got\n\t%s\nwant\n\t%s", test.input, got, test.want)
		}
	}
}

func TestConditionalElementOrAttributeAllowed(t *testing.T) {
	data := struct {
		A    safehtml.Identifier
		B    []string
		C, D bool
		X    string
	}{
		A: safehtml.IdentifierFromConstant("id"),
		B: []string{"foo", "bar"},
		C: false,
		D: true,
		X: "hello",
	}
	for _, test := range [...]string{
		// Conditional element names that lead to the same element content sanitization
		// contexts are allowed.
		`{{if .C}}<object></object>{{end}}{{ .X }}`,
		`{{if .C}}<a>{{end}}{{ .X }}`,
		`{{if .C}}<a{{else}}<b{{end}}>{{ .X }}`,
		`{{if .C}}<a>{{else}}<b>{{end}}{{ .X }}`,
		`{{if .C}}<a>{{else if .D}}<b>{{else}}<h1>{{end}}{{ .X }}`,
		`{{range .B}}<object></object>{{end}}{{ .X }}`,
		`{{range .B}}<a>{{end}}{{ .X }}`,
		`{{range .B}}<a>{{else}}<b>{{end}}{{ .X }}`,
		`{{with .C}}<object></object>{{end}}{{ .X }}`,
		`{{with .C}}<a>{{end}}{{ .X }}`,
		`{{with .C}}<a{{else}}<b{{end}}>{{ .X }}`,
		`{{with .C}}<a>{{else}}<b>{{end}}{{ .X }}`,

		// Conditional element or attribute names that lead to the same attribute value sanitization
		// contexts are allowed.
		`<input{{if .C}} checked{{end}} name="foo">`,
		`{{if .C}}<img{{else}}<audio{{end}} src="{{ .X }}">`,
		`{{if .C}}<img{{else if .D}}<audio{{else}}<input{{end}} src="{{ .X }}">`,
		`<label {{if .C}}lang{{else}}translate{{end}}="{{ .A }}">`,
		`<label {{if .C}}lang{{else if .D}}translate{{else}}spellcheck{{end}}="{{ .A }}">`,
		`<label {{range .B}}lang{{else}}translate{{end}}="{{ .A }}">`,
		`{{with .C}}<img{{else}}<audio{{end}} src="{{ .A }}">`,
		`<label {{with .C}}lang{{else}}translate{{end}}="{{ .A }}">`,

		// Conditional insertion of an attribute-value pair with a fixed attribute name is allowed, even if the
		// attributes have different sanitization contexts.
		`<a {{if .C}}id="{{ .A }}"{{end}}>foo</a>`,
		`<a {{if .C}}id="{{ .A }}"{{else}}href="{{ .X }}"{{end}}>foo</a>`,
		`<a {{if .C}}id="{{ .A }}"{{else if .D}}href="{{ .X }}"{{else}}class="{{ .X }}"{{end}}>foo</a>`,
		`<a {{with .C}}id="{{ .A }}"{{end}}>foo</a>`,
		`<a {{with .C}}id="{{ .A }}"{{else}}href="{{ .X }}"{{end}}>foo</a>`,
	} {
		tmpl := Must(New(test).Parse(stringConstant(test)))
		var b bytes.Buffer
		err := tmpl.Execute(&b, data)
		if err != nil {
			t.Errorf("unexpected an error for template %s :\n\t%s", test, err)
			continue
		}
	}
}

func TestConditionalElementOrAttributeErrorMessage(t *testing.T) {
	// The conditional branch error message prefix should only be emitted for templates with
	// a conditional branching error.
	conditionalBranchMsgPattern := regexp.MustCompile(`conditional branch with .* results in sanitization error: `)
	for _, test := range [...]struct {
		in, err                 string
		hasConditionalBranchMsg bool
	}{
		{
			`<option foo="{{ . }}">`,
			`actions must not occur in the "foo" attribute value context of a "option" element`,
			false,
		},
		{
			`<option {{if .X}}foo{{else}}bar{{end}}="{{ . }}">`,
			`actions must not occur in the "foo" attribute value context of a "option" element`,
			true,
		},
		{
			`<foo>{{ . }}</foo>`,
			`actions must not occur in the element content context of a "foo" element`,
			false,
		},
		{
			`{{if .X}}<foo>{{else}}<bar>{{end}}{{ . }}</imaginaryelement>`,
			`actions must not occur in the element content context of a "foo" element`,
			true,
		},
	} {
		tmpl := Must(New("").Parse(stringConstant(test.in)))
		var b bytes.Buffer
		err := tmpl.Execute(&b, nil)
		if err == nil {
			t.Fatalf("expected an error")
		}
		got := err.Error()
		hasConditionalBranchMsg := conditionalBranchMsgPattern.MatchString(got)
		if hasConditionalBranchMsg && !test.hasConditionalBranchMsg {
			t.Errorf("%s : error message unexpectedly reports conditional branching failure:\n\t%q", test.in, got)
		} else if !hasConditionalBranchMsg && test.hasConditionalBranchMsg {
			t.Errorf("%s : error message does not report conditional branching failure:\n\t%q", test.in, got)
		}
		if !strings.Contains(got, test.err) {
			t.Errorf("error\n\t%s\ndoes not contain expected string\n\t%s", got, test.err)
		}
	}
}

// TestNilEmptySliceData test that nil and empty slice data is rendered sensibly and without error.
func TestNilEmptySliceData(t *testing.T) {
	tt := Must(New("").Parse(`<b>{{ . }}</b>`))
	for _, test := range [...]struct {
		desc string
		in   interface{}
		want string
	}{
		{"nil", nil, `<b>&lt;nil&gt;</b>`},
		{"zero-length slice", []string{}, "<b>[]</b>"},
		{"slice containing nil", []interface{}{nil}, "<b>[&lt;nil&gt;]</b>"},
	} {
		var b bytes.Buffer
		err := tt.Execute(&b, test.in)
		if err != nil {
			t.Fatalf("%s: unexpected error: %s", test.desc, err)
		}
		if got := b.String(); got != test.want {
			t.Errorf("%s: got %q, want %q", test.desc, got, test.want)
		}
	}
}

var testConversionFuncs = FuncMap{
	"makeHTMLForTest":               func(s string) safehtml.HTML { return testconversions.MakeHTMLForTest(s) },
	"makeURLForTest":                func(s string) safehtml.URL { return testconversions.MakeURLForTest(s) },
	"makeTrustedResourceURLForTest": func(s string) safehtml.TrustedResourceURL { return testconversions.MakeTrustedResourceURLForTest(s) },
	"makeStyleForTest":              func(s string) safehtml.Style { return testconversions.MakeStyleForTest(s) },
	"makeStyleSheetForTest":         func(s string) safehtml.StyleSheet { return testconversions.MakeStyleSheetForTest(s) },
	"makeScriptForTest":             func(s string) safehtml.Script { return testconversions.MakeScriptForTest(s) },
	"makeIdentifierForTest":         func(s string) safehtml.Identifier { return testconversions.MakeIdentifierForTest(s) },
}

func TestSanitize(t *testing.T) {
	data := struct {
		T           bool
		A, E        []string
		QueryParams map[string]string
	}{
		T:           true,
		A:           []string{"<a>", "<b>"},
		E:           []string{},
		QueryParams: map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"},
	}
	for _, test := range [...]struct {
		input  string
		output string
		err    string
	}{
		// Overescaping.
		{
			input:  `Hello, {{"<Cincinnati>" | html}}!`,
			output: "Hello, &lt;Cincinnati&gt;!",
			err:    ``,
		},
		{
			input:  `Hello, {{html "<Cincinnati>"}}!`,
			output: "Hello, &lt;Cincinnati&gt;!",
			err:    ``,
		},
		{
			input:  `{{with "<Cincinnati>"}}{{$msg := .}}Hello, {{$msg}}!{{end}}`,
			output: "Hello, &lt;Cincinnati&gt;!",
			err:    ``,
		},
		// Assignment.
		{
			input:  `{{if $x := "<Hello>"}}{{$x}}{{end}}`,
			output: "&lt;Hello&gt;",
			err:    ``,
		},
		// if else.
		{
			input:  `{{if 1}}Hello{{end}}, {{"<Cincinnati>"}}!`,
			output: `Hello, &lt;Cincinnati&gt;!`,
			err:    ``,
		},
		{
			input:  `{{if 0}}{{"<Hello>"}}{{else}}{{"<Goodbye>"}}{{end}}!`,
			output: `&lt;Goodbye&gt;!`,
			err:    ``,
		},
		// with body.
		{
			input:  `{{with "<Hello>"}}{{.}}{{end}}`,
			output: "&lt;Hello&gt;",
			err:    ``,
		},
		// with-else.
		{
			input:  `{{with .E}}{{.}}{{else}}{{"<Hello>"}}{{end}}`,
			output: "&lt;Hello&gt;",
			err:    ``,
		},
		// range body.
		{
			input:  "{{range .A}}{{.}}{{end}}",
			output: "&lt;a&gt;&lt;b&gt;",
			err:    ``,
		},
		// range-else.
		{
			input:  `{{range .E}}{{.}}{{else}}{{"<Hello>"}}{{end}}`,
			output: "&lt;Hello&gt;",
			err:    ``,
		},
		// Non-string value.
		{
			input:  "{{.T}}",
			output: "true",
			err:    ``,
		},
		// Multiple attributes.
		{
			input:  `<a width="1" value="{{"<Hello>"}}">`,
			output: `<a width="1" value="&lt;Hello&gt;">`,
			err:    ``,
		},
		// HTML comment ignored.
		{
			input:  `<b>Hello, <!-- name of world -->{{"<Cincinnati>"}}</b>`,
			output: "<b>Hello, &lt;Cincinnati&gt;</b>",
			err:    ``,
		},
		{
			input:  `<!-- -{{""}}-> <script -->{{"doEvil()//"}}<!-- -{{""}}-> </script -->`,
			output: `doEvil()//`,
			err:    ``,
		},
		// HTML comment not first < in text node.
		{
			input:  "<<!-- -->!--",
			output: "&lt;!--",
			err:    ``,
		},
		{
			input:  `<<!-- -->script>{{"doEvil()"}}<<!-- -->/script>`,
			output: `&lt;script>doEvil()&lt;/script>`,
			err:    ``,
		},
		// Split HTML comment.
		{
			input:  `<b>Hello, <!-- name of {{if 1}}city -->{{"<Cincinnati>"}}{{else}}world -->{{"<Boston>"}}{{end}}</b>`,
			output: "<b>Hello, &lt;Cincinnati&gt;</b>",
			err:    ``,
		},
		// No comment injection.
		{
			input:  `<{{"!--"}}`,
			output: `&lt;!--`,
			err:    ``,
		},
		// No RCDATA end tag injection.
		{
			input:  `<textarea><{{"/textarea "}}...</textarea>`,
			output: `<textarea>&lt;/textarea ...</textarea>`,
			err:    ``,
		},
		// Template-author-controlled '<' <script> body not overescaped.
		{
			input:  `<script>var b = 1 < 2</script>`,
			output: `<script>var b = 1 < 2</script>`,
			err:    ``,
		},
		// Template-author controlled HTML metacharacters in <style>.
		{
			input:  `<style>a[href=~"<"] > b { color: blue }</style>`,
			output: `<style>a[href=~"<"] > b { color: blue }</style>`,
			err:    ``,
		},
		// HTML substitution commented out.
		{
			input:  `<p><!-- {{"<Hello>"}} --></p>`,
			output: "<p></p>",
			err:    ``,
		},
		// Comment ends flush with start.
		{
			input:  `<!--{{.}}--><p>Hello</p>`,
			output: "<p>Hello</p>",
			err:    ``,
		},
		// HTML normalization.
		{
			input:  "a < b",
			output: "a &lt; b",
			err:    ``,
		},
		{
			input:  "a << b",
			output: "a &lt;&lt; b",
			err:    ``,
		},
		{
			input:  "a<<!-- --><!-- -->b",
			output: "a&lt;b",
			err:    ``,
		},
		// HTML doctype not normalized.
		{
			input:  "<!DOCTYPE html>Hello, World!",
			output: "<!DOCTYPE html>Hello, World!",
			err:    ``,
		},
		// HTML doctype not case-insensitive.
		{
			input:  "<!doCtYPE htMl>Hello, World!",
			output: "<!doCtYPE htMl>Hello, World!",
			err:    ``,
		},
		// No doctype injection.
		{
			input:  `<!{{"DOCTYPE"}}`,
			output: "&lt;!DOCTYPE",
			err:    ``,
		},
		// range values sanitized.
		{
			input:  "<textarea>{{range .A}}{{.}}{{end}}</textarea>",
			output: "<textarea>&lt;a&gt;&lt;b&gt;</textarea>",
			err:    ``,
		},
		// Actions outside of HTML element expect HTML.
		{
			input:  `<head>title</head>{{ "<b>foo</b>" }}`,
			output: `<head>title</head>&lt;b&gt;foo&lt;/b&gt;`,
			err:    ``,
		},
		{
			input:  `<head>title</head>{{ makeHTMLForTest "<b>foo</b>" }}`,
			output: `<head>title</head><b>foo</b>`,
			err:    ``,
		},
		{
			input:  `{{ "<b>foo</b>" }}`,
			output: `&lt;b&gt;foo&lt;/b&gt;`,
			err:    ``,
		},
		{
			input:  `{{ makeHTMLForTest "<b>foo</b>" }}`,
			output: `<b>foo</b>`,
			err:    ``,
		},
		// Attribute value contexts that allow untrusted strings.
		{
			input:  `<link media="{{ "print" }}">`,
			output: `<link media="print">`,
			err:    ``,
		},
		{
			input:  `<form method="{{ "get<" }}"></form>`,
			output: `<form method="get&lt;"></form>`, // untrusted string is still HTML-escaped
			err:    ``,
		},
		// Element content contexts that expect HTML.
		{
			input:  `<span>{{ "<b>foo</b>" }}</span>`,
			output: `<span>&lt;b&gt;foo&lt;/b&gt;</span>`,
			err:    ``,
		},
		{
			input:  `<span>{{ makeHTMLForTest "<b>foo</b>" }}</span>`,
			output: `<span><b>foo</b></span>`,
			err:    ``,
		},
		// Attribute value contexts that expect HTML.
		{
			input:  `<iframe srcdoc="{{ "<a href=\"https://www.foo.com\">foo</a>" }}">{{ "<b>bar</b>" }}</iframe>`,
			output: ``,
			err:    `expected a safehtml.HTML value`,
		},
		{
			input:  `<iframe srcdoc="{{ makeHTMLForTest "<a href=\"https://www.foo.com\">foo</a>" }}">{{ makeHTMLForTest "<b>bar</b>" }}</iframe>`,
			output: `<iframe srcdoc="&lt;a href=&#34;https://www.foo.com&#34;&gt;foo&lt;/a&gt;"><b>bar</b></iframe>`,
			err:    ``,
		},
		// Attribute value contexts that expect URL.
		// safehtml.URL values should still be HTML-escaped even after bypassing URL sanitization.
		{
			input:  `<q cite="{{ "data:,\"><script>alert('pwned!')</script>" }}">foo</q>`,
			output: `<q cite="about:invalid#zGoSafez">foo</q>`,
			err:    ``,
		},
		{
			input:  `<q cite="{{ makeURLForTest "data:,\"><script>alert('pwned!')</script>" }}">foo</q>`,
			output: `<q cite="data:,%22%3e%3cscript%3ealert%28%27pwned!%27%29%3c/script%3e">foo</q>`,
			err:    ``,
		},
		{
			input:  `<link rel="alternate" href="{{ "data:,\"><script>alert('pwned!')</script>" }}">`,
			output: `<link rel="alternate" href="about:invalid#zGoSafez">`,
			err:    ``,
		},
		{
			input:  `<link rel="alternate" href="{{ makeURLForTest "data:,\"><script>alert('pwned!')</script>" }}">`,
			output: `<link rel="alternate" href="data:,%22%3e%3cscript%3ealert%28%27pwned!%27%29%3c/script%3e">`,
			err:    ``,
		},
		{
			input:  `<q cite="{{ "data:,\"><script>alert('pwned!')</script>" }}my/path">foo</q>`,
			output: `<q cite="about:invalid#zGoSafezmy/path">foo</q>`,
			err:    ``,
		},
		{
			input:  `<q cite="{{ makeURLForTest "http://www.foo.com/" }}my/path">foo</q>`,
			output: `<q cite="http://www.foo.com/my/path">foo</q>`,
			err:    ``,
		},
		{
			input:  `<q cite="{{ makeURLForTest "http://www.foo.com/" }}main?a={{ "b&c=d" }}">foo</q>`,
			output: `<q cite="http://www.foo.com/main?a=b%26c%3dd">foo</q>`,
			err:    ``,
		},
		{
			input:  `<q cite="{{ makeURLForTest "http://www.foo.com/" }}main?a={{ "w&x" }}&b={{ "y#z" }}">foo</q>`,
			output: `<q cite="http://www.foo.com/main?a=w%26x&b=y%23z">foo</q>`,
			err:    ``,
		},
		{
			input:  `<q cite="http://www.foo.com/{{ "multiple/path/segments" }}">foo</q>`,
			output: `<q cite="http://www.foo.com/multiple/path/segments">foo</q>`,
			err:    ``,
		},
		{
			input:  `<q cite="/foo?q={{ "bar&x=baz" }}">foo</q>`,
			output: `<q cite="/foo?q=bar%26x%3dbaz">foo</q>`,
			err:    ``,
		},
		{
			input:  `<q cite="/foo?q={{ "bar&x=baz" }}&j={{ "bar&x=baz" }}">foo</q>`,
			output: `<q cite="/foo?q=bar%26x%3dbaz&j=bar%26x%3dbaz">foo</q>`,
			err:    ``,
		},
		{
			input:  `<q cite="http://www.foo.com/{{ "multiple/path/segments" }}?q={{ "bar&x=baz" }}">foo</q>`,
			output: `<q cite="http://www.foo.com/multiple/path/segments?q=bar%26x%3dbaz">foo</q>`,
			err:    ``,
		},
		{
			input:  `<q cite="?q={{ "myQuery" }}&hl={{ "en" }}">foo</q>`,
			output: `<q cite="?q=myQuery&hl=en">foo</q>`,
			err:    ``,
		},
		{
			input:  `<q cite="j{{ "avascript:alert(1)" }}">foo</q>`,
			output: ``,
			err:    `action cannot be interpolated into the "cite" URL attribute value of this "q" element: URL prefix "j" is unsafe; it might be interpreted as part of a scheme`,
		},
		{
			input:  `<q cite="javascript:{{ "alert(1)" }}">foo</q>`,
			output: ``,
			err:    `action cannot be interpolated into the "cite" URL attribute value of this "q" element: URL prefix "javascript:" contains an unsafe scheme`,
		},
		{
			input:  `<q cite="  {{ "not interpreted as a URL prefix" }}">foo</q>`,
			output: ``,
			err:    `action cannot be interpolated into the "cite" URL attribute value of this "q" element: URL prefix "  " contains whitespace or control characters`,
		},
		{
			input:  `<q cite="{{ "http://www.foo.com/?q=hello\\.world" }}">foo</q>`,
			output: `<q cite="http://www.foo.com/?q=hello%5c.world">foo</q>`,
			err:    ``,
		},
		{
			input:  `<q cite="/path/{{ ".." }}/{{ "foo" }}?n1=v1">foo</q>`,
			output: `<q cite="/path/../foo?n1=v1">foo</q>`,
			err:    ``,
		},
		{
			input:  `<q cite="/foo?a=b{{range $k, $v := .QueryParams}}&amp;{{$k}}={{$v}}{{end}}">foo</q>`,
			output: `<q cite="/foo?a=b&amp;k1=v1&amp;k2=v2&amp;k3=v3">foo</q>`,
			err:    ``,
		},
		// Safe type values that are inappropriate for the HTML context get
		// unpacked into string and sanitized at run-time.
		{
			input:  `<span>{{ makeScriptForTest "alert(\"foo\");" }}</span>`,
			output: `<span>alert(&#34;foo&#34;);</span>`,
			err:    ``,
		},
		{
			input:  `<q cite="{{ makeStyleForTest "width: 1em;height: 1em;" }}">foo</q>`,
			output: `<q cite="about:invalid#zGoSafez">foo</q>`,
			err:    ``,
		},
		// Attribute value contexts that expect TrustedResouceURL.
		{
			input:  `<link href="{{ "data:,foo" }}">`,
			output: ``,
			err:    `expected a safehtml.TrustedResourceURL value`,
		},
		{
			input:  `<link href="{{ makeTrustedResourceURLForTest "data:,foo" }}">`,
			output: `<link href="data:,foo">`,
			err:    ``,
		},
		{
			input:  `<iframe src="{{ "data:,foo" }}"></iframe>`,
			output: ``,
			err:    `expected a safehtml.TrustedResourceURL value`,
		},
		{
			input:  `<iframe src="{{ makeTrustedResourceURLForTest "data:,foo" }}"></iframe>`,
			output: `<iframe src="data:,foo"></iframe>`,
			err:    ``,
		},
		{
			input:  `<link href="{{ "data:,foo" }}my/path">`,
			output: ``,
			err:    `expected a safehtml.TrustedResourceURL value`,
		},
		{
			input:  `<link href="{{ makeTrustedResourceURLForTest "https://www.foo.com/" }}my/path">`,
			output: `<link href="https://www.foo.com/my/path">`,
			err:    ``,
		},
		{
			input:  `<link href="  {{ "not interpreted as a URL prefix" }}">`,
			output: ``,
			err:    `action cannot be interpolated into the "href" URL attribute value of this "link" element: URL prefix "  " contains whitespace or control characters`,
		},
		{
			// Note: the error message here is confusing, since "main?a=" isn't actually the prefix of the URL.
			// However, having a URL with a dynamic prefix and suffix and a static middle portion is probably
			// never a valid use case, so we are ok with this failing confusingly.
			input:  `<link href="{{ makeTrustedResourceURLForTest "https://www.foo.com/" }}main?a={{ "b&c=d" }}">`,
			output: ``,
			err:    `action cannot be interpolated into the "href" URL attribute value of this "link" element: "main?a=" is a disallowed TrustedResourceURL prefix`,
		},
		{
			input:  `<link href="/foo?q={{ "myQuery" }}&hl={{ "en" }}">`,
			output: `<link href="/foo?q=myQuery&hl=en">`,
			err:    ``,
		},
		{
			input:  `<link href="/path/{{ ".." }}/{{ "foo" }}?n1=v1">`,
			output: ``,
			err:    `cannot substitute ".." after TrustedResourceURL prefix: ".." is disallowed`,
		},
		{
			// Invalid UTF-8.
			input:  `<link href="/foo?{{ "\xFF\xFE\xFD" }}">`,
			output: `<link href="/foo?%ff%fe%fd">`,
			err:    ``,
		},
		{
			// Supplementary codepoints.
			input:  `<link href="/foo?{{ "\U00012345" }}">`,
			output: `<link href="/foo?%f0%92%8d%85">`,
			err:    ``,
		},
		{
			input:  `<link href="https://www.foo.com/{{ "main.html" }}">`,
			output: `<link href="https://www.foo.com/main.html">`,
			err:    ``,
		},
		{
			input:  `<link href="https://www.foo.com/{{ "multiple/path/segments" }}">`,
			output: `<link href="https://www.foo.com/multiple%2fpath%2fsegments">`,
			err:    ``,
		},
		{
			input:  `<link href="/foo?q={{ "bar&x=baz" }}">`,
			output: `<link href="/foo?q=bar%26x%3dbaz">`,
			err:    ``,
		},
		{
			input:  `<link href="https://www.foo.com/{{ "multiple/path/segments" }}?q={{ "bar&x=baz" }}">`,
			output: `<link href="https://www.foo.com/multiple%2fpath%2fsegments?q=bar%26x%3dbaz">`,
			err:    ``,
		},
		{
			input:  `<link href="http://www.foo.com/{{ "main.html" }}">`,
			output: ``,
			err:    `action cannot be interpolated into the "href" URL attribute value of this "link" element: "http://www.foo.com/" is a disallowed TrustedResourceURL prefix`,
		},
		{
			input:  `<link href="j{{ "avascript:alert(1)" }}">`,
			output: ``,
			err:    `action cannot be interpolated into the "href" URL attribute value of this "link" element: "j" is a disallowed TrustedResourceURL prefix`,
		},
		{
			input:  `<link href="javascript:{{ "alert(1)" }}">`,
			output: ``,
			err:    `action cannot be interpolated into the "href" URL attribute value of this "link" element: "javascript:" is a disallowed TrustedResourceURL prefix`,
		},
		// Attribute value contexts that accept both URL and TrustedResouceURL.
		{
			// URL sanitization applied to untrusted string.
			input:  `<source src="{{ "data:,\"><script>alert('pwned!')</script>" }}">`,
			output: `<source src="about:invalid#zGoSafez">`,
			err:    ``,
		},
		{
			input:  `<source src="{{ makeURLForTest "data:,\"><script>alert('pwned!')</script>" }}"> <source src="{{ makeTrustedResourceURLForTest "data:,foo" }}">`,
			output: `<source src="data:,%22%3e%3cscript%3ealert%28%27pwned!%27%29%3c/script%3e"> <source src="data:,foo">`,
			err:    ``,
		},
		{
			input:  `<source src="{{ "data:,\"><script>alert('pwned!')</script>" }}my/path">`,
			output: `<source src="about:invalid#zGoSafezmy/path">`,
			err:    ``,
		},
		{
			input:  `<source src="{{ makeURLForTest "http://www.foo.com/" }}my/path">`,
			output: `<source src="http://www.foo.com/my/path">`,
			err:    ``,
		},
		{
			input:  `<source src="{{ makeURLForTest "http://www.foo.com/" }}main?a={{ "b&c=d" }}">`,
			output: `<source src="http://www.foo.com/main?a=b%26c%3dd">`,
			err:    ``,
		},
		{
			input:  `<source src="{{ makeURLForTest "http://www.foo.com/" }}main?a={{ "w&x" }}&b={{ "y#z" }}">`,
			output: `<source src="http://www.foo.com/main?a=w%26x&b=y%23z">`,
			err:    ``,
		},
		{
			input:  `<source src="http://www.foo.com/{{ "multiple/path/segments" }}">`,
			output: `<source src="http://www.foo.com/multiple/path/segments">`,
			err:    ``,
		},
		{
			input:  `<source src="/foo?q={{ "bar&x=baz" }}">`,
			output: `<source src="/foo?q=bar%26x%3dbaz">`,
			err:    ``,
		},
		{
			input:  `<source src="/foo?q={{ "bar&x=baz" }}&j={{ "bar&x=baz" }}">`,
			output: `<source src="/foo?q=bar%26x%3dbaz&j=bar%26x%3dbaz">`,
			err:    ``,
		},
		{
			input:  `<source src="http://www.foo.com/{{ "multiple/path/segments" }}?q={{ "bar&x=baz" }}">`,
			output: `<source src="http://www.foo.com/multiple/path/segments?q=bar%26x%3dbaz">`,
			err:    ``,
		},
		{
			input:  `<source src="?q={{ "myQuery" }}&hl={{ "en" }}">`,
			output: `<source src="?q=myQuery&hl=en">`,
			err:    ``,
		},
		{
			input:  `<source src="{{ "http://www.foo.com/main" }}?q={{ "param" }}">`,
			output: `<source src="http://www.foo.com/main?q=param">`,
			err:    ``,
		},
		{
			input:  `<source src="j{{ "avascript:alert(1)" }}">`,
			output: ``,
			err:    `action cannot be interpolated into the "src" URL attribute value of this "source" element: URL prefix "j" is unsafe; it might be interpreted as part of a scheme`,
		},
		{
			input:  `<source src="javascript:{{ "alert(1)" }}">`,
			output: ``,
			err:    `action cannot be interpolated into the "src" URL attribute value of this "source" element: URL prefix "javascript:" contains an unsafe scheme`,
		},
		{
			input:  `<source src="{{ "http://www.foo.com/?q=hello\\.world" }}">`,
			output: `<source src="http://www.foo.com/?q=hello%5c.world">`,
			err:    ``,
		},
		{
			input:  `<source src="  {{ "not interpreted as a URL prefix" }}">`,
			output: ``,
			err:    `action cannot be interpolated into the "src" URL attribute value of this "source" element: URL prefix "  " contains whitespace or control characters`,
		},
		{
			input:  `<source src="/path/{{ ".." }}/{{ "foo" }}?n1=v1">`,
			output: `<source src="/path/../foo?n1=v1">`,
			err:    ``,
		},
		{
			input:  `<source src="/foo?a=b{{range $k, $v := .QueryParams}}&amp;{{$k}}={{$v}}{{end}}">`,
			output: `<source src="/foo?a=b&amp;k1=v1&amp;k2=v2&amp;k3=v3">`,
			err:    ``,
		},
		// Attribute value contexts that expect Style.
		{
			input:  `<p style="{{ "width: 1em;height: 1em;" }}">foo</p>`,
			output: ``,
			err:    `expected a safehtml.Style value`,
		},
		{
			input:  `<p style="{{ makeStyleForTest "width: 1em;height: 1em;" }}">foo</p>`,
			output: `<p style="width: 1em;height: 1em;">foo</p>`,
			err:    ``,
		},
		{
			input:  `<p style="color:green; &{{ "gt;<script>alert(1);</script>" }}">foo</p>`,
			output: ``,
			err:    `action cannot be interpolated into the "style" attribute value of this "p" element: prefix "color:green; &" ends with an incomplete HTML character reference; did you mean "&amp;" instead of "&"?`,
		},
		// Element content contexts that expect StyleSheet.
		{
			input:  `<style>{{ "P.special { color:red ; }" }}</style>`,
			output: ``,
			err:    `expected a safehtml.StyleSheet value`,
		},
		{
			input:  `<style>{{ makeStyleSheetForTest "P.special { color:red ; }" }}</style>`,
			output: `<style>P.special { color:red ; }</style>`,
			err:    ``,
		},
		{
			input:  `<style>// {{"cannot insert dynamic comment"}}</style>`,
			output: ``,
			err:    `expected a safehtml.StyleSheet value`,
		},
		{
			input:  `<style>/* </b{{"notParsedAsTagName"}} */</style>`,
			output: ``,
			err:    `expected a safehtml.StyleSheet value`,
		},
		// Element content contexts that expect Script.
		{
			input:  `<script>{{ "alert(1);" }}</script>`,
			output: ``,
			err:    `expected a safehtml.Script value`,
		},
		{
			input:  `<script>{{ makeScriptForTest "alert(1);" }}</script>`,
			output: `<script>alert(1);</script>`,
			err:    ``,
		},
		{
			input:  `<script>// {{"cannot insert dynamic comment"}}</script>`,
			output: ``,
			err:    `expected a safehtml.Script value`,
		},
		// Attribute value contexts that expect enumerated string values.
		{
			input:  `<a target="{{ "blah" }}">foo</a>`,
			output: ``,
			err:    `expected one of the following strings: ["_blank" "_self"]`,
		},
		{
			input:  `<a target="{{ "_blank" }}">foo</a>`,
			output: `<a target="_blank">foo</a>`,
			err:    ``,
		},
		{
			input:  `<a target="{{ "_self" }}">foo</a>`,
			output: `<a target="_self">foo</a>`,
			err:    ``,
		},
		{
			input:  `<a target="prefix{{ "_self" }}">foo</a>`,
			output: ``,
			err:    `partial substitutions are disallowed in the "target" attribute value context of a "a" element`,
		},
		// Attribute value contexts that expect Identifiers.
		{
			input:  `<p name="{{ "my-identifier" }}" id="{{ "my-identifier" }}">foo</p>`,
			output: ``,
			err:    `expected a safehtml.Identifier value`,
		},
		{
			input:  `<p name="{{ makeIdentifierForTest "my-identifier" }}" id="{{ makeIdentifierForTest "my-identifier" }}">foo</p>`,
			output: `<p name="my-identifier" id="my-identifier">foo</p>`,
			err:    ``,
		},
		// Element content contexts that expect RCDATA.
		{
			input:  `<textarea>{{ "</textarea><script>alert('pwned!');</script>" }}</textarea>`,
			output: `<textarea>&lt;/textarea&gt;&lt;script&gt;alert(&#39;pwned!&#39;);&lt;/script&gt;</textarea>`,
			err:    ``,
		},
		{
			input:  `<title>{{ "</title><script>alert('pwned!');</script>" }}</title>`,
			output: `<title>&lt;/title&gt;&lt;script&gt;alert(&#39;pwned!&#39;);&lt;/script&gt;</title>`,
			err:    ``,
		},
		// data-* attributes values.
		{
			input:  `<p data-foo="{{ "foo" }}" data-bar="{{ "b<a>r" }}">baz</p>`,
			output: `<p data-foo="foo" data-bar="b&lt;a&gt;r">baz</p>`,
			err:    ``,
		},
		{
			input:  `<p data-4badname="{{ "foo" }}">baz</p>`,
			output: ``,
			err:    `actions must not occur in the "data-4badname" attribute value context of a "p" element`,
		},
		// Attribute sanitization contexts propagate correctly over conditionals.
		// Notice that the if and else branches are sanitized differently and correctly.
		{
			input:  `<a {{if 1}}id="{{ "foo:bar" }}"{{else}}href="{{ "foo:bar" }}"{{end}}>foo</a>`,
			output: ``,
			err:    `expected a safehtml.Identifier value`,
		},
		{
			input:  `<a {{if 0}}id="{{ "foo:bar" }}"{{else}}href="{{ "foo:bar" }}"{{end}}>foo</a>`,
			output: `<a href="about:invalid#zGoSafez">foo</a>`,
			err:    ``,
		},
		// Conditional valueless attribute name.
		{
			input: `<img class="{{"iconClass"}}"` +
				`{{if 1}} color="{{"<iconColor>"}}"{{end}}` +
				// Double quotes inside if/else.
				` src=` +
				`{{if 1}}"/foo?{{"<iconPath>"}}"` +
				`{{else}}"images/cleardot.gif"{{end}}` +
				// Missing space before title, but it is not a
				// part of the src attribute.
				`{{if .T}}title="{{"<title>"}}"{{end}}` +
				// Quotes outside if/else.
				` alt="` +
				`{{if .T}}{{"<alt>"}}` +
				`{{else}}{{if .F}}{{"<title>"}}{{end}}` +
				`{{end}}"` +
				`>`,
			output: `<img class="iconClass" color="&lt;iconColor&gt;" src="/foo?%3ciconPath%3e"title="&lt;title&gt;" alt="&lt;alt&gt;">`,
			err:    ``,
		},
	} {
		tmpl := Must(New("").Funcs(testConversionFuncs).Parse(stringConstant(test.input)))
		var b bytes.Buffer
		err := tmpl.Execute(&b, data)
		if test.err != "" {
			if err == nil {
				t.Errorf("%s : expected error", test.input)
				continue
			}
			if got := err.Error(); !strings.Contains(got, test.err) {
				t.Errorf("%s : error\n\t%q\ndoes not contain expected string\n\t%q", test.input, got, test.err)
			}
			continue
		}
		if test.err == "" && err != nil {
			t.Errorf("%s : template execution failed:\n%s", test.input, err)
			continue
		}
		if want, got := test.output, b.String(); want != got {
			t.Errorf("%s : escaped output: got\n\t%s\nwant\n\t%s", test.input, got, want)
			continue
		}
	}
}

func TestConditionalURLPrefixError(t *testing.T) {
	data := struct {
		B         []string
		C, D      bool
		URLSuffix string
	}{
		B:         []string{"foo", "bar"},
		C:         false,
		D:         true,
		URLSuffix: "suffix",
	}
	for _, test := range [...]struct {
		input, want string
	}{
		// Conditonal URL prefix in attribute value contexts that expect URLs.
		{
			`<q cite="{{if .C}}mailto:{{end}}{{ .URLSuffix }}">foo</q>`,
			`actions must not occur after an ambiguous URL prefix`,
		},
		{
			`<q cite="{{if .C}}mailto:{{else}}javascript:{{end}}{{ .URLSuffix }}">foo</q>`,
			`actions must not occur after an ambiguous URL prefix`,
		},
		{
			`<q cite="{{if .C}}mailto{{else}}javascript{{end}}:{{ .URLSuffix }}">foo</q>`,
			`actions must not occur after an ambiguous URL prefix`,
		},
		{
			`<q cite="{{if .C}}mailto:{{else if .D}}javascript:{{else}}tel:{{end}}{{ .URLSuffix }}">foo</q>`,
			`actions must not occur after an ambiguous URL prefix`,
		},
		{
			`<q cite="{{range .B}}mailto:{{end}}{{ .URLSuffix }}">foo</q>`,
			`actions must not occur after an ambiguous URL prefix`,
		},
		{
			`<q cite="{{range .B}}mailto:{{else}}javascript:{{end}}{{ .URLSuffix }}">foo</q>`,
			`actions must not occur after an ambiguous URL prefix`,
		},
		{
			`<q cite="{{with .C}}mailto:{{end}}{{ .URLSuffix }}">foo</q>`,
			`actions must not occur after an ambiguous URL prefix`,
		},
		{
			`<q cite="{{with .C}}mailto:{{else}}javascript:{{end}}{{ .URLSuffix }}">foo</q>`,
			`actions must not occur after an ambiguous URL prefix`,
		},
		// Conditonal URL prefix in attribute value contexts that expect TrustedResourceURLs.
		{
			`<link href="{{if .C}}mailto:{{end}}{{ .URLSuffix }}">`,
			`actions must not occur after an ambiguous URL prefix`,
		},
		{
			`<link href="{{if .C}}mailto:{{else}}javascript:{{end}}{{ .URLSuffix }}">`,
			`actions must not occur after an ambiguous URL prefix`,
		},
		{
			`<link href="{{if .C}}mailto{{else}}javascript{{end}}:{{ .URLSuffix }}">`,
			`actions must not occur after an ambiguous URL prefix`,
		},
		{
			`<link href="{{if .C}}mailto:{{else if .D}}javascript:{{else}}tel:{{end}}{{ .URLSuffix }}">`,
			`actions must not occur after an ambiguous URL prefix`,
		},
		{
			`<link href="{{range .B}}mailto:{{end}}{{ .URLSuffix }}">`,
			`actions must not occur after an ambiguous URL prefix`,
		},
		{
			`<link href="{{range .B}}mailto:{{else}}javascript:{{end}}{{ .URLSuffix }}">`,
			`actions must not occur after an ambiguous URL prefix`,
		},
		{
			`<link href="{{with .C}}mailto:{{end}}{{ .URLSuffix }}">`,
			`actions must not occur after an ambiguous URL prefix`,
		},
		{
			`<link href="{{with .C}}mailto:{{else}}javascript:{{end}}{{ .URLSuffix }}">`,
			`actions must not occur after an ambiguous URL prefix`,
		},
		// Conditonal URL prefix in attribute value contexts that expect URLs or TrustedResourceURLs.
		{
			`<source src="{{if .C}}mailto:{{end}}{{ .URLSuffix }}">`,
			`actions must not occur after an ambiguous URL prefix`,
		},
		{
			`<source src="{{if .C}}mailto:{{else}}javascript:{{end}}{{ .URLSuffix }}">`,
			`actions must not occur after an ambiguous URL prefix`,
		},
		{
			`<source src="{{if .C}}mailto{{else}}javascript{{end}}:{{ .URLSuffix }}">`,
			`actions must not occur after an ambiguous URL prefix`,
		},
		{
			`<source src="{{if .C}}mailto:{{else if .D}}javascript:{{else}}tel:{{end}}{{ .URLSuffix }}">`,
			`actions must not occur after an ambiguous URL prefix`,
		},
		{
			`<source src="{{range .B}}mailto:{{end}}{{ .URLSuffix }}">`,
			`actions must not occur after an ambiguous URL prefix`,
		},
		{
			`<source src="{{range .B}}mailto:{{else}}javascript:{{end}}{{ .URLSuffix }}">`,
			`actions must not occur after an ambiguous URL prefix`,
		},
		{
			`<source src="{{with .C}}mailto:{{end}}{{ .URLSuffix }}">`,
			`actions must not occur after an ambiguous URL prefix`,
		},
		{
			`<source src="{{with .C}}mailto:{{else}}javascript:{{end}}{{ .URLSuffix }}">`,
			`actions must not occur after an ambiguous URL prefix`,
		},
	} {
		tmpl := Must(New(test.input).Parse(stringConstant(test.input)))
		var b bytes.Buffer
		err := tmpl.Execute(&b, data)
		if err == nil {
			t.Errorf("expected an error for template %s", test.input)
			continue
		}
		if got := err.Error(); !strings.Contains(got, test.want) {
			t.Errorf("got error:\n\t%s\nwant:\n\t%s", got, test.want)

		}
	}
}

func TestValidateDoesNotEndsWithCharRefPrefix(t *testing.T) {
	const wantErr = `ends with an incomplete HTML character reference; did you mean "&amp;" instead of "&"?`
	for _, test := range [...]struct {
		in    string
		valid bool
	}{
		// Incomplete HTML character escape sequences.
		{`&`, false},
		{`javascript&`, false},
		{`javascript&c`, false},
		{`javascript&colon`, false},
		{`javascript&blk1`, false},
		{`javascript&#`, false},
		{`javascript&#5`, false},
		{`javascript&#x`, false},
		{`javascript&#xa`, false},
		{`javascript&#XA`, false},
		{`javascript&#X3`, false},
		// Invalid HTML character references.
		{`javascript&x3A;`, true},
		{`javascript&x3a;`, true},
		{`javascript&X3A;`, true},
		{`javascript&X3a;`, true},
		// Complete HTML character references.
		{`javascript&colon;`, true},
		{`javascript&#58;`, true},
	} {
		err := validateDoesNotEndsWithCharRefPrefix(test.in)
		switch {
		case err != nil:
			if test.valid {
				t.Errorf("validateDoesNotEndsWithCharRefPrefix(%q) failed: %s", test.in, err)
			} else if !strings.Contains(err.Error(), wantErr) {
				t.Errorf("validateDoesNotEndsWithCharRefPrefix(%q) error\n\t%s\ndoes not contain expected string\n\t%s", test.in, err.Error(), wantErr)
			}
		case !test.valid:
			t.Errorf("validateDoesNotEndsWithCharRefPrefix(%q) succeeded unexpectedly", test.in)
		}
	}
}

func TestDataAttributeNamePattern(t *testing.T) {
	for _, test := range [...]struct {
		in   string
		want bool
	}{
		{`data-a`, true},
		{`data-foo`, true},
		{`data-foo-bar`, true},
		{`data-f0o-b4r`, true},
		{`data-_foo`, true},
		// Does not begin with "data-".
		{`data`, false},
		{`foo`, false},
		// No characters after hyphen.
		{`data-`, false},
		// Suffix starts with a digit.
		{`data-4oo`, false},
		// Contains ACSII upper alphas.
		// Note: this test case isn't strictly necessary, since sanitizerForContext is given
		// lower-case attribute names.
		{`data-Foo`, false},
		// Contains colon characters.
		{`data-foo:bar`, false},
		// Contains unicode characters that are allowed in XML names
		// (https://www.w3.org/TR/xml/#NT-Name), but conservatively rejected
		// by our regexp pattern.
		{"data-\u037Fbar", false},
		{"data-fo\u0300", false},
	} {
		if got := dataAttributeNamePattern.MatchString(test.in); got != test.want {
			t.Errorf("dataAttributeNamePattern.MatchString(%q) = %t", test.in, got)
		}
	}
}

const testSanitizationLogicWant = `cannot escape action {{.}}: unquoted attribute values disallowed`

// TestSanitizationLogic ensures that the underlying html/template sanitization logic is
// replaced by the safehtml/template sanitization logic no matter how the template is parsed
// or executed.
func TestSanitizationLogic(t *testing.T) {
	// This template will be accepted by html/template but not by safehtml/template
	// since the latter does not allow data to be substituted into unquoted attribute
	// value contexts.
	const templateText = `<a href={{.}}>unquoted href attribute value</a>`

	// Create temp file containing the template text for constructors that parse templates
	// from files.
	tmpfile, err := ioutil.TempFile("", "path")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.WriteString(templateText); err != nil {
		t.Fatal(err)
	}
	filename := stringConstant(tmpfile.Name())
	templateName := filepath.Base(tmpfile.Name())

	for _, test := range [...]struct {
		parseFuncName string
		tmpl          *Template
	}{
		{"Parse", Must(New(templateName).Parse(templateText))},
		{"ParseFromTrustedTemplate", Must(New(templateName).ParseFromTrustedTemplate(MakeTrustedTemplate(templateText)))},
		{"ParseFiles (method)", Must(New(templateName).ParseFiles(filename))},
		{"ParseFiles (function)", Must(ParseFiles(filename))},
		{"ParseFilesFromTrustedSources (method)", Must(New(templateName).ParseFilesFromTrustedSources(TrustedSourceFromConstant(filename)))},
		{"ParseFilesFromTrustedSources (function)", Must(ParseFilesFromTrustedSources(TrustedSourceFromConstant(filename)))},
		{"ParseGlob (method)", Must(New(templateName).ParseGlob(filename))},
		{"ParseGlob (function)", Must(ParseGlob(filename))},
		{"ParseGlobFromTrustedSource (method)", Must(New(templateName).ParseGlobFromTrustedSource(TrustedSourceFromConstant(filename)))},
		{"ParseGlobFromTrustedSource (function)", Must(ParseGlobFromTrustedSource(TrustedSourceFromConstant(filename)))},
	} {
		var b bytes.Buffer
		err := test.tmpl.Execute(&b, nil)
		testSanitizationLogicCheckError(t, err, test.parseFuncName, "Execute")
		_, err = test.tmpl.ExecuteToHTML(nil)
		testSanitizationLogicCheckError(t, err, test.parseFuncName, "ExecuteToHTML")
		err = test.tmpl.ExecuteTemplate(&b, templateName, nil)
		testSanitizationLogicCheckError(t, err, test.parseFuncName, "ExecuteTemplate")
		_, err = test.tmpl.ExecuteTemplateToHTML(templateName, nil)
		testSanitizationLogicCheckError(t, err, test.parseFuncName, "ExecuteTemplateToHTML")
	}
}

func testSanitizationLogicCheckError(t *testing.T, err error, parseFuncName, executeFuncName string) {
	prefix := parseFuncName + ", " + executeFuncName
	if err == nil {
		t.Errorf("%s : expected execution error", prefix)
		return
	}
	if got := err.Error(); !strings.Contains(got, testSanitizationLogicWant) {
		t.Errorf("%s : the error message:\n\t%s\ndoes not contain:\n\t%s", prefix, got, testSanitizationLogicWant)
	}
}

func TestCannotCallInternalSanitizers(t *testing.T) {
	const templateName = "test"
	// Programmatically generate templates that call each sanitizer function in the internal
	// function map.
	for sanitizerName := range funcs {
		tmplText := `{{ "foo" | ` + sanitizerName + ` }}`
		_, err := New(templateName).Parse(stringConstant(tmplText))
		if err == nil {
			t.Errorf("expected error parsing template which calls internal sanitizer %q", sanitizerName)
		}
	}
}

func TestExecuteErrors(t *testing.T) {
	for _, test := range [...]struct {
		desc      string
		tmpl      stringConstant
		data      interface{}
		want      string
		fullMatch bool
	}{
		{
			desc: `invalid template`,
			tmpl: `{{template "foo"}}`,
			want: `no such template "foo"`,
		},
		{
			desc: `missing '"' after recursive call`,
			tmpl: `<select size="{{template "y"}}></select>` +
				`{{define "y"}}{{if .Tail}}{{template "y" .Tail}}{{end}}3"{{end}}`,
			want: `cannot compute output context for template y$htmltemplate_StateAttr_DelimDoubleQuote_attrSize_elementSelect`,
		},
		{
			desc: `element and attribute name confused`,
			tmpl: `<a=foo>`,
			want: `: expected space, attr name, or end of tag, but got "=foo>"`,
		},
		{
			desc: `urlquery is disallowed if it is not the last command in the pipeline`,
			tmpl: `Hello, {{. | urlquery | print}}!`,
			want: `predefined escaper "urlquery" disallowed in template`,
		},
		{
			desc: `html is disallowed if it is not the last command in the pipeline`,
			tmpl: `Hello, {{. | html | print}}!`,
			want: `predefined escaper "html" disallowed in template`,
		},
		{
			desc: `direct call to html is disallowed if it is not the last command in the pipeline`,
			tmpl: `Hello, {{html . | print}}!`,
			want: `predefined escaper "html" disallowed in template`,
		},
		{
			desc: `html is disallowed in a pipeline that is in an unquoted attribute context, even if it is the last command in the pipeline`,
			tmpl: `<div class={{. | html}}>Hello<div>`,
			want: `predefined escaper "html" disallowed in template`,
		},
		{
			desc: `html is allowed since it is the last command in the pipeline, but urlquery is not`,
			tmpl: `Hello, {{. | urlquery | html}}!`,
			want: `predefined escaper "urlquery" disallowed in template`,
		},
		{
			desc: `unquoted attribute value disallowed`,
			tmpl: `<a title={{ . }}>bar</a>`,
			want: `unquoted attribute values disallowed`,
		},
		{
			desc: `dynamic element name suffix 1`,
			tmpl: `<a{{ "foo" }} title="foo">`,
			want: `actions must not affect element or attribute names`,
		},
		{
			desc: `dynamic element name suffix 2`,
			tmpl: `<div{{template "y"}}>` +
				// Illegal starting in stateTag but not in stateText.
				`{{define "y"}} foo<b{{end}}`,
			want: `"<" in attribute name: " foo<b"`,
		},
		{
			desc: `dynamic whole attribute name 1`,
			tmpl: `<area {{ "foo" }}>`,
			want: `actions must not affect element or attribute names`,
		},
		{
			desc: `dynamic whole attribute name 2`,
			tmpl: `<area {{ "foo" }} title="foo">`,
			want: `actions must not affect element or attribute names`,
		},
		{
			desc: `dynamic whole attribute name 3`,
			tmpl: `<area title="foo" {{ "foo" }}>`,
			want: `actions must not affect element or attribute names`,
		},
		{
			desc: `dynamic whole attribute name 4`,
			tmpl: `<area {{ "foo" }}="foo">`,
			want: `actions must not affect element or attribute names`,
		},
		{
			desc: `dynamic attribute name suffix`,
			tmpl: `<area t{{ "foo" }}="foo">`,
			want: `actions must not affect element or attribute names`,
		},
		{
			desc: `dynamic attribute name prefix`,
			tmpl: `<area {{ "foo" }}t="foo">`,
			want: `actions must not affect element or attribute names`,
		},
		{
			desc: `missing quote in the else branch`,
			tmpl: `{{if .Cond}}<a href="foo">{{else}}<a href="bar>{{end}}`,
			want: `{{if}} branches end in different contexts`,
		},
		// When we have a conditional element or attribute name suffix, the
		// sanitizer only sees the name prefix when performing template sanitization.
		// The prefix is very unlikely to be an allowed element or attribute name
		// on its own, and will therefore be rejected.
		{
			desc: `conditional element name suffix`,
			tmpl: `<me{{if 1}}ta{{else}}nuitem{{end}}>{{ "foo" }}`,
			want: `actions must not occur in the element content context of a "me" element`,
		},
		{
			desc: `conditional attribute name suffix`,
			tmpl: `<area d{{if 1}}raggabl{{else}}ropzon{{end}}e="{{ "foo" }}">`,
			want: `actions must not occur in the "d" attribute value context of a "area" element`,
		},
		{
			desc: `if = disallowed element, else = HTML, safehtml/template conditonal branch error`,
			tmpl: `{{if 0}}<object>{{end}}{{ "hello" }}`,
			want: `conditional branch with element "object" results in sanitization error: ` +
				`actions must not occur in the element content context of a "object" element`,
		},
		{
			desc: `if = Script, else = HTML, html/template conditonal branch error`,
			tmpl: `{{if 0}}<script>{{end}}{{ "hello" }}`,
			want: `branches end in different contexts`,
		},
		{
			desc: `if = Script, else = HTML, safehtml/template conditonal branch error`,
			tmpl: `{{if 0}}<script{{else}}<span{{end}}>{{ "hello" }}`,
			want: `conditional branches end in different element content sanitization contexts: ` +
				`element "script" has sanitization context "Script", ` +
				`element "span" has sanitization context "HTML"`,
		},
		{
			desc: `if = Script, else = HTML, html/template conditonal branch error`,
			tmpl: `{{if 0}}<script>{{else}}<span>{{end}}{{ "hello" }}`,
			want: `branches end in different contexts`,
		},
		{
			desc: `if = Script, else if = HTML, else = HTML, html/template conditonal branch error`,
			tmpl: `{{if 0}}<script>{{else if 1}}<span>{{else}}<b>{{end}}{{ "hello" }}`,
			want: `branches end in different contexts`,
		},
		{
			desc: `range = disallowed element, else = HTML, safehtml/template conditonal branch error`,
			tmpl: `{{range .}}<object>{{end}}{{ "hello" }}`,
			data: []string{"foo", "bar"},
			want: `conditional branch with element "object" results in sanitization error: ` +
				`actions must not occur in the element content context of a "object" element`,
		},
		{
			desc: `range = Script, else = HTML, html/template conditonal branch error`,
			tmpl: `{{range .}}<script>{{end}}{{ "hello" }}`,
			data: []string{"foo", "bar"},
			want: `branches end in different contexts`,
		},
		{
			desc: `range = Script, else = HTML, html/template conditonal branch error`,
			tmpl: `{{range .}}<script>{{else}}<area>{{end}}{{ "hello" }}`,
			data: []string{"foo", "bar"},
			want: `branches end in different contexts`,
		},
		{
			desc: `with = disallowed element, else = HTML, safehtml/template conditonal branch error`,
			tmpl: `{{with 0}}<object>{{end}}{{ "hello" }}`,
			want: `conditional branch with element "object" results in sanitization error: ` +
				`actions must not occur in the element content context of a "object" element`,
		},
		{
			desc: `with = Script, else = HTML, html/template conditonal branch error`,
			tmpl: `{{with 0}}<script>{{end}}{{ "hello" }}`,
			want: `branches end in different contexts`,
		},
		{
			desc: `with = Script, else = HTML, safehtml/template conditonal branch error`,
			tmpl: `{{with 0}}<script{{else}}<span{{end}}>{{ "hello" }}`,
			want: `conditional branches end in different element content sanitization contexts: ` +
				`element "script" has sanitization context "Script", ` +
				`element "span" has sanitization context "HTML"`,
		},
		{
			desc: `with = Script, else = HTML, html/template conditonal branch error`,
			tmpl: `{{with 0}}<script>{{else}}<span>{{end}}{{ "hello" }}`,
			want: `branches end in different contexts`,
		},
		{
			desc: `if = disallowed attribute, else = no sanitization, safehtml/template conditonal branch error`,
			tmpl: `<p {{if 0}}customattr{{else}}class{{end}}="{{ "hello" }}">`,
			want: `conditional branch with {element="p", attribute="customattr"} results in sanitization error: ` +
				`actions must not occur in the "customattr" attribute value context of a "p" element`,
		},
		{
			desc: `if = TrustedResourceURLOrURL, else = TrustedResourceURL, safehtml/template conditonal branch error`,
			tmpl: `{{if 0}}<img{{else}}<track{{end}} src="{{ "hello" }}">`,
			want: `conditional branches end in different attribute value sanitization contexts: ` +
				`{element="img", attribute="src"} has sanitization context "TrustedResourceURLOrURL", ` +
				`{element="track", attribute="src"} has sanitization context "TrustedResourceURL"`,
		},
		{
			desc: `if = TrustedResourceURLOrURL, else if = TrustedResourceURLOrURL, else = TrustedResourceURL, html/template conditonal branch error`,
			tmpl: `{{if 0}}<img{{else if 1}}<audio{{else}}<track{{end}} src="{{ "hello" }}">`,
			want: `conditional branches end in different attribute value sanitization contexts: ` +
				`{element="img", attribute="src"} has sanitization context "TrustedResourceURLOrURL", ` +
				`{element="track", attribute="src"} has sanitization context "TrustedResourceURL"`,
		},
		{
			desc: `if = TrustedResourceURLOrURL, else = Identifier, safehtml/template conditonal branch error`,
			tmpl: `<a {{if 0}}href{{else}}id{{end}}="{{ "hello" }}">`,
			want: `conditional branches end in different attribute value sanitization contexts: ` +
				`{element="a", attribute="href"} has sanitization context "TrustedResourceURLOrURL", ` +
				`{element="a", attribute="id"} has sanitization context "Identifier"`,
		},
		{
			desc: `if = TrustedResourceURLOrURL, else if = Identifier, else = TargetEnum, safehtml/template conditonal branch error`,
			tmpl: `<a {{if 0}}href{{else if .D}}id{{else}}target{{end}}="{{ "hello" }}">`,
			want: `conditional branches end in different attribute value sanitization contexts: ` +
				`{element="a", attribute="href"} has sanitization context "TrustedResourceURLOrURL", ` +
				`{element="a", attribute="target"} has sanitization context "TargetEnum"`,
		},
		{
			desc: `range = disallowed attribute, else = no sanitization, safehtml/template conditonal branch error`,
			tmpl: `<p {{range .}}customattr{{else}}class{{end}}="{{ "hello" }}">`,
			data: []string{"foo", "bar"},
			want: `conditional branch with {element="p", attribute="customattr"} results in sanitization error: ` +
				`actions must not occur in the "customattr" attribute value context of a "p" element`,
		},
		{
			desc: `range = TrustedResourceURLOrURL, else = Identifier, safehtml/template conditonal branch error`,
			tmpl: `<a {{range .}}href{{else}}id{{end}}="{{ "hello" }}">`,
			data: []string{"foo", "bar"},
			want: `conditional branches end in different attribute value sanitization contexts: ` +
				`{element="a", attribute="href"} has sanitization context "TrustedResourceURLOrURL", ` +
				`{element="a", attribute="id"} has sanitization context "Identifier"`,
		},
		{
			desc: `with = disallowed attribute, else = no sanitization, safehtml/template conditonal branch error`,
			tmpl: `<p {{with 0}}customattr{{else}}class{{end}}="{{ "hello" }}">`,
			want: `conditional branch with {element="p", attribute="customattr"} results in sanitization error: ` +
				`actions must not occur in the "customattr" attribute value context of a "p" element`,
		},
		{
			desc: `with = TrustedResourceURLOrURL, else = HTML, html/template conditonal branch error`,
			tmpl: `{{with 0}}<img{{end}} src="{{ "hello" }}">`,
			want: `branches end in different contexts`,
		},
		{
			desc: `with = TrustedResourceURLOrURL, else = TrustedResourceURL, safehtml/template conditonal branch error`,
			tmpl: `{{with 0}}<img{{else}}<track{{end}} src="{{ "hello" }}">`,
			want: `conditional branches end in different attribute value sanitization contexts: ` +
				`{element="img", attribute="src"} has sanitization context "TrustedResourceURLOrURL", ` +
				`{element="track", attribute="src"} has sanitization context "TrustedResourceURL"`,
		},
		{
			desc: `with = TrustedResourceURLOrURL, else = Identifier, safehtml/template conditonal branch error`,
			tmpl: `<a {{with 0}}href{{else}}id{{end}}="{{ "hello" }}">`,
			want: `conditional branches end in different attribute value sanitization contexts: ` +
				`{element="a", attribute="href"} has sanitization context "TrustedResourceURLOrURL", ` +
				`{element="a", attribute="id"} has sanitization context "Identifier"`,
		},
		{
			desc: `disallowed attributes disallowed`,
			tmpl: `<option customattr="{{ . }}">`,
			want: `actions must not occur in the "customattr" attribute value context of a "option" element`,
		},
		{
			desc: `disallowed element name disallowed 1`,
			tmpl: `<imaginaryelement>{{ . }}</imaginaryelement>`,
			want: `actions must not occur in the element content context of a "imaginaryelement" element`,
		},
		{
			desc: `disallowed element name disallowed 2`,
			tmpl: `<base title="{{ . }}">`,
			want: `actions must not occur in the "title" attribute value context of a "base" element`,
		},
		{
			desc: `disallowed element name disallowed 3`,
			tmpl: `<meta title="{{ . }}">`,
			want: `actions must not occur in the "title" attribute value context of a "meta" element`,
		},
		{
			desc: `disallowed element name disallowed 4`,
			tmpl: `<object src="{{ . }}"></object>`,
			want: `actions must not occur in the "src" attribute value context of a "object" element`,
		},
		{
			desc: `disallowed element name disallowed 5`,
			tmpl: `<object>{{ . }}</object>`,
			want: `actions must not occur in the element content context of a "object" element`,
		},
		{
			desc: `conditional URL prefix error in URL attribute value sanitization context 1`,
			tmpl: `<q cite="{{if 0}}mailto:{{end}}{{ "suffix" }}">foo</q>`,
			want: `actions must not occur after an ambiguous URL prefix`,
		},
		{
			desc: `conditional URL prefix error in URL attribute value sanitization context 2`,
			tmpl: `<q cite="{{if 0}}mailto:{{else}}javascript:{{end}}{{ "suffix" }}">foo</q>`,
			want: `actions must not occur after an ambiguous URL prefix`,
		},
		{
			desc: `conditional URL prefix error in URL attribute value sanitization context 3`,
			tmpl: `<q cite="{{if 0}}mailto{{else}}javascript{{end}}:{{ "suffix" }}">foo</q>`,
			want: `actions must not occur after an ambiguous URL prefix`,
		},
		{
			desc: `conditional URL prefix error in URL attribute value sanitization context 4`,
			tmpl: `<q cite="{{if 0}}mailto:{{else if 1}}javascript:{{else}}tel:{{end}}{{ "suffix" }}">foo</q>`,
			want: `actions must not occur after an ambiguous URL prefix`,
		},
		{
			desc: `conditional URL prefix error in URL attribute value sanitization context 5`,
			tmpl: `<q cite="{{range .B}}mailto:{{end}}{{ "suffix" }}">foo</q>`,
			data: []string{"foo", "bar"},
			want: `actions must not occur after an ambiguous URL prefix`,
		},
		{
			desc: `conditional URL prefix error in URL attribute value sanitization context 6`,
			tmpl: `<q cite="{{range .B}}mailto:{{else}}javascript:{{end}}{{ "suffix" }}">foo</q>`,
			data: []string{"foo", "bar"},
			want: `actions must not occur after an ambiguous URL prefix`,
		},
		{
			desc: `conditional URL prefix error in URL attribute value sanitization context 7`,
			tmpl: `<q cite="{{with 0}}mailto:{{end}}{{ "suffix" }}">foo</q>`,
			want: `actions must not occur after an ambiguous URL prefix`,
		},
		{
			desc: `conditional URL prefix error in URL attribute value sanitization context 8`,
			tmpl: `<q cite="{{with 0}}mailto:{{else}}javascript:{{end}}{{ "suffix" }}">foo</q>`,
			want: `actions must not occur after an ambiguous URL prefix`,
		},
		{
			desc: `conditional URL prefix error in TrustedResourceURL attribute value sanitization context 1`,
			tmpl: `<link href="{{if 0}}mailto:{{end}}{{ "suffix" }}">`,
			want: `actions must not occur after an ambiguous URL prefix`,
		},
		{
			desc: `conditional URL prefix error in TrustedResourceURL attribute value sanitization context 2`,
			tmpl: `<link href="{{if 0}}mailto:{{else}}javascript:{{end}}{{ "suffix" }}">`,
			want: `actions must not occur after an ambiguous URL prefix`,
		},
		{
			desc: `conditional URL prefix error in TrustedResourceURL attribute value sanitization context 3`,
			tmpl: `<link href="{{if 0}}mailto{{else}}javascript{{end}}:{{ "suffix" }}">`,
			want: `actions must not occur after an ambiguous URL prefix`,
		},
		{
			desc: `conditional URL prefix error in TrustedResourceURL attribute value sanitization context 4`,
			tmpl: `<link href="{{if 0}}mailto:{{else if 1}}javascript:{{else}}tel:{{end}}{{ "suffix" }}">`,
			want: `actions must not occur after an ambiguous URL prefix`,
		},
		{
			desc: `conditional URL prefix error in TrustedResourceURL attribute value sanitization context 5`,
			tmpl: `<link href="{{range .B}}mailto:{{end}}{{ "suffix" }}">`,
			data: []string{"foo", "bar"},
			want: `actions must not occur after an ambiguous URL prefix`,
		},
		{
			desc: `conditional URL prefix error in TrustedResourceURL attribute value sanitization context 6`,
			tmpl: `<link href="{{range .B}}mailto:{{else}}javascript:{{end}}{{ "suffix" }}">`,
			data: []string{"foo", "bar"},
			want: `actions must not occur after an ambiguous URL prefix`,
		},
		{
			desc: `conditional URL prefix error in TrustedResourceURL attribute value sanitization context 7`,
			tmpl: `<link href="{{with 0}}mailto:{{end}}{{ "suffix" }}">`,
			want: `actions must not occur after an ambiguous URL prefix`,
		},
		{
			desc: `conditional URL prefix error in TrustedResourceURL attribute value sanitization context 8`,
			tmpl: `<link href="{{with 0}}mailto:{{else}}javascript:{{end}}{{ "suffix" }}">`,
			want: `actions must not occur after an ambiguous URL prefix`,
		},
		{
			desc: `conditional URL prefix error in TrustedResourceURLOrURL attribute value sanitization context 1`,
			tmpl: `<source src="{{if 0}}mailto:{{end}}{{ "suffix" }}">`,
			want: `actions must not occur after an ambiguous URL prefix`,
		},
		{
			desc: `conditional URL prefix error in TrustedResourceURLOrURL attribute value sanitization context 2`,
			tmpl: `<source src="{{if 0}}mailto:{{else}}javascript:{{end}}{{ "suffix" }}">`,
			want: `actions must not occur after an ambiguous URL prefix`,
		},
		{
			desc: `conditional URL prefix error in TrustedResourceURLOrURL attribute value sanitization context 3`,
			tmpl: `<source src="{{if 0}}mailto{{else}}javascript{{end}}:{{ "suffix" }}">`,
			want: `actions must not occur after an ambiguous URL prefix`,
		},
		{
			desc: `conditional URL prefix error in TrustedResourceURLOrURL attribute value sanitization context 4`,
			tmpl: `<source src="{{if 0}}mailto:{{else if 1}}javascript:{{else}}tel:{{end}}{{ "suffix" }}">`,
			want: `actions must not occur after an ambiguous URL prefix`,
		},
		{
			desc: `conditional URL prefix error in TrustedResourceURLOrURL attribute value sanitization context 5`,
			tmpl: `<source src="{{range .B}}mailto:{{end}}{{ "suffix" }}">`,
			data: []string{"foo", "bar"},
			want: `actions must not occur after an ambiguous URL prefix`,
		},
		{
			desc: `conditional URL prefix error in TrustedResourceURLOrURL attribute value sanitization context 6`,
			tmpl: `<source src="{{range .B}}mailto:{{else}}javascript:{{end}}{{ "suffix" }}">`,
			data: []string{"foo", "bar"},
			want: `actions must not occur after an ambiguous URL prefix`,
		},
		{
			desc: `conditional URL prefix error in TrustedResourceURLOrURL attribute value sanitization context 7`,
			tmpl: `<source src="{{with 0}}mailto:{{end}}{{ "suffix" }}">`,
			want: `actions must not occur after an ambiguous URL prefix`,
		},
		{
			desc: `conditional URL prefix error in TrustedResourceURLOrURL attribute value sanitization context 8`,
			tmpl: `<source src="{{with 0}}mailto:{{else}}javascript:{{end}}{{ "suffix" }}">`,
			want: `actions must not occur after an ambiguous URL prefix`,
		},
		{
			desc: `error message reports accurate line number`,
			tmpl: `<html>Line 1
Line 2
Line 3
Line 4<script>{{ "this will cause a run-time failure" }}</script>
Line 5
Line 6</html>`,
			want:      `template: error message reports accurate line number:4:17: executing "error message reports accurate line number" at <_sanitizeScript>: error calling _sanitizeScript: expected a safehtml.Script value`,
			fullMatch: true,
		},
		{
			desc: `ends in non-text context 1`,
			tmpl: `<a width=1 title={{"hello"}}`,
			want: `ends in non-text context`,
		},
		{
			desc: `ends in non-text context 2`,
			tmpl: "<script>foo();",
			want: `ends in non-text context`,
		},
		{
			desc: `unquoted static attribute value 1`,
			tmpl: `<input type=button value= 1+1=2>`,
			want: `"=" in unquoted attr: "1+1=2"`,
		},
		{
			desc: `unquoted static attribute value 2`,
			tmpl: "<a class=`foo>",
			want: "\"`\" in unquoted attr: \"`foo\"",
		},
	} {
		tmpl := Must(New(test.desc).Parse(test.tmpl))
		var b bytes.Buffer
		err := tmpl.Execute(&b, test.data)
		if err == nil {
			t.Errorf("%s: expected an error", test.desc)
			return
		}
		got := err.Error()
		if test.fullMatch && got != test.want {
			t.Errorf("%s: got error:\n\t%q\nwant:\n\t%q", test.desc, got, test.want)
			return
		}
		if !test.fullMatch && !strings.Contains(got, test.want) {
			t.Errorf("%s: error\n\t%q\ndoes not contain expected string\n\t%q", test.desc, got, test.want)
		}
	}
}
