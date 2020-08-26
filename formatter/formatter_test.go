package formatter

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestFormatter(t *testing.T) {
	c := qt.New(t)

	var f Formatter

	for _, test := range []struct {
		name   string
		input  string
		output string
	}{
		{
			"Basic",
			`<div>{{ range .Foo }}<div>{{ . }}</div>{{ end }}</div>`, `
<div>
  {{ range .Foo }}
    <div>
      {{ . }}
    </div>
  {{ end }}
</div>
`},
		{
			"Inline 1",
			`<div class='{{ printf "%s" .Foo  }}'>Foo</div>`,
			`
<div class='{{ printf "%s" .Foo  }}'>
  Foo
</div>
`},
		{
			"Inline 2",
			`<small>v{{ $.Version }}</small>`, `
<small>
  v{{ $.Version }}
</small>
`},
		{
			"Preserve space above block",
			`<h1>Hugo</h1>
			
{{ range .Foo }}{{ . }}{{ end }}
 `, `
<h1>
  Hugo
</h1>

{{ range .Foo }}
  {{ . }}
{{ end }}
`},
		{
			"Preserve some space above comment",
			`<h1>Hugo</h1>







			{{/* comment */}}
			{{ range .Foo }}{{ . }}{{ end }}
 `, `
<h1>
  Hugo
</h1>

{{/* comment */}}
{{ range .Foo }}
  {{ . }}
{{ end }}
`}, {
			"Preserve space in pre",
			`<pre>     {{ range .Foo }}        {{ . }}  {{ end }}    </pre>`,
			`<pre>     {{ range .Foo }}        {{ . }}  {{ end }}    </pre>`,
		},
		{
			"else",
			`{{ if .Foo }}Foo{{ else }}Bar{{ end }}`,
			`
{{ if .Foo }}
  Foo
{{ else }}
  Bar
{{ end }}
`,
		},
		{
			"else inside",
			`{{ define "main" }}{{ if .Foo }}Foo{{ else }}Bar{{ end }}{{ end }}`,
			`
{{ define "main" }}
  {{ if .Foo }}
    Foo
  {{ else }}
    Bar
  {{ end }}
{{ end }}
`,
		},
		{
			"else if",
			`{{ if .Foo }}Foo{{ else if .Bar }}Bar{{ end }}`,
			`
{{ if .Foo }}
  Foo
{{ else if .Bar }}
  Bar
{{ end }}
`,
		},
		{
			"with else",
			`{{ with .Foo }}Foo{{else }}Bar{{ end }}`,
			`
{{ with .Foo }}
  Foo
{{else }}
  Bar
{{ end }}
`,
		},
		{
			"with 1, no newline added",
			`{{ with .Enum }}
Enum:
{{ range . }}
{{ end }}
{{ end }}`,
			`
{{ with .Enum }}
  Enum:
  {{ range . }}{{ end }}
{{ end }}
`,
		},
	} {
		c.Run(test.name, func(c *qt.C) {
			res, err := f.Format(test.input)
			c.Assert(err, qt.IsNil)
			// Make the testdata a little easier to construct.
			res = strings.Trim(res, "\n")
			expect := strings.Trim(test.output, "\n")
			c.Assert(res, qt.Equals, expect, qt.Commentf("%s", res))
		})

	}
}
