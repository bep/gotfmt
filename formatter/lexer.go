package formatter

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
)

var (
	startDelim = []byte("{{")
	endDelim   = []byte("}}")

	isStartKeywordRe = regexp.MustCompile(`{{\s*(define|if|range|with)`)
	isEndKeywordRe   = regexp.MustCompile(`{{\s*end`)
)

const (
	tError itemType = iota
	tEOF
	tAction      // Standalone action.
	tActionStart // Start of: range, with, if
	tActionEnd   // End of block.
	tOther       // HTML etc.
)

func main(l *lexer) stateFunc {
	if l.isEOF() {
		return lexDone
	}

	// Fast forward to next Go template command.
	i := l.index(startDelim)

	if i != -1 {
		l.pos += i
		l.emit(tOther)
		return handleAction
	}

	l.pos = len(l.input)
	return lexDone

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

type itemType int

type lexer struct {
	input []byte
	items []item

	state stateFunc

	pos   int // input position
	start int // item start position
}

func (l *lexer) isEOF() bool {
	return l.pos >= len(l.input)
}

func (l *lexer) emit(t itemType) {
	l.items = append(l.items, item{pos: l.pos, typ: t, val: l.input[l.start:l.pos]})
	l.start = l.pos
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

// Begins with "{{"
func handleAction(l *lexer) stateFunc {
	idxEndDelim := l.index(endDelim)
	if idxEndDelim == -1 {
		return l.errorf("missing closing delimiter")
	}

	skip := idxEndDelim + len(endDelim)

	command := l.input[l.pos : l.pos+skip]

	l.pos += skip

	if isEndKeywordRe.Match(command) {
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
