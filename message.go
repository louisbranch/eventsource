package eventsource

type message struct {
	content  []byte
	channels []string
}
