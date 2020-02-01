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
    <link href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css" rel="stylesheet">
    <link href="all.css" rel="stylesheet">
    <style>
      #map {
        display: block;
        width: 100%;
        height: 700;
      }
    </style>
  </head>
  <body>
    <div class="container" style="margin-left: 30px; margin-right: 30px; width: 95%">
      <div class="row">
        <div class="col-xs-12">
				  <div class="row">
					  <p>
						  <div class="col-xs-2"></div>
						  <div class="col-xs-2" style="padding-bottom: 10px">Mapped Points: {{.PathLength}}</div>
						  <div class="col-xs-2">Strongest DB: {{.HighDB}}</div>
						  <div class="col-xs-2">Weakest DB: {{.LowDB}}</div>
					  </p>
          </div>
        </div>
      </div>
			<div class="row">
        <div class="col-xs-1">
            <div class="row">
              <div style="padding-top: 10px"></div>
              <p>
                <button title="Toggle Heatmap" onclick="toggleHeatmap()"><i class="fas fa-fire fa-2x" style="padding-top: 3px"></i></button>
                &nbsp;
                <button title="Add Ruler" onclick="addRuler()" style="width: 38px; height: 36; padding-top: 4; padding-left: 0"><i class="fas fa-ruler fa-2x"></i></i></button>
              </p>
            </div>
            <div class="row">
              <p>
                <button title="Toggle Overlay" onclick="toggleOverlay()" style="width: 38px; padding-left: 3"><i class="fab fa-battle-net fa-2x" style="padding-top: 3px"></i></button>
                &nbsp;
                <button title="Edit Overlay" onclick="overlayEditable()" style="width: 38px; height: 36; padding-top: 4; padding-left: 3"><i class="fas fa-edit fa-2x"></i></button>
              </p>
            </div>
            <div class="row">
              <p>
                <button title="Toggle Drive" onclick="toggleDrive()" style="width: 38px; height: 36; padding-top: 4; padding-left: 1"><i class="fas fa-car-crash fa-2x"></i></button>
                &nbsp;
                <button title="Edit Drive" onclick="driveEditable()" style="width: 38px; height: 36; padding-top: 4; padding-left: 3"><i class="fas fa-edit fa-2x"></i></button>
              </p>
            </div>
        </div>
				<div class="col-xs-10">
          <div style="height: 85%" id="map"></div>
        </div>
      </div>
    </div>
  </body>
  <script type="text/javascript" src="https://maps.googleapis.com/maps/api/js?key={{.Apikey}}&libraries=visualization"></script>
  <script type="text/javascript" src="labels.js"></script>
<script>
var heatMapData = {{.Heatmap}};
var overlayCoords = {{.ConvexHull}};
var overlayDrive = {{.Drive}};

var map = new google.maps.Map(document.getElementById('map'), {
  zoom: 16,
  center: {lat: {{.Lat}}, lng: {{.Lng}}},
  mapTypeId: 'satellite',
	mapTypeControlOptions: {style: google.maps.MapTypeControlStyle.DROPDOWN_MENU},
	controlSize: 30,
	streetViewControl: false
});

var heatmap = new google.maps.visualization.HeatmapLayer({
  data: heatMapData
});

var convexHull =  new google.maps.Polygon({
					paths: overlayCoords,
					editable: false,
          strokeColor: '#3366FF',
          strokeOpacity: 0.8,
          strokeWeight: 2,
          fillColor: '#3366FF',
          fillOpacity: 0.35
        });

var drive =  new google.maps.Polyline({
		path: overlayDrive,
		editable: false,
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
function driveEditable() {
	if (drive.editable) {
		drive.setEditable(false);
	}
	else {
		drive.setEditable(true);
	}
}
function overlayEditable() {
	if (convexHull.editable) {
		convexHull.setEditable(false);
	}
	else {
		convexHull.setEditable(true);
	}
}
var lines = new Array();
function addRuler() {
  var ruler1 = new google.maps.Marker({
    position: map.getCenter(),
    map: map,
    draggable: true
  });
  var ruler2 = new google.maps.Marker({
    position: map.getCenter(),
    map: map,
    draggable: true
  });
  var ruler1label = new Label({
    map: map
  });
  var ruler2label = new Label({
    map: map
  });
  ruler1label.bindTo('position', ruler1, 'position');
  ruler2label.bindTo('position', ruler2, 'position');
  var rulerpoly = new google.maps.Polyline({
    path: [ruler1.position, ruler2.position],
    strokeColor: "#FFFF00",
    strokeOpacity: .7,
    strokeWeight: 7
  });
  rulerpoly.setMap(map);
  ruler1label.set('text', distance(ruler1.getPosition().lat(), ruler1.getPosition().lng(), ruler2.getPosition().lat(), ruler2.getPosition().lng()));
  ruler2label.set('text', distance(ruler1.getPosition().lat(), ruler1.getPosition().lng(), ruler2.getPosition().lat(), ruler2.getPosition().lng()));
  google.maps.event.addListener(ruler1, 'drag', function() {
    rulerpoly.setPath([ruler1.getPosition(), ruler2.getPosition()]);
    ruler1label.set('text', distance(ruler1.getPosition().lat(), ruler1.getPosition().lng(), ruler2.getPosition().lat(), ruler2.getPosition().lng()));
    ruler2label.set('text', distance(ruler1.getPosition().lat(), ruler1.getPosition().lng(), ruler2.getPosition().lat(), ruler2.getPosition().lng()));
  });
  google.maps.event.addListener(ruler2, 'drag', function() {
    rulerpoly.setPath([ruler1.getPosition(), ruler2.getPosition()]);
    ruler1label.set('text', distance(ruler1.getPosition().lat(), ruler1.getPosition().lng(), ruler2.getPosition().lat(), ruler2.getPosition().lng()));
    ruler2label.set('text', distance(ruler1.getPosition().lat(), ruler1.getPosition().lng(), ruler2.getPosition().lat(), ruler2.getPosition().lng()));
  });

  google.maps.event.addListener(ruler1, 'dblclick', function() {
    ruler1.setMap(null);
    ruler2.setMap(null);
    ruler1label.setMap(null);
    ruler2label.setMap(null);
    rulerpoly.setMap(null);
  });

  google.maps.event.addListener(ruler2, 'dblclick', function() {
    ruler1.setMap(null);
    ruler2.setMap(null);
    ruler1label.setMap(null);
    ruler2label.setMap(null);
    rulerpoly.setMap(null);
  });

  // Add our new ruler to an array for later reference
  lines.push([ruler1, ruler2, ruler1label, ruler2label, rulerpoly]);
}

function distance(lat1, lon1, lat2, lon2) {
  var R = 3959; // Here's the right settings for miles and feet
  var dLat = (lat2 - lat1) * Math.PI / 180;
  var dLon = (lon2 - lon1) * Math.PI / 180;
  var a = Math.sin(dLat / 2) * Math.sin(dLat / 2) +
    Math.cos(lat1 * Math.PI / 180) * Math.cos(lat2 * Math.PI / 180) *
    Math.sin(dLon / 2) * Math.sin(dLon / 2);
  var c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
  var d = R * c;
  if (d > 1) return Math.round(d) + "mi";
  else if (d <= 1) return Math.round(d * 5280) + "ft";
  return d;
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
	HighDB     template.JS
	LowDB      template.JS
	PathLength int
	Apikey     *string
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

type KismetDatabase struct {
	Dot11_device struct {
		Dot11_device_advertisedSsidMap struct {
			Five3936447 struct {
				Dot11_advertisedssid_beacon                       int           `json:"dot11.advertisedssid.beacon"`
				Dot11_advertisedssid_beaconInfo                   string        `json:"dot11.advertisedssid.beacon_info"`
				Dot11_advertisedssid_beaconrate                   int           `json:"dot11.advertisedssid.beaconrate"`
				Dot11_advertisedssid_beaconsSec                   int           `json:"dot11.advertisedssid.beacons_sec"`
				Dot11_advertisedssid_channel                      string        `json:"dot11.advertisedssid.channel"`
				Dot11_advertisedssid_cloaked                      int           `json:"dot11.advertisedssid.cloaked"`
				Dot11_advertisedssid_cryptSet                     int           `json:"dot11.advertisedssid.crypt_set"`
				Dot11_advertisedssid_dot11dCountry                string        `json:"dot11.advertisedssid.dot11d_country"`
				Dot11_advertisedssid_dot11dList                   []interface{} `json:"dot11.advertisedssid.dot11d_list"`
				Dot11_advertisedssid_dot11eChannelUtilizationPerc float64       `json:"dot11.advertisedssid.dot11e_channel_utilization_perc"`
				Dot11_advertisedssid_dot11eQbss                   int           `json:"dot11.advertisedssid.dot11e_qbss"`
				Dot11_advertisedssid_dot11eQbssStations           int           `json:"dot11.advertisedssid.dot11e_qbss_stations"`
				Dot11_advertisedssid_dot11rMobility               int           `json:"dot11.advertisedssid.dot11r_mobility"`
				Dot11_advertisedssid_dot11rMobilityDomainID       int           `json:"dot11.advertisedssid.dot11r_mobility_domain_id"`
				Dot11_advertisedssid_firstTime                    int           `json:"dot11.advertisedssid.first_time"`
				Dot11_advertisedssid_htCenter1                    int           `json:"dot11.advertisedssid.ht_center_1"`
				Dot11_advertisedssid_htCenter2                    int           `json:"dot11.advertisedssid.ht_center_2"`
				Dot11_advertisedssid_htMode                       string        `json:"dot11.advertisedssid.ht_mode"`
				Dot11_advertisedssid_ietagChecksum                int           `json:"dot11.advertisedssid.ietag_checksum"`
				Dot11_advertisedssid_lastTime                     int           `json:"dot11.advertisedssid.last_time"`
				Dot11_advertisedssid_location                     struct {
					Kismet_common_location_avgAlt    int `json:"kismet.common.location.avg_alt"`
					Kismet_common_location_avgAltNum int `json:"kismet.common.location.avg_alt_num"`
					Kismet_common_location_avgLat    int `json:"kismet.common.location.avg_lat"`
					Kismet_common_location_avgLoc    struct {
						Kismet_common_location_alt      float64 `json:"kismet.common.location.alt"`
						Kismet_common_location_fix      int     `json:"kismet.common.location.fix"`
						Kismet_common_location_heading  float64 `json:"kismet.common.location.heading"`
						Kismet_common_location_lat      float64 `json:"kismet.common.location.lat"`
						Kismet_common_location_lon      float64 `json:"kismet.common.location.lon"`
						Kismet_common_location_speed    float64 `json:"kismet.common.location.speed"`
						Kismet_common_location_timeSec  int     `json:"kismet.common.location.time_sec"`
						Kismet_common_location_timeUsec int     `json:"kismet.common.location.time_usec"`
						Kismet_common_location_valid    int     `json:"kismet.common.location.valid"`
					} `json:"kismet.common.location.avg_loc"`
					Kismet_common_location_avgLon   int `json:"kismet.common.location.avg_lon"`
					Kismet_common_location_avgNum   int `json:"kismet.common.location.avg_num"`
					Kismet_common_location_locFix   int `json:"kismet.common.location.loc_fix"`
					Kismet_common_location_locValid int `json:"kismet.common.location.loc_valid"`
					Kismet_common_location_maxLoc   struct {
						Kismet_common_location_alt      float64 `json:"kismet.common.location.alt"`
						Kismet_common_location_fix      int     `json:"kismet.common.location.fix"`
						Kismet_common_location_heading  float64 `json:"kismet.common.location.heading"`
						Kismet_common_location_lat      float64 `json:"kismet.common.location.lat"`
						Kismet_common_location_lon      float64 `json:"kismet.common.location.lon"`
						Kismet_common_location_speed    float64 `json:"kismet.common.location.speed"`
						Kismet_common_location_timeSec  int     `json:"kismet.common.location.time_sec"`
						Kismet_common_location_timeUsec int     `json:"kismet.common.location.time_usec"`
						Kismet_common_location_valid    int     `json:"kismet.common.location.valid"`
					} `json:"kismet.common.location.max_loc"`
					Kismet_common_location_minLoc struct {
						Kismet_common_location_alt      float64 `json:"kismet.common.location.alt"`
						Kismet_common_location_fix      int     `json:"kismet.common.location.fix"`
						Kismet_common_location_heading  float64 `json:"kismet.common.location.heading"`
						Kismet_common_location_lat      float64 `json:"kismet.common.location.lat"`
						Kismet_common_location_lon      float64 `json:"kismet.common.location.lon"`
						Kismet_common_location_speed    float64 `json:"kismet.common.location.speed"`
						Kismet_common_location_timeSec  int     `json:"kismet.common.location.time_sec"`
						Kismet_common_location_timeUsec int     `json:"kismet.common.location.time_usec"`
						Kismet_common_location_valid    int     `json:"kismet.common.location.valid"`
					} `json:"kismet.common.location.min_loc"`
				} `json:"dot11.advertisedssid.location"`
				Dot11_advertisedssid_maxrate         float64 `json:"dot11.advertisedssid.maxrate"`
				Dot11_advertisedssid_probeResponse   int     `json:"dot11.advertisedssid.probe_response"`
				Dot11_advertisedssid_ssid            string  `json:"dot11.advertisedssid.ssid"`
				Dot11_advertisedssid_ssidlen         int     `json:"dot11.advertisedssid.ssidlen"`
				Dot11_advertisedssid_wpsManuf        string  `json:"dot11.advertisedssid.wps_manuf"`
				Dot11_advertisedssid_wpsModelName    string  `json:"dot11.advertisedssid.wps_model_name"`
				Dot11_advertisedssid_wpsModelNumber  string  `json:"dot11.advertisedssid.wps_model_number"`
				Dot11_advertisedssid_wpsSerialNumber string  `json:"dot11.advertisedssid.wps_serial_number"`
				Dot11_advertisedssid_wpsState        int     `json:"dot11.advertisedssid.wps_state"`
			} `json:"53936447"`
		} `json:"dot11.device.advertised_ssid_map"`
		Dot11_device_associatedClientMap      struct{}      `json:"dot11.device.associated_client_map"`
		Dot11_device_bssTimestamp             int           `json:"dot11.device.bss_timestamp"`
		Dot11_device_clientDisconnects        int           `json:"dot11.device.client_disconnects"`
		Dot11_device_clientMap                struct{}      `json:"dot11.device.client_map"`
		Dot11_device_datasize                 int           `json:"dot11.device.datasize"`
		Dot11_device_datasizeRetry            int           `json:"dot11.device.datasize_retry"`
		Dot11_device_lastBeaconTimestamp      int           `json:"dot11.device.last_beacon_timestamp"`
		Dot11_device_lastBeaconedSsid         string        `json:"dot11.device.last_beaconed_ssid"`
		Dot11_device_lastBeaconedSsidChecksum int           `json:"dot11.device.last_beaconed_ssid_checksum"`
		Dot11_device_lastBssid                string        `json:"dot11.device.last_bssid"`
		Dot11_device_lastProbedSsidCsum       int           `json:"dot11.device.last_probed_ssid_csum"`
		Dot11_device_lastSequence             int           `json:"dot11.device.last_sequence"`
		Dot11_device_numAdvertisedSsids       int           `json:"dot11.device.num_advertised_ssids"`
		Dot11_device_numAssociatedClients     int           `json:"dot11.device.num_associated_clients"`
		Dot11_device_numClientAps             int           `json:"dot11.device.num_client_aps"`
		Dot11_device_numFragments             int           `json:"dot11.device.num_fragments"`
		Dot11_device_numProbedSsids           int           `json:"dot11.device.num_probed_ssids"`
		Dot11_device_numRetries               int           `json:"dot11.device.num_retries"`
		Dot11_device_probedSsidMap            struct{}      `json:"dot11.device.probed_ssid_map"`
		Dot11_device_typeset                  int           `json:"dot11.device.typeset"`
		Dot11_device_wpaAnonceList            []interface{} `json:"dot11.device.wpa_anonce_list"`
		Dot11_device_wpaHandshakeList         []interface{} `json:"dot11.device.wpa_handshake_list"`
		Dot11_device_wpaNonceList             []interface{} `json:"dot11.device.wpa_nonce_list"`
		Dot11_device_wpaPresentHandshake      int           `json:"dot11.device.wpa_present_handshake"`
		Dot11_device_wpsM3Count               int           `json:"dot11.device.wps_m3_count"`
		Dot11_device_wpsM3Last                int           `json:"dot11.device.wps_m3_last"`
	} `json:"dot11.device"`
	Kismet_device_base_basicCryptSet int    `json:"kismet.device.base.basic_crypt_set"`
	Kismet_device_base_basicTypeSet  int    `json:"kismet.device.base.basic_type_set"`
	Kismet_device_base_channel       string `json:"kismet.device.base.channel"`
	Kismet_device_base_commonname    string `json:"kismet.device.base.commonname"`
	Kismet_device_base_crypt         string `json:"kismet.device.base.crypt"`
	Kismet_device_base_datasize      int    `json:"kismet.device.base.datasize"`
	Kismet_device_base_firstTime     int    `json:"kismet.device.base.first_time"`
	Kismet_device_base_freqKhzMap    struct {
		Two452000_000000 float64 `json:"2452000.000000"`
	} `json:"kismet.device.base.freq_khz_map"`
	Kismet_device_base_frequency float64 `json:"kismet.device.base.frequency"`
	Kismet_device_base_key       string  `json:"kismet.device.base.key"`
	Kismet_device_base_lastTime  int     `json:"kismet.device.base.last_time"`
	Kismet_device_base_location  struct {
		Kismet_common_location_avgAlt    int `json:"kismet.common.location.avg_alt"`
		Kismet_common_location_avgAltNum int `json:"kismet.common.location.avg_alt_num"`
		Kismet_common_location_avgLat    int `json:"kismet.common.location.avg_lat"`
		Kismet_common_location_avgLoc    struct {
			Kismet_common_location_alt      float64 `json:"kismet.common.location.alt"`
			Kismet_common_location_fix      int     `json:"kismet.common.location.fix"`
			Kismet_common_location_heading  float64 `json:"kismet.common.location.heading"`
			Kismet_common_location_lat      float64 `json:"kismet.common.location.lat"`
			Kismet_common_location_lon      float64 `json:"kismet.common.location.lon"`
			Kismet_common_location_speed    float64 `json:"kismet.common.location.speed"`
			Kismet_common_location_timeSec  int     `json:"kismet.common.location.time_sec"`
			Kismet_common_location_timeUsec int     `json:"kismet.common.location.time_usec"`
			Kismet_common_location_valid    int     `json:"kismet.common.location.valid"`
		} `json:"kismet.common.location.avg_loc"`
		Kismet_common_location_avgLon   int `json:"kismet.common.location.avg_lon"`
		Kismet_common_location_avgNum   int `json:"kismet.common.location.avg_num"`
		Kismet_common_location_locFix   int `json:"kismet.common.location.loc_fix"`
		Kismet_common_location_locValid int `json:"kismet.common.location.loc_valid"`
		Kismet_common_location_maxLoc   struct {
			Kismet_common_location_alt      float64 `json:"kismet.common.location.alt"`
			Kismet_common_location_fix      int     `json:"kismet.common.location.fix"`
			Kismet_common_location_heading  float64 `json:"kismet.common.location.heading"`
			Kismet_common_location_lat      float64 `json:"kismet.common.location.lat"`
			Kismet_common_location_lon      float64 `json:"kismet.common.location.lon"`
			Kismet_common_location_speed    float64 `json:"kismet.common.location.speed"`
			Kismet_common_location_timeSec  int     `json:"kismet.common.location.time_sec"`
			Kismet_common_location_timeUsec int     `json:"kismet.common.location.time_usec"`
			Kismet_common_location_valid    int     `json:"kismet.common.location.valid"`
		} `json:"kismet.common.location.max_loc"`
		Kismet_common_location_minLoc struct {
			Kismet_common_location_alt      float64 `json:"kismet.common.location.alt"`
			Kismet_common_location_fix      int     `json:"kismet.common.location.fix"`
			Kismet_common_location_heading  float64 `json:"kismet.common.location.heading"`
			Kismet_common_location_lat      float64 `json:"kismet.common.location.lat"`
			Kismet_common_location_lon      float64 `json:"kismet.common.location.lon"`
			Kismet_common_location_speed    float64 `json:"kismet.common.location.speed"`
			Kismet_common_location_timeSec  int     `json:"kismet.common.location.time_sec"`
			Kismet_common_location_timeUsec int     `json:"kismet.common.location.time_usec"`
			Kismet_common_location_valid    int     `json:"kismet.common.location.valid"`
		} `json:"kismet.common.location.min_loc"`
	} `json:"kismet.device.base.location"`
	Kismet_device_base_locationCloud struct {
		Kis_gps_rrd_lastSampleTs int `json:"kis.gps.rrd.last_sample_ts"`
		Kis_gps_rrd_samples100   []struct {
			Kismet_historic_location_alt       float64 `json:"kismet.historic.location.alt"`
			Kismet_historic_location_frequency int     `json:"kismet.historic.location.frequency"`
			Kismet_historic_location_heading   float64 `json:"kismet.historic.location.heading"`
			Kismet_historic_location_lat       float64 `json:"kismet.historic.location.lat"`
			Kismet_historic_location_lon       float64 `json:"kismet.historic.location.lon"`
			Kismet_historic_location_signal    int     `json:"kismet.historic.location.signal"`
			Kismet_historic_location_speed     float64 `json:"kismet.historic.location.speed"`
			Kismet_historic_location_timeSec   int     `json:"kismet.historic.location.time_sec"`
		} `json:"kis.gps.rrd.samples_100"`
		Kis_gps_rrd_samples10k []interface{} `json:"kis.gps.rrd.samples_10k"`
		Kis_gps_rrd_samples1m  []interface{} `json:"kis.gps.rrd.samples_1m"`
	} `json:"kismet.device.base.location_cloud"`
	Kismet_device_base_macaddr          string `json:"kismet.device.base.macaddr"`
	Kismet_device_base_manuf            string `json:"kismet.device.base.manuf"`
	Kismet_device_base_modTime          int    `json:"kismet.device.base.mod_time"`
	Kismet_device_base_name             string `json:"kismet.device.base.name"`
	Kismet_device_base_numAlerts        int    `json:"kismet.device.base.num_alerts"`
	Kismet_device_base_packets_crypt    int    `json:"kismet.device.base.packets.crypt"`
	Kismet_device_base_packets_data     int    `json:"kismet.device.base.packets.data"`
	Kismet_device_base_packets_error    int    `json:"kismet.device.base.packets.error"`
	Kismet_device_base_packets_filtered int    `json:"kismet.device.base.packets.filtered"`
	Kismet_device_base_packets_llc      int    `json:"kismet.device.base.packets.llc"`
	Kismet_device_base_packets_rrd      struct {
		Kismet_common_rrd_aggregator string    `json:"kismet.common.rrd.aggregator"`
		Kismet_common_rrd_blankVal   int       `json:"kismet.common.rrd.blank_val"`
		Kismet_common_rrd_dayVec     []int     `json:"kismet.common.rrd.day_vec"`
		Kismet_common_rrd_hourVec    []int     `json:"kismet.common.rrd.hour_vec"`
		Kismet_common_rrd_lastTime   int       `json:"kismet.common.rrd.last_time"`
		Kismet_common_rrd_minuteVec  []float64 `json:"kismet.common.rrd.minute_vec"`
	} `json:"kismet.device.base.packets.rrd"`
	Kismet_device_base_packets_rx    int    `json:"kismet.device.base.packets.rx"`
	Kismet_device_base_packets_total int    `json:"kismet.device.base.packets.total"`
	Kismet_device_base_packets_tx    int    `json:"kismet.device.base.packets.tx"`
	Kismet_device_base_phyname       string `json:"kismet.device.base.phyname"`
	Kismet_device_base_seenby        struct {
		_1975842976 struct {
			Kismet_common_seenby_firstTime  int `json:"kismet.common.seenby.first_time"`
			Kismet_common_seenby_freqKhzMap struct {
				Two452000_000000 float64 `json:"2452000.000000"`
			} `json:"kismet.common.seenby.freq_khz_map"`
			Kismet_common_seenby_lastTime   int `json:"kismet.common.seenby.last_time"`
			Kismet_common_seenby_numPackets int `json:"kismet.common.seenby.num_packets"`
			Kismet_common_seenby_signal     struct {
				Kismet_common_signal_carrierset  int     `json:"kismet.common.signal.carrierset"`
				Kismet_common_signal_encodingset int     `json:"kismet.common.signal.encodingset"`
				Kismet_common_signal_lastNoise   int     `json:"kismet.common.signal.last_noise"`
				Kismet_common_signal_lastSignal  int     `json:"kismet.common.signal.last_signal"`
				Kismet_common_signal_maxNoise    int     `json:"kismet.common.signal.max_noise"`
				Kismet_common_signal_maxSignal   int     `json:"kismet.common.signal.max_signal"`
				Kismet_common_signal_maxseenrate float64 `json:"kismet.common.signal.maxseenrate"`
				Kismet_common_signal_minNoise    int     `json:"kismet.common.signal.min_noise"`
				Kismet_common_signal_minSignal   int     `json:"kismet.common.signal.min_signal"`
				Kismet_common_signal_signalRrd   struct {
					Kismet_common_rrd_aggregator string    `json:"kismet.common.rrd.aggregator"`
					Kismet_common_rrd_blankVal   int       `json:"kismet.common.rrd.blank_val"`
					Kismet_common_rrd_lastTime   int       `json:"kismet.common.rrd.last_time"`
					Kismet_common_rrd_minuteVec  []float64 `json:"kismet.common.rrd.minute_vec"`
				} `json:"kismet.common.signal.signal_rrd"`
				Kismet_common_signal_type string `json:"kismet.common.signal.type"`
			} `json:"kismet.common.seenby.signal"`
			Kismet_common_seenby_uuid string `json:"kismet.common.seenby.uuid"`
		} `json:"-1975842976"`
	} `json:"kismet.device.base.seenby"`
	Kismet_device_base_serverUUID string `json:"kismet.device.base.server_uuid"`
	Kismet_device_base_signal     struct {
		Kismet_common_signal_carrierset  int     `json:"kismet.common.signal.carrierset"`
		Kismet_common_signal_encodingset int     `json:"kismet.common.signal.encodingset"`
		Kismet_common_signal_lastNoise   int     `json:"kismet.common.signal.last_noise"`
		Kismet_common_signal_lastSignal  int     `json:"kismet.common.signal.last_signal"`
		Kismet_common_signal_maxNoise    int     `json:"kismet.common.signal.max_noise"`
		Kismet_common_signal_maxSignal   int     `json:"kismet.common.signal.max_signal"`
		Kismet_common_signal_maxseenrate float64 `json:"kismet.common.signal.maxseenrate"`
		Kismet_common_signal_minNoise    int     `json:"kismet.common.signal.min_noise"`
		Kismet_common_signal_minSignal   int     `json:"kismet.common.signal.min_signal"`
		Kismet_common_signal_signalRrd   struct {
			Kismet_common_rrd_aggregator string    `json:"kismet.common.rrd.aggregator"`
			Kismet_common_rrd_blankVal   int       `json:"kismet.common.rrd.blank_val"`
			Kismet_common_rrd_lastTime   int       `json:"kismet.common.rrd.last_time"`
			Kismet_common_rrd_minuteVec  []float64 `json:"kismet.common.rrd.minute_vec"`
		} `json:"kismet.common.signal.signal_rrd"`
		Kismet_common_signal_type string `json:"kismet.common.signal.type"`
	} `json:"kismet.device.base.signal"`
	Kismet_device_base_tags struct{} `json:"kismet.device.base.tags"`
	Kismet_device_base_type string   `json:"kismet.device.base.type"`
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
func populateTemplate(points Points, apikey *string) []byte {
	var page Page
	var convexData string
	var heatmap string
	var driveData string
	high, low := 0, 0
	if apikey != nil {
		page.Apikey = apikey
	} else {
		page.Apikey = nil
	}
	var tplBuffer bytes.Buffer
	convexPoints := findConvexHull(points)
	for _, point := range convexPoints {
		convexData += fmt.Sprintf("(new google.maps.LatLng(%g, %g)), ", point.Y, point.X)
	}
	for _, point := range points {
		heatmap += fmt.Sprintf("{location: new google.maps.LatLng(%g, %g), weight: %f}, ", point.Y, point.X, (float64(point.Dbm)/10.0)+9.0)
		if high == 0 || low == 0 {
			high = point.Dbm
			low = point.Dbm
		}
		if point.Dbm > high {
			high = point.Dbm
		}
		if point.Dbm < low {
			low = point.Dbm
		}
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
	page.HighDB = template.JS(fmt.Sprintf("%v", high))
	page.LowDB = template.JS(fmt.Sprintf("%v", low))
	t, err := template.New("webpage").Parse(tpl)
	checkError(err)
	err = t.Execute(&tplBuffer, page)
	checkError(err)
	return tplBuffer.Bytes()
}

func printPoints(file string, points *Points) {
	list := make(map[string]int)
	var data bytes.Buffer
	if file == "" {
		return
	}

	for _, p := range *points {
		if _, ok := list[p.BSSID]; !ok {
			list[p.BSSID] = p.Dbm
		} else if db, ok := list[p.BSSID]; ok && db < p.Dbm {
			list[p.BSSID] = p.Dbm
		}
	}
	for k, v := range list {
		data.WriteString(fmt.Sprintf("%s,%v\n", k, v))
	}
	ioutil.WriteFile(file, data.Bytes(), 0644)
}

func main() {
	//Parse command line arguments
	var gpsFile = flag.String("f", "", "GPS input file")
	var bssid = flag.String("b", "", "File or comma seperated list of bssids")
	var outFile = flag.String("o", "", "Html Output file")
	var aerodump = flag.Bool("a", false, "Switch to specify aerodump gps file")
	var kismet = flag.Bool("k", false, "Switch to specify kismet database")
	var googleapi = flag.String("api", "", "Google Maps API key")
	var points = flag.String("p", "", "CSV Output file for reported BSSID values")
	flag.Parse()
	if !flag.Parsed() || !(flag.NFlag() >= 3) {
		fmt.Println("Usage: warmap -f <Kismet gpsxml file> -b <File or List of BSSIDs> -o <HTML output file>")
		os.Exit(1)
	}
	var gpsPoints Points
	if *aerodump {
		gpsPoints = parseAeroGPS(*gpsFile)
	} else if *kismet {
		bssids := parseBssid(*bssid)
		gpsPoints = parseKismet(*gpsFile, bssids)
	} else {
		bssids := parseBssid(*bssid)
		gpsPoints = parseXML(*gpsFile, bssids)
	}
	printPoints(*points, &gpsPoints)
	templateBuffer := populateTemplate(gpsPoints, googleapi)
	ioutil.WriteFile(*outFile, templateBuffer, 0644)
}
