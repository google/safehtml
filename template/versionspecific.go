// +build !go1.11 appengine

package template

// testNilEmptySliceDataNilWant is the expected value for the "nil" test case in
// TestNilEmptySliceData.
//
// nil is not printed by text/template before Go 1.11.
const testNilEmptySliceDataNilWant = `<b></b>`
