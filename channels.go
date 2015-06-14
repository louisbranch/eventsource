package eventsource

import (
	"net/http"
	"strings"
)

// The ChannelSubscriber interface is used to determine which channels a client
// has subscribed to. This package has two built-in implementations: NoChannels
// and QueryStringChannels, but you can implement your own.
type ChannelSubscriber interface {
	ParseRequest(*http.Request) []string
}

// A NoChannels implements the ChannelSubscriber interface by always returning
// an empty list of channels. This is useful for implementing an eventsource
// with global messages only.
type NoChannels struct{}

// The ParseRequest function returns an empty list of channels.
func (n NoChannels) ParseRequest(req *http.Request) []string {
	return []string{}
}

// A QueryStringChannels implements the ChannelSubscriber interface by parsing
// the request querystring and extracting channels separated by commas. Eg.:
// /?channels=a,b,c
type QueryStringChannels struct {
	Name string
}

// The ParseRequest function parses they querystring and extracts the channels
// params, spliting it by commas.
func (n QueryStringChannels) ParseRequest(req *http.Request) []string {
	channels := req.URL.Query().Get(n.Name)
	return strings.Split(channels, ",")
}
