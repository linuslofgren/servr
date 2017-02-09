package main

import (
  "fmt"
  "net/http"
  _ "net"
  "os"
  "os/signal"
  _ "errors"
  _ "time"
  _ "strings"
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
  // close the socket before exit

  http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
    handleReq(w, r, sessionByUserIdStmt)
  })

  http.ListenAndServe(":8080", nil)
}

func handleExit(fun func ()) {
  sig := make(chan os.Signal, 1)
  signal.Notify(sig, os.Interrupt)
  <- sig // wait for signal
  fun() // call the function
  os.Exit(1) // exit the program
}
func handleReq(w http.ResponseWriter, r *http.Request,  stmt *sql.Stmt) {
  fmt.Println("REQ")
  // Set timeout to 10 sec
  err := r.ParseForm()
  if err != nil {
    // FIXME: handle error
    panic(err)
  }
  var user, password = r.Form.Get("user"), r.Form.Get("password")
  fmt.Println(user, password)

  var session sql.NullString
  var salt sql.NullString
  err = stmt.QueryRow(user).Scan(&session, &salt)
  if err != nil {
    if err == sql.ErrNoRows {
      fmt.Fprintf(w, "NOT FOUND")
    } else {
      fmt.Fprintf(w, "NOT GOOD")
    }
    panic(err)
  }
  if session.Valid && salt.Valid {
    fmt.Println(session.String)
    if session.String == password + salt.String {
      fmt.Fprintf(w, "CLEAR")
    } else {
      fmt.Fprintf(w, "WRONG")
    }
  } else {
    fmt.Println("NULL")
  }
}
