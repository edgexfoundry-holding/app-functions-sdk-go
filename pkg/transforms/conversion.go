/*
Copyright 2017 The Nuclio Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// @nuclio.configure
// function.yaml:
//   spec:
//     platform:
//       attributes:
//         network: compose-files_edgex-network

package transforms

import (
	"encoding/json"
	"encoding/xml"

	"github.com/edgexfoundry/edgex-go/pkg/models"
)

// Conversion houses various built in conversion transforms (XML, JSON, CSV)
type Conversion struct {
}

// TransformToXML ...
func (f Conversion) TransformToXML(params ...interface{}) interface{} {
	if len(params) < 1 {
		return nil
	}
	println("TRANSFORMING TO XML")
	if result, ok := params[0].(*models.Event); ok {
		b, err := xml.Marshal(result)
		if err != nil {
			// LoggingClient.Error(fmt.Sprintf("Error parsing XML. Error: %s", err.Error()))
			return nil
		}
		// should we return a byte[] or string?
		// return b
		return string(b)
	}
	return nil
}

// TransformToJSON ...
func (f Conversion) TransformToJSON(params ...interface{}) interface{} {
	if len(params) < 1 {
		return nil
	}
	println("TRANSFORMING TO JSON")

	if result, ok := params[0].(*models.Event); ok {
		b, err := json.Marshal(result)
		if err != nil {
			// LoggingClient.Error(fmt.Sprintf("Error parsing XML. Error: %s", err.Error()))
			return nil
		}
		// should we return a byte[] or string?
		// return b
		return string(b)
	}
	return nil
}