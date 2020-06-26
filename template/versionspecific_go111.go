// +build go1.11,!appengine

package template

// testNilEmptySliceDataNilWant is the expected value for the "nil" test case in
// TestNilEmptySliceData.
//
// nil is printed as `<nil>` by text/template in Go 1.11, which then gets HTML-escaped
// by safehtml/template to `&lt;nil&gt;`.
const testNilEmptySliceDataNilWant = `<b>&lt;nil&gt;</b>`
