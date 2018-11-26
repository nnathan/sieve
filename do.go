package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"sieve/lexer"
)

func main() {
	fn := os.Args[1]

	var in []byte
	var err error

	if fn == "-" {
		in, err = ioutil.ReadAll(os.Stdin)
	} else {
		in, err = ioutil.ReadFile(fn)
	}

	if err != nil {
		fmt.Printf("error reading file %s: %v\n", fn, err)
		os.Exit(1)
	}

	l := lexer.NewLexer(in)

	for {
		pos, tok, val := l.Scan()

		fmt.Printf("%d:%d %s `%s`\n", pos.Line, pos.Column, tok, val)

		if tok == lexer.ItemIllegal || tok == lexer.ItemEOF {
			break
		}
	}
}
