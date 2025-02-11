package po

import (
	"errors"
	"fmt"
	"io"
	"polinco/com"
	"strconv"
	"text/scanner"
)

// /////////////////////////////////////////////////////////////
// NODE
// /////////////////////////////////////////////////////////////
type pNode struct {
	cmd   int
	extra int
	str   string
	pos   scanner.Position
}

func newPNode(str string, cmd, extra int, pos scanner.Position) pNode {
	return pNode{cmd: cmd, str: str, extra: extra, pos: pos}
}

func (n *pNode) String() string {
	return n.str + ":" + strconv.Itoa(n.cmd)
}

// /////////////////////////////////////////////////////////////
// LEXer
// /////////////////////////////////////////////////////////////
type pLexer struct {
	scanner.Scanner
	com.Lexer
	lnum        int
	s           string
	err         error
	varmap      map[string]string
	print_trace bool
}

func (l *pLexer) skip_space() {
	for {
		for l.IsSpace(l.Peek()) {
			l.Next()
		}
		if l.Peek() != '#' {
			break
		}
		for l.Peek() != '\n' { // 改行までコメント
			l.Next()
		}
	}
}

// 字句解析機
func (l *pLexer) Lex(lval *yySymType) int {
	l.skip_space()

	c := l.Peek()
	l.trace(fmt.Sprintf("lex: go! %c", c))
	if l.IsAlpha(l.Peek()) { // 英字
		var ret []rune
		for l.IsDigit(l.Peek()) || l.IsLetter(l.Peek()) {
			ret = append(ret, l.Next())
		}
		str := string(ret)
		if str == "msgid" {
			l.trace("lex:msgid")
			lval.node = newPNode(str, MSGID, 0, l.Pos())
			return MSGID
		} else if str == "msgstr" {
			l.trace("lex:msgstr")
			lval.node = newPNode(str, MSGSTR, 0, l.Pos())
			return MSGSTR
		} else {
			// panic!
			l.trace("lex:unknown:" + str)
			return int(c)
		}
	}
	if l.Peek() == '"' {
		sbuf := make([]rune, 0)
		l.Next()
		escape := false
		for (l.Peek() != '"' || escape) && l.Peek() != scanner.EOF {
			d := l.Next()
			if d == '\\' {
				escape = !escape
			} else {
				escape = false
			}
			sbuf = append(sbuf, d)
		}
		l.Next()
		str := string(sbuf)
		l.trace("lex:string:" + str)
		lval.node = newPNode(str, STRING, 0, l.Pos())
		return STRING
	}
	l.trace("lex:unknown:" + string(c))
	return int(c)
}

func (l *pLexer) Error(s string) {
	pos := l.Pos()
	if l.err == nil {
		l.err = errors.New(fmt.Sprintf("%s:Error:%s \n", pos.String(), s))
	}
}

func (l *pLexer) trace(s string) {
	if l.print_trace {
		fmt.Printf(s + "\n")
	}
}

func newLexer(r io.Reader) *pLexer {
	p := new(pLexer)
	p.Init(r)
	p.print_trace = false
	return p
}
