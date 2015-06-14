package eventsource

import (
	"net/http"
	"reflect"
	"testing"
)

var req, _ = http.NewRequest("GET", "/?channels=a,b,c", nil)

func TestNoChannelsParseRequest(t *testing.T) {
	sub := NoChannels{}
	results := sub.ParseRequest(req)

	if len(results) > 0 {
		t.Errorf("expected:\n%q\nto be empty\n", results)
	}
}

func TestQueryStringChannelsParseRequestEmpty(t *testing.T) {
	sub := QueryStringChannels{}
	results := sub.ParseRequest(req)

	if len(results) > 0 {
		t.Errorf("expected:\n%q\nto be empty\n", results)
	}
}

func TestQueryStringChannelsParseRequest(t *testing.T) {
	sub := QueryStringChannels{Name: "channels"}
	results := sub.ParseRequest(req)
	expected := []string{"a", "b", "c"}

	if !reflect.DeepEqual(expected, results) {
		t.Errorf("expected:\n%q\nto be equal to:\n%q\n", results, expected)
	}
}
