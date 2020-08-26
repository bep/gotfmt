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

	const (
		placeholderBase    = "gotfmt__id"
		newlinePlaceholder = "gotfmt__newline"
	)
	state := &formattingState{}
	var withPlaceholders strings.Builder
	var numPrecedingNewlines int

	inlinePlaceholder := func() string {
		return fmt.Sprintf("INLINE_%s%d_", placeholderBase, state.nextAction())
	}

	actionPlaceholder := func(end bool) string {
		close := ""
		if end {
			close = "/"
		}
		return fmt.Sprintf("<%sdiv %s%d>", close, placeholderBase, state.nextAction())
	}

	for _, it := range items {
		var v string
		addPlaceholder := func() {
			state.addPlaceholder(
				itemPlaceholder{item: it, placeholder: v},
			)
		}

		switch it.typ {
		case tAction:
			v = inlinePlaceholder()
			addPlaceholder()
		case tComment:
			v = fmt.Sprintf("<!-- COMMENT_%s%d_ -->", placeholderBase, state.nextAction())
			addPlaceholder()
		case tActionStart:
			v = actionPlaceholder(false)
			addPlaceholder()
		case tActionEnd:
			v = actionPlaceholder(true)
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

		// Preserve some intentional whitespace above template blocks and comments.
		switch it.typ {
		case tActionStart, tComment:
			if numPrecedingNewlines > 1 {
				withPlaceholders.WriteString(newlinePlaceholder + "\n")
			}
		}

		if !it.isWhiteSpace() {
			numPrecedingNewlines = 0
		}

		withPlaceholders.WriteString(v)

	}

	s := withPlaceholders.String()

	formatted := gohtml.Format(s)

	// Sanity check.
	numPlaceholders := strings.Count(formatted, placeholderBase)
	if numPlaceholders != len(state.placeholders) {
		return input, fmt.Errorf("placeholder mismatch: expected %d, got %d", len(state.placeholders), numPlaceholders)
	}

	oldnew := make([]string, len(state.placeholders)*2)
	i := 0
	for _, p := range state.placeholders {
		oldnew[i] = p.placeholder
		valStr := string(p.item.val)
		replacement := valStr
		replacement = strings.TrimSpace(replacement)
		if replacement != valStr {
			// TODO(bep) check this fmt.Println("===>", valStr, "=>", replacement, "<")
		}
		oldnew[i+1] = replacement
		i += 2
	}

	formatted = strings.ReplaceAll(formatted, newlinePlaceholder, "")

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
	item        item
	placeholder string
}
