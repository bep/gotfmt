package formatter

import (
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
			"<div>{{ range .Foo }}<div>{{ . }}</div>{{ end }}</div>\n", `<div>
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
			`<div class='{{ printf "%s" .Foo  }}'>
  Foo
</div>`},
		{
			"Inline 2",
			`<small>v{{ $.Version }}</small>`, `<small>
  v{{ $.Version }}
</small>`},
		{
			"Preserve space above block",
			`<h1>Hugo</h1>
			
{{ range .Foo }}{{ . }}{{ end }}
`, `<h1>
  Hugo
</h1>

{{ range .Foo }}
  {{ . }}
{{ end }}
`},
		{
			"Preserve space below block",
			`{{ range .Foo }}{{ . }}{{ end }}

<h1>Hugo</h1>
`, "{{ range .Foo }}\n  {{ . }}\n{{ end }}\n\n<h1>\n  Hugo\n</h1>\n"},

		{
			"Preserve some space above comment",
			`<h1>Hugo</h1>







			{{/* comment */}}
			{{ range .Foo }}{{ . }}{{ end }}
`, `<h1>
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
			`{{ if .Foo }}
  Foo
{{ else }}
  Bar
{{ end }}`,
		},
		{
			"else inside",
			`{{ define "main" }}{{ if .Foo }}Foo{{ else }}Bar{{ end }}{{ end }}`,
			`{{ define "main" }}
  {{ if .Foo }}
    Foo
  {{ else }}
    Bar
  {{ end }}
{{ end }}`,
		},
		{
			"else if",
			`{{ if .Foo }}Foo{{ else if .Bar }}Bar{{ end }}`,
			`{{ if .Foo }}
  Foo
{{ else if .Bar }}
  Bar
{{ end }}`,
		},
		{
			"with else",
			`{{ with .Foo }}Foo{{else }}Bar{{ end }}`,
			`{{ with .Foo }}
  Foo
{{else }}
  Bar
{{ end }}`,
		},
		{
			"with 1, no newline added",
			`{{ with .Enum }}
Enum:
{{ range . }}
{{ end }}
{{ end }}`,
			`{{ with .Enum }}
  Enum:
  {{ range . }}{{ end }}
{{ end }}`,
		},
		{
			"Template blocks in HTML attribute",
			`<body class="{{ with .Foo }}{{ . }}{{ end }}">`,
			`<body class="{{ with .Foo }}{{ . }}{{ end }}">`,
		},
		{
			"Deeply nested template blocks",
			`{{ range .Boo }}
{{ if .Moo }}
{{ if .FOO }}
{{ end }}
{{ end }}
{{ end }}`,
			`{{ range .Boo }}
  {{ if .Moo }}
    {{ if .FOO }}{{ end }}
  {{ end }}
{{ end }}`,
		},
	} {
		c.Run(test.name, func(c *qt.C) {
			res, err := f.Format(test.input)
			c.Assert(err, qt.IsNil)
			c.Assert(res, qt.Equals, test.output, qt.Commentf("%s", res))
		})
	}
}
