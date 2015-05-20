package eventsource

import (
	"bytes"
)

type Event struct {
	Name     string
	Message  string
	Channels []string
}

func (e *Event) Bytes() []byte {
	var buf bytes.Buffer
	if e.Name != "" {
		buf.WriteString("event: ")
		buf.WriteString(e.Name)
		buf.WriteString("\n")
	}
	buf.WriteString("data: ")
	buf.WriteString(e.Message)
	buf.WriteString("\n")
	return buf.Bytes()
}
