package kanji

import _ "embed"

// N1 is the sourced JLPT N1 kanji curriculum.
//go:embed n1.json
var N1 []byte
