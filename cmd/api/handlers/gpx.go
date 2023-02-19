package handlers

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
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
	FileName string `json:"fileName"`
	RouteID  int    `json:"routeID"`
}

type GMapToGPXResponse struct {
	URL string `json:"url"`
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

type ResponseData struct {
	Data  any    `json:"data"`
	Error string `json:"error,omitempty"`
}

func (h *Handlers) ConvertGMAPToGPX(w http.ResponseWriter, r *http.Request) http.Handler {
	w.Header().Set("Content-Type", "application/json")

	var routeContext *GMapToGPXRequest
	err := json.NewDecoder(r.Body).Decode(&routeContext)
	if err != nil {
		return Error(fmt.Errorf("failed to decode request JSON: %q", err), http.StatusBadRequest)
	}

	if routeContext == nil {
		return Error(fmt.Errorf("invalid request payload"), http.StatusBadRequest)
	}
	gMapURL := GMapURL + "/gp/ajaxRoute/get"
	if routeContext.RouteID < 5000000 {
		return Error(fmt.Errorf("invalid route ID: %d, must be greater than 5000000", routeContext.RouteID), http.StatusBadRequest)
	}

	reqData := url.Values{"rId": {fmt.Sprint(routeContext.RouteID)}}.Encode()
	gMapReq, err := http.NewRequest("POST", gMapURL, strings.NewReader(reqData))
	if err != nil {
		return Error(fmt.Errorf("failed to create gMap api request: %q", err), http.StatusInternalServerError)
	}
	gMapReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(gMapReq)
	if err != nil {
		return Error(fmt.Errorf("failed gMap api request: %q", err), http.StatusInternalServerError)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			h.Log.Errorf("failed to close response body, %v", err)
		}
	}(resp.Body)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return Error(fmt.Errorf("failed to read the gMap response body: %q", err), http.StatusInternalServerError)
	}
	query, err := url.ParseQuery(string(respBody))
	if err != nil {
		return Error(fmt.Errorf("failed to parse query parameters: %q", err), http.StatusInternalServerError)
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
		return Error(fmt.Errorf("no route data for %v", routeContext.RouteID), http.StatusBadRequest)
	}
	for stringIndex := 0; stringIndex < len(polyStrings); stringIndex = stringIndex + 2 {
		latitude, _ := strconv.ParseFloat(polyStrings[stringIndex], 64)
		longitude, _ := strconv.ParseFloat(polyStrings[stringIndex+1], 64)
		resultGPX.Track.TrackSegment.TrackPoint = append(resultGPX.Track.TrackSegment.TrackPoint, trackPoint{
			Latitude:  latitude,
			Longitude: longitude,
		})
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

	elevations, err := h.Environment.GCP.MapsClient.Elevation(r.Context(), eleReq)
	if err != nil {
		return Error(fmt.Errorf("failed to fetch elevation data: %q", err), http.StatusInternalServerError)
	}

	for i, elevationResp := range elevations {
		if elevationResp.Elevation != 0 {
			resultGPX.Track.TrackSegment.TrackPoint[i].Elevation = elevationResp.Elevation
		}
	}

	gpxRes, err := xml.Marshal(resultGPX)
	if err != nil {
		return Error(fmt.Errorf("failed to marshal xml: %s", err), http.StatusInternalServerError)
	}
	key := data.CreateNewObjectKey(20)
	uUrl, err := h.Environment.GCP.GetSignedUploadURL(key)
	if err != nil {
		return Error(fmt.Errorf("failed to create upload request: %q", err), http.StatusInternalServerError)
	}

	uploadReq, err := http.NewRequest(http.MethodPut, *uUrl, bytes.NewReader(gpxRes))
	if err != nil {
		return Error(fmt.Errorf("failed to create put request"), http.StatusInternalServerError)
	}
	uploadReq.Header.Set("Content-Type", "binary/octet-stream")

	uploadRes, err := client.Do(uploadReq)
	if err != nil {
		return Error(fmt.Errorf("failed to upload payload: %s", err), http.StatusInternalServerError)
	}

	if uploadRes.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(uploadRes.Body)
		return Error(fmt.Errorf("failed to upload file: %q", errBody), http.StatusInternalServerError)
	}

	dUrl, err := h.Environment.GCP.GetSignedDownloadURL(key)
	if err != nil {
		return Error(fmt.Errorf("failed to get download url: %q", err), http.StatusInternalServerError)
	}

	if dUrl == nil {
		return Error(fmt.Errorf("failed to download GPX file"), http.StatusInternalServerError)
	}

	return JSON(ResponseData{Data: GMapToGPXResponse{URL: *dUrl}})
}
