package helper

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"
)

// ErrMessage struct
type ErrMessage struct {
	Code       int    `json:"code"`
	Status     string `json:"status"`
	Message    string `json:"message"`
	Error      string `json:"error"`
	Requesturl string `json:"requesturl"`
}

//PostDataToWM method
func PostDataToWM(url string, msg []byte) (errmsg error) {

	var i = 0
	var retries = 2
	for i <= retries {
		errmsg = nil
		var err *ErrMessage
		response, terr := http.Post(url, "application/json", bytes.NewBuffer(msg))
		if terr != nil {
			var errOBJ ErrMessage
			err = &errOBJ
			err.Error = "Error in processing the request"
			err.Message = terr.Error()
			//err.Status = response.Status

		} else {
			// You can have your own error handling logic here
			data, terr := ioutil.ReadAll(response.Body)
			if terr != nil {
				var errOBJ ErrMessage
				err = &errOBJ
				err.Error = "Error reading the response"
				err.Message = "Error reading the response"
				err.Code = response.StatusCode
				err.Status = response.Status

			}

			if response.StatusCode != 200 {
				if json.Unmarshal(data, &err) != nil {
					var errOBJ ErrMessage
					err = &errOBJ
					err.Error = "Error response"
					err.Message = string(data)
					err.Code = response.StatusCode
					err.Status = response.Status
				}
			}

		}
		if err != nil {
			if err.Code == 502 && i < retries {
				i++
				time.Sleep(2 * time.Second)
				continue
			}
			b, err := json.Marshal(err)
			if err != nil {
				return err
			}
			errmsg = errors.New(string(b))
		}
		i = retries + 1
	}

	return errmsg
}
