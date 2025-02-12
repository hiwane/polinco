package php

import (
	"strings"
	"testing"
)

func TestGetArgNum(t *testing.T) {
	for _, s := range []struct {
		input  string
		idx    int
		end    string
		expect int
	}{
		{"__d('cake', 'a', 'b')", 2, ")", 3},
		{"__d('cake', 'a', 'b')", 4, ")", 2},
		{"__d('cake', 'a', 'b')", 6, ")", 1},
		{"__d('cake', 'a', 'b',)", 6, ")", 1},
		{"__d('cake', 'a', __d('boo', 'ruuuu', 3))", 4, ")", 2},
		{"__d('cake', 'a', __d('boo', 'ruuuu', 3))", 6, ")", 1},
		{"__d('cake', 'a', __d('boo', 'ruuuu', 3,))", 6, ")", 1},
		{"__d('access_counters', 'The number of ', [ __d('access_counters', 'Starting Value'), 0, AccessCounterConst::MAX_COUNT_START, ],", 7, "]", 3},
	} {
		lexer := NewLexer(strings.NewReader(s.input))
		tokens := getTokens(lexer)

		v := getArgNum(tokens[s.idx:], s.end)
		if v != s.expect {
			t.Errorf("\ninput =%d: %v [%s]\nexpect=%v\nactual=%v\n", s.idx, s.input, tokens[s.idx].Value, s.expect, v)
		}
	}
}

func TestGetTokens(t *testing.T) {

	type Expect struct {
		Type  TokenType
		Value string
	}

	for _, s := range []struct {
		input  string
		expect []Expect
	}{
		{"__d('cake', 'abc', \"baaaaaa\")", []Expect{
			{TOKEN_IDENTIFIER, "__d"},
			{TOKEN_SYMBOL, "("},
			{TOKEN_STRING1, "cake"},
			{TOKEN_SYMBOL, ","},
			{TOKEN_STRING1, "abc"},
			{TOKEN_SYMBOL, ","},
			{TOKEN_STRING2, "baaaaaa"},
			{TOKEN_SYMBOL, ")"},
		}},
		{"__d('cake', 'ab' . 'c', \"baaaa\" \"aa\")", []Expect{
			{TOKEN_IDENTIFIER, "__d"},
			{TOKEN_SYMBOL, "("},
			{TOKEN_STRING1, "cake"},
			{TOKEN_SYMBOL, ","},
			{TOKEN_STRING1, "abc"},
			{TOKEN_SYMBOL, ","},
			{TOKEN_STRING2, "baaaaaa"},
			{TOKEN_SYMBOL, ")"},
		}},
		{"__d('access_counters', 'The number of ', [ __d('access_counters', 'Starting Value'), 0, AccessCounterConst::MAX_COUNT_START, ],", []Expect{
			{TOKEN_IDENTIFIER, "__d"},
			{TOKEN_SYMBOL, "("},
			{TOKEN_STRING1, "access_counters"},
			{TOKEN_SYMBOL, ","},
			{TOKEN_STRING1, "The number of "},
			{TOKEN_SYMBOL, ","},
			{TOKEN_SYMBOL, "["},
			{TOKEN_IDENTIFIER, "__d"},
			{TOKEN_SYMBOL, "("},
			{TOKEN_STRING1, "access_counters"},
			{TOKEN_SYMBOL, ","},
			{TOKEN_STRING1, "Starting Value"},
			{TOKEN_SYMBOL, ")"},
			{TOKEN_SYMBOL, ","},
			{TOKEN_NUMBER, "0"},
			{TOKEN_SYMBOL, ","},
			{TOKEN_IDENTIFIER, "AccessCounterConst"},
			{TOKEN_SYMBOL, ":"},
			{TOKEN_SYMBOL, ":"},
			{TOKEN_IDENTIFIER, "MAX_COUNT_START"},
			{TOKEN_SYMBOL, ","},
			{TOKEN_SYMBOL, "]"},
			{TOKEN_SYMBOL, ","},
		}},
	} {
		lexer := NewLexer(strings.NewReader(s.input))
		tokens := getTokens(lexer)

		if len(tokens) != len(s.expect) {
			t.Errorf("\ninput =%v\nexpect=%v\nactual=%v\n", s.input, s.expect, tokens)
			continue
		}

		for i, v := range s.expect {
			if tokens[i].Type != v.Type || tokens[i].Value != v.Value {
				t.Errorf("\ninput =%v\nexpect=%v\nactual=%v\n", s.input, s.expect, tokens)
				break
			}
		}
	}
}
