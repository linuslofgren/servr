package main

import (
  "net"
  _ "os"
  "fmt"
  _ "net/http"
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
  conn, err := net.DialUnix(sType, nil,
    &net.UnixAddr{serversocket, sType})
  if err != nil {
    panic(err)
  }

  // Write the request to the socket
  _, err = conn.Write([]byte(""/*"calendar:Thim --H:S"*/))
  if err != nil {
    panic(err)
  }

  // Read the response
  var buf [1024]byte
   _, err = conn.Read(buf[:]) // Read
  if err != nil {
    panic(err)
  }
  fmt.Printf("%s\n", string(buf[:]))

  // Close the connection
  conn.Close()
}
