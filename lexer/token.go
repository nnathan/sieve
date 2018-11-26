package lexer

// Token is the type of a single token.
type Token int

const (
	ItemIllegal Token = iota
	ItemEOF
	ItemMultilineString
	ItemIdentifier
	ItemNumber
	ItemTag
	ItemString
	ItemLeftBracket
	ItemRightBracket
	ItemLeftSquareBracket
	ItemRightSquareBracket
	ItemLeftCurlyBrace
	ItemRightCurlyBrace
	ItemSemicolon
	ItemComma
	ItemLast
)

var tokenNames = map[Token]string{
	ItemIllegal:            "<illegal>",
	ItemEOF:                "EOF",
	ItemMultilineString:    "multline",
	ItemIdentifier:         "identifier",
	ItemNumber:             "number",
	ItemTag:                "tag",
	ItemString:             "string",
	ItemLeftBracket:        "(",
	ItemRightBracket:       ")",
	ItemLeftSquareBracket:  "[",
	ItemRightSquareBracket: "]",
	ItemLeftCurlyBrace:     "{",
	ItemRightCurlyBrace:    "}",
	ItemSemicolon:          ";",
	ItemComma:              ",",
}

// String returns string name of the token.
func (t Token) String() string {
	return tokenNames[t]
}
