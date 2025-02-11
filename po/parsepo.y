%{
package po

import (
//	"fmt"
	"polint/com"
)

var poEntries []*com.PoEntry
%}

%union{
	node pNode
}

%token MSGID MSGSTR STRING

%%

pofile
	  : entries {}
	  ;

entries:
	   | entries entry
	   ;


entry:
	 MSGID strings MSGSTR strings {
	 	poEntries = append(poEntries, &com.PoEntry{MsgID: $2.node.str, MsgStr: $4.node.str, Pos: $1.node.pos})
	}



strings
	: STRING
	| strings STRING {
		$$.node.str = $1.node.str + $2.node.str
	}
	;

%%
