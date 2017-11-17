//Rewrite of Tom Steele's node.JS version

package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
)

//To-Do
//Add Signal bleed based signal strength

//Type Definitions and variables

//tpl defines the HTML template
const tpl = `
<!DOCTYPE HTML SYSTEM>
<html xmlns="http://www.w3.org/1999/xhtml" xml:lang="en">
  <head>
    <title>WarMap</title>
    <script type="text/javascript" src="http://ajax.googleapis.com/ajax/libs/jquery/1.6.4/jquery.min.js"></script>
    <script type="text/javascript" src="http://hpneo.github.io/gmaps/gmaps.js"></script>
    <link href="http://netdna.bootstrapcdn.com/bootstrap/3.0.0/css/bootstrap.min.css" rel="stylesheet">
    <style>
      #map {
        display: block;
        width: 100%;
        height: 700;
      }
    </style>
  </head>
  <body>
    <div class="container" style="padding-top: 80px">
      <div class="row">
        <div class="col-xs-12">
          <p>Displaying {{.PathLength}} points.</p>
					<p><button onclick="toggleHeatmap()">Toggle Heatmap</button><p>
					<p><button onclick="toggleOverlay()">Toggle Overlay</button><p>
					<p><button onclick="toggleDrive()">Toggle Drive</button><p>
        </div>
      </div>
      <div class="row">
        <div class="col-xs-11">
          <div id="map"></div>
        </div>
      </div>
    </div>
  </body>
  <script type="text/javascript" src="https://maps.googleapis.com/maps/api/js?key=AIzaSyAruUeKxiz_cGDM5OPGWX6DlAhHCe1xRas&libraries=visualization">  
  </script>
<script>
var heatMapData = {{.Heatmap}};
var overlayCoords = {{.ConvexHull}};
var overlayDrive = {{.Drive}};

var map = new google.maps.Map(document.getElementById('map'), {
  zoom: 17,
  center: {lat: {{.Lat}}, lng: {{.Lng}}},
  mapTypeId: 'satellite'
});

var heatmap = new google.maps.visualization.HeatmapLayer({
  data: heatMapData
});

var convexHull =  new google.maps.Polygon({
          paths: overlayCoords,
          strokeColor: '#3366FF',
          strokeOpacity: 0.8,
          strokeWeight: 2,
          fillColor: '#3366FF',
          fillOpacity: 0.35
        });

var drive =  new google.maps.Polyline({
	    path: overlayDrive,
		geodesic: true,
		strokeColor: '#3366FF',
		strokeOpacity: 1.0,
	});

function toggleHeatmap() {
    heatmap.setMap(heatmap.getMap() ? null : map);
}
function toggleOverlay() {
	convexHull.setMap(convexHull.getMap() ? null : map)
}
function toggleDrive() {
	drive.setMap(drive.getMap() ? null : map)
}
</script>
</html>
`

////////////////////
//Type Definitions//
///////////////////

// Points defines a []Point array
type Points []Point

//Point holds X, Y coordinates
type Point struct {
	X, Y  float64
	Dbm   int
	BSSID string
}

//Page Holds the Values for html template
type Page struct {
	Lat        float64
	Lng        float64
	Heatmap    template.JS
	ConvexHull template.JS
	Drive      template.JS
	PathLength int
}

//GPSXMLPoint defines a struct to hold the values
//of the kismet generated gps
type GPSXMLPoint struct {
	Bssid     string  `xml:"bssid,attr"`
	Lat       float64 `xml:"lat,attr"`
	Lon       float64 `xml:"lon,attr"`
	Source    string  `xml:"source,attr"`
	TimeSec   int     `xml:"time-sec,attr"`
	TimeUSec  int     `xml:"time-usec,attr"`
	Spd       float64 `xml:"spd,attr"`
	Heading   float64 `xml:"heading,attr"`
	Fix       int     `xml:"fix,attr"`
	Alt       float64 `xml:"alt,attr"`
	SignalDbm int     `xml:"signal_dbm,attr"`
	NoiseDbm  int     `xml:"noise_dbm,attr"`
}

type GPSAeroPoint struct {
	Class  string
	Tag    string
	Device string
	Mode   int
	Time   string
	Ept    float64
	Lat    float64
	Lon    float64
	Alt    float64
	Epx    float64
	Epy    float64
	Epv    float64
	Track  float64
	Speed  float64
	Climb  float64
	Eps    float64
	Epc    float64
}

/////////////
//Functions//
////////////

//checkError is a generic error check function
func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

//populateTemplate populates the html template
func populateTemplate(points Points) []byte {
	var page Page
	var convexData string
	var heatmap string
	var driveData string
	var tplBuffer bytes.Buffer
	convexPoints := findConvexHull(points)
	for _, point := range convexPoints {
		convexData += fmt.Sprintf("(new google.maps.LatLng(%g, %g)), ", point.Y, point.X)
	}
	for _, point := range points {
		heatmap += fmt.Sprintf("{location: new google.maps.LatLng(%g, %g), weight: %f}, ", point.Y, point.X, (float64(point.Dbm)/10.0)+9.0)
	}
	for _, point := range points {
		driveData += fmt.Sprintf("(new google.maps.LatLng(%g, %g)), ", point.Y, point.X)
	}
	page.Lat = points[0].Y
	page.Lng = points[0].X
	page.PathLength = len(driveData)
	page.ConvexHull = template.JS("[" + convexData[:len(convexData)-2] + "]")
	page.Heatmap = template.JS("[" + heatmap[:len(heatmap)-2] + "]")
	page.Drive = template.JS("[" + driveData[:len(driveData)-2] + "]")
	t, err := template.New("webpage").Parse(tpl)
	checkError(err)
	err = t.Execute(&tplBuffer, page)
	checkError(err)
	return tplBuffer.Bytes()
}

func main() {
	//Parse command line arguments
	var gpsFile = flag.String("f", "", "GPS input file")
	var bssid = flag.String("b", "", "File or comma seperated list of bssids")
	var outFile = flag.String("o", "", "Html Output file")
	var aerodump = flag.Bool("a", false, "Switch to specify aerodump gps file")
	flag.Parse()
	if !flag.Parsed() || !(flag.NFlag() >= 3) {
		fmt.Println("Usage: warmap -f <Kismet gpsxml file> -b <File or List of BSSIDs> -o <HTML output file>")
		os.Exit(1)
	}
	var gpsPoints Points
	if *aerodump {
		gpsPoints = parseAeroGPS(*gpsFile)
	} else {
		bssids := parseBssid(*bssid)
		gpsPoints = parseXML(*gpsFile, bssids)
	}
	templateBuffer := populateTemplate(gpsPoints)
	ioutil.WriteFile(*outFile, templateBuffer, 0644)
}
