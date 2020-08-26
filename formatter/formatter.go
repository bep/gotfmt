package formatter

import (
	"fmt"
	"regexp"
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

	nextNonWhiteSpace := func(i int) (it item) {
		for j := i; j < len(items); j++ {
			it = items[j]
			if !it.isWhiteSpace() {
				return
			}
		}
		return
	}

	var newlineCounter int
	var newlineInserted bool

	for i, it := range items {
		var v string
		addPlaceholder := func() {
			state.addReplacement(
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
		case tActionEndStart:
			v = fmt.Sprintf("</div %s%d>", placeholderBase, state.nextAction())
			state.addTemporary(v)
			withPlaceholders.WriteString(v)
			v = fmt.Sprintf("<div %s%d>", placeholderBase, state.nextAction())
			addPlaceholder()
		case tNewline:
			newlineCounter++
			var next item
			if !newlineInserted && newlineCounter > 1 {
				next = nextNonWhiteSpace(i)
			}
			if next.preserveNewlineBefore() {
				// Preserve a newline before template comments and blocks.
				v = newlinePlaceholder
				newlineInserted = true
			} else {
				v = string(it.val)
			}
		case tOther, tSpace:
			v = string(it.val)
		case tEOF:
			if len(it.val) > 0 {
				panic("eof with value")
			}
		default:
			panic(fmt.Sprintf("unsupported type: %s", it.typ))
		}

		if !it.isWhiteSpace() {
			// Reset state for next action/comment.
			newlineInserted = false
			newlineCounter = 0
		}

		withPlaceholders.WriteString(v)

	}

	s := withPlaceholders.String()
	formatted := gohtml.Format(s)

	//fmt.Println(formatted)

	// Sanity check.
	numPlaceholders := strings.Count(formatted, placeholderBase)
	if numPlaceholders != state.numPlaceholders() {
		return input, fmt.Errorf("placeholder mismatch: expected %d, got %d", state.numPlaceholders(), numPlaceholders)
	}

	oldnew := make([]string, len(state.toReplace)*2)
	i := 0
	for _, p := range state.toReplace {
		oldnew[i] = p.placeholder
		var replacement string
		replacement = string(p.item.val)
		replacement = strings.TrimSpace(replacement) // TODO(bep) check this

		oldnew[i+1] = replacement
		i += 2
	}

	formatted = strings.ReplaceAll(formatted, newlinePlaceholder, "")

	for _, s := range state.toRemove {
		// Remove the entire line
		re := regexp.MustCompile(fmt.Sprintf(`(?m)\n+^.*%s.*$`, s))
		formatted = re.ReplaceAllString(formatted, "")
	}

	replacer := strings.NewReplacer(oldnew...)

	return replacer.Replace(formatted), nil

}

type formattingState struct {
	actionCounter int
	toReplace     []itemPlaceholder
	toRemove      []string
}

func (s *formattingState) addReplacement(p itemPlaceholder) {
	s.toReplace = append(s.toReplace, p)
}

func (s *formattingState) numPlaceholders() int {
	return len(s.toReplace) + len(s.toRemove)
}

func (s *formattingState) addTemporary(str string) {
	s.toRemove = append(s.toRemove, str)
}

func (s *formattingState) nextAction() int {
	s.actionCounter++
	return s.actionCounter
}

type itemPlaceholder struct {
	item        item
	placeholder string
}
