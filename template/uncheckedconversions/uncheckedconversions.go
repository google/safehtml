// Copyright (c) 2017 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

// Package uncheckedconversions provides functions to create values of
// safehtml/template types from plain strings. Use of these
// functions could potentially result in safehtml/template type values that
// violate their type contract, and hence result in security vulnerabilties.
//
package uncheckedconversions

import (
	"github.com/google/safehtml/internal/template/raw"
	"github.com/google/safehtml/template"
)

var trustedSource = raw.TrustedSource.(func(string) template.TrustedSource)
var trustedTemplate = raw.TrustedTemplate.(func(string) template.TrustedTemplate)

// TrustedSourceFromStringKnownToSatisfyTypeContract converts a string into a TrustedSource.
//
func TrustedSourceFromStringKnownToSatisfyTypeContract(s string) template.TrustedSource {
	return trustedSource(s)
}

// TrustedTemplateFromStringKnownToSatisfyTypeContract converts a string into a TrustedTemplate.
//
func TrustedTemplateFromStringKnownToSatisfyTypeContract(s string) template.TrustedTemplate {
	return trustedTemplate(s)
}
