package eventsource

type message struct {
	id       uint
	content  []byte
	channels []string
}
