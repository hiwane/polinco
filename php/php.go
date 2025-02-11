package php

import (
	"fmt"
	"os"
	"path/filepath"
	"polinco/com"
	"strings"
)

func ParsePHPDir(linter *com.Linter, dirname string, entriesDict map[string]map[string]*com.PoEntry) error {

	// dirname 配下のファイル/ディレクトリを取得
	files, err := filepath.Glob(dirname + "/*")
	if err != nil {
		linter.Logger.Fatal(err)
		return err
	}

	for _, file := range files {
		// ディレクトリなら再起に
		if fi, err := os.Stat(file); err == nil && fi.IsDir() {
			err = ParsePHPDir(linter, file, entriesDict)
			if err != nil {
				linter.Logger.Fatal(err)
				return err
			}
			continue
		}

		// *.php ファイルなら解析する
		if strings.HasSuffix(file, ".php") {
			err = parsePHPFile(linter, file, entriesDict)
			if err != nil {
				linter.Logger.Fatal(err)
				return err
			}
		}
	}
	return nil
}

func parsePHPFile(linter *com.Linter, filename string, entriesDict map[string]map[string]*com.PoEntry) error {
	file, err := os.Open(filename)
	if err != nil {
		linter.Logger.Fatal(err)
		return err
	}
	defer file.Close()

	lexer := NewLexer(file)
	lexer.scanner.Position.Filename = filename
	nofile := false
	if nofile {
		sss := `
        $data = [
            'title_icon' => '',
            //'content_id' => '',
            'title' => "あい\u{00A0}\u{00A0}\u{00A0}\u{00A0} う&え<お>か<hr>き",
        ];`
		lexer = NewLexer(strings.NewReader(sss))
	}

	tokens := getTokens(lexer)
	linter.Dprintf("start parsePHPFile(%s): %d tokens\n", filename, len(tokens))
	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		if tok.is(TOKEN_IDENTIFIER, "__d") {
			if i+6 >= len(tokens) {
				linter.Reporter.ReportError(filename, tok.Lnum, tok.Col, com.LevelError, "Invalid __d function")
				continue
			}

			if !tokens[i+1].is(TOKEN_SYMBOL, "(") {
				linter.Reporter.ReportError(filename, tok.Lnum, tok.Col, com.LevelError, "Invalid __d function: missing '('")
				continue
			}
			if tokens[i+2].isType(TOKEN_STRING2) {
				linter.Reporter.ReportError(filename, tok.Lnum, tok.Col, com.LevelWarning, "1st argument of __d() should be a single quoted string not a double quoted string.")
			} else if tokens[i+2].is(TOKEN_SYMBOL, "$") {
				// 解析不可能故逃げる
				continue
			} else if !tokens[i+2].isType(TOKEN_STRING1) {
				linter.Reporter.ReportError(filename, tok.Lnum, tok.Col, com.LevelWarning, "1st argument of __d() should be a single quoted string: "+tokens[i+2].Value)
				continue
			}

			if !tokens[i+3].is(TOKEN_SYMBOL, ",") {
				linter.Reporter.ReportError(filename, tok.Lnum, tok.Col, com.LevelError, "Invalid __d function: missing ','")
				continue
			}

			if tokens[i+4].isType(TOKEN_STRING2) {
				linter.Reporter.ReportError(filename, tok.Lnum, tok.Col, com.LevelWarning, "2nd argument of __d() should be a single quoted string not a double quoted string.")
			} else if tokens[i+4].is(TOKEN_SYMBOL, "$") {
				// 解析不可能故逃げる
				continue
			} else if !tokens[i+4].isType(TOKEN_STRING1) {
				linter.Reporter.ReportError(filename, tok.Lnum, tok.Col, com.LevelWarning, "2nd argument of __d() should be a single quoted string: "+tokens[i+4].Value)
				continue
			}

			if !tokens[i+5].is(TOKEN_SYMBOL, ",") && !tokens[i+5].is(TOKEN_SYMBOL, ")") {
				linter.Reporter.ReportError(filename, tok.Lnum, tok.Col, com.LevelError, "Invalid __d function: missing ',' or ')'")
				continue
			}

			entries, ok := entriesDict[tokens[i+2].Value]
			if !ok {
				linter.Reporter.ReportError(filename, tok.Lnum, tok.Col, com.LevelError, "Unknown domain: "+tokens[i+2].Value+", msgid="+tokens[i+4].Value)
				continue
			}

			entry, ok := entries[tokens[i+4].Value]
			if !ok {
				linter.Reporter.ReportError(filename, tok.Lnum, tok.Col, com.LevelError, "Unknown msgid: __d("+tokens[i+2].Value+","+tokens[i+4].Value+")")
				continue
			}

			var i int
			for i = 9; i >= 0; i-- {
				if strings.Contains(entry.MsgStr, fmt.Sprintf("{%d}", i)) {
					break
				}
			}

			argnum := getArgNum(tokens, i+5)
			if argnum < i {
				linter.Reporter.ReportError(filename, tok.Lnum, tok.Col, com.LevelError, fmt.Sprintf("Invalid __d function: missing %d-th argument for {%d}", i+1, i))
			}
		}
	}

	// Error handler
	return nil
}

func getArgNum(tokens []*Token, start int) int {
	depth := 0
	argnum := 0
	empty := true

	for i := start; i < len(tokens); i++ {
		if tokens[i].is(TOKEN_SYMBOL, "(") {
			depth++
		} else if tokens[i].is(TOKEN_SYMBOL, ")") {
			depth--
			if depth < 0 {
				if !empty {
					argnum++
				}
				break
			}
		} else if tokens[i].is(TOKEN_SYMBOL, ",") {
			if depth == 0 {
				argnum++
				empty = true
			}
		} else {
			empty = false
		}
	}

	return argnum
}

func getTokens(lexer *Lexer) []*Token {

	tokens := make([]*Token, 0)
	for {
		tok := lexer.NextToken()

		if tok.Type == TOKEN_EOF {
			break
		}

		tokens = append(tokens, tok)
	}

	// 文字列を連結する
	ret := make([]*Token, 0, len(tokens))
	for i := 0; i < len(tokens); i++ {
		if len(ret) > 0 && ret[len(ret)-1].isString() && tokens[i].isString() {
			// 文字列が続いた
			ret[len(ret)-1].concat(tokens[i])
			continue
		}

		if len(ret) > 0 && i+1 < len(tokens) && ret[len(ret)-1].isString() && tokens[i].is(TOKEN_SYMBOL, ".") && tokens[i+1].isString() {
			// . で連結された文字列
			ret[len(ret)-1].concat(tokens[i+1])
			i++
			continue
		}

		ret = append(ret, tokens[i])
	}

	return ret
}

/**
 * 破壊的
 */
func (t *Token) concat(s *Token) *Token {
	t.Value += s.Value
	if t.Type != s.Type {
		t.Type = TOKEN_STRING2
	}
	return t
}

func (t *Token) isString() bool {
	return t.isType(TOKEN_STRING1) || t.isType(TOKEN_STRING2)
}

func (t *Token) isType(tokenType TokenType) bool {
	return t.Type == tokenType
}

func (t *Token) isValue(val string) bool {
	return t.Value == val
}

func (t *Token) is(tokenType TokenType, val string) bool {
	return t.isType(tokenType) && t.isValue(val)
}
