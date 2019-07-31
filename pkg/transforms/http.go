//
// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package transforms

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/edgexfoundry/app-functions-sdk-go/pkg/util"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
)

// HTTPSender ...
type HTTPSender struct {
	URL      string
	MimeType string
}

// NewHTTPSender creates, initializes and returns a new instance of HTTPSender
func NewHTTPSender(url string, mimeType string) HTTPSender {
	return HTTPSender{
		URL:      url,
		MimeType: mimeType,
	}
}

// HTTPPost ...
func (sender HTTPSender) HTTPPost(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
	if len(params) < 1 {
		// We didn't receive a result
		return false, errors.New("No Data Received")
	}
	if sender.MimeType == "" {
		sender.MimeType = "application/json"
	}
	data, err := util.CoerceType(params[0])
	if err != nil {
		return false, err
	}

	edgexcontext.LoggingClient.Info("POSTing data")
	response, err := http.Post(sender.URL, sender.MimeType, bytes.NewReader(data))
	if err != nil {
		//LoggingClient.Error(err.Error())
		return false, err
	}
	defer response.Body.Close()
	edgexcontext.LoggingClient.Info(fmt.Sprintf("Response: %s", response.Status))
	edgexcontext.LoggingClient.Debug(fmt.Sprintf("Sent data: %s", string(data)))
	bodyBytes, errReadingBody := ioutil.ReadAll(response.Body)
	if errReadingBody != nil {
		return false, errReadingBody
	}

	edgexcontext.LoggingClient.Trace("Data exported", "Transport", "HTTP", clients.CorrelationHeader, edgexcontext.CorrelationID)

	// continues the pipeline if we get a 2xx response, stops pipeline if non-2xx response
	isSuccessfulPost := response.StatusCode >= 200 && response.StatusCode < 300
	if isSuccessfulPost == true {
		err = edgexcontext.MarkAsPushed()
		if err != nil {
			edgexcontext.LoggingClient.Error(err.Error())
		}
	}
	return isSuccessfulPost, bodyBytes

}