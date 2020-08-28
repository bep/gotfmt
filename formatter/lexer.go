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

	isStartKeywordRe    = regexp.MustCompile(`^{{-?\s*(block|define|if|range|with)`)
	isEndStartKeywordRe = regexp.MustCompile(`^{{-?\s*else`)
	isEndKeywordRe      = regexp.MustCompile(`^{{-?\s*end`)
	isCommentRe         = regexp.MustCompile(`^{{/\*`)
	// There may be some fale positives here, but for this purpose it needs to
	// be a little coarse grained to match "<body {{ with .Type }}{{ . }}{{ end }}>" etc.
	htmlReBase   = `<\/?[a-zA-Z].*\s*\/?>`
	isHTMLTagRe  = regexp.MustCompile("^" + htmlReBase)
	wasHTMLTagRe = regexp.MustCompile(htmlReBase + "$")
)

const (
	tZero itemType = iota
	tError
	tEOF

	tBracketOpen  // HTML opening bracket, '<'.
	tBracketClose // HTML closing bracket, '>'.
	tSpace        // Any whitespace that's not \n.
	tNewline      // Newline (\n).
	tOther        // HTML etc.

	// Types above here are template tokens.
	tAction         // Standalone action.
	tComment        // {{/* Comment */}}.
	tActionStart    // Start of: range, with, if
	tActionEndStart // Start of: else, else if
	tActionEnd      // End of block.

)

var zeroIt item

func main(l *lexer) stateFunc {

	for {
		switch r := l.next(); {
		case r == '{' && l.peek() == '{':
			l.backup()
			if l.pos > l.start {
				l.emit(tOther)
			}
			return lexAction
		case r == '<' && isHTMLTagRe.Match(l.input[l.pos-1:]):
			l.inHTMLElement = true
			l.emit(tBracketOpen)
		case r == '>' && wasHTMLTagRe.Match(l.input[:l.pos]):
			l.inHTMLElement = false
			l.emit(tBracketClose)
		case r == '\n':
			l.emit(tNewline)
		case unicode.IsSpace(r):
			l.emit(tSpace)
		case r == eof:
			return lexDone
		default:
			return lexOther
		}
	}
}

// The runes that we care about for some reason.
func (l *lexer) hasSpecialMeaning(r rune) bool {
	switch r {
	case '{':
		return true
	case '\'', '"':
		return true
	case '>', '<':
		return true
	}
	return unicode.IsSpace(r)
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

func (it item) isTemplateToken() bool {
	return it.typ >= tAction
}

func (it item) shouldPreserveNewlineBefore() bool {
	return it.typ == tComment || it.typ == tActionStart
}

func (it item) shouldPreserveNewlineAfter() bool {
	return it.typ == tActionEnd || it.typ == tComment
}

func (it item) IsZero() bool {
	return it.typ == tZero
}

func (it item) IsEOF() bool {
	return it.typ == tEOF
}

type itemType int

type lexer struct {
	input []byte
	items []item

	state stateFunc

	pos   int // input position
	start int // item start position
	width int // width of last element

	// Content state
	inHTMLElement bool // whether we're inside a HTML element (opening or closing).

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

// Consume until we arrive at one of the characters we care about.
func lexOther(l *lexer) stateFunc {
	skip := bytes.IndexFunc(l.input[l.pos:], func(r rune) bool {
		return l.hasSpecialMeaning(r)
	})

	if skip == -1 {
		l.pos = len(l.input)
		return lexDone
	}

	l.pos += skip
	l.emit(tOther)

	return main
}

// Begins with "{{"
func lexAction(l *lexer) stateFunc {
	idxEndDelim := l.index(endDelim)
	if idxEndDelim == -1 {
		return l.errorf("missing closing delimiter")
	}

	skip := idxEndDelim + len(endDelim)

	command := l.input[l.pos : l.pos+skip]

	l.pos += skip

	if isCommentRe.Match(command) {
		l.emit(tComment)
	} else if l.inHTMLElement {
		l.emit(tAction)
	} else if isEndKeywordRe.Match(command) {
		l.emit(tActionEnd)
	} else if isStartKeywordRe.Match(command) {
		l.emit(tActionStart)
	} else if isEndStartKeywordRe.Match(command) {
		l.emit(tActionEndStart)
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
