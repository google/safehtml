// Copyright (c) 2017 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package safehtml

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestStyleFromConstantPanic(t *testing.T) {
	for _, test := range [...]struct {
		desc, input, want string
	}{
		{
			desc:  "angle brackets 1",
			input: `width: x<;`,
			want:  `contains angle brackets`,
		},
		{
			desc:  "angle brackets 2",
			input: `width: x>;`,
			want:  `contains angle brackets`,
		},
		{
			desc:  "angle brackets 3",
			input: `</style><script>alert('pwned')</script>`,
			want:  `contains angle brackets`,
		},
		{
			desc:  "no ending semicolon",
			input: `width: 1em`,
			want:  `must end with ';'`,
		},
		{
			desc:  "no colon",
			input: `width= 1em;`,
			want:  `must contain at least one ':' to specify a property-value pair`,
		},
	} {
		tryStyleFromConstant := func() (ret string) {
			defer func() {
				r := recover()
				if r == nil {
					ret = ""
					return
				}
				ret = fmt.Sprint(r)
			}()
			StyleFromConstant(stringConstant(test.input))
			return ""
		}
		errMsg := tryStyleFromConstant()
		if errMsg == "" {
			t.Errorf("%s: expected panic", test.desc)
			continue
		}
		if !strings.Contains(errMsg, test.want) {
			t.Errorf("%s: error message does not contain\n\t%q\ngot:\n\t%q", test.desc, test.want, errMsg)
		}
	}
}

func TestStyleFromProperties(t *testing.T) {
	for _, test := range [...]struct {
		desc  string
		input StyleProperties
		want  string
	}{
		{
			desc: "BackgroundImageURLs single URL",
			input: StyleProperties{
				BackgroundImageURLs: []string{"http://goodUrl.com/a"},
			},
			want: `background-image:url("http://goodUrl.com/a");`,
		},
		{
			desc: "BackgroundImageURLs multiple URLs",
			input: StyleProperties{
				BackgroundImageURLs: []string{"http://goodUrl.com/a", "http://goodUrl.com/b"},
			},
			want: `background-image:url("http://goodUrl.com/a"), url("http://goodUrl.com/b");`,
		},
		{
			desc: "BackgroundImageURLs invalid runes in URL escaped",
			input: StyleProperties{
				BackgroundImageURLs: []string{"http://goodUrl.com/a\"\\\n"},
			},
			want: `background-image:url("http://goodUrl.com/a\000022\00005C\00000A");`,
		},
		{
			desc: "FontFamily unquoted names",
			input: StyleProperties{
				FontFamily: []string{"serif", "sans-serif", "GulimChe"},
			},
			want: `font-family:serif, sans-serif, GulimChe;`,
		},
		{
			desc: "FontFamily quoted names",
			input: StyleProperties{
				FontFamily: []string{"\nserif", "serif\n", "Goudy Bookletter 1911", "New Century Schoolbook", `"sans-serif"`},
			},
			want: `font-family:"\00000Aserif", "serif\00000A", "Goudy Bookletter 1911", "New Century Schoolbook", "sans-serif";`,
		},
		{
			desc: "FontFamily quoted and unquoted names",
			input: StyleProperties{
				FontFamily: []string{"sans-serif", "Goudy Bookletter 1911", "GulimChe", `"fantasy"`, "Times New Roman"},
			},
			want: `font-family:sans-serif, "Goudy Bookletter 1911", GulimChe, "fantasy", "Times New Roman";`,
		},
		{
			desc: "Display",
			input: StyleProperties{
				Display: "inline",
			},
			want: "display:inline;",
		},
		{
			desc: "BackgroundColor",
			input: StyleProperties{
				BackgroundColor: "red",
			},
			want: "background-color:red;",
		},
		{
			desc: "BackgroundPosition",
			input: StyleProperties{
				BackgroundPosition: "100px -110px",
			},
			want: "background-position:100px -110px;",
		},
		{
			desc: "BackgroundRepeat",
			input: StyleProperties{
				BackgroundRepeat: "no-repeat",
			},
			want: "background-repeat:no-repeat;",
		},
		{
			desc: "BackgroundSize",
			input: StyleProperties{
				BackgroundSize: "10px",
			},
			want: "background-size:10px;",
		},
		{
			desc: "Color",
			input: StyleProperties{
				Color: "#000",
			},
			want: "color:#000;",
		},
		{
			desc: "Height",
			input: StyleProperties{
				Height: "100px",
			},
			want: "height:100px;",
		},
		{
			desc: "Width",
			input: StyleProperties{
				Width: "120px",
			},
			want: "width:120px;",
		},
		{
			desc: "Left",
			input: StyleProperties{
				Left: "140px",
			},
			want: "left:140px;",
		},
		{
			desc: "Right",
			input: StyleProperties{
				Right: "160px",
			},
			want: "right:160px;",
		},
		{
			desc: "Top",
			input: StyleProperties{
				Top: "180px",
			},
			want: "top:180px;",
		},
		{
			desc: "Bottom",
			input: StyleProperties{
				Bottom: "200px",
			},
			want: "bottom:200px;",
		},
		{
			desc: "FontWeight",
			input: StyleProperties{
				FontWeight: "100",
			},
			want: "font-weight:100;",
		},
		{
			desc: "Padding",
			input: StyleProperties{
				Padding: "5px 1em 0 2em",
			},
			want: "padding:5px 1em 0 2em;",
		},
		{
			desc: "ZIndex",
			input: StyleProperties{
				ZIndex: "-2",
			},
			want: "z-index:-2;",
		},
		{
			desc: "multiple properties",
			input: StyleProperties{
				BackgroundImageURLs: []string{"http://goodUrl.com/a", "http://goodUrl.com/b"},
				FontFamily:          []string{"serif", "Goudy Bookletter 1911", "Times New Roman", "monospace"},
				BackgroundColor:     "#bbff10",
				BackgroundPosition:  "100px -110px",
				BackgroundRepeat:    "no-repeat",
				BackgroundSize:      "10px",
				Width:               "12px",
				Height:              "10px",
			},
			want: `background-image:url("http://goodUrl.com/a"), url("http://goodUrl.com/b");` +
				`font-family:serif, "Goudy Bookletter 1911", "Times New Roman", monospace;` +
				`background-color:#bbff10;` +
				`background-position:100px -110px;` +
				`background-repeat:no-repeat;` +
				`background-size:10px;` +
				`height:10px;` +
				`width:12px;`,
		},
		{
			desc: "multiple properties, some empty and unset",
			input: StyleProperties{
				BackgroundImageURLs: []string{"http://goodUrl.com/a", "http://goodUrl.com/b"},
				BackgroundPosition:  "100px -110px",
				BackgroundSize:      "",
				Width:               "12px",
				Height:              "10px",
			},
			want: `background-image:url("http://goodUrl.com/a"), url("http://goodUrl.com/b");` +
				`background-position:100px -110px;` +
				`height:10px;` +
				`width:12px;`,
		},
		{
			desc:  "no properties set",
			input: StyleProperties{},
			want:  "",
		},
		{
			desc: "sanitize comment in regular value",
			input: StyleProperties{
				BackgroundRepeat:   "// This is bad",
				BackgroundPosition: "/* This is bad",
				BackgroundSize:     "This is bad */",
			},
			want: "background-position:zGoSafezInvalidPropertyValue;" +
				`background-repeat:zGoSafezInvalidPropertyValue;` +
				`background-size:zGoSafezInvalidPropertyValue;`,
		},
		{
			desc: "sanitize comment in middle of regular value",
			input: StyleProperties{
				BackgroundRepeat:   "10px /* This is bad",
				BackgroundPosition: "10px // This is bad",
				BackgroundSize:     "10px */ This is bad",
			},
			want: "background-position:zGoSafezInvalidPropertyValue;" +
				`background-repeat:zGoSafezInvalidPropertyValue;` +
				`background-size:zGoSafezInvalidPropertyValue;`,
		},
		{
			desc: "sanitize bad rune in regular value",
			input: StyleProperties{
				BackgroundSize: "This&is$bad",
			},
			want: "background-size:zGoSafezInvalidPropertyValue;",
		},
		{
			desc: "sanitize invalid enum value",
			input: StyleProperties{
				Display: "badValue123",
			},
			want: "display:zGoSafezInvalidPropertyValue;",
		},
		{
			desc: "sanitize unsafe URL value",
			input: StyleProperties{
				BackgroundImageURLs: []string{"javascript:badJavascript();"},
			},
			want: `background-image:url("about:invalid#zGoSafez");`,
		},
		{
			desc: "sanitize regular and enum properties with newline prefix",
			input: StyleProperties{
				Display:         "\nfoo",
				BackgroundColor: "\nfoo",
			},
			want: "display:zGoSafezInvalidPropertyValue;background-color:zGoSafezInvalidPropertyValue;",
		},
		{
			desc: "sanitize regular and enum properties with newline suffix",
			input: StyleProperties{
				Display:         "foo\n",
				BackgroundColor: "foo\n",
			},
			want: "display:zGoSafezInvalidPropertyValue;background-color:zGoSafezInvalidPropertyValue;",
		},
		{
			desc: "regular value symbols in value",
			input: StyleProperties{
				BackgroundSize: "*+/-.!#%_ \t",
			},
			want: "background-size:*+/-.!#%_ \t;",
		},
		{
			desc: "quoted and unquoted font family names CSS-escaped",
			input: StyleProperties{
				FontFamily: []string{
					`"`,
					`""`,
					`serif\`,
					`"Gulim\Che"`,
					`"Gulim"Che"`,
					`New Century Schoolbook"`,
					`"New Century Schoolbook`,
					`New Century" Schoolbook`,
					`sans-"serif`,
				},
			},
			want: `font-family:"\000022", ` +
				`"\000022\000022", ` +
				`"serif\00005C", ` +
				`"Gulim\00005CChe", ` +
				`"Gulim\000022Che", ` +
				`"New Century Schoolbook\000022", ` +
				`"\000022New Century Schoolbook", ` +
				`"New Century\000022 Schoolbook", ` +
				`"sans-\000022serif";`,
		},
		{
			desc: "less-than rune CSS-escaped",
			input: StyleProperties{
				BackgroundImageURLs: []string{`</style><script>evil()</script>`},
				FontFamily:          []string{`</style><script>evil()</script>`},
			},
			want: `background-image:url("\00003C/style>\00003Cscript>evil()\00003C/script>");` +
				`font-family:"\00003C/style>\00003Cscript>evil()\00003C/script>";`,
		},
	} {
		got := StyleFromProperties(test.input).String()
		if got != test.want {
			t.Errorf("%s:\ngot:\n\t%s\nwant\n\t%s", test.desc, got, test.want)
		}
	}
}

// TestStyleFromPropertiesAllFieldsValidated will fail if any fields in
// StyleProperties are not safely validated by StyleFromProperties.
//
// This is a sanity check to make sure that all fields are validated and tested.
// If a new field is added but not validated, this test will most likely fail.
func TestStyleFromPropertiesAllFieldsValidated(t *testing.T) {
	// Use reflection to set all fields in StyleProperties.
	var style StyleProperties
	v := reflect.ValueOf(&style).Elem()
	const badValue = `</style><script>evil()</script>`
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		switch f.Type().Kind() {
		case reflect.String:
			f.SetString(badValue)
		case reflect.Slice:
			if f.Type().Elem().Kind() == reflect.String {
				f.Set(reflect.ValueOf([]string{badValue}))
			} else {
				t.Fatalf("unknown slice type for field %q in StyleProperties", v.Type().Field(i).Name)
			}
		default:
			t.Fatalf("unknown %s field %q in StyleProperties", f.Type().Kind(), v.Type().Field(i).Name)
		}
	}
	const want = `background-image:url("\00003C/style>\00003Cscript>evil()\00003C/script>");` +
		`font-family:"\00003C/style>\00003Cscript>evil()\00003C/script>";` +
		`display:zGoSafezInvalidPropertyValue;` +
		`background-color:zGoSafezInvalidPropertyValue;` +
		`background-position:zGoSafezInvalidPropertyValue;` +
		`background-repeat:zGoSafezInvalidPropertyValue;` +
		`background-size:zGoSafezInvalidPropertyValue;` +
		`color:zGoSafezInvalidPropertyValue;` +
		`height:zGoSafezInvalidPropertyValue;` +
		`width:zGoSafezInvalidPropertyValue;` +
		`left:zGoSafezInvalidPropertyValue;` +
		`right:zGoSafezInvalidPropertyValue;` +
		`top:zGoSafezInvalidPropertyValue;` +
		`bottom:zGoSafezInvalidPropertyValue;` +
		`font-weight:zGoSafezInvalidPropertyValue;` +
		`padding:zGoSafezInvalidPropertyValue;` +
		`z-index:zGoSafezInvalidPropertyValue;`
	got := StyleFromProperties(style).String()
	if got != want {
		t.Errorf("got:\n\t%s\nwant\n\t%s", got, want)
	}
}

func TestCSSEscapeString(t *testing.T) {
	for _, test := range [...]struct {
		desc, input, output string
	}{
		{
			desc:   "escape disallowed codepoints in <string-token>",
			input:  "\"\\\n",
			output: `\000022\00005C\00000A`,
		},
		{
			desc:   "escape control characters",
			input:  "\u0001\u001F\u007F\u0080\u0090\u009F\u2028\u2029",
			output: `\000001\00001F\00007F\000080\000090\00009F\002028\002029`,
		},
		{
			desc:   "escape '<'",
			input:  "<",
			output: `\00003C`,
		},
		{
			desc:   "substitute NULL",
			input:  "\u0000",
			output: "\uFFFD",
		},
		{
			desc:   "no escaping required",
			input:  `this(can_BE$s4fely:Quoted`,
			output: `this(can_BE$s4fely:Quoted`,
		},
	} {
		escaped := cssEscapeString(test.input)
		if escaped != test.output {
			t.Errorf("%s:\ngot:\n\t%s\nwant\n\t%s", test.desc, escaped, test.output)
		}
	}
}
