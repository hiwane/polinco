package com

import (
	"fmt"
)

const LevelFatal = 0
const LevelError = 1
const LevelWarning = 2
const LevelInfo = 3
const LevelNone = 4

func lv2str(level int) string {
	switch level {
	case LevelFatal:
		return "FATAL"
	case LevelError:
		return "ERR"
	case LevelWarning:
		return "WRN"
	case LevelInfo:
		return "INF"
	case LevelNone:
		return "NON"
	default:
		return fmt.Sprintf("=%d=", level)
	}
}

/////////////////////////////
// reportCounter
/////////////////////////////

type reportCounter struct {
	num int
}

func (r *reportCounter) addError(level int) {
	r.num++
}

func (r *reportCounter) CountError() int {
	return r.num
}

/////////////////////////////
// PlainReporter
/////////////////////////////

type PlainReporter struct {
	reportCounter
}

func NewReporter(reporter string) Reporter {
	return &PlainReporter{}
}

func (r *PlainReporter) ReportError(filename string, lnum, col, level int, msg string) {
	fmt.Printf("%s:%d:%d:%s: %s\n", filename, lnum, col, lv2str(level), msg)
	r.addError(level)
}
