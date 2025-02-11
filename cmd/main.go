package main

import (
	"flag"
	"fmt"
	"github.com/hiwane/flagvar"
	"log"
	"os"
	"path/filepath"
	"polinco/com"
	"polinco/php"
	"polinco/po"
	"strings"
)

var gitCommit string

func main() {

	var (
		locale_dir    = flag.String("locale", "", "locale directory")
		src_dir       = flag.String("src", "", "source directory")
		package_name  = flag.String("package", "", "package name")
		parse_level   = flag.Int("parse-level", 0, "debug level")
		parse_verbose = flag.Bool("parse-verbose", false, "verbose mode")
	)
	reporters := []string{"plain"} // , "json", "csv"}
	opt_reporter := flagvar.NewChoiceVar(reporters[0], reporters)
	flag.Var(opt_reporter, "reporter", fmt.Sprintf("reporter (choose from %v)", reporters))

	flag.Parse()

	logger := log.New(log.Writer(), "", log.LstdFlags|log.Lshortfile|log.Lmsgprefix)
	logger.Printf("start polint! version=%s\n", gitCommit)

	reporter := com.NewReporter(opt_reporter.String())
	linter := &com.Linter{Reporter: reporter, Logger: logger}

	po.Debug(*parse_level, *parse_verbose)
	if *locale_dir == "" {
		// po.Debug(8, true)
		sep := "@@@@@@@@@@@@@@@@@@\n"
		for _, s := range []string{
			"msgid \"hoge\"\nmsgstr \"fuga\"\n",
			"msgid \"\"\nmsgstr \"\"\n",
			"msgid \"gaogao\"\nmsgstr \"hoghoge\"\n\"guruguru\"\n",
		} {
			fmt.Printf("%s%s%s", sep, s, sep)
			r := strings.NewReader(s)
			po.ParsePo(r)
		}
		os.Exit(0)
	}

	// locale_dir/ja_JP/*.po を読み込む
	entries, err := parsePoLocale(linter, *locale_dir)
	if err != nil {
		os.Exit(1)
	}

	if linter.Reporter.CountError() > 0 {
		os.Exit(1)
	}

	if *src_dir != "" {
		entriesDict := make(map[string]map[string]*com.PoEntry)
		entriesDict[*package_name] = entries

		err = php.ParsePHPDir(linter, *src_dir, entriesDict)
		if err != nil {
			os.Exit(1)
		}
	}

	os.Exit(0)
}

func parsePoLocale(linter *com.Linter, locale_dir string) (map[string]*com.PoEntry, error) {
	// locale_dir/ja_JP/*.po を読み込む
	// fmt.Printf("parse_dir  '%s'\n", locale_dir)
	dirs, err := filepath.Glob(locale_dir + "/*")
	if err != nil {
		linter.Logger.Fatal(err)
		return nil, err
	}
	var ret map[string]*com.PoEntry
	for _, dir := range dirs {
		// fmt.Printf("parse_dir  %s\n", dir)
		files, err := filepath.Glob(dir + "/*.po")
		if err != nil {
			linter.Logger.Fatal(err)
			return nil, err
		}

		if len(files) == 0 {
			continue
		}

		poret, err := parsePoDir(linter, files)
		if err != nil {
			linter.Logger.Fatal(err)
			return nil, err
		}

		if ret != nil {
			if !comparePo(ret, poret) {
				linter.Logger.Fatalf("po files are different: %s", dir)
				return nil, err
			}
			panic("not implemented!")
		}

		ret = poret
	}
	return ret, nil
}

func parsePoDir(linter *com.Linter, files []string) (map[string]*com.PoEntry, error) {

	id2entry := make(map[string]*com.PoEntry)
	str2entry := make(map[string]*com.PoEntry)
	for _, file := range files {
		fmt.Printf("parse_file %s\n", file)
		_id2, _str2, err := parsePoFile(linter, file)
		if err != nil {
			linter.Reporter.ReportError(file, 0, 0, com.LevelError, err.Error())
			return nil, err
		}

		for msgid, entry := range _id2 {
			if e, ok := id2entry[msgid]; ok {
				linter.Reporter.ReportError(file, entry.Pos.Line, entry.Pos.Column, com.LevelError, fmt.Sprintf("1duplicate msgid: %s=%s [%s:%d:%s]", msgid, entry.MsgStr, e.Pos.Filename, e.Pos.Line, e.MsgStr))
			} else {
				id2entry[msgid] = entry
			}
		}

		for msgstr, entry := range _str2 {
			if e, ok := str2entry[msgstr]; ok {
				linter.Reporter.ReportError(file, entry.Pos.Line, entry.Pos.Column, com.LevelWarning, fmt.Sprintf("1duplicate msgstr: %s=%s [%s:%d:%s]", msgstr, entry.MsgID, e.Pos.Filename, e.Pos.Line, e.MsgID))
			} else {
				str2entry[msgstr] = entry
			}
		}
	}

	return id2entry, nil

}

/***
 * 単一ファイル内でのチェック
 */
func parsePoFile(linter *com.Linter, filename string) (map[string]*com.PoEntry, map[string]*com.PoEntry, error) {
	r, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}

	poEntries, err := po.ParsePo(r)
	if err != nil {
		return nil, nil, err
	}

	str2entry := make(map[string]*com.PoEntry)
	id2entry := make(map[string]*com.PoEntry)
	for _, entry := range poEntries {
		entry.Filename = filename
	}

	for _, entry := range poEntries {

		if entry.Filename == "" {
			panic("bug!")
		}

		if e, ok := str2entry[entry.MsgStr]; ok {
			if e.Filename == "" {
				panic("bug!")
			}
			linter.Reporter.ReportError(
				filename, entry.Pos.Line, entry.Pos.Column, com.LevelWarning,
				fmt.Sprintf("2duplicate msgstr: %s=%s [%s:%d:%s]", entry.MsgID, entry.MsgStr, e.Filename, e.Pos.Line, e.MsgID))
		} else {
			str2entry[entry.MsgStr] = entry
		}

		if e, ok := id2entry[entry.MsgID]; ok {
			linter.Reporter.ReportError(
				filename, entry.Pos.Line, entry.Pos.Column, com.LevelError,
				fmt.Sprintf("2duplicate msgid: %s=%s [%s:%d:%s]", entry.MsgID, entry.MsgStr, e.Filename, e.Pos.Line, e.MsgID))
		} else {
			id2entry[entry.MsgID] = entry
		}
	}

	return id2entry, str2entry, nil
}

func comparePo(po1, po2 map[string]*com.PoEntry) bool {
	if len(po1) != len(po2) {
		return false
	}

	for k, _ := range po1 {
		_, ok := po2[k]
		if !ok {
			return false
		}
	}
	return true
}
