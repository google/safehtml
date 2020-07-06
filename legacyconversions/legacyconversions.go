// Copyright (c) 2017 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

// Package legacyconversions provides functions to create values of package
// safehtml types from plain strings. This package is functionally equivalent
// to package uncheckedconversions, but is only intended for temporary use
// when upgrading code to use package safehtml types.
//
// New code must not use the conversion functions in this package. Instead, new code
// should create package safehtml type values using the functions provided in package
// safehtml or package safehtml/template. If neither of these options are feasible,
// new code should request a security review to use the conversion functions in package
// safehtml/uncheckedconversions instead.
package legacyconversions

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

// RiskilyAssumeHTML converts a plain string into a HTML.
// This function must only be used for refactoring legacy code.
func RiskilyAssumeHTML(s string) safehtml.HTML {
	return html(s)
}

// RiskilyAssumeScript converts a plain string into a Script.
// This function must only be used for refactoring legacy code.
func RiskilyAssumeScript(s string) safehtml.Script {
	return script(s)
}

// RiskilyAssumeStyle converts a plain string into a Style.
// This function must only be used for refactoring legacy code.
func RiskilyAssumeStyle(s string) safehtml.Style {
	return style(s)
}

// RiskilyAssumeStyleSheet converts a plain string into a StyleSheet.
// This function must only be used for refactoring legacy code.
func RiskilyAssumeStyleSheet(s string) safehtml.StyleSheet {
	return styleSheet(s)
}

// RiskilyAssumeURL converts a plain string into a URL.
// This function must only be used for refactoring legacy code.
func RiskilyAssumeURL(s string) safehtml.URL {
	return url(s)
}

// RiskilyAssumeTrustedResourceURL converts a plain string into a TrustedResourceURL.
// This function must only be used for refactoring legacy code.
func RiskilyAssumeTrustedResourceURL(s string) safehtml.TrustedResourceURL {
	return trustedResourceURL(s)
}

// RiskilyAssumeIdentifier converts a plain string into an Identifier.
// This function must only be used for refactoring legacy code.
func RiskilyAssumeIdentifier(s string) safehtml.Identifier {
	return identifier(s)
}
