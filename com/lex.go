package com

// /////////////////////////////////////////////////////////////
// LEXer
// /////////////////////////////////////////////////////////////
type Lexer struct {
}

func (p *Lexer) IsUpper(ch rune) bool {
	return 'A' <= ch && ch <= 'Z'
}
func (p *Lexer) IsLower(ch rune) bool {
	return 'a' <= ch && ch <= 'z'
}
func (p *Lexer) IsAlpha(ch rune) bool {
	return p.IsUpper(ch) || p.IsLower(ch)
}
func (p *Lexer) IsDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}
func (p *Lexer) IsAlnum(ch rune) bool {
	return p.IsAlpha(ch) || p.IsDigit(ch)
}
func (p *Lexer) IsLetter(ch rune) bool {
	return p.IsAlpha(ch) || ch == '_'
}
func (p *Lexer) IsSpace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}
