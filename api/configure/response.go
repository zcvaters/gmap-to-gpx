package configure

import (
	"github.com/zcvaters/gmap-to-gpx/cmd/data"
)

type ResponseData struct {
	Data  any    `json:"data"`
	Error string `json:"errors"`
}

type Response struct {
	Data  any        `json:"data"`
	Error data.Error `json:"errors,omitempty"`
}

//func WriteAPIResponse(w http.ResponseWriter, res ResponseData) {
//	//var resErr data.Error
//	//if errVal, ok := data.Errors[res.Error]; ok {
//	//	resErr = errVal
//	//}
//	//statusCode := http.StatusOK
//	//if resErr.Code != 0 {
//	//	statusCode = resErr.StatusCode
//	//}
//	//w.WriteHeader(statusCode)
//	err := data.WriteJSONBytes(Response{
//		StatusCode: statusCode,
//		Data:       res.Data,
//		Error:      resErr,
//	}, w)
//	if err != nil {
//		Log.Errorf("failed to write JSON response: %v", err)
//	}
//}
