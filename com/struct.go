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
	reportError(fullpath, filename string, lnum, col, level int, msg string)
	CountError() int
	SetStripPrefix(prefix string)
}

type Linter struct {
	Reporter Reporter
	Logger   *log.Logger
	verbose  bool
}

func (l *Linter) SetVerbose(verbose bool) {
	l.verbose = verbose
}

func (l *Linter) Dprintf(fmt string, args ...any) {
	if l.verbose {
		l.Logger.Printf(fmt, args...)
	}
}
