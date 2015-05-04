package main

import (
  "bufio"
  "fmt"
  "os"
)

func read() {
  for {
    reader := bufio.NewReader(os.Stdin)
    tweet, _ := reader.ReadString('\n')
    fmt.Println(string(tweet))
  }
}

func main() {
  done := make(chan bool)
  go read()

  _ = <-done
}