package eventsource

import (
	"bytes"
)

type event struct {
	name     string
	message  string
	channels []string
}

func (e *event) Bytes() []byte {
	var buf bytes.Buffer
	if e.name != "" {
		buf.WriteString("event: ")
		buf.WriteString(e.name)
		buf.WriteString("\n")
	}
	buf.WriteString("data: ")
	buf.WriteString(e.message)
	buf.WriteString("\n")
	return buf.Bytes()
}
