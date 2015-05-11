package eventsource

import (
  "fmt"
  "net/http"
)

var header string = `HTTP/1.1 200 OK
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
Access-Control-Allow-Origin: %s
Access-Control-Allow-Credentials: true

retry: 2000

`
func Handler (res http.ResponseWriter, req *http.Request) {
  hj, ok := res.(http.Hijacker)
  if !ok {
    http.Error(res, "webserver doesn't support hijacking", http.StatusInternalServerError)
    return
  }

  conn, _, err := hj.Hijack()
  if err != nil {
    http.Error(res, err.Error(), http.StatusInternalServerError)
    return
  }

  origin := req.Header.Get("origin")
  h := fmt.Sprintf(header, origin)
  _, err = conn.Write([]byte(h))

  if err != nil {
    conn.Close()
  }
}
