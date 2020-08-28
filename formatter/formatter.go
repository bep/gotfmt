package formatter

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/yosssi/gohtml"
)

const (
	placeholderBase    = "gotfmt__id"
	newlinePlaceholder = "<br gotfmt__newline />"
)

// Formatter supports formatting some Go template string.
// We currently only support Go HTML templates.
type Formatter struct {
}

// Format formats input.
func (f Formatter) Format(input string) (string, error) {

	items, err := parseTemplate([]byte(input))
	if err != nil {
		return "", err
	}

	state := &parser{
		pos:   -1,
		items: items,
	}

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

	for {
		it := state.Next()
		if it.IsEOF() {
			break
		}

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
			state.withPlaceholders.WriteString(v)
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
						state.withPlaceholders.WriteString(newlinePlaceholder)
					} else {
						prev, _ := state.prevTemplateToken()
						if prev.shouldPreserveNewlineAfter() {
							state.withPlaceholders.WriteString(newlinePlaceholder)
						}
					}
				}
				state.newlineCounter = 0
			}
		}

		state.withPlaceholders.WriteString(v)
	}

	withPlaceholders := state.withPlaceholders.String()
	formatted := gohtml.Format(withPlaceholders)

	numPlaceholders := strings.Count(formatted, placeholderBase)
	if numPlaceholders != state.numPlaceholders() {
		log.Printf("placeholder mismatch: expected %d, got %d", state.numPlaceholders(), numPlaceholders)
		return input, fmt.Errorf("failed to format, most likely because your HTML is not well formed (check for unclosed divs etc.)")
	}

	oldnew := make([]string, len(state.toReplace)*2)
	i := 0
	for _, p := range state.toReplace {
		oldnew[i] = p.placeholder
		replacement := string(p.item.val)
		oldnew[i+1] = replacement
		i += 2
	}

	// Note that all of these repeated replacements isn't a particulary effective
	// way of doing this,
	// but it assumes relatively small text documents, so it should
	// be plenty fast enough.
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

	if state.hadTrailingNewline() {
		formatted += "\n"
	}

	return formatted, nil

}

type itemPlaceholder struct {
	item        item
	placeholder string
}

type parser struct {
	items []item
	pos   int // current item's index in items

	actionCounter  int // used as ID in placeholders
	newlineCounter int // used to track and preserve some newlines.

	toReplace []itemPlaceholder // temporary placeholders for template tokens.
	toRemove  []string          // used to insert temporary </div> to preserve if/else indentation.

	// This is what we send to the HTML formatter.
	withPlaceholders strings.Builder
}

// Next moves the cursor one step ahead and returns that item.
func (p *parser) Next() item {
	if p.pos < len(p.items)-1 {
		p.pos++
		return p.items[p.pos]
	}
	panic("next called after EOF")
}

func (p *parser) addReplacement(ph itemPlaceholder) {
	p.toReplace = append(p.toReplace, ph)
}

func (p *parser) addTemporary(str string) {
	p.toRemove = append(p.toRemove, str)
}

// Returns whether the input source had a trailing newline.
// We really need to preserve those.
// This is a big thing for editors.
func (p *parser) hadTrailingNewline() bool {
	for i := len(p.items) - 2; i >= 0; i-- {
		it := p.items[i]
		if it.typ == tSpace {
			continue
		}
		if it.typ == tNewline {
			return true
		}
		break
	}

	return false
}

func (p *parser) nextAction() int {
	p.actionCounter++
	return p.actionCounter
}

func (p *parser) numPlaceholders() int {
	return len(p.toReplace) + len(p.toRemove)
}

// prevTemplateToken walks back and looks for the previous template token.
// It also returns number of newlines between.
func (p *parser) prevTemplateToken() (item, int) {
	numNewlines := 0
	for j := p.pos - 1; j >= 0; j-- {
		it := p.items[j]
		if it.isTemplateToken() {
			return it, numNewlines
		}
		if it.typ == tNewline {
			numNewlines++
		}
	}
	return zeroIt, 0

}
