package handlers

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/joomcode/errorx"
	"github.com/zcvaters/gmap-to-gpx/cmd/configure/logging"
	"github.com/zcvaters/gmap-to-gpx/cmd/data"
	"googlemaps.github.io/maps"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const GMapURL string = "https://www.gmap-pedometer.com"

type GMapToGPXRequest struct {
	RouteID int `json:"routeID"`
}

type MapDataResp struct {
	CenterX         string `json:"centerX"`
	CenterY         string `json:"centerY"`
	ZoomLevel       string `json:"zl"`
	ZoomView        string `json:"zv"`
	Filter          string `json:"fl"`
	Polyline        string `json:"polyline"`
	Elevation       string `json:"elev"`
	ResourceID      string `json:"rId"`
	RandomValue     string `json:"rdm"`
	PointOfInterest string `json:"pta"`
	Distance        string `json:"distance"`
	ShowName        string `json:"show_name_description"`
	Name            string `json:"name"`
	Description     string `json:"description"`
}

type ResultingGPX struct {
	XMLName xml.Name `xml:"gpx"`
	Creator string   `xml:"creator,attr,omitempty"`
	Track   struct {
		Name         string `xml:"name,omitempty"`
		TrackSegment struct {
			TrackPoint []struct {
				Latitude  float64 `xml:"lat,attr"`
				Longitude float64 `xml:"lon,attr"`
				Elevation float64 `xml:"ele,omitempty"`
			} `xml:"trkpt"`
		} `xml:"trkseg"`
	} `xml:"trk"`
}

func (h *Handlers) ConvertGMAPToGPX(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	var routeContext *GMapToGPXRequest
	err := json.NewDecoder(r.Body).Decode(&routeContext)
	if err == io.EOF {
		return errorx.AssertionFailed.New("request contains no body contents")
	} else if err != nil || routeContext == nil {
		return errorx.AssertionFailed.New("request json malformed")
	}

	if routeContext.RouteID == 0 {
		return errorx.AssertionFailed.New("invalid route id. Cannot be: %d", routeContext.RouteID)
	}

	gMapURL := GMapURL + "/gp/ajaxRoute/get"
	if routeContext.RouteID < 5000000 {
		// TODO: Determine why/if route id's this low still exist.
		//gMapURL = GMapURL + "/getRoute.php"
		return errorx.AssertionFailed.New("invalid route ID, must be greater than 5000000.")
	}

	reqData := url.Values{"rId": {fmt.Sprint(routeContext.RouteID)}}.Encode()
	gMapReq, err := http.NewRequest("POST", gMapURL, strings.NewReader(reqData))
	if err != nil {
		return errorx.InternalError.New("failed to create gmap api request: %s", err)
	}
	gMapReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(gMapReq)
	if err != nil {
		return errorx.Decorate(err, "failed gMap api request")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logging.Log.Errorf("failed to close response body, %v", err)
		}
	}(resp.Body)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return errorx.InternalError.New("failed to read the gMap response body: %s", err)
	}
	query, err := url.ParseQuery(string(respBody))
	if err != nil {
		return errorx.InternalError.New("failed to parse query parameters: %s", err)
	}

	mapDataResp := MapDataResp{}
	mapDataResp.CenterX = query.Get("centerX")
	mapDataResp.CenterY = query.Get("centerY")
	mapDataResp.ZoomLevel = query.Get("zl")
	mapDataResp.ZoomView = query.Get("zv")
	mapDataResp.Filter = query.Get("fl")
	mapDataResp.Polyline = query.Get("polyline")
	mapDataResp.Elevation = query.Get("elev")
	mapDataResp.ResourceID = query.Get("rId")
	mapDataResp.RandomValue = query.Get("rdm")
	mapDataResp.PointOfInterest = query.Get("pta")
	mapDataResp.Distance = query.Get("distance")
	mapDataResp.ShowName = query.Get("show_name_description")
	mapDataResp.Name = query.Get("name")
	mapDataResp.Description = query.Get("description")

	resultGPX := &ResultingGPX{Creator: "gMapToGPX"}

	if mapDataResp.Name != "" {
		resultGPX.Track.Name = mapDataResp.Name
	} else {
		resultGPX.Track.Name = "gMapToGPX"
	}

	type trackPoint struct {
		Latitude  float64 `xml:"lat,attr"`
		Longitude float64 `xml:"lon,attr"`
		Elevation float64 `xml:"ele,omitempty"`
	}
	polyStrings := strings.Split(mapDataResp.Polyline, "a")
	if len(polyStrings) < 2 {
		return errorx.InternalError.New("received poly strings less than two points.")
	}
	for stringIndex := 0; stringIndex < len(polyStrings); stringIndex = stringIndex + 2 {
		latitude, _ := strconv.ParseFloat(polyStrings[stringIndex], 64)
		longitude, _ := strconv.ParseFloat(polyStrings[stringIndex+1], 64)
		resultGPX.Track.TrackSegment.TrackPoint = append(resultGPX.Track.TrackSegment.TrackPoint, trackPoint{
			Latitude:  latitude,
			Longitude: longitude,
		})
	}

	gClient, err := maps.NewClient(maps.WithAPIKey(h.Env.ElevationAPIKey))
	if err != nil {
		return errorx.InternalError.New("failed to configure maps api client: %s", err)
	}

	eleReq := &maps.ElevationRequest{
		Locations: nil,
	}

	for _, trackP := range resultGPX.Track.TrackSegment.TrackPoint {
		eleReq.Locations = append(eleReq.Locations, maps.LatLng{
			Lat: trackP.Latitude,
			Lng: trackP.Longitude,
		})
	}

	elevations, err := gClient.Elevation(r.Context(), eleReq)
	if err != nil {
		return errorx.InternalError.New("failed to fetch elevation data: %s", err)
	}

	for i, elevationResp := range elevations {
		if elevationResp.Elevation != 0 {
			resultGPX.Track.TrackSegment.TrackPoint[i].Elevation = elevationResp.Elevation
		}
	}

	gpxRes, err := xml.Marshal(resultGPX)
	if err != nil {
		return errorx.InternalError.New("failed to marshal xml: %s", err)
	}

	logging.Log.Debug(string(gpxRes))

	if err := data.WriteJSONBytes(data.ResponseData{
		Data:  "",
		Error: "",
	}, w); err != nil {
		return errorx.InternalError.New("failed to write json response: %s", err)
	}
	return nil
}
