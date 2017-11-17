# warmap-go

Warmap takes a Kismet gpsxml or Aerodump gps file and a set of BSSIDs and creates a polygon of coordinates using the convex hull algorithm. This polygon is overlayed over a Google Maps generated map to show the coverage area of the specified BSSID. In addition, a heatmap is produced which indicates the intensity of the signal strength at all discovered points.

##Usage:##

go run warmap -f [Kismet gpsxml or Aerodump gps file] -a [boolean switch if youre using Aerodump output] -b [File of Comma-seperated List of BSSIDs] -o [HTML output file]

Binaries for all platforms can be found <a href="https://github.com/rmikehodges/warmap-go/releases">here</a>
