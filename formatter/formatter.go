package formatter

import (
	"fmt"
	"strings"

	"github.com/yosssi/gohtml"
)

type Formatter struct {
}

type itemPlaceholder struct {
	item        item
	placeholder string
}

type formattingState struct {
	actionCounter int
	placeholders  []itemPlaceholder
}

func (s *formattingState) addPlaceholder(it item, p string) {
	s.placeholders = append(s.placeholders, itemPlaceholder{item: it, placeholder: p})
}

func (s *formattingState) nextAction() int {
	s.actionCounter++
	return s.actionCounter
}

func (f Formatter) Format(input string) (string, error) {

	items, err := parseTemplate([]byte(input))
	if err != nil {
		return "", err
	}

	state := &formattingState{}
	var withPlaceholders strings.Builder

	for _, it := range items {
		var v string
		switch it.typ {
		case tAction:
			v = fmt.Sprintf(`<span gotfmtid%d/>`, state.nextAction())
			state.addPlaceholder(it, v)
		case tActionStart:
			v = fmt.Sprintf(`<div gotfmtid%d>`, state.nextAction())
			state.addPlaceholder(it, v)
		case tActionEnd:
			v = fmt.Sprintf(`</div gotfmtid%d>`, state.nextAction())
			state.addPlaceholder(it, v)
		case tOther:
			v = string(it.val)
		case tEOF:
			if len(it.val) > 0 {
				panic("eof with value")
			}
		default:
			panic(fmt.Sprintf("unsupported type: %s", it.typ))
		}
		withPlaceholders.WriteString(v)
	}

	s := withPlaceholders.String()
	formatted := gohtml.Format(s)

	// Sanity check.
	numPlaceholders := strings.Count(formatted, "gotfmtid")
	if numPlaceholders != len(state.placeholders) {
		//fmt.Println(formatted)
		//fmt.Println(s)
		return input, fmt.Errorf("placeholder mismatch: expected %d, got %d", len(state.placeholders), numPlaceholders)

	}

	oldnew := make([]string, len(state.placeholders)*2)
	i := 0
	for _, p := range state.placeholders {
		oldnew[i] = p.placeholder
		oldnew[i+1] = string(p.item.val)
		i += 2
	}

	replacer := strings.NewReplacer(oldnew...)

	return replacer.Replace(formatted), nil

}
