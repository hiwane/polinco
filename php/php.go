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
				linter.Reporter.ReportError(filename, tok.Lnum, tok.Col, com.LevelWarning, "1st argument of __d() should be a single quoted string")
			} else if !tokens[i+2].isType(TOKEN_STRING1) {
				linter.Reporter.ReportError(filename, tok.Lnum, tok.Col, com.LevelWarning, "1st argument of __d() should be a single quoted string")
				continue
			}

			if !tokens[i+3].is(TOKEN_SYMBOL, ",") {
				linter.Reporter.ReportError(filename, tok.Lnum, tok.Col, com.LevelError, "Invalid __d function: missing ','")
				continue
			}

			if tokens[i+4].isType(TOKEN_STRING2) {
				linter.Reporter.ReportError(filename, tok.Lnum, tok.Col, com.LevelWarning, "2nd argument of __d() should be a single quoted string")
			} else if !tokens[i+4].isType(TOKEN_STRING1) {
				linter.Reporter.ReportError(filename, tok.Lnum, tok.Col, com.LevelWarning, "2nd argument of __d() should be a single quoted string")
				continue
			}

			if !tokens[i+5].is(TOKEN_SYMBOL, ",") && !tokens[i+5].is(TOKEN_SYMBOL, ")") {
				linter.Reporter.ReportError(filename, tok.Lnum, tok.Col, com.LevelError, "Invalid __d function: missing ',' or ')'")
				continue
			}

			entries, ok := entriesDict[tokens[i+2].Value]
			if !ok {
				linter.Reporter.ReportError(filename, tok.Lnum, tok.Col, com.LevelError, "Unknown pluginname: "+tokens[i+2].Value)
				continue
			}

			entry, ok := entries[tokens[i+4].Value]
			if !ok {
				linter.Reporter.ReportError(filename, tok.Lnum, tok.Col, com.LevelError, "Unknown msgid: "+tokens[i+4].Value)
				continue
			}

			// {0} が含まれているか
			sep := ")"
			if strings.Contains(entry.MsgStr, "{0}") {
				sep = ","
			}

			if !tokens[i+5].is(TOKEN_SYMBOL, sep) {
				linter.Reporter.ReportError(filename, tok.Lnum, tok.Col, com.LevelError, fmt.Sprintf("Invalid __d function: missing '%s': %s", sep, entry.MsgID))
			}
		}
	}

	// Error handler
	return nil
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

	return tokens

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
