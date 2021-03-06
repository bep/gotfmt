package formatter

import (
	"fmt"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestLexer(t *testing.T) {
	c := qt.New(t)

	parse := func(c *qt.C, input string) []item {
		items, err := parseTemplate([]byte(input))
		c.Assert(err, qt.IsNil)
		c.Assert(items, qt.Not(qt.IsNil))
		return items
	}

	assertTypes := func(c *qt.C, got []item, expect ...itemType) {
		c.Helper()
		var expectStr string
		var gotStr string
		for _, it := range expect {
			expectStr += fmt.Sprint(it) + ","
		}
		for _, it := range got {
			gotStr += fmt.Sprint(it.typ) + ","
		}
		c.Assert(gotStr, qt.Equals, expectStr)

	}

	c.Run("One define", func(c *qt.C) {
		items := parse(c, `{{ define "main" }}<div>Main</div>{{ end }}`)
		assertTypes(c, items, tActionStart, tBracketOpen, tOther, tBracketClose, tOther,
			tBracketOpen, tOther, tBracketClose, tActionEnd, tEOF)
	})

	c.Run("Two define", func(c *qt.C) {
		items := parse(c, "{{ define \"main\" }}<div>Main</div>{{ end }}\n\n{{ define \"other\" }}<div>Other</div>{{ end }}")

		assertTypes(c, items,
			tActionStart, tBracketOpen, tOther, tBracketClose, tOther, tBracketOpen, tOther, tBracketClose, tActionEnd,
			tNewline, tNewline,
			tActionStart, tBracketOpen, tOther, tBracketClose, tOther, tBracketOpen, tOther, tBracketClose, tActionEnd, tEOF,
		)
	})

	c.Run("Whitespace", func(c *qt.C) {
		items := parse(c, " \n \n{{ range .Foo }}{{ end }}")

		assertTypes(c, items,
			tSpace,
			tNewline,
			tSpace,
			tNewline,
			tActionStart,
			tActionEnd,
			tEOF)
	})

	c.Run("Template comment", func(c *qt.C) {
		items := parse(c, "{{/* Comment */}}{{ range .Foo }}{{ end }}")

		assertTypes(c, items,
			tComment,
			tActionStart,
			tActionEnd,
			tEOF)
	})

	c.Run("else", func(c *qt.C) {
		items := parse(c, "{{ if .Foo }}FOO{{ else }}BAR{{ end }}")

		assertTypes(c, items,
			tActionStart,
			tOther,
			tActionEndStart,
			tOther,
			tActionEnd,
			tEOF)
	})

	c.Run("else if", func(c *qt.C) {
		items := parse(c, "{{ if .Foo }}FOO{{ else if .Bar }}BAR{{ end }}")

		assertTypes(c, items,
			tActionStart,
			tOther,
			tActionEndStart,
			tOther,
			tActionEnd,
			tEOF)
	})

	c.Run("nested blocks", func(c *qt.C) {
		items := parse(c, `{{ with .Enum }}
Enum:
{{ range . }}
{{ end }}
{{ end }}`)

		assertTypes(c, items,
			tActionStart,
			tNewline,
			tOther,
			tNewline,
			tActionStart,
			tNewline,
			tActionEnd,
			tNewline,
			tActionEnd,
			tEOF)
	})

	c.Run("Template blocks inside HTML attribute", func(c *qt.C) {
		items := parse(c, `<body {{ with .Type }}{{ . }}{{ end }}>`)

		assertTypes(c, items,
			tBracketOpen, tOther, tSpace, tAction, tAction, tAction, tBracketClose, tEOF,
		)
	})

	c.Run("Brackets in text", func(c *qt.C) {
		items := parse(c, `32 < 52 a > b`)

		assertTypes(c, items,
			tOther, tSpace, tOther, tSpace, tOther, tSpace, tOther, tSpace, tOther, tSpace, tOther, tEOF,
		)
	})

	c.Run("Template commands with trim markers", func(c *qt.C) {
		items := parse(c, `{{- with .Type .}}{{ . }}{{- else -}}{{ . }}{{- end -}}`)

		assertTypes(c, items,
			tActionStart, tAction, tActionEndStart, tAction, tActionEnd, tEOF,
		)
	})

	c.Run("block keyword", func(c *qt.C) {
		items := parse(c, `{{ block "title" . }}Title{{ end }}`)

		assertTypes(c, items,
			tActionStart, tOther, tActionEnd, tEOF,
		)
	})

}
