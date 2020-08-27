package formatter

import (
	"fmt"
	"log"
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

	nextLineItem := func(i, skip int, matches func(item) bool) item {
		skipCount := 0
		for j := i + 1; j < len(items); j++ {
			it := items[j]
			if it.typ == tNewline {
				skipCount++
			} else if matches(it) {
				if skipCount >= skip {
					return it
				}
			} else if !it.isWhiteSpace() {
				return zeroIt
			}
		}
		return zeroIt
	}

	prevLineItem := func(i, skip int, matches func(it item) bool) item {
		skipCount := 0
		for j := i - 1; j >= 0; j-- {
			it := items[j]
			if it.typ == tNewline {
				skipCount++
			} else if matches(it) {
				if skipCount >= skip {
					return it
				}
			} else if !it.isWhiteSpace() {
				return zeroIt
			}
		}
		return zeroIt
	}

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
			prev := prevLineItem(i, 1, preserveNewlineAfter)
			if !prev.IsZero() {
				v = newlinePlaceholder
			} else {
				next := nextLineItem(i, 1, preserveNewlineBefore)
				if !next.IsZero() {
					v = newlinePlaceholder
				} else {
					v = string(it.val)
				}
			}
		case tOther, tSpace, tBracketOpen, tBracketClose, tQuoteStart, tQuoteEnd:
			v = string(it.val)
		case tEOF:
			if len(it.val) > 0 {
				panic("eof with value")
			}
		default:
			panic(fmt.Sprintf("unsupported item type: %s", it.typ))
		}

		withPlaceholders.WriteString(v)

	}

	s := withPlaceholders.String()
	formatted := gohtml.Format(s)

	numPlaceholders := strings.Count(formatted, placeholderBase)
	if numPlaceholders != state.numPlaceholders() {
		log.Printf("placeholder mismatch: expected %d, got %d", state.numPlaceholders(), numPlaceholders)
		return input, fmt.Errorf("failed to format, most likely because your HTML is not well formed (check for unclosed divs etc.)")
	}

	oldnew := make([]string, len(state.toReplace)*2)
	i := 0
	for _, p := range state.toReplace {
		oldnew[i] = p.placeholder
		var replacement string
		replacement = string(p.item.val)
		replacement = strings.TrimSpace(replacement) // TODO(bep) check this, add test.

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

	formatted = replacer.Replace(formatted)

	// We really need to preserve any trailing newline.
	// This is a big thing for editors.
	var hadTrailingNewline bool
	for i := len(items) - 2; i >= 0; i-- {
		it := items[i]
		if it.typ == tSpace {
			continue
		}
		if it.typ == tNewline {
			hadTrailingNewline = true
		}
		break
	}

	if hadTrailingNewline {
		formatted += "\n"
	}

	return formatted, nil

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
