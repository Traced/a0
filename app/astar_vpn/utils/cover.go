package utils

import (
	"bytes"
	"io"
)

func StringMapToBufferReader(m map[string]string) io.Reader {
	var buf bytes.Buffer
	for k, v := range m {
		buf.WriteString(k + "=" + v + "&")
	}
	return &buf
}
