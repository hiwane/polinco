package po

import (
	"io"
	"polinco/com"
)

func ParsePo(r io.Reader) ([]*com.PoEntry, error) {
	lexer := newLexer(r)
	poEntries = make([]*com.PoEntry, 0)
	yyParse(lexer)
	if lexer.err != nil {
		return nil, lexer.err
	}
	return poEntries, nil
}

/**

	str2id := make(map[string]*PoEntry)
	id2str := make(map[string]*PoEntry)
	for _, entry := range poEntries {
		if e, ok := str2id[entry.MsgStr]; ok {
			fmt.Printf("duplicate msgstr: %s [file=%s, %d:%s, %d:%s]\n", entry.MsgStr, e.Pos.Filename, e.Pos.Line, e.MsgID, entry.Pos.Line, entry.MsgID)
		}

		str2id[entry.MsgStr] = entry
		id2str[entry.MsgID] = entry

		fmt.Printf("msgid: %s\n", entry.MsgID)
		fmt.Printf("msgstr: %s\n", entry.MsgStr)
	}

	return nil, nil
}
*/

func Debug(level int, verbose bool) {
	yyDebug = level
	yyErrorVerbose = verbose
}
