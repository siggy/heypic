heypic.me, in go
===============

Local
-----

    $ go get github.com/araddon/httpstream
    $ go install
    $ websocketd --port=8080 --staticdir=public heypic \
      --ck=$TWITTER_CONSUMER_KEY \
      --cs=$TWITTER_CONSUMER_SECRET \
      --ot=$TWITTER_OAUTH_TOKEN \
      --os=$TWITTER_OAUTH_TOKEN_SECRET
