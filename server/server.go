package main

import (
  "fmt"
  _ "net/http"
  "net"
  "os"
  "os/signal"
  "errors"
  "time"
  "strings"
  "database/sql"
  _ "github.com/go-sql-driver/mysql"
)

const (
  sockettype = "unix"
)



func main() {
  // mysql conn
  db, err := sql.Open("mysql", os.Getenv("SQL_LOGIN"))
  if err != nil {
    // This should not happen so a panic is appropriate
    panic(err)
  }

  sessionByUserIdStmt, err := db.Prepare("SELECT Sessions.Session, Sessions.Salt FROM Sessions JOIN Users ON Users.Id=Sessions.User_Id WHERE Users.Name=?;")
  if err != nil {
    // This should not happen so a panic is appropriate
    panic(err)
  }


  // setup socket
  socketAddr, err := net.ResolveUnixAddr(sockettype, "/tmp/server")
  if err != nil {
    // This should not happen so a panic is appropriate
    panic(err)
  }
  var listner *net.UnixListener
  // FIXME: DO...WHILE
  for err = errors.New("");err != nil; {
    listner, err = connectSocket(socketAddr)
    if err != nil {
      fmt.Println("A SOCKET CONNECTION COULD NOT BE ESTABLISHED.")
      os.Remove(socketAddr.String())
    } else {
      fmt.Println("CONNECTED TO SOCKET")
      break
    }
  }


  // close the socket before exit
  go handleExit(func (){listner.Close()})

  for {
    conn, err := listner.AcceptUnix()
    if err != nil {
      fmt.Println(err)
      conn.Close()
      continue
    }
    go func() {
      defer func() {
        if r := recover(); r != nil {
          fmt.Println("Panic in connection handler:", r)
        }
      }()
      handleConn(conn, sessionByUserIdStmt)
    }()
  }

}

func connectSocket(socket *net.UnixAddr) (*net.UnixListener, error) {
  return net.ListenUnix(socket.Network(), socket)
}

func handleExit(fun func ()) {
  sig := make(chan os.Signal, 1)
  signal.Notify(sig, os.Interrupt)
  <- sig // wait for signal
  fun() // call the function
  os.Exit(1) // exit the program
}

func handleConn (conn *net.UnixConn, stmt *sql.Stmt) {
  // Close the connection
  defer func(){conn.Write([]byte("EOF"));conn.Close()}()
  // Set timeout to 10 sec
  conn.SetReadDeadline(time.Now().Add(time.Duration(10)*time.Second))
  // Wait to for a request on the socket
  var buf = make([]byte, 1024)
  n, err := conn.Read(buf) // Read
  if err != nil {
    panic(err)
  }
  buf = buf[:n]

  // divide the request [request, user, password]
  var res [3]string
  for i := 0; i < 2; i++ {
    if index := strings.IndexRune(string(buf), ':'); index != -1 {
      res[i], buf = string(buf[:index]), buf[index+1:] // +1 to skip the following ':'
    } else {
      panic(fmt.Sprintf("Syntax error '%s'", string(buf)))
      return
    }
  }
  res[2] = string(buf)
  var request, user, password = res[0], res[1], res[2]
  fmt.Println(request, user, password)

  var session sql.NullString
  var salt sql.NullString
  err = stmt.QueryRow(user).Scan(&session, &salt)
  if err != nil {
    if err == sql.ErrNoRows {
      conn.Write([]byte("NOT FOUND"))
    } else {
      conn.Write([]byte("NOT GOOD"))
    }
    panic(err)
  }
  if session.Valid && salt.Valid {
    fmt.Println(session.String)
    if session.String == password + salt.String {
      conn.Write([]byte("CLEAR"))
    }
  } else {
    fmt.Println("NULL")
  }
}
