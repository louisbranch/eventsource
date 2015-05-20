package eventsource

import (
	"bytes"
	"errors"
	"net/url"
	"strings"
)

type Event struct {
	Name     string
	Message  string
	Channels []string
}

func ParseEventQuery(query url.Values) (Event, error) {
	e := Event{}
	message := query.Get("message")
	if message == "" {
		return e, errors.New("Event message can't be blank")
	}

	e.Message = message
	e.Name = query.Get("event")
	e.Channels = strings.Split(query.Get("channels"), ",")
	return e, nil
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
