package php

import (
	_ "fmt"
	"io"
	"text/scanner"
)

type TokenType string

const (
	TOKEN_EOF        TokenType = "EOF"
	TOKEN_IDENTIFIER TokenType = "IDENTIFIER"
	TOKEN_NUMBER     TokenType = "NUMBER"
	TOKEN_KEYWORD    TokenType = "KEYWORD"
	TOKEN_SYMBOL     TokenType = "SYMBOL"
	TOKEN_STRING1    TokenType = "STRING1"
	TOKEN_STRING2    TokenType = "STRING2"
)

var keywords = map[string]bool{
	"if": true, "else": true, "while": true, "function": true,
	"return": true, "echo": true, "class": true, "public": true,
}

type Token struct {
	Type      TokenType
	Value     string
	Lnum, Col int
}

func NewToken(typ TokenType, value string, pos scanner.Position) *Token {
	return &Token{Type: typ, Value: value, Lnum: pos.Line, Col: pos.Column}
}

type Lexer struct {
	scanner scanner.Scanner
}

func NewLexer(reader io.Reader) *Lexer {
	var s scanner.Scanner
	s.Init(reader)
	s.Mode = scanner.ScanIdents | scanner.ScanInts | scanner.ScanFloats | scanner.ScanStrings | scanner.ScanChars | scanner.ScanRawStrings | scanner.ScanComments

	s.Mode = scanner.ScanIdents | scanner.ScanInts | scanner.ScanComments

	return &Lexer{scanner: s}
}

func (l *Lexer) NextToken() *Token {
	for {
		tok := l.scanner.Scan()
		if tok == scanner.EOF {
			return NewToken(TOKEN_EOF, "", l.scanner.Pos())
		}

		// コメントを無視
		if tok == scanner.Comment {
			continue
		}

		if tok == '"' || tok == '\'' {
			runes := make([]rune, 0)
			escape := false
			for {
				r := l.scanner.Next()
				if r == tok && !escape || r == scanner.EOF {
					break
				}

				if r == '\\' {
					escape = !escape
				} else {
					escape = false
				}

				runes = append(runes, r)
			}

			var t TokenType
			if tok == '"' {
				t = TOKEN_STRING2
			} else {
				t = TOKEN_STRING1
			}

			return NewToken(t, string(runes), l.scanner.Pos())
		}

		text := l.scanner.TokenText()
		if keywords[text] {
			return NewToken(TOKEN_KEYWORD, text, l.scanner.Pos())
		} else if tok == scanner.Int || tok == scanner.Float {
			return NewToken(TOKEN_NUMBER, text, l.scanner.Pos())
		} else if tok == scanner.Ident {
			return NewToken(TOKEN_IDENTIFIER, text, l.scanner.Pos())
		} else {
			return NewToken(TOKEN_SYMBOL, text, l.scanner.Pos())
		}
	}
}
