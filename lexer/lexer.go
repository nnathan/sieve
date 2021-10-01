package lexer

import (
	"bytes"
)

type Lexer struct {
	src     []byte
	offset  int
	ch      byte
	pos     Position
	nextPos Position
}

type Position struct {
	Line   int
	Column int
}

func NewLexer(src []byte) *Lexer {
	l := &Lexer{src: src}
	l.nextPos.Line = 1
	l.nextPos.Column = 1
	l.next()
	return l
}

// Load the next character into l.ch (or 0 on end of input) and update
// line and column position.
func (l *Lexer) next() {
	l.pos = l.nextPos
	if l.offset >= len(l.src) {
		l.ch = 0
		l.offset++
		return
	}
	ch := l.src[l.offset]
	if ch == '\n' {
		l.nextPos.Line++
		l.nextPos.Column = 1
	} else {
		l.nextPos.Column++
	}
	l.ch = ch
	l.offset++
}

func (l *Lexer) Scan() (pos Position, tok Token, val string) {
	isEOF := func() bool { return (l.offset - 1) >= len(l.src) }

	// Consume whitespace and hash comments: RFC specifies space, tabs, CRLF. Relaxed to include single \n.
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' || l.ch == '\n' || l.ch == '#' || l.ch == '/' {

		if l.ch == '/' {
			l.next()

			if l.ch != '*' {
				return l.pos, ItemIllegal, "expecting '*' to begin bracketed comment"
			}

			l.next()

			for !isEOF() {
				if l.ch == 0 {
					return l.pos, ItemIllegal, "invalid NUL character encountered in bracketed comment"
				}

				if l.ch == '\r' {
					l.next()

					if l.ch != '\n' {
						return l.pos, ItemIllegal, "expected \\n after \\r"
					}
				}

				if l.ch == '*' {
					l.next()
					break
				}

				l.next()
			}

			noSlash := true

			for !isEOF() {
				if l.ch != '*' && l.ch != 0 && l.ch != '\r' {
					if noSlash && l.ch == '/' {
						break
					}

					l.next()
					noSlash = false
					continue
				}

				if l.ch == 0 {
					return l.pos, ItemIllegal, "invalid NUL character encountered in bracketed comment"
				}

				if l.ch == '\r' {
					l.next()

					if l.ch != '\n' {
						return l.pos, ItemIllegal, "expected \\n after \\r"
					}
				}

				if l.ch == '*' {
					noSlash = true
				}

				l.next()
			}

			if isEOF() {

				return l.pos, ItemIllegal, "premature EOF trying to read bracketed comment"
			}

			l.next()

			continue
		}

		// Consume hash comments
		if l.ch == '#' {
			l.next()

			for !isEOF() && l.ch != '\r' && l.ch != '\n' {
				if l.ch == 0 {
					return l.pos, ItemIllegal, "invalid NUL character encountered in bracketed comment"
				}

				l.next()
			}
		}

		// handle case of \r\n
		if l.ch == '\r' {
			l.next()

			if l.ch != '\n' {
				return l.pos, ItemIllegal, "expected \\n after \\r"
			}

			// fallthrough to read '\n'
		}

		l.next()
	}

	if isEOF() {
		// l.next() reached end of input
		return l.pos, ItemEOF, ""
	}

	pos = l.pos
	tok = ItemIllegal
	val = string([]byte{l.ch})

	ch := l.ch
	start := l.offset - 1
	l.next()

	// consume identifier
	if isIdentifierStart(ch) {
		for isIdentifierStart(l.ch) || (l.ch >= '0' && l.ch <= '9') {
			l.next()
		}

		name := string(l.src[start : l.offset-1])

		// handle multiline strings
		if name == "text" && l.ch == ':' {
			l.next()
			tok = ItemMultilineString
			goto multiline
		}

		tok = ItemIdentifier
		val = name

		return pos, tok, val
	}

	// consume number
	if ch >= '0' && ch <= '9' {
		for l.ch >= '0' && l.ch <= '9' {
			l.next()
		}

		if l.ch == 'k' || l.ch == 'K' || l.ch == 'm' || l.ch == 'M' || l.ch == 'g' || l.ch == 'G' {
			l.next()
		}

		tok = ItemNumber
		val = string(l.src[start : l.offset-1])

		return pos, tok, val
	}

	if ch == ':' {
		if isIdentifierStart(l.ch) {
			l.next()
		} else {
			return l.pos, ItemIllegal, "expected identifier character ([a-zA-Z_])"
		}

		for isIdentifierStart(l.ch) || (l.ch >= '0' && l.ch <= '9') {
			l.next()
		}

		tok = ItemTag
		val = string(l.src[start : l.offset-1])

		return pos, tok, val
	}

	if ch == '"' {
		buf := bytes.NewBuffer(nil)

		for !isEOF() && (l.ch == '\\' || l.ch == '\r' || l.ch == '\n' || (l.ch != 0 && l.ch != '"')) {
			if l.ch == '\\' {
				l.next()

				switch {
				case !isEOF() && l.ch != 0 && l.ch != '\r' && l.ch != '\n':
					fallthrough
				case l.ch == '\\':
					fallthrough
				case l.ch == '"':
					buf.WriteByte(l.ch)
					l.next()
				}

				continue
			}

			if l.ch == '\r' {
				buf.WriteByte(l.ch)
				l.next()

				if l.ch != '\n' {
					return l.pos, ItemIllegal, "expected \\n after \\r"
				}
			}

			if l.ch == '\n' {
				buf.WriteByte(l.ch)
				l.next()

				continue
			}

			// at this point we're guaranteed that we're not EOF and l.ch != 0
			start = l.offset - 1
			for !isEOF() && l.ch != '\\' && l.ch != '\r' && l.ch != '\n' && l.ch != '"' && l.ch != 0 {
				l.next()
			}

			buf.Write(l.src[start : l.offset-1])
		}

		if isEOF() {
			return l.pos, ItemIllegal, "premature EOF trying to read string"
		} else if l.ch == 0 {
			return l.pos, ItemIllegal, "invalid NUL character encountered in string"
		}

		// l.ch is guaranteed to be '"'
		l.next()

		tok = ItemString
		val = buf.String()

		return pos, tok, val
	}

	switch ch {
	case '(':
		tok = ItemLeftBracket
	case ')':
		tok = ItemRightBracket
	case '[':
		tok = ItemLeftSquareBracket
	case ']':
		tok = ItemRightSquareBracket
	case '{':
		tok = ItemLeftCurlyBrace
	case '}':
		tok = ItemRightCurlyBrace
	case ';':
		tok = ItemSemicolon
	case ',':
		tok = ItemComma
	}

	return pos, tok, val

multiline:
	// resume from scanning "text:"
	for l.ch == ' ' || l.ch == '\t' {
		l.next()
	}

	if l.ch == '#' {
		l.next()
		for l.ch != '\r' && l.ch != '\n' && !isEOF() {
			l.next()
		}
	}

	if l.ch == '\r' {
		l.next()

		if l.ch != '\n' {
			return l.pos, ItemIllegal, "expected \\n after \\r"
		}
	}

	if l.ch == '\n' {
		l.next()
	}

	if isEOF() {
		return l.pos, ItemIllegal, "premature EOF trying to read multiline string"
	}

	buf := bytes.NewBuffer(nil)

	// loop to detect four possible cases:
	// - a line with a single dot signalling end of multline string
	// - a "dot-stuffed" line, that is, a line with two dots where preceding dot is removed
	// - a line with a single dot followed by octet-not-crlf characters
	// - a line with beginning with a non-dot
	for {
		start = l.offset - 1

		// state used to detect if we encounter only a single dot on a line
		singleDot := false

		if l.ch == '.' {
			l.next()

			singleDot = true

			// handle dot-stuffed line
			if l.ch == '.' {
				l.next()
				singleDot = false
				start++
			}

			// consume octet-not-crlf
			for l.ch != 0 && l.ch != '\r' && l.ch != '\n' {
				l.next()
				singleDot = false
			}
		} else if l.ch != 0 && l.ch != '\r' && l.ch != '\n' {
			l.next()

			for l.ch != 0 && l.ch != '\r' && l.ch != '\n' {
				l.next()
			}
		}

		if isEOF() {
			return l.pos, ItemIllegal, "premature EOF trying to read multiline string"
		} else if l.ch == 0 {
			return l.pos, ItemIllegal, "invalid NUL character encountered in multiline string"
		}

		if l.ch == '\r' {
			l.next()

			if l.ch != '\n' {
				return l.pos, ItemIllegal, "expected \\n after \\r"
			}
		}

		if l.ch == '\n' {
			l.next()
		}

		// encountered a line with a single '.' followed by newline
		if singleDot {
			break
		}

		buf.Write(l.src[start : l.offset-1])
	}

	val = buf.String()

	return pos, tok, val

}

func isIdentifierStart(ch byte) bool {
	return ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}
