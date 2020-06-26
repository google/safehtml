// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package template_test

import (
	"log"
	"os"
	"strings"

	"github.com/google/safehtml/template"
)

func Example() {
	const tpl = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>{{.Title}}</title>
	</head>
	<body>
		{{range .Items}}<div>{{ . }}</div>{{else}}<div><strong>no rows</strong></div>{{end}}
	</body>
</html>`

	check := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}
	t, err := template.New("webpage").Parse(tpl)
	check(err)

	data := struct {
		Title string
		Items []string
	}{
		Title: "My page",
		Items: []string{
			"My photos",
			"My blog",
		},
	}

	err = t.Execute(os.Stdout, data)
	check(err)

	noItems := struct {
		Title string
		Items []string
	}{
		Title: "My another page",
		Items: []string{},
	}

	err = t.Execute(os.Stdout, noItems)
	check(err)

	// Output:
	// <!DOCTYPE html>
	// <html>
	// 	<head>
	// 		<meta charset="UTF-8">
	// 		<title>My page</title>
	// 	</head>
	// 	<body>
	// 		<div>My photos</div><div>My blog</div>
	// 	</body>
	// </html>
	// <!DOCTYPE html>
	// <html>
	// 	<head>
	// 		<meta charset="UTF-8">
	// 		<title>My another page</title>
	// 	</head>
	// 	<body>
	// 		<div><strong>no rows</strong></div>
	// 	</body>
	// </html>

}

func Example_autosanitization() {
	t := template.Must(template.New("foo").Parse(`<a href="{{ .X }}">{{ .Y }}</a>`))
	renderHTML := func(x, y string) {
		if err := t.Execute(os.Stdout, struct{ X, Y string }{x, y}); err != nil {
			log.Fatal(err)
		}
	}
	renderHTML("javascript:evil()", "</a><script>alert('pwned')</script><a>")
	// Output:
	// <a href="about:invalid#zGoSafez">&lt;/a&gt;&lt;script&gt;alert(&#39;pwned&#39;)&lt;/script&gt;&lt;a&gt;</a>
}

// The following example is duplicated in html/template; keep them in sync.

func ExampleTemplate_block() {
	const (
		master  = `Names:{{block "list" .}}{{"\n"}}{{range .}}{{println "-" .}}{{end}}{{end}}`
		overlay = `{{define "list"}} {{join . ", "}}{{end}} `
	)
	var (
		funcs     = template.FuncMap{"join": strings.Join}
		guardians = []string{"Gamora", "Groot", "Nebula", "Rocket", "Star-Lord"}
	)
	masterTmpl, err := template.New("master").Funcs(funcs).Parse(master)
	if err != nil {
		log.Fatal(err)
	}
	overlayTmpl, err := template.Must(masterTmpl.Clone()).Parse(overlay)
	if err != nil {
		log.Fatal(err)
	}
	if err := masterTmpl.Execute(os.Stdout, guardians); err != nil {
		log.Fatal(err)
	}
	if err := overlayTmpl.Execute(os.Stdout, guardians); err != nil {
		log.Fatal(err)
	}
	// Output:
	// Names:
	// - Gamora
	// - Groot
	// - Nebula
	// - Rocket
	// - Star-Lord
	// Names: Gamora, Groot, Nebula, Rocket, Star-Lord
}
