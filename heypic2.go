package main

import (
  "fmt"
  wsd "github.com/joewalnes/websocketd/libwebsocketd"
  "net/http"
  "os"
  "time"
)

const MAXFORKS = 10

func main() {
  // A log scope allows you to customize the logging that websocketd performs.
  //You can provide your own log scope with a log func.
  logScope := wsd.RootLogScope(wsd.LogAccess, func(l *wsd.LogScope,
    level wsd.LogLevel, levelName string,
    category string, msg string, args ...interface{}) {
    fmt.Println(args...)
  })
  // Configuration options tell websocketd where to look for programs to
  // run as WebSockets.
  config := &wsd.Config{
    // ScriptDir:    "./ws-bin",
    // UsingScriptDir: true,
    StartupTime:  time.Now(),
    DevConsole:   true,
    // StaticDir:    "./public",
    CommandName:  "server",
  }
  // Register your route and handler.
  // os.ClearEnv();
  http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
    handler := http.StripPrefix("/", wsd.NewWebsocketdServer(config, logScope, MAXFORKS))
    handler.ServeHTTP(rw, req)
  })
  if err := http.ListenAndServe(fmt.Sprintf(":%d", 8080), nil); err != nil {
    fmt.Println("could not start server!", err)
    os.Exit(1)
  }
}