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
	"regexp"
	"strings"
)

var gitCommit string

type strsslice []string

func (s *strsslice) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *strsslice) Set(value string) error {
	for _, v := range *s {
		if v == value {
			return nil
		}
	}

	*s = append(*s, value)
	return nil
}

func main() {

	var (
		// locale_dir    = flag.String("locale", "", "locale directory")
		src_dir       = flag.String("src", "", "source directory")
		parse_level   = flag.Int("parse-level", 0, "debug level")
		parse_verbose = flag.Bool("parse-verbose", false, "verbose mode of parser")
		verbose       = flag.Bool("verbose", false, "verbose mode of polinco")
		strip_prefix  = flag.String("strip-prefix", "", "strip the specified prefix from file path in the report")
	)
	reporters := []string{"plain", "github"} // , "json", "csv"}
	opt_reporter := flagvar.NewChoiceVar(reporters[0], reporters)
	flag.Var(opt_reporter, "reporter", fmt.Sprintf("reporter (choose from %v)", reporters))

	var plugins strsslice
	// *.po ファイルを読み込むプラグイン名
	flag.Var(&plugins, "plugin", "plugin name")

	flag.Parse()

	logger := log.New(log.Writer(), "", log.LstdFlags|log.Lshortfile|log.Lmsgprefix)
	logger.Printf("start polint! version=%s\n", gitCommit)

	reporter := com.NewReporter(opt_reporter.String())
	reporter.SetStripPrefix(*strip_prefix)
	linter := &com.Linter{Reporter: reporter, Logger: logger}
	linter.SetVerbose(*verbose)

	po.Debug(*parse_level, *parse_verbose)

	entriesDict := make(map[string]map[string]*com.PoEntry)
	for _, plugin := range plugins {
		locale_dir := plugin + "/resources/locales/"
		pentries, err := parsePoLocale(linter, locale_dir)
		if err != nil {
			os.Exit(1)
		}

		for k, v := range pentries {
			if _, ok := entriesDict[k]; ok {
				linter.Reporter.ReportError("", 0, 0, com.LevelError, fmt.Sprintf("duplicate plugin: %s", k))
			}
			entriesDict[k] = v
		}
	}

	// locale_dir/ja_JP/*.po を読み込む

	if linter.Reporter.CountError() > 0 {
		fmt.Printf("exit ... pofile err %d\n", linter.Reporter.CountError())
		os.Exit(1)
	}

	if *src_dir != "" {
		err := php.ParsePHPDir(linter, *src_dir, entriesDict)
		if err != nil {
			os.Exit(1)
		}
	}

	if linter.Reporter.CountError() > 0 {
		fmt.Printf("exit ... phpfile err %d\n", linter.Reporter.CountError())
		os.Exit(1)
	}

	os.Exit(0)
}

func countEntry(entries map[string]map[string]*com.PoEntry) int {
	n := 0
	for _, v := range entries {
		n += len(v)
	}
	return n
}

func parsePoLocale(linter *com.Linter, locale_dir string) (map[string]map[string]*com.PoEntry, error) {
	// locale_dir/ja_JP/*.po を読み込む
	// fmt.Printf("parse_dir  '%s'\n", locale_dir)
	dirs, err := filepath.Glob(locale_dir + "/*")
	if err != nil {
		linter.Logger.Fatal(err)
		return nil, err
	}
	var ret map[string]map[string]*com.PoEntry
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
		}

		if ret == nil || len(ret) < len(poret) || len(ret) == len(poret) && countEntry(ret) < countEntry(poret) {
			ret = poret
		}
	}
	return ret, nil
}

func parsePoDir(linter *com.Linter, files []string) (map[string]map[string]*com.PoEntry, error) {

	retmaps := make(map[string]map[string]*com.PoEntry)
	retmapi := make(map[string]map[string]*com.PoEntry)

	for _, file := range files {
		// file = $path/plugin_name.po から plugin_name を取得
		plugin_name := strings.TrimSuffix(filepath.Base(file), ".po")

		linter.Dprintf("%s: %s start\n", plugin_name, file)

		id2entry, ok := retmapi[plugin_name]
		if !ok {
			id2entry = make(map[string]*com.PoEntry)
			retmapi[plugin_name] = id2entry
		}

		str2entry, ok := retmaps[plugin_name]
		if !ok {
			str2entry = make(map[string]*com.PoEntry)
			retmaps[plugin_name] = str2entry
		}

		_id2, _str2, err := parsePoFile(linter, file)
		if err != nil {
			linter.Reporter.ReportError(file, 0, 0, com.LevelError, err.Error())
			return nil, err
		}

		for msgid, entry := range _id2 {
			if e, ok := id2entry[msgid]; ok {
				linter.Reporter.ReportError(file, entry.Pos.Line, entry.Pos.Column, com.LevelError, fmt.Sprintf("1duplicate msgid: %s=%s [%s:%d:%s]", msgid, entry.MsgStr, e.Filename, e.Pos.Line, e.MsgStr))
			} else {
				id2entry[msgid] = entry
			}
		}

		for msgstr, entry := range _str2 {
			if e, ok := str2entry[msgstr]; ok && entry.MsgID != e.MsgID {
				// 英語の場合複数形と最後にピリオドがあるかもしれない
				if entry.MsgID+"." != e.MsgID &&
					entry.MsgID != e.MsgID+"." &&
					entry.MsgID+"s" != e.MsgID &&
					entry.MsgID != e.MsgID+"s" {
					linter.Reporter.ReportError(file, entry.Pos.Line, entry.Pos.Column, com.LevelWarning, fmt.Sprintf("1duplicate msgstr: %s=%s [%s:%d:%s]", msgstr, entry.MsgID, e.Filename, e.Pos.Line, e.MsgID))
					continue
				}
			}
			str2entry[msgstr] = entry
		}
	}

	return retmapi, nil
}

/***
 * 単一ファイル内でのチェック
 */
func parsePoFile(linter *com.Linter, filename string) (map[string]*com.PoEntry, map[string]*com.PoEntry, error) {
	r, err := os.Open(filename)
	if err != nil {
		linter.Logger.Fatal(err)
		return nil, nil, err
	}

	poEntries, err := po.ParsePo(r)
	if err != nil {
		linter.Logger.Fatal(err)
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
			if e.Filename == "" || e.Filename != entry.Filename {
				panic("bug!")
			}
			if entry.MsgID+"." != e.MsgID &&
				entry.MsgID != e.MsgID+"." &&
				entry.MsgID+"s" != e.MsgID &&
				entry.MsgID != e.MsgID+"s" {
				linter.Reporter.ReportError(
					filename, entry.Pos.Line, entry.Pos.Column, com.LevelWarning,
					fmt.Sprintf("2duplicate msgstr: %s=%s [%d:%s]", entry.MsgID, entry.MsgStr, e.Pos.Line, e.MsgID))
			} else {
				linter.Reporter.ReportError(
					filename, entry.Pos.Line, entry.Pos.Column, com.LevelInfo,
					fmt.Sprintf("2duplicate msgstr x similar msgid: %s=%s [%d:%s]", entry.MsgID, entry.MsgStr, e.Pos.Line, e.MsgID))
			}
		} else {
			str2entry[entry.MsgStr] = entry
		}

		if e, ok := id2entry[entry.MsgID]; ok {
			linter.Reporter.ReportError(
				filename, entry.Pos.Line, entry.Pos.Column, com.LevelError,
				fmt.Sprintf("2duplicate msgid: %s=%s [%d:%s]", entry.MsgID, entry.MsgStr, e.Pos.Line, e.MsgID))
		} else {
			id2entry[entry.MsgID] = entry
		}

		b, err := regexp.MatchString(`^[a-zA-Z0-9 {}()<>:/=%[\]'"?,._\\-]*$`, entry.MsgID)
		if err != nil {
			return nil, nil, err
		}

		if !b {
			linter.Reporter.ReportError(
				filename, entry.Pos.Line, entry.Pos.Column, com.LevelError,
				fmt.Sprintf("invalid msgid: '%s'", entry.MsgID))
		}

		for i := 0; i < 10; i++ {
			tag := fmt.Sprintf("{%d}", i)
			if strings.Contains(entry.MsgID, tag) != strings.Contains(entry.MsgStr, tag) {
				linter.Reporter.ReportError(
					filename, entry.Pos.Line, entry.Pos.Column, com.LevelWarning,
					fmt.Sprintf("missing `%s` in msgid<%s> or msgstr<%s>", tag, entry.MsgID, entry.MsgStr))
			}
		}
	}

	return id2entry, str2entry, nil
}

func comparePo(po1, po2 map[string]map[string]*com.PoEntry) bool {
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
