package template_test

import (
	"os"

	"github.com/google/safehtml"
	"github.com/google/safehtml/template"
)

type MyPageData struct {
	Message string
	Script  safehtml.Script
}

var tmplMyPage = template.Must(template.New("myPage").Parse(
	`<strong>{{.Message}}</strong>` +
		// include scripts for page render
		`<script>{{.Script}}</script>`,
))

// Using safehtml.Script to safely inject script content
func Example_script() {
	err := tmplMyPage.Execute(os.Stdout, MyPageData{
		Message: "welcome to my cool website!!",
		Script:  safehtml.ScriptFromConstant(`alert("hello world!")`),
	})

	if err != nil {
		panic(err)
	}

	// Output:
	// <strong>welcome to my cool website!!</strong><script>alert("hello world!")</script>
}
