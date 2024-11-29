package tags

import (
	"fmt"
	"strings"
)

var lineBreakBytes = []byte{0x0D, 0x0A}

func LineBreakString() string {
	lb := fmt.Sprintf("\\x%02X\\x%02X", lineBreakBytes[0], lineBreakBytes[1])
	return strings.ToLower(lb)
}

/*
\x0B\c01={u0B:\h01}
\x0B\xF0\c10=\x0d\x0a{x0BF0\h10}
\x0B\xF1=\x0d\x0a{x0BF1}
\x0B\xF2=\x0d\x0a{x0BF2}
\x0B\xF3=\x0d\x0a{x0BF3}
\x0B\xF4\c08=\x0d\x0a{x0BF4\h08}
\x0B\xF6\c0C=\x0d\x0a{x0BF6\h0C}
\x0B\xF7\c04=\x0d\x0a{x0BF7\h04}
\x0B\xF8\c0C=\x0d\x0a{x0BF8\h0C}
\x0B\xF9=\x0d\x0a{x0BF9}
\x0B\xFA\c04=\x0d\x0a{x0BFA\h04}
*/