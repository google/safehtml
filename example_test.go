// Copyright (c) 2017 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package safehtml_test

import (
	"bytes"
	"testing"

	"github.com/google/safehtml"
	"github.com/google/safehtml/template"
)

func TestExampleScriptFromDataAndConstant(t *testing.T) {
	const script = `alert(msg['Greeting'] + ' ' + msg['Names'] + '! The year is ' + msg['Year'])`
	type WelcomeMessage struct {
		Greeting string
		Names    []string
		Year     int
	}
	data := WelcomeMessage{
		Greeting: "Hello",
		Names:    []string{"Alice", "Bob"},
		Year:     3055,
	}
	s, err := safehtml.ScriptFromDataAndConstant("msg", data, script)
	if err != nil {
		t.Fatalf("while building script from data: %v", err)
	}
	t.Log(s)
	// Output:
	// var msg = {"Greeting":"Hello","Names":["Alice","Bob"],"Year":3055};
	// alert(msg['Greeting'] + ' ' + msg['Names'] + '! The year is ' + msg['Year'])
}

// ScriptFromDataAndConstant can be used to pass dynamic data from
// a Go program to inline scripts in a safehtml/template Template.
func TestExampleScriptFromDataAndConstant_safeHTMLTemplate(t *testing.T) {
	tmpl := template.Must(template.New("").Parse(`
		<!DOCTYPE html>
		<html><body>
			{{ if .IsIntro }}
				<h1>Welcome!</h1>
			{{ else }}
				<h1>Nice to see you again!</h1>
			{{ end }}
			<ul>
			{{ range .GoodPoints }}
				<li>{{ . }}</li>
			{{ end }}
			</ul>
			<script>{{ .Script }}</script>
		</body></html>`))
	data := struct {
		Name string
		ID   int
	}{
		getName(),
		getID(),
	}
	script, err := safehtml.ScriptFromDataAndConstant(
		"myArgs", data, `  my.functionCall(myArgs[‘Name’], myArgs[‘ID’])`)
	if err != nil {
		t.Fatalf("while building script from data: %v", err)
	}
	out := &bytes.Buffer{}
	if err := tmpl.Execute(out, struct {
		IsIntro    bool
		GoodPoints []string
		Script     safehtml.Script
	}{false, []string{"foo", "bar"}, script}); err != nil {
		t.Fatalf("while rendering template: %v", err)
	}
	t.Log(out)
	// Output:
	// <!DOCTYPE html>
	// <html><body>
	//	 ...
	//   <script>
	//   var myArgs = {"Name":"Sam","ID":14};
	//   my.functionCall(myArgs[‘Name’], myArgs[‘ID’])
	//   </script>
	// </body></html>
}

func getName() string {
	return "Sam"
}

func getID() int {
	return 14
}

func TestExampleTrustedResourceURLFormat(t *testing.T) {
	tru, err := safehtml.TrustedResourceURLFormatFromConstant(`//www.youtube.com/v/%{id}?hl=%{lang}`, map[string]string{
		"id":   "abc0def1",
		"lang": "en",
	})
	if err != nil {
		t.Fatalf("while building URL: %v", err)
	}
	t.Log(tru)
	// Output:
	// //www.youtube.com/v/abc0def1?hl=en
}
