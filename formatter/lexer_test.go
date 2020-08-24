package formatter

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestLexer(t *testing.T) {
	c := qt.New(t)

	input := []byte(`
		<div>{{ range .Foo }}<div>{{ . }}</div>{{ end }}</div>
	`)

	assertTypes := func(got []item, expect ...itemType) {
		c.Assert(got, qt.HasLen, len(expect))
		for i, it := range got {
			c.Assert(it.typ, qt.Equals, expect[i])
		}
	}

	items, err := parseTemplate(input)
	c.Assert(err, qt.IsNil)
	c.Assert(items, qt.Not(qt.IsNil))
	assertTypes(items, tOther, tActionStart, tOther, tAction, tOther, tActionEnd, tOther, tEOF)
}
