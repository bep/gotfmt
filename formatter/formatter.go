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

const (
	placeholderBase    = "gotfmt__id"
	newlinePlaceholder = "<br gotfmt__newline />"
)

func (f Formatter) Format(input string) (string, error) {

	items, err := parseTemplate([]byte(input))
	if err != nil {
		return "", err
	}

	state := &parser{}

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

	// prevTemplateToken walks back and looks for the previous template token.
	// It also returns number of newlines between.
	prevTemplateToken := func(i int) (item, int) {
		numNewlines := 0
		for j := i - 1; j >= 0; j-- {
			it := items[j]
			if it.isTemplateToken() {
				return it, numNewlines
			}
			if it.typ == tNewline {
				numNewlines++
			}
		}
		return zeroIt, 0
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
			state.newlineCounter++
			v = string(it.val)
		case tOther, tSpace, tBracketOpen, tBracketClose:
			v = string(it.val)
		case tEOF:
			if len(it.val) > 0 {
				panic("eof with value")
			}
		default:
			panic(fmt.Sprintf("unsupported item type: %s", it.typ))
		}

		if it.typ != tEOF && state.newlineCounter > 0 {
			if state.newlineCounter > 0 && !it.isWhiteSpace() {
				if state.newlineCounter > 1 {
					if it.shouldPreserveNewlineBefore() {
						withPlaceholders.WriteString(newlinePlaceholder)
					} else {
						prev, _ := prevTemplateToken(i)
						if prev.shouldPreserveNewlineAfter() {
							withPlaceholders.WriteString(newlinePlaceholder)
						}
					}
				}
				state.newlineCounter = 0
			}
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
		// Remove the entire line if it's on its own.
		re := regexp.MustCompile(fmt.Sprintf(`(?m)\n+^\s*%s\s*$`, s))
		formatted = re.ReplaceAllString(formatted, "")
		// In case it's left on the same line as others.
		formatted = strings.ReplaceAll(formatted, s, "")
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

type parser struct {
	actionCounter  int
	newlineCounter int
	toReplace      []itemPlaceholder
	toRemove       []string
}

func (s *parser) addReplacement(p itemPlaceholder) {
	s.toReplace = append(s.toReplace, p)
}

func (s *parser) numPlaceholders() int {
	return len(s.toReplace) + len(s.toRemove)
}

func (s *parser) addTemporary(str string) {
	s.toRemove = append(s.toRemove, str)
}

func (s *parser) nextAction() int {
	s.actionCounter++
	return s.actionCounter
}

type itemPlaceholder struct {
	item        item
	placeholder string
}
