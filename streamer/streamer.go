// based on:
// https://github.com/araddon/httpstream/blob/master/examples/twoauth.go

package main

// twitter oauth

import (
  "bufio"
  "encoding/json"
  "fmt"
  "flag"
  "github.com/araddon/httpstream"
  "github.com/mrjones/oauth"
  "log"
  "net"
  "net/rpc"
  "strings"
  "os"
)

var (
  consumerKey      *string = flag.String("ck", "TWITTER_CONSUMER_KEY", "Consumer Key")
  consumerSecret   *string = flag.String("cs", "TWITTER_CONSUMER_SECRET", "Consumer Secret")
  oauthToken       *string = flag.String("ot", "TWITTER_OAUTH_TOKEN", "Oauth Token")
  oauthTokenSecret *string = flag.String("os", "TWITTER_OAUTH_TOKEN_SECRET", "OAuthTokenSecret")

  logLevel         *string = flag.String("logging", "debug", "Which log level: [debug,info,warn,error,fatal]")
)

type heypic struct {
  Lat    float64  `json:"lat"`
  Lon    float64  `json:"lon"`
  ImgUrl string   `json:"img_url"`
}

func read(params chan<- string) {
  type keyword struct {
    key       string
    timestamp int
  }

  for {
    reader := bufio.NewReader(os.Stdin)
    text, _ := reader.ReadString('\n')
    params <- text
  }
}

func keywording(params <-chan string) {
  keywords := make([]string, 0)
  for param := range params {
    keywords = append(keywords, param)
    fmt.Println(strings.Join(keywords,","))
  }
}

func write(stream <-chan []byte) {
  for tw := range stream {
    var f interface{}
    err := json.Unmarshal(tw, &f)
    if err != nil {
      httpstream.Log(httpstream.ERROR, err.Error())
      continue
    }

    tweet, ok := f.(map[string]interface{})
    if !ok ||
      tweet["geo"] == nil && tweet["place"] == nil ||
      tweet["possibly_sensitive"].(bool) {
      continue
    }

    entities, ok := tweet["entities"].(map[string]interface{})
    if !ok {
      continue
    }

    mediaArray, ok := entities["media"].([]interface{})
    if !ok || len(mediaArray) == 0 {
      continue
    }

    media, ok := mediaArray[0].(map[string]interface{})
    if !ok {
      continue
    }

    var lat float64
    var lon float64

    // todo: use user.location
    if tweet["geo"] != nil {
      geo, ok := tweet["geo"].(map[string]interface{})
      if !ok {
        continue
      }

      coordinates, ok := geo["coordinates"].([]interface{})
      if !ok {
        continue
      }

      lat = coordinates[0].(float64)
      lon = coordinates[1].(float64)
    } else if tweet["place"] != nil {
      place, ok := tweet["place"].(map[string]interface{})
      if !ok {
        continue
      }

      boundingBox, ok := place["bounding_box"].(map[string]interface{})
      if !ok {
        continue
      }

      coordinates, ok := boundingBox["coordinates"].([]interface{})
      if !ok {
        continue
      }
      coordsArr, ok := coordinates[0].([]interface{})

      var minLat float64 = 90
      var minLon float64 = 180
      var maxLat float64 = -90
      var maxLon float64 = -180

      for _,coordInterface := range coordsArr {
        coord := coordInterface.([]interface{})
        lon = coord[0].(float64)
        lat = coord[1].(float64)

        if (lon < minLon) {
          minLon = lon
        }
        if (lon > maxLon) {
          maxLon = lon
        }

        if (lat < minLat) {
          minLat = lat
        }
        if (lat > maxLat) {
          maxLat = lat
        }
      }

      lon = (maxLon + minLon) / 2
      lat = (maxLat + minLat) / 2
    }

    output := make(map[string]interface{})
    output["tweet"] = tweet
    output["heypic"] = &heypic{
      Lat: lat,
      Lon: lon,
      ImgUrl: media["media_url_https"].(string),
    }
    jsonOutput, err := json.Marshal(output)
    if err != nil {
      httpstream.Log(httpstream.ERROR, err.Error())
      continue
    }

    fmt.Println(string(jsonOutput))
  }
}

func main() {

  rpc.Register(NewRPC())
  l, e := net.Listen("tcp", ":9876")
  if e != nil {
    log.Fatal("listen error:", e)
  }
  rpc.Accept(l)



  flag.Parse()
  httpstream.SetLogger(log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile), *logLevel)

  stream := make(chan []byte, 1000)
  params := make(chan string)
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

  go read(params)
  go keywording(params)

  // err := client.Filter(nil, []string{"photo"}, nil, []string{"-180,-90,180,90"}, false, done)
  err := client.Filter(nil, nil, nil, []string{"-180,-90,180,90"}, false, done)
  if err != nil {
    httpstream.Log(httpstream.ERROR, err.Error())
  } else {
    go write(stream)
  }

  _ = <-done
}
