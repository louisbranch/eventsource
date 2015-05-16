package eventsource

type event struct {
	name     string
	message  []byte
	channels []string
}
