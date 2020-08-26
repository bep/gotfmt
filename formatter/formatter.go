package formatter

import (
	"fmt"
	"strings"

	"github.com/yosssi/gohtml"
)

type Formatter struct {
}

func (f Formatter) Format(input string) (string, error) {

	items, err := parseTemplate([]byte(input))
	if err != nil {
		return "", err
	}

	const placeholderBase = "gotfmtid"
	state := &formattingState{}
	var withPlaceholders strings.Builder
	var numPrecedingNewlines int

	for _, it := range items {
		var v string
		addPlaceholder := func() {
			state.addPlaceholder(
				itemPlaceholder{item: it, placeholder: v, numPrecedingNewlines: numPrecedingNewlines},
			)
		}

		switch it.typ {
		case tAction:
			v = fmt.Sprintf("INLINE_%s%d_", placeholderBase, state.nextAction())
			addPlaceholder()
		case tActionStart:
			v = fmt.Sprintf("<div %s%d>", placeholderBase, state.nextAction())
			addPlaceholder()
		case tActionEnd:
			v = fmt.Sprintf("</div %s%d>", placeholderBase, state.nextAction())
			addPlaceholder()
		case tNewline:
			v = string(it.val)
			numPrecedingNewlines++
		case tOther, tSpace:
			v = string(it.val)
		case tEOF:
			if len(it.val) > 0 {
				panic("eof with value")
			}
		default:
			panic(fmt.Sprintf("unsupported type: %s", it.typ))
		}

		if it.typ != tNewline && it.typ != tSpace {
			numPrecedingNewlines = 0
		}

		withPlaceholders.WriteString(v)

	}

	s := withPlaceholders.String()

	formatted := gohtml.Format(s)

	// Sanity check.
	numPlaceholders := strings.Count(formatted, placeholderBase)
	if numPlaceholders != len(state.placeholders) {
		fmt.Println(s)
		for i := 1; i <= len(state.placeholders); i++ {
			pid := fmt.Sprintf("%s%d", placeholderBase, i)
			if !(strings.Contains(formatted, pid)) {
				fmt.Println(pid, " missing.")
			}
		}
		return input, fmt.Errorf("placeholder mismatch: expected %d, got %d", len(state.placeholders), numPlaceholders)
	}

	oldnew := make([]string, len(state.placeholders)*2)
	i := 0
	for _, p := range state.placeholders {
		oldnew[i] = p.placeholder
		replacement := string(p.item.val)
		if p.numPrecedingNewlines > 1 {
			replacement = "\n" + replacement
		}
		if p.item.typ == tActionStart {
			replacement = replacement
		}
		oldnew[i+1] = replacement
		i += 2
	}

	replacer := strings.NewReplacer(oldnew...)

	return replacer.Replace(formatted), nil

}

type formattingState struct {
	actionCounter int
	placeholders  []itemPlaceholder
}

func (s *formattingState) addPlaceholder(p itemPlaceholder) {
	s.placeholders = append(s.placeholders, p)
}

func (s *formattingState) nextAction() int {
	s.actionCounter++
	return s.actionCounter
}

type itemPlaceholder struct {
	item                 item
	placeholder          string
	numPrecedingNewlines int
}
