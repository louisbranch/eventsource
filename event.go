package eventsource

import (
	"bytes"
	"errors"
	"net/url"
	"strings"
)

type event struct {
	name     string
	message  string
	channels []string
}

func ParseEventUrl(query url.Values) (event, error) {
	e := event{}
	message := query.Get("message")
	if message == "" {
		return e, errors.New("Event message can't be blank")
	}

	channels := query.Get("channels")
	if channels == "" {
		return e, errors.New("Event channels can't be blank")
	}

	e.message = message
	e.name = query.Get("event")
	e.channels = strings.Split(channels, ",")
	return e, nil
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
