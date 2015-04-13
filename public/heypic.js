var map         = null;
var markers     = [];
var infoWindow  = new google.maps.InfoWindow({maxWidth: 200});
var MAX_MARKERS = 100;
var INTERVAL    = 2000;

function initialize() {
  map = new google.maps.Map(document.getElementById('map-canvas'), {
    center: { lat: 30, lng: 0},
    zoom: 2
  });

  var interval = setInterval(
    function() {
      if (markers.length > 0) {
        google.maps.event.trigger(markers[markers.length - 1], 'click');
      }
    },
    INTERVAL
  );

  google.maps.event.addListener(map, 'click', function() {
    clearInterval(interval);
  });
}
google.maps.event.addDomListener(window, 'load', initialize);

var ws = new WebSocket("ws://localhost:8080");

ws.onopen = function(evt) {
  console.log("Connection open. Sending message...");
  ws.send("Hello WebSockets!");
};

ws.onmessage = function(evt) {
  var json = JSON.parse(evt.data);

  var marker = new google.maps.Marker({
    position: {lat: json.heypic.lat, lng: json.heypic.lon},
    map: map,
    animation: google.maps.Animation.DROP,
    title: json.tweet.text,
    icon: {
      url: json.tweet.user.profile_image_url_https,
      size: new google.maps.Size(20, 20),
      origin: new google.maps.Point(0,0),
      anchor: new google.maps.Point(0, 20)
    }
  });

  markers.push(marker);
  if (markers.length > MAX_MARKERS) {
    markers.shift().setMap(null);
  }

  var text = twttr.txt.autoLink(json.tweet.text, {urlEntities: json.tweet.entities.media});
  text = text.replace(/<a /g, '<a target="_blank" ');

  google.maps.event.addListener(marker, 'click', function() {
    infoWindow.setContent(
      '<a target="_blank" href=\"' + json.heypic.img_url + '\">' +
        '<img width="150" height="150" src=\"' + json.heypic.img_url +
          ':thumb\" alt="original image" title="original image">' +
      '</a>' +
      '<div>' + text + '</div>');
    infoWindow.open(map, marker);
  });
};

ws.onclose = function(evt) {
  console.log("Connection closed.");
};

ws.onerror = function(err) {
  console.log(err.name + " => " + err.message);
}
