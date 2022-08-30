// based on:
// https://github.com/araddon/httpstream/blob/master/examples/twoauth.go

package main

// twitter oauth

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/araddon/httpstream"
	"github.com/mrjones/oauth"
)

var (
	consumerKey      *string = flag.String("ck", "TWITTER_CONSUMER_KEY", "Consumer Key")
	consumerSecret   *string = flag.String("cs", "TWITTER_CONSUMER_SECRET", "Consumer Secret")
	oauthToken       *string = flag.String("ot", "TWITTER_OAUTH_TOKEN", "Oauth Token")
	oauthTokenSecret *string = flag.String("os", "TWITTER_OAUTH_TOKEN_SECRET", "OAuthTokenSecret")

	logLevel *string = flag.String("logging", "debug", "Which log level: [debug,info,warn,error,fatal]")
)

type Heypic struct {
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	ImgUrl string  `json:"img_url"`
}

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

	// -119.655185,40.270889,-119.532276,40.329018

	// 40.247893,-119.717244,40.352106,-119.470052

	// -119.712587,40.391256,-118.218446,41.414651

	err := client.Filter(nil, []string{"burning man", "BurningMan", "BMwebcast"}, nil, []string{"-119.712587,40.391256,-118.218446,41.414651"}, false, done)
	// err := client.Filter(nil, nil, nil, []string{"-180,-90,180,90"}, false, done)
	if err != nil {
		httpstream.Log(httpstream.ERROR, err.Error())
	} else {

		go func() {

			for tw := range stream {
				var f interface{}
				err := json.Unmarshal(tw, &f)
				if err != nil {
					httpstream.Log(httpstream.ERROR, err.Error())
					continue
				}

				tweet, ok := f.(map[string]interface{})
				if !ok {
					continue
				}

				// fmt.Println(tweet)
				// os.Exit(0)

				if tweet["geo"] == nil && tweet["place"] == nil {
					continue
				}

				if tweet["possibly_sensitive"] != nil && tweet["possibly_sensitive"].(bool) {
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
					coordsArr, _ := coordinates[0].([]interface{})

					var minLat float64 = 90
					var minLon float64 = 180
					var maxLat float64 = -90
					var maxLon float64 = -180

					for _, coordInterface := range coordsArr {
						coord := coordInterface.([]interface{})
						lon = coord[0].(float64)
						lat = coord[1].(float64)

						if lon < minLon {
							minLon = lon
						}
						if lon > maxLon {
							maxLon = lon
						}

						if lat < minLat {
							minLat = lat
						}
						if lat > maxLat {
							maxLat = lat
						}
					}

					lon = (maxLon + minLon) / 2
					lat = (maxLat + minLat) / 2
				}

				output := make(map[string]interface{})
				output["tweet"] = tweet
				output["heypic"] = &Heypic{
					Lat:    lat,
					Lon:    lon,
					ImgUrl: media["media_url_https"].(string),
				}
				jsonOutput, err := json.Marshal(output)
				if err != nil {
					httpstream.Log(httpstream.ERROR, err.Error())
					continue
				}

				fmt.Println(string(jsonOutput))
			}
		}()
	}

	<-done
}
