package formatter

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"unicode"
	"unicode/utf8"
)

const eof = -1

var (
	endDelim = []byte("}}")

	isStartKeywordRe = regexp.MustCompile(`^{{\s*(define|if|range|with)`)
	isEndKeywordRe   = regexp.MustCompile(`^{{\s*end`)
	isCommentRe      = regexp.MustCompile(`^{{/\*`)
)

const (
	tError itemType = iota
	tEOF

	tAction      // Standalone action.
	tComment     // {{/* Comment */}}.
	tActionStart // Start of: range, with, if
	tActionEnd   // End of block.
	tSpace       // Any whitespace that's not \n.
	tNewline     // Newline (\n).
	tOther       // HTML etc.
)

func main(l *lexer) stateFunc {
	for {
		switch r := l.next(); {
		case r == '{' && l.peek() == '{':
			l.backup()
			if l.pos > l.start {
				l.emit(tOther)
			}
			return handleAction
		case r == '\n':
			l.emit(tNewline)
		case unicode.IsSpace(r):
			l.emit(tSpace)
		case r == eof:
			return lexDone
		}
	}

}

func newLexer(input []byte) *lexer {
	return &lexer{
		input: input,
	}
}

type item struct {
	typ itemType
	pos int
	val []byte
}

func (it item) isWhiteSpace() bool {
	return it.typ == tNewline || it.typ == tSpace
}

type itemType int

type lexer struct {
	input []byte
	items []item

	state stateFunc

	pos   int // input position
	start int // item start position
	width int // width of last element
}

func (l *lexer) emit(t itemType) {
	l.items = append(l.items, item{pos: l.pos, typ: t, val: l.input[l.start:l.pos]})
	l.start = l.pos
}

func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}

	runeValue, runeWidth := utf8.DecodeRune(l.input[l.pos:])
	l.width = runeWidth
	l.pos += l.width
	return runeValue
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) index(b []byte) int {
	return bytes.Index(l.input[l.pos:], b)
}

func (l *lexer) hasPrefix(b []byte) bool {
	return bytes.HasPrefix(l.input[l.pos:], b)
}

func (l *lexer) errorf(format string, args ...interface{}) stateFunc {
	l.items = append(l.items, item{tError, l.start, []byte(fmt.Sprintf(format, args...))})
	// nil terminates the parser
	return nil
}

func (l *lexer) run() *lexer {
	for l.state = main; l.state != nil; {
		l.state = l.state(l)
	}
	return l
}

type stateFunc func(*lexer) stateFunc

// Begins with "{{"
func handleAction(l *lexer) stateFunc {
	idxEndDelim := l.index(endDelim)
	if idxEndDelim == -1 {
		return l.errorf("missing closing delimiter")
	}

	skip := idxEndDelim + len(endDelim)

	command := l.input[l.pos : l.pos+skip]

	l.pos += skip

	if isCommentRe.Match(command) {
		l.emit(tComment)
	} else if isEndKeywordRe.Match(command) {
		l.emit(tActionEnd)
	} else if isStartKeywordRe.Match(command) {
		l.emit(tActionStart)
	} else {
		l.emit(tAction)
	}

	return main
}

func lexDone(l *lexer) stateFunc {
	if l.pos > l.start {
		l.emit(tOther)
	}
	l.emit(tEOF)
	return nil
}

func parseTemplate(input []byte) (items []item, err error) {
	l := newLexer(input)
	l.run()

	for _, item := range l.items {
		items = append(items, item)
		if item.typ == tError {
			err = errors.New(string(item.val))
			break
		}
		if item.typ == tEOF {
			break
		}
	}
	return
}
