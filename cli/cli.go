package main

import (
  _ "net"
  _ "os"
  "io"
  "fmt"
  "net/http"
  "net/url"
  "flag"
)

var (
  serversocket = "serversocket"
  sType = "unix" // socket type
)

func init() {
  // the sockets can be overridden by the -serversocket flags
  flag.StringVar(&serversocket, "serversocket",
    "/tmp/server", "path for the unix server socket file")
  flag.Parse()
}

func main() {
  res, _ := http.PostForm("http://127.0.0.1:8080/", url.Values{"user": {"Thim H"}, "password": {"SE"}})
  buf := make([]byte, 200)
  io.ReadFull(res.Body, buf)
  fmt.Printf("%s\n",buf)
}
