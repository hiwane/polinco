package com

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func NewReporter(reporter string) Reporter {
	if reporter == "github" {
		return NewGithubReporter()
	} else {
		return NewPlainReporter()
	}
}

// ///////////////////////////
// reportCounter
// ///////////////////////////
type reportCounter struct {
	self        Reporter
	num         int
	stripPrefix string
}

func (r *reportCounter) addError(level int) {
	if level <= LevelError {
		r.num++
	}
}

func (r *reportCounter) CountError() int {
	return r.num
}

func (r *reportCounter) SetStripPrefix(prefix string) {
	r.stripPrefix = prefix
}

func (r *reportCounter) stripFilename(filename string) string {
	if strings.HasPrefix(filename, r.stripPrefix) {
		filename = filename[len(r.stripPrefix):]
	}
	return filename
}

func (r *reportCounter) ReportError(filename string, lnum, col, level int, msg string) {
	slimname := r.stripFilename(filename)
	r.self.reportError(filename, slimname, lnum, col, level, msg)
	r.addError(level)
}

/////////////////////////////
// PlainReporter
/////////////////////////////

type PlainReporter struct {
	reportCounter
}

func NewPlainReporter() *PlainReporter {
	p := &PlainReporter{}
	p.self = p
	return p
}

func (r *PlainReporter) reportError(fullpath, filename string, lnum, col, level int, msg string) {
	fmt.Printf("%s:%d:%d:%s: %s\n", filename, lnum, col, lv2str(level), msg)
}

/////////////////////////////
// GithubReporter
/////////////////////////////

type gitInfo struct {
	url      string
	branch   string
	hash     string
	username string
	reponame string
}

type GithubReporter struct {
	reportCounter
	cache map[string]*gitInfo
}

func (r *GithubReporter) getGitDir(filename string) string {
	n := len(strings.Split(filename, "/"))
	for i := 0; i <= n; i++ {
		filename = filepath.Dir(filename)
		if filename == "/" || filename == "." || filename == "" {
			return ""
		}
		if st, err := os.Stat(filepath.Join(filename, ".git")); err == nil && st.IsDir() {
			return filename
		}
	}
	return ""
}

func execGitCommand(dirname string, args ...string) (string, error) {

	cmd := exec.Command("git", args...)
	cmd.Dir = dirname
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func getCurrentCommit(dirname string) (string, error) {
	return execGitCommand(dirname, "rev-parse", "HEAD")
}

func getFileURL(dirname string) (string, error) {
	// GitHubリポジトリのURLを取得するために `git config --get remote.origin.url` を実行

	// リポジトリURLを整形
	repoURL, err := execGitCommand(dirname, "config", "--get", "remote.origin.url")
	if err != nil {
		return "", err
	}

	// SSHの場合: git@github.com:user/repository.git -> https://github.com/user/repository
	if strings.HasPrefix(repoURL, "git@github.com:") {
		repoURL = "https://github.com/" + strings.TrimPrefix(repoURL, "git@github.com:")
		repoURL = strings.TrimSuffix(repoURL, ".git")
	} else {
		// HTTPS形式の場合、そのまま処理
		repoURL = strings.TrimSuffix(repoURL, ".git")
	}

	return repoURL, nil
}

func (r *GithubReporter) getGitInfo(dirname string) (*gitInfo, error) {
	var err error

	gi := &gitInfo{}
	gi.branch, err = execGitCommand(dirname, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		fmt.Printf("getGitInfo: branch %v\n", err)
		return nil, err
	}

	gi.hash, err = getCurrentCommit(dirname)
	if err != nil {
		fmt.Printf("getGitInfo: hash %v\n", err)
		return nil, err
	}

	gi.url, err = getFileURL(dirname)
	if err != nil {
		fmt.Printf("getGitInfo: url %v\n", err)
		return nil, err
	}

	vv := strings.Split(gi.url, "/")
	gi.username = vv[len(vv)-2]
	gi.reponame = vv[len(vv)-1]

	return gi, nil
}

func (r *GithubReporter) reportPlain(filename string, lnum, col, level int, msg string) {
	fmt.Printf("- [ ] %s:%d %s\n", filename, lnum, msg)
}

func (r *GithubReporter) reportError(fullpath, filename string, lnum, col, level int, msg string) {
	// username/Cabinets/resources/locales/ja_JP/cabinets.po から，Cabinets と cabinets.po
	dir := r.getGitDir(fullpath)
	if dir == "" {
		r.reportPlain(filename, lnum, col, level, msg)
		return
	}

	if _, ok := r.cache[dir]; !ok {
		r.cache[dir], _ = r.getGitInfo(dir)
	}

	gi := r.cache[dir]
	if gi == nil {
		r.reportPlain(filename, lnum, col, level, msg)
		return
	}

	basename := filepath.Base(filename)
	if strings.Contains(filename, gi.reponame) {
		filename = filename[strings.Index(filename, gi.reponame)+len(gi.reponame)+1:]
	}

	fmt.Printf("- [ ] [%s:%s:%d](%s/blob/%s/%s#L%d) %s\n", gi.reponame, basename, lnum, gi.url, gi.hash, filename, lnum, msg)
}

func NewGithubReporter() *GithubReporter {
	p := &GithubReporter{}
	p.self = p
	p.cache = make(map[string]*gitInfo)
	return p
}
