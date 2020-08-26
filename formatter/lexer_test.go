package formatter

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestLexer(t *testing.T) {
	c := qt.New(t)

	assertTypes := func(c *qt.C, got []item, expect ...itemType) {
		c.Helper()
		for i, it := range got {
			c.Assert(it.typ, qt.Equals, expect[i], qt.Commentf("item %d: %s", i+1, it.val))
		}
		c.Assert(got, qt.HasLen, len(expect))
	}

	c.Run("One define", func(c *qt.C) {
		input := []byte(`{{ define "main" }}<div>Main</div>{{ end }}`)
		items, err := parseTemplate(input)
		c.Assert(err, qt.IsNil)
		c.Assert(items, qt.Not(qt.IsNil))
		assertTypes(c, items, tActionStart, tOther, tActionEnd, tEOF)

	})

	c.Run("Two define", func(c *qt.C) {
		input := []byte("{{ define \"main\" }}<div>Main</div>{{ end }}\n\n{{ define \"other\" }}<div>Other</div>{{ end }}")
		items, err := parseTemplate(input)
		c.Assert(err, qt.IsNil)
		c.Assert(items, qt.Not(qt.IsNil))

		assertTypes(c, items,
			tActionStart,
			tOther,
			tActionEnd,
			tNewline, tNewline,
			tActionStart,
			tOther,
			tActionEnd,
			tEOF)
	})

	c.Run("Whitespace", func(c *qt.C) {
		input := []byte(" \n \n{{ range .Foo }}{{ end }}")
		items, err := parseTemplate(input)
		c.Assert(err, qt.IsNil)
		c.Assert(items, qt.Not(qt.IsNil))

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
		input := []byte("{{/* Comment */}}{{ range .Foo }}{{ end }}")
		items, err := parseTemplate(input)
		c.Assert(err, qt.IsNil)
		c.Assert(items, qt.Not(qt.IsNil))

		assertTypes(c, items,
			tComment,
			tActionStart,
			tActionEnd,
			tEOF)
	})

}
