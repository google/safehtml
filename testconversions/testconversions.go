// Copyright (c) 2017 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

// Package testconversions provides functions to to create arbitrary values of
// package safehtml types for use by tests only. Note that the created values may
// violate type contracts.
//
// These functions are useful when types are constructed in a manner where using
// the package safehtml API is too inconvenient. Please use the package safehtml
// API whenever possible; there is value in having tests reflect common usage.
// Using the package safehtml API also avoids, by design, non-contract complying
// instances from being created.
package testconversions

import (
	"github.com/google/safehtml/internal/raw"
	"github.com/google/safehtml"
)

var html = raw.HTML.(func(string) safehtml.HTML)
var script = raw.Script.(func(string) safehtml.Script)
var style = raw.Style.(func(string) safehtml.Style)
var styleSheet = raw.StyleSheet.(func(string) safehtml.StyleSheet)
var url = raw.URL.(func(string) safehtml.URL)
var trustedResourceURL = raw.TrustedResourceURL.(func(string) safehtml.TrustedResourceURL)
var identifier = raw.Identifier.(func(string) safehtml.Identifier)

// MakeHTMLForTest converts a plain string into a HTML.
// This function must only be used in tests.
func MakeHTMLForTest(s string) safehtml.HTML {
	return html(s)
}

// MakeScriptForTest converts a plain string into a Script.
// This function must only be used in tests.
func MakeScriptForTest(s string) safehtml.Script {
	return script(s)
}

// MakeStyleForTest converts a plain string into a Style.
// This function must only be used in tests.
func MakeStyleForTest(s string) safehtml.Style {
	return style(s)
}

// MakeStyleSheetForTest converts a plain string into a StyleSheet.
// This function must only be used in tests.
func MakeStyleSheetForTest(s string) safehtml.StyleSheet {
	return styleSheet(s)
}

// MakeURLForTest converts a plain string into a URL.
// This function must only be used in tests.
func MakeURLForTest(s string) safehtml.URL {
	return url(s)
}

// MakeTrustedResourceURLForTest converts a plain string into a TrustedResourceURL.
// This function must only be used in tests.
func MakeTrustedResourceURLForTest(s string) safehtml.TrustedResourceURL {
	return trustedResourceURL(s)
}

// MakeIdentifierForTest converts a plain string into an Identifier.
// This function must only be used in tests.
func MakeIdentifierForTest(s string) safehtml.Identifier {
	return identifier(s)
}
