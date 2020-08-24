package formatter

import (
	"fmt"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestFormatter(t *testing.T) {
	c := qt.New(t)

	f := Formatter{}

	format := func(input string) {
		res, err := f.Format(input)
		c.Assert(err, qt.IsNil)

		fmt.Println(res)
		fmt.Println("-------")
	}

	format(`<div>{{ range .Foo }}<div>{{ . }}</div>{{ end }}</div>`)
	format(`<div class='{{ printf "%s" .Foo  }}'>  <div>inner</div> </div>`)
	format(`{{ define "main" }}<div>Main</div>{{ end }}`)

}
