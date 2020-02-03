package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

//parseKismet parses the specified Kismet file and returns a Points array with all the values
func parseKismet(file string, bssids []string, filter map[string]interface{}) (points Points) {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		panic("Ensure the Kismit file exists")
	}

	rows, err := db.Query("SELECT device FROM devices")
	if err != nil {
		panic("Kismet database corrupt")
	}
	defer rows.Close()
	var device string
	for rows.Next() {
		err = rows.Scan(&device)
		var data KismetDatabase
		if err := json.Unmarshal([]byte(device), &data); err != nil {
			fmt.Println(err)
		}
		//fmt.Println(data.Kismet_device_base_location.Kismet_common_location_avgLat)
		points = append(points, Point{data.Kismet_device_base_location.Kismet_common_location_avgLoc.Kismet_common_location_lon, data.Kismet_device_base_location.Kismet_common_location_avgLoc.Kismet_common_location_lat, data.Kismet_device_base_signal.Kismet_common_signal_minSignal, data.Kismet_device_base_macaddr})
	}
	points = filterBSSID(points, bssids, filter)
	return
}

//parseXML parses the specified XML file and returns a Points array with all the values
func parseXML(file string, bssids []string, filter map[string]interface{}) (points Points) {
	xmlFile, err := os.Open(file)
	if err != nil {
		panic("Ensure the GPSXML file exists")
	}
	defer xmlFile.Close()
	xmlScanner := bufio.NewScanner(xmlFile)
	for xmlScanner.Scan() {
		line := xmlScanner.Text()
		if strings.Contains(line, "<gps-point") {
			var gpsxml GPSXMLPoint
			xml.Unmarshal([]byte(line), &gpsxml)
			points = append(points, Point{gpsxml.Lon, gpsxml.Lat, gpsxml.SignalDbm, gpsxml.Bssid})
		}
	}
	points = filterBSSID(points, bssids, filter)
	return

}

// parse the json file for Aerodump
func parseAeroGPS(file string) (points Points) {
	gpsFile, err := os.Open(file)
	if err != nil {
		panic("Ensure the Aero GPS file exists")
	}
	defer gpsFile.Close()
	jsonScanner := bufio.NewScanner(gpsFile)
	for jsonScanner.Scan() {
		line := jsonScanner.Text()
		if strings.Contains(line, "class") {
			var gpsaero GPSAeroPoint
			json.Unmarshal([]byte(line), &gpsaero)
			points = append(points, Point{gpsaero.Lon, gpsaero.Lat, 0, ""})
		}
	}
	return
}

//filterBSSID returns all GPSXMLPoint structs that have a particular bssid field
func filterBSSID(points Points, bssid []string, filter map[string]interface{}) (filteredPoints Points) {
	for _, i := range points {
		for _, n := range bssid {
			if i.BSSID == n {
				if filter["filter"].(bool) {
					if (i.Dbm * -1) < filter["max"].(int) {
						filteredPoints = append(filteredPoints, i)
					}
				} else {
					filteredPoints = append(filteredPoints, i)
				}
			}
		}
	}
	if len(filteredPoints) == 0 {
		log.Fatal("Your BSSID was not found in the file")
	}
	return
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
