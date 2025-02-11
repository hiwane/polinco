package php

import (
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

	_, err := os.ReadFile(filename)
	if err != nil {
		linter.Logger.Fatal(err)
		return err
	}

	// Error handler
	return nil
}
