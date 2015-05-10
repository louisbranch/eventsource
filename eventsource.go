package eventsource

import (
  "net/http"
)

var retry []byte = []byte("retry: 2000\n")

func setHeader (res http.ResponseWriter, req *http.Request) {
  origin := req.Header.Get("origin")
  header := res.Header()
  header.Set("Access-Control-Allow-Origin", origin)
  header.Set("Access-Control-Allow-Credentials", "true")
  header.Set("Content-Type", "text/event-stream")
  header.Set("Cache-Control", "no-cache")
  header.Set("Connection", "keep-alive")
}

func Handler (res http.ResponseWriter, req *http.Request) {
  setHeader(res, req)

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

  _, err = conn.Write(retry)

  if err != nil {
    conn.Close()
  }
}
