// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package template_test

import (
	"log"
	"os"

	"github.com/google/safehtml/template"
)

// Here we demonstrate loading a set of templates from a directory.
func ExampleTemplate_glob() {
	// Here we load three template files with the following contents:
	// 		testdata/glob_t0.tmpl: `T0 invokes T1: ({{template "T1"}})`
	// 		testdata/glob_t1.tmpl: `{{define "T1"}}T1 invokes T2: ({{template "T2"}}){{end}}`
	// 		testdata/glob_t2.tmpl: `{{define "T2"}}This is T2{{end}}`
	// Note that ParseGlob only accepts an untyped string constant.
	// glob_t0.tmpl is the first name matched, so it becomes the starting template,
	// the value returned by ParseGlob.
	tmpl := template.Must(template.ParseGlob("testdata/glob_*.tmpl"))

	err := tmpl.Execute(os.Stdout, nil)
	if err != nil {
		log.Fatalf("template execution: %s", err)
	}
	// Output:
	// T0 invokes T1: (T1 invokes T2: (This is T2))
}

// Here we demonstrate loading a set of templates from files in different directories
func ExampleTemplate_parsefiles() {
	// Here we load two template files from different directories with the following contents:
	// 		testdata/dir1/parsefiles_t1.tmpl: `T1 invokes T2: ({{template "T2"}})`
	// 		testdata/dir2/parsefiles_t2.tmpl: `{{define "T2"}}This is T2{{end}}`
	// Note that ParseFiles only accepts an untyped string constants.
	tmpl := template.Must(template.ParseFiles("testdata/dir1/parsefiles_t1.tmpl", "testdata/dir2/parsefiles_t2.tmpl"))

	err := tmpl.Execute(os.Stdout, nil)
	if err != nil {
		log.Fatalf("template execution: %s", err)
	}
	// Output:
	// T1 invokes T2: (This is T2)
}

// This example demonstrates one way to share some templates
// and use them in different contexts. In this variant we add multiple driver
// templates by hand to an existing bundle of templates.
func ExampleTemplate_helpers() {
	// Here we load the helpers from two template files with the following contents:
	// 		testdata/helpers_t1.tmpl: `{{define "T1"}}T1 invokes T2: ({{template "T2"}}){{end}}`
	// 		testdata/helpers_t2.tmpl: `{{define "T2"}}This is T2{{end}}`
	// Note that ParseGlob only accepts an untyped string constant.
	templates := template.Must(template.ParseGlob("testdata/helpers_*.tmpl"))
	// Add one driver template to the bunch; we do this with an explicit template definition.
	_, err := templates.Parse("{{define `driver1`}}Driver 1 calls T1: ({{template `T1`}})\n{{end}}")
	if err != nil {
		log.Fatal("parsing driver1: ", err)
	}
	// Add another driver template.
	_, err = templates.Parse("{{define `driver2`}}Driver 2 calls T2: ({{template `T2`}})\n{{end}}")
	if err != nil {
		log.Fatal("parsing driver2: ", err)
	}
	// We load all the templates before execution. This package does not require
	// that behavior but html/template's escaping does, so it's a good habit.
	err = templates.ExecuteTemplate(os.Stdout, "driver1", nil)
	if err != nil {
		log.Fatalf("driver1 execution: %s", err)
	}
	err = templates.ExecuteTemplate(os.Stdout, "driver2", nil)
	if err != nil {
		log.Fatalf("driver2 execution: %s", err)
	}
	// Output:
	// Driver 1 calls T1: (T1 invokes T2: (This is T2))
	// Driver 2 calls T2: (This is T2)
}

// This example demonstrates how to use one group of driver
// templates with distinct sets of helper templates.
func ExampleTemplate_share() {
	// Here we load the helpers from two template files with the following contents:
	// 		testdata/share_t0.tmpl: "T0 ({{.}} version) invokes T1: ({{template `T1`}})\n"
	// 		testdata/share_t1.tmpl: `{{define "T1"}}T1 invokes T2: ({{template "T2"}}){{end}}`
	// Note that ParseGlob only accepts an untyped string constant.
	drivers := template.Must(template.ParseGlob("testdata/share_*.tmpl"))

	// We must define an implementation of the T2 template. First we clone
	// the drivers, then add a definition of T2 to the template name space.

	// 1. Clone the helper set to create a new name space from which to run them.
	first, err := drivers.Clone()
	if err != nil {
		log.Fatal("cloning helpers: ", err)
	}
	// 2. Define T2, version A, and parse it.
	_, err = first.Parse("{{define `T2`}}T2, version A{{end}}")
	if err != nil {
		log.Fatal("parsing T2: ", err)
	}

	// Now repeat the whole thing, using a different version of T2.
	// 1. Clone the drivers.
	second, err := drivers.Clone()
	if err != nil {
		log.Fatal("cloning drivers: ", err)
	}
	// 2. Define T2, version B, and parse it.
	_, err = second.Parse("{{define `T2`}}T2, version B{{end}}")
	if err != nil {
		log.Fatal("parsing T2: ", err)
	}

	// Execute the templates in the reverse order to verify the
	// first is unaffected by the second.
	err = second.ExecuteTemplate(os.Stdout, "share_t0.tmpl", "second")
	if err != nil {
		log.Fatalf("second execution: %s", err)
	}
	err = first.ExecuteTemplate(os.Stdout, "share_t0.tmpl", "first")
	if err != nil {
		log.Fatalf("first: execution: %s", err)
	}

	// Output:
	// T0 (second version) invokes T1: (T1 invokes T2: (T2, version B))
	// T0 (first version) invokes T1: (T1 invokes T2: (T2, version A))
}
