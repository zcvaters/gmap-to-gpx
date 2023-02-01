package handlers

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/joomcode/errorx"
	"github.com/zcvaters/gmap-to-gpx/api/configure"
	. "github.com/zcvaters/gmap-to-gpx/cmd/configure/logging"
	"github.com/zcvaters/gmap-to-gpx/cmd/data"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const GMapURL string = "https://www.gmap-pedometer.com"

//const OpenElevationAPI string = "https://api.open-elevation.com/api/v1/lookup"

type GMapToGPXRequest struct {
	RouteID int `json:"routeID"`
}

type GMapToGPXResponse struct {
	PreSignedURL string `json:"url,omitempty"`
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

type OpenElevationRequest struct {
	Locations []struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"locations"`
}

type OpenElevationResponse struct {
	Results []struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Elevation float64 `json:"elevation"`
		Errors    string  `json:"error,omitempty"`
	} `json:"results"`
}

func ConvertGMAPToGPX(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		return errorx.InternalError.New("failed to read the request body: %s", err)
	}

	if len(reqBody) == 0 {
		return errorx.AssertionFailed.New("request missing body")
	}

	var routeContext GMapToGPXRequest
	err = json.Unmarshal(reqBody, &routeContext)
	if err != nil {
		return errorx.IllegalArgument.New("failed to unmarshal request body %s", err)
	}

	if routeContext.RouteID == 0 {
		return errorx.AssertionFailed.New("no routeID present: %d", routeContext.RouteID)
	}

	var gMapURL string
	if routeContext.RouteID < 5000000 {
		// TODO: Determine why/if route id's this low still exist.
		//gMapURL = GMapURL + "/getRoute.php"
		return errorx.AssertionFailed.New("invalid route ID, must be greater than 5000000.")
	} else {
		gMapURL = GMapURL + "/gp/ajaxRoute/get"
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
			Log.Errorf("failed to close response body, %v", err)
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

	//allPoints := new(OpenElevationRequest)
	//type openElev struct {
	//	Latitude  float64 `json:"latitude"`
	//	Longitude float64 `json:"longitude"`
	//}
	//for _, trackP := range resultGPX.Track.TrackSegment.TrackPoint {
	//	allPoints.Locations = append(allPoints.Locations, openElev{
	//		Latitude:  trackP.Latitude,
	//		Longitude: trackP.Longitude,
	//	})
	//}
	//
	//marshal, err := json.Marshal(allPoints)
	//if err != nil {
	//	return errorx.InternalError.New("failed to marshal json: %s", err)
	//}
	//
	//eleReq, err := http.NewRequest("POST", fmt.Sprintf(environment.Variables.OpenElevationAPIURL+"/api/v1/lookup"), bytes.NewBuffer(marshal))
	//if err != nil {
	//	return errorx.InternalError.New("failed to create request: %s", err)
	//}
	//eleReq.Header.Set("Content-Type", "application/json")
	//eleResp, err := client.Do(eleReq)
	//if err != nil {
	//	return errorx.InternalError.New("failed to make request to elevation api: %s", err)
	//}
	//if eleResp.StatusCode != http.StatusOK {
	//	return errorx.InternalError.New("elevation api status %s", eleResp.Status)
	//}
	//
	//defer func(Body io.ReadCloser) {
	//	err := Body.Close()
	//	if err != nil {
	//		Log.Errorf("failed to close eleResBody, %v", err)
	//	}
	//}(eleResp.Body)
	//
	//eleResBody, err := io.ReadAll(eleResp.Body)
	//if err != nil {
	//	return errorx.InternalError.New("failed to read the response elevation api response body: %s", err)
	//}
	//
	//var openElevationResp OpenElevationResponse
	//err = json.Unmarshal(eleResBody, &openElevationResp)
	//if err != nil {
	//	return errorx.InternalError.New("failed to unmarshal response from elevation api: %s", err)
	//}
	//
	//for i, elevationResp := range openElevationResp.Results {
	//	if elevationResp.Elevation != 0 {
	//		resultGPX.Track.TrackSegment.TrackPoint[i].Elevation = elevationResp.Elevation
	//	}
	//}

	//out, err := xml.MarshalIndent(resultGPX, " ", "  ")
	//if err != nil {
	//	return err
	//}

	gpxRes, err := xml.Marshal(resultGPX)
	if err != nil {
		return errorx.InternalError.New("failed to marshal xml: %s", err)
	}

	Log.Debug(string(gpxRes))

	if err := data.WriteJSONBytes(configure.ResponseData{
		Data:  gpxRes,
		Error: "",
	}, w); err != nil {
		return errorx.InternalError.New("failed to write json response: %s", err)
	}
	return nil
}
