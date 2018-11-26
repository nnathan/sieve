package lexer_test

import (
	"fmt"
	"strings"
	"testing"

	"sieve/lexer"
)

func TestLexer(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// whitespace, hash comments, line endings
		{"", ""},
		{"\r", `1:2 <illegal> "expected \\n after \\r"`},
		{"#foo\xff\x00", `1:6 <illegal> "invalid NUL character encountered in bracketed comment"`},
		{"#\r", `1:3 <illegal> "expected \\n after \\r"`},
		{" \t#foo\xff\n\r\n", ""},

		// identifiers and numbers
		{"foo", `1:1 identifier "foo"`},
		{"fo\xff", `1:1 identifier "fo", 1:3 <illegal> "\xff"`},
		{"100", `1:1 number "100"`},
		{"100\xff", `1:1 number "100", 1:4 <illegal> "\xff"`},
		{"100m", `1:1 number "100m"`},

		// multline strings
		{"text:\n.\n", `1:1 multline ""`},
		{"text: \t#foo\x00\n.\n", `1:1 multline ""`},
		{"text:\r\nfoo\nbar\r\n.\n", `1:1 multline "foo\nbar\r\n"`},
		{"text:", `1:6 <illegal> "premature EOF trying to read multiline string"`},
		{"text:\r", `1:7 <illegal> "expected \\n after \\r"`},
		{"text:#\r", `1:8 <illegal> "expected \\n after \\r"`},
		{"text:#", `1:7 <illegal> "premature EOF trying to read multiline string"`},
		{"text:\n..\n...\n.foo\n..foo\n...foo\n.\n", `1:1 multline ".\n..\n.foo\n.foo\n..foo\n"`},

		// quoted strings
		{`""`, `1:1 string ""`},
		{"\"x\xff\"", `1:1 string "x\xff"`},
		{"\"x\r\n\"", `1:1 string "x\r\n"`},
		{"\"x\r\"", `1:4 <illegal> "expected \\n after \\r"`},
		{"\"\x00", `1:2 <illegal> "invalid NUL character encountered in string"`},
		{"\"", `1:2 <illegal> "premature EOF trying to read string"`},

		// bracketed comments
		{"/**/", ""},
		{"/*foo\r\nbar\nbaz*/", ""},
		{"/* /* **/", ""},
		{"/*foo*/ /*baz*/", ""},
		{"/*foo\rbar*/", `1:7 <illegal> "expected \\n after \\r"`},
		{"/*foo\x00\xff*/", `1:6 <illegal> "invalid NUL character encountered in bracketed comment"`},
		{"/* */ */", `1:7 <illegal> "*"`},

		// invalid characters
		{"%", `1:1 <illegal> "%"`},
	}

	for _, tt := range tests {
		l := lexer.NewLexer([]byte(tt.input))

		ss := []string{}

		for {
			pos, tok, val := l.Scan()

			if tok == lexer.ItemEOF {
				break
			}

			ss = append(ss, fmt.Sprintf("%d:%d %s %q", pos.Line, pos.Column, tok, val))

			if tok == lexer.ItemIllegal {
				break
			}
		}

		actual := strings.Join(ss, ", ")

		if actual != tt.expected {
			t.Errorf("got %q expected %q", actual, tt.expected)
		}

	}
}

func TestAllTokens(t *testing.T) {
	input := " \t /**/ # comment\n" +
		"text:\nfoo\n.\n foo 100g :foo \"foo\"" +
		"( ) [ ] { } ; , %"

	expected := "multline identifier number tag string ( ) [ ] { } ; , <illegal> EOF"

	ss := make([]string, 0, lexer.ItemLast)
	seen := make([]bool, lexer.ItemLast)

	l := lexer.NewLexer([]byte(input))

	for {
		_, tok, _ := l.Scan()
		ss = append(ss, tok.String())
		seen[int(tok)] = true

		if tok == lexer.ItemEOF {
			break
		}
	}

	actual := strings.Join(ss, " ")

	if actual != expected {
		t.Errorf("got %q expected %q", actual, expected)
	}

	for i, s := range seen {
		if !s {
			t.Errorf("token %s (%d) not seen", lexer.Token(i), i)
		}
	}
}
