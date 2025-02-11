package com

import (
	//	"fmt"
	"log"
	"text/scanner"
)

type PoEntry struct {
	MsgID    string
	MsgStr   string
	Pos      scanner.Position
	Filename string
	Called   int
}

type Reporter interface {
	ReportError(filename string, lnum, col, level int, msg string)
	CountError() int
}

type Linter struct {
	Reporter Reporter
	Logger   *log.Logger
}
