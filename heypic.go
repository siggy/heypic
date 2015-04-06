// based on:
// https://github.com/araddon/httpstream/blob/master/examples/twoauth.go

package main

// twitter oauth

import (
  "encoding/json"
  "fmt"
  "flag"
  "github.com/araddon/httpstream"
  "github.com/mrjones/oauth"
  "log"
  "os"
)

var (
  consumerKey      *string = flag.String("ck", "TWITTER_CONSUMER_KEY", "Consumer Key")
  consumerSecret   *string = flag.String("cs", "TWITTER_CONSUMER_SECRET", "Consumer Secret")
  oauthToken       *string = flag.String("ot", "TWITTER_OAUTH_TOKEN", "Oauth Token")
  oauthTokenSecret *string = flag.String("os", "TWITTER_OAUTH_TOKEN_SECRET", "OAuthTokenSecret")

  logLevel         *string = flag.String("logging", "debug", "Which log level: [debug,info,warn,error,fatal]")
)

func main() {

  flag.Parse()
  httpstream.SetLogger(log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile), *logLevel)

  stream := make(chan []byte, 1000)
  done := make(chan bool)

  httpstream.OauthCon = oauth.NewConsumer(
    *consumerKey,
    *consumerSecret,
    oauth.ServiceProvider{
      RequestTokenUrl:   "http://api.twitter.com/oauth/request_token",
      AuthorizeTokenUrl: "https://api.twitter.com/oauth/authorize",
      AccessTokenUrl:    "https://api.twitter.com/oauth/access_token",
    })

  at := oauth.AccessToken{
    Token:  *oauthToken,
    Secret: *oauthTokenSecret,
  }

  client := httpstream.NewOAuthClient(&at, httpstream.OnlyTweetsFilter(func(line []byte) {
    stream <- line
  }))

  err := client.Filter(nil, []string{"photo"}, nil, []string{"-180,-90,180,90"}, false, done)
  if err != nil {
    httpstream.Log(httpstream.ERROR, err.Error())
  } else {

    go func() {

      for tw := range stream {
        var f interface{}
        err := json.Unmarshal(tw, &f)
        if err != nil {
          httpstream.Log(httpstream.ERROR, err.Error())
        } else {
          tweet, ok := f.(map[string]interface{})
          if ok {
            user, ok := tweet["user"].(map[string]interface{})
            if ok {
              if tweet["geo"] != nil ||
                  tweet["place"] != nil ||
                  (user["location"] != nil && len(user["location"].(string)) != 0) {
                fmt.Println("geo:", tweet["id_str"])

                entities, ok := tweet["entities"].(map[string]interface{})
                if ok {
                  mediaArray, ok := entities["media"].([]interface{})
                  if ok && len(mediaArray) > 0 {
                    media, ok := mediaArray[0].(map[string]interface{})
                    if ok {
                      fmt.Println(media["media_url_https"])
                    }
                  }
                }
              }
            }
          }
        }
      }
    }()
  }

  _ = <-done
}
