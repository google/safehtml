// Copyright (c) 2017 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package safehtml_test

import (
	"testing"

	"github.com/google/safehtml/legacyconversions"
	"github.com/google/safehtml/testconversions"
	"github.com/google/safehtml/uncheckedconversions"
)

const html = `<script>this is not a valid safehtml.HTML`

func TestHTMLFromStringKnownToSatisfyTypeContract(t *testing.T) {
	if out := uncheckedconversions.HTMLFromStringKnownToSatisfyTypeContract(html).String(); html != out {
		t.Errorf("uncheckedconversions.HTMLFromStringKnownToSatisfyTypeContract(%q).String() = %q, want %q",
			html, out, html)
	}
	if out := legacyconversions.RiskilyAssumeHTML(html).String(); html != out {
		t.Errorf("legacyconversions.RiskilyAssumeHTML(%q).String() = %q, want %q", html, out, html)
	}
	if out := testconversions.MakeHTMLForTest(html).String(); html != out {
		t.Errorf("testconversions.MakeHTMLForTest(%q).String() = %q, want %q",
			html, out, html)
	}
}

const script = `</script>this is not a valid safehtml.Script`

func TestScriptFromStringKnownToSatisfyTypeContract(t *testing.T) {
	if out := uncheckedconversions.ScriptFromStringKnownToSatisfyTypeContract(script).String(); script != out {
		t.Errorf("uncheckedconversions.ScriptFromStringKnownToSatisfyTypeContract(%q).String() = %q, want %q",
			script, out, script)
	}
	if out := legacyconversions.RiskilyAssumeScript(script).String(); script != out {
		t.Errorf("legacyconversions.RiskilyAssumeScript(%q).String() = %q, want %q",
			script, out, script)
	}
	if out := testconversions.MakeScriptForTest(script).String(); script != out {
		t.Errorf("testconversions.MakeScriptForTest(%q).String() = %q, want %q",
			script, out, script)
	}
}

const style = `width:expression(this is not valid safehtml.Style`

func TestStyleFromStringKnownToSatisfyTypeContract(t *testing.T) {
	if out := uncheckedconversions.StyleFromStringKnownToSatisfyTypeContract(style).String(); style != out {
		t.Errorf("uncheckedconversions.StyleFromStringKnownToSatisfyTypeContract(%q).String() = %q, want %q",
			style, out, style)
	}
	if out := legacyconversions.RiskilyAssumeStyle(style).String(); style != out {
		t.Errorf("legacyconversions.RiskilyAssumeStyle(%q).String() = %q, want %q",
			style, out, style)
	}
	if out := testconversions.MakeStyleForTest(style).String(); style != out {
		t.Errorf("testconversions.MakeStyleForTest(%q).String() = %q, want %q",
			style, out, style)
	}
}

const styleSheet = `P { text: <not a valid safehtml.StyleSheet> }`

func TestStyleSheetFromStringKnownToSatisfyTypeContract(t *testing.T) {
	if out := uncheckedconversions.StyleSheetFromStringKnownToSatisfyTypeContract(styleSheet).String(); styleSheet != out {
		t.Errorf("uncheckedconversions.StyleSheetFromStringKnownToSatisfyTypeContract(%q).String() = %q, want %q",
			styleSheet, out, styleSheet)
	}
	if out := legacyconversions.RiskilyAssumeStyleSheet(styleSheet).String(); styleSheet != out {
		t.Errorf("legacyconversions.RiskilyAssumeStyleSheet(%q).String() = %q, want %q",
			styleSheet, out, styleSheet)
	}
	if out := testconversions.MakeStyleSheetForTest(styleSheet).String(); styleSheet != out {
		t.Errorf("testconversions.MakeStyleSheetForTest(%q).String() = %q, want %q",
			styleSheet, out, styleSheet)
	}
}

const url = `data:this will not be sanitized`

func TestURLFromStringKnownToSatisfyTypeContract(t *testing.T) {
	if out := uncheckedconversions.URLFromStringKnownToSatisfyTypeContract(url).String(); url != out {
		t.Errorf("uncheckedconversions.URLFromStringKnownToSatisfyTypeContract(%q).String() = %q, want %q",
			url, out, url)
	}
	if out := legacyconversions.RiskilyAssumeURL(url).String(); url != out {
		t.Errorf("legacyconversions.RiskilyAssumeURL(%q).String() = %q, want %q",
			url, out, url)
	}
	if out := testconversions.MakeURLForTest(url).String(); url != out {
		t.Errorf("testconversions.MakeURLForTest(%q).String() = %q, want %q",
			url, out, url)
	}
}

const tru = `data:this will not be sanitized`

func TestTrustedResourceURLFromStringKnownToSatisfyTypeContract(t *testing.T) {
	if out := uncheckedconversions.TrustedResourceURLFromStringKnownToSatisfyTypeContract(tru).String(); tru != out {
		t.Errorf("uncheckedconversions.TrustedResourceURLFromStringKnownToSatisfyTypeContract(%q).String() = %q, want %q",
			tru, out, tru)
	}
	if out := legacyconversions.RiskilyAssumeTrustedResourceURL(tru).String(); tru != out {
		t.Errorf("legacyconversions.RiskilyAssumeTrustedResourceURL(%q).String() = %q, want %q",
			tru, out, tru)
	}
	if out := testconversions.MakeTrustedResourceURLForTest(tru).String(); tru != out {
		t.Errorf("testconversions.MakeTrustedResourceURLForTest(%q).String() = %q, want %q",
			tru, out, tru)
	}
}

const identifier = `1nvalid-identifier-starting-with-a-digit`

func TestIdentifierFromStringKnownToSatisfyTypeContract(t *testing.T) {
	if out := uncheckedconversions.IdentifierFromStringKnownToSatisfyTypeContract(identifier).String(); identifier != out {
		t.Errorf("uncheckedconversions.IdentifierFromStringKnownToSatisfyTypeContract(%q).String() = %q, want %q",
			identifier, out, identifier)
	}
	if out := legacyconversions.RiskilyAssumeIdentifier(identifier).String(); identifier != out {
		t.Errorf("legacyconversions.RiskilyAssumeIdentifier(%q).String() = %q, want %q",
			identifier, out, identifier)
	}
	if out := testconversions.MakeIdentifierForTest(identifier).String(); identifier != out {
		t.Errorf("testconversions.MakeIdentifierForTest(%q).String() = %q, want %q",
			identifier, out, identifier)
	}
}
