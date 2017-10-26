//Rewrite of Tom Steele's node.JS version

package main

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"

	"io/ioutil"

	"golang.org/x/net/html/charset"
)

//To-Do
//Add Signal bleed based signal strength

//Type Definitions and variables

//tpl defines the HTML template
const tpl = `
<html>
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
        <div class="col-xs-1">
          <button type="button" id="editButton" class="btn btn-primary" data-toggle="button">Editable</button>
        </div>
      </div>
    </div>
  </body>
<script type="text/javascript"
  src="https://maps.googleapis.com/maps/api/js?libraries=visualization">
</script>
<script>
var heatMapData = {{.Heatmap}};
var overlayCoords = {{.Paths}};
var overlayDrive = {{.Drive}};

var map = new google.maps.Map(document.getElementById('map'), {
  zoom: 17,
  center: {lat: {{.Lat}}, lng: {{.Lng}}},
  mapTypeId: 'satellite'
});

var heatmap = new google.maps.visualization.HeatmapLayer({
  data: heatMapData
});

var polygon =  new google.maps.Polygon({
          paths: overlayCoords,
          strokeColor: '#3366FF',
          strokeOpacity: 0.8,
          strokeWeight: 2,
          fillColor: '#3366FF',
          fillOpacity: 0.35
        });

var drive =  new google.maps.Polygon({
          paths: overlayDrive,
          strokeColor: '#3366FF',
          strokeOpacity: 1.0,
          strokeWeight: 2,
          fillOpacity: 0.0
        });

function toggleHeatmap() {
        heatmap.setMap(heatmap.getMap() ? null : map);
}
function toggleOverlay() {
	polygon.setMap(polygon.getMap() ? null : map)
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
	X, Y float64
	Dbm  int
}

//Page Holds the Values for html template
type Page struct {
	Lat        float64
	Lng        float64
	Heatmap    template.JS
	Paths      template.JS
	Drive      template.JS
	PathLength int
}

//GPSRun defines a struct to hold the top level of the
//xml tree
type GPSRun struct {
	XMLName    xml.Name   `xml:"gps-run"`
	GPSVersion string     `xml:"gps-version,attr"`
	StartTime  string     `xml:"start-time,attr"`
	GPSPoints  []GPSPoint `xml:"gps-point"`
}

//GPSPoint defines a struct to hold the values
//of gps-point
type GPSPoint struct {
	XMLName   xml.Name `xml:"gps-point"`
	Bssid     string   `xml:"bssid,attr"`
	Lat       float64  `xml:"lat,attr"`
	Lon       float64  `xml:"lon,attr"`
	Source    string   `xml:"source,attr"`
	TimeSec   int      `xml:"time-sec,attr"`
	TimeUSec  int      `xml:"time-usec,attr"`
	Spd       float64  `xml:"spd,attr"`
	Heading   float64  `xml:"heading,attr"`
	Fix       int      `xml:"fix,attr"`
	Alt       float64  `xml:"alt,attr"`
	SignalDbm int      `xml:"signal_dbm,attr"`
	NoiseDbm  int      `xml:"noise_dbm,attr"`
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

//parseXML parses the specified XML file and returns a GPSRun struct
//containing its values
func parseXML(file string) (target GPSRun) {
	xmlFile, err := os.Open(file)
	if err != nil {
		panic("Ensure the GPSXML file exists")
	}
	defer xmlFile.Close()
	var xmlData io.Reader = bufio.NewReader(xmlFile)
	decoder := xml.NewDecoder(xmlData)
	decoder.CharsetReader = charset.NewReaderLabel
	err = decoder.Decode(&target)
	checkError(err)
	return
}

//parseCoords parses a GPSPoint array and puts them into a Point
//struct that is then placed into a Points struct
func processCoords(gpspoints []GPSPoint) (points Points) {
	for i := 0; i < len(gpspoints); i++ {
		points = append(points, Point{gpspoints[i].Lon, gpspoints[i].Lat, gpspoints[i].SignalDbm})
	}
	return
}

//filterBSSID returns all GPSPoint structs that have a particular bssid field
func filterBSSID(gpsPoints []GPSPoint, bssid []string) (filteredPoints []GPSPoint) {
	for i := 0; i < len(gpsPoints); i++ {
		for n := 0; n < len(bssid); n++ {
			if gpsPoints[i].Bssid == bssid[n] {
				filteredPoints = append(filteredPoints, gpsPoints[i])
			}
		}
	}
	if len(filteredPoints) == 0 {
		log.Fatal("Your BSSID was not found in the file")
	}
	return
}

//Swap swaps the location of two Point structs in a Points struct
func (points Points) Swap(i, j int) {
	points[i], points[j] = points[j], points[i]
}

//Len is custom length definition for Points
func (points Points) Len() int {
	return len(points)
}

//Less sorts Points by x and, if equal, by y
func (points Points) Less(i, j int) bool {
	if points[i].X == points[j].X {
		return points[i].Y < points[j].Y
	}
	return points[i].X < points[j].X
}

//crossProduct returns the modulo (and sign) of the cross product between vetors OA and OB
func crossProduct(A, B, O Point) float64 {
	return (A.X-O.X)*(B.Y-O.Y) - (A.Y-O.Y)*(B.X-O.X)
}

// findConvexHull returns a slice of Point with a convex hull
// it is counterclockwise and starts and ends at the same point
func findConvexHull(points Points) Points {
	n := len(points)  // number of points to find convex hull
	var result Points // final result
	count := 0        // size of our convex hull (number of points added)
	// lets sort our points by x and if equal by y
	sort.Sort(points)
	if n == 0 {
		return result
	}
	// add the first element:
	result = append(result, points[0])
	count++
	//find the lower hull
	for i := 1; i < n; i++ {
		// remove points which are not part of the lower hull
		for count > 1 && crossProduct(result[count-2], result[count-1], points[i]) < 0.00000000000000001 {
			count--
			result = result[:count]
		}
		// add a new better point than the removed ones
		result = append(result, points[i])
		count++
	}
	count0 := count // our base counter for the upper hull
	// find the upper hull
	for i := n - 2; i >= 0; i-- {
		// remove points which are not part of the upper hull
		for count-count0 > 0 && crossProduct(result[count-2], result[count-1], points[i]) < 0.00000000000000001 {
			count--
			result = result[:count]
		}
		// add a new better point than the removed ones
		result = append(result, points[i])
		count++
	}
	return result
}

//parseBssid takes a filename or comma seperated list of BSSIDs
//and outputs an array containing the parsed BSSIDs
func parseBssid(bssids string) []string {
	var (
		bssidSlice     []string
		tempBssidSlice []string
	)
	r, err := regexp.Compile("(([a-zA-Z0-9]{2}:)){5}[a-zA-Z0-9]{2}")
	checkError(err)
	file, err := os.Open(bssids)
	if err == nil {
		defer file.Close()
		var lines []string
		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		bssidSlice = lines
	} else {
		bssidSlice = strings.Split(bssids, ",")
	}
	for i := 0; i < len(bssidSlice); i++ {
		if r.MatchString(bssidSlice[i]) {
			tempBssidSlice = append(tempBssidSlice, strings.ToUpper(bssidSlice[i]))
		}
	}
	if len(tempBssidSlice) == 0 {
		log.Fatal("Looks like you didn't have any correctly formatted SSIDs")
	}
	return tempBssidSlice
}

//genHeatmap generates a string to insert a list of heatmap objects
//into the html template

//populateTemplate populates the html template
func populateTemplate(points Points, allPoints Points, gpsPoints []GPSPoint) []byte {
	var page Page
	var heatmap string
	var pathsData string
	var driveData string
	var tplBuffer bytes.Buffer
	for i := 0; i < len(points); i++ {
		pathsData += fmt.Sprintf("(new google.maps.LatLng(%g, %g)), ", points[i].Y, points[i].X)
	}
	for i := 0; i < len(gpsPoints); i++ {
		driveData += fmt.Sprintf("(new google.maps.LatLng(%g, %g)), ", gpsPoints[i].Lat, gpsPoints[i].Lon)
	}
	for n := 0; n < len(allPoints); n++ {
		heatmap += fmt.Sprintf("{location: new google.maps.LatLng(%g, %g), weight: %f}, ", allPoints[n].Y, allPoints[n].X, (float64(allPoints[n].Dbm)/10.0)+9.0)
	}
	page.Lat = points[0].Y
	page.Lng = points[0].X
	page.PathLength = len(pathsData)
	page.Paths = template.JS("[" + pathsData[:len(pathsData)-2] + "]")
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
	var gpsxmlFile = flag.String("f", "", "Gpsxml input file")
	var bssid = flag.String("b", "", "File or comma seperated list of bssids")
	var outFile = flag.String("o", "", "Html Output file")
	flag.Parse()
	if !flag.Parsed() || !(flag.NFlag() == 3) {
		fmt.Println("Usage: warmap -f <Kismet gpsxml file> -b <File or List of BSSIDs> -o <HTML output file>")
		os.Exit(1)
	}
	xmlTree := parseXML(*gpsxmlFile)
	bssidData := parseBssid(*bssid)
	filteredPoints := filterBSSID(xmlTree.GPSPoints, bssidData)
	parsedCoords := processCoords(filteredPoints)
	templateBuffer := populateTemplate(findConvexHull(parsedCoords), parsedCoords, xmlTree.GPSPoints)
	ioutil.WriteFile(*outFile, templateBuffer, 0644)
}
