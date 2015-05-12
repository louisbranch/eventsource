package eventsource

import (
  "bytes"
  "fmt"
  "net/http"
)

const header string = `HTTP/1.1 200 OK
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
Access-Control-Allow-Credentials: true`

const body string = "\n\nretry: 2000\n"

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

  _, err = conn.Write(initialResponse(req))

  if err != nil {
    conn.Close()
  }
}

func initialResponse(req *http.Request) []byte {
  var buf bytes.Buffer
  buf.WriteString(header)
  if origin := req.Header.Get("origin"); origin != "" {
    cors:= fmt.Sprintf("Access-Control-Allow-Origin: %s", origin)
    buf.WriteString(cors)
  }
  buf.WriteString(body)
  return buf.Bytes()
}
