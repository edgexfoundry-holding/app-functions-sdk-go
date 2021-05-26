//
// Copyright (c) 2021 Intel Corporation
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

package app

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/transforms"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/util"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

const (
	ProfileNames        = "profilenames"
	DeviceNames         = "devicenames"
	SourceNames         = "sourcenames"
	ResourceNames       = "resourcenames"
	FilterOut           = "filterout"
	EncryptionKey       = "key"
	InitVector          = "initvector"
	Url                 = "url"
	ExportMethod        = "method"
	ExportMethodPost    = "post"
	ExportMethodPut     = "put"
	MimeType            = "mimetype"
	PersistOnError      = "persistonerror"
	ContinueOnSendError = "continueonsenderror"
	ReturnInputData     = "returninputdata"
	SkipVerify          = "skipverify"
	Qos                 = "qos"
	Retain              = "retain"
	AutoReconnect       = "autoreconnect"
	ConnectTimeout      = "connecttimeout"
	DeviceName          = "devicename"
	ReadingName         = "readingname"
	Rule                = "rule"
	BatchThreshold      = "batchthreshold"
	TimeInterval        = "timeinterval"
	HeaderName          = "headername"
	SecretPath          = "secretpath"
	SecretName          = "secretname"
	BrokerAddress       = "brokeraddress"
	ClientID            = "clientid"
	KeepAlive           = "keepalive"
	Topic               = "topic"
	TransformType       = "type"
	TransformXml        = "xml"
	TransformJson       = "json"
	AuthMode            = "authmode"
	Tags                = "tags"
	ResponseContentType = "responsecontenttype"
	Algorithm           = "algorithm"
	CompressGZIP        = "gzip"
	CompressZLIB        = "zlib"
	EncryptAES          = "aes"
	Mode                = "mode"
	BatchByCount        = "bycount"
	BatchByTime         = "bytime"
	BatchByTimeAndCount = "bytimecount"
)

type postPutParameters struct {
	method              string
	url                 string
	mimeType            string
	persistOnError      bool
	continueOnSendError bool
	returnInputData     bool
	headerName          string
	secretPath          string
	secretName          string
}

// Configurable contains the helper functions that return the function pointers for building the configurable function pipeline.
// They transform the parameters map from the Pipeline configuration in to the actual actual parameters required by the function.
type Configurable struct {
	lc logger.LoggingClient
}

// NewConfigurable returns a new instance of Configurable
func NewConfigurable(lc logger.LoggingClient) *Configurable {
	return &Configurable{
		lc: lc,
	}
}

// FilterByProfileName - Specify the profile names of interest to filter for data coming from certain sensors.
// The Filter by Profile Name transform looks at the Event in the message and looks at the profile names of interest list,
// provided by this function, and filters out those messages whose Event is for profile names not in the
// profile names of interest.
// This function will return an error and stop the pipeline if a non-edgex
// event is received or if no data is received.
// For example, data generated by a motor does not get passed to functions only interested in data from a thermostat.
// This function is a configuration function and returns a function pointer.
func (app *Configurable) FilterByProfileName(parameters map[string]string) interfaces.AppFunction {
	transform, ok := app.processFilterParameters("FilterByProfileName", parameters, ProfileNames)
	if !ok {
		return nil
	}

	return transform.FilterByProfileName
}

// FilterByDeviceName - Specify the device names of interest to filter for data coming from certain sensors.
// The Filter by Device Name transform looks at the Event in the message and looks at the device names of interest list,
// provided by this function, and filters out those messages whose Event is for device names not in the
// device names of interest.
// This function will return an error and stop the pipeline if a non-edgex
// event is received or if no data is received.
// For example, data generated by a motor does not get passed to functions only interested in data from a thermostat.
// This function is a configuration function and returns a function pointer.
func (app *Configurable) FilterByDeviceName(parameters map[string]string) interfaces.AppFunction {
	transform, ok := app.processFilterParameters("FilterByDeviceName", parameters, DeviceNames)
	if !ok {
		return nil
	}

	return transform.FilterByDeviceName
}

// FilterBySourceName - Specify the source names (resources and/or commands) of interest to filter for data coming from certain sensors.
// The Filter by Source Name transform looks at the Event in the message and looks at the source names of interest list,
// provided by this function, and filters out those messages whose Event is for source names not in the
// source names of interest.
// This function will return an error and stop the pipeline if a non-edgex
// event is received or if no data is received.
// For example, data generated by a motor does not get passed to functions only interested in data from a thermostat.
// This function is a configuration function and returns a function pointer.
func (app *Configurable) FilterBySourceName(parameters map[string]string) interfaces.AppFunction {
	transform, ok := app.processFilterParameters("FilterBySourceName", parameters, SourceNames)
	if !ok {
		return nil
	}

	return transform.FilterBySourceName
}

// FilterByResourceName - Specify the resource name of interest to filter for data from certain types of IoT objects,
// such as temperatures, motion, and so forth, that may come from an array of sensors or devices. The Filter by resource name assesses
// the data in each Event and Reading, and removes readings that have a resource name that is not in the list of
// resource names of interest for the application.
// This function will return an error and stop the pipeline if a non-edgex
// event is received or if no data is received.
// For example, pressure reading data does not go to functions only interested in motion data.
// This function is a configuration function and returns a function pointer.
func (app *Configurable) FilterByResourceName(parameters map[string]string) interfaces.AppFunction {
	transform, ok := app.processFilterParameters("FilterByResourceName", parameters, ResourceNames)
	if !ok {
		return nil
	}

	return transform.FilterByResourceName
}

// Transform transforms an EdgeX event to XML or JSON based on specified transform type.
// It will return an error and stop the pipeline if a non-edgex event is received or if no data is received.
// This function is a configuration function and returns a function pointer.
func (app *Configurable) Transform(parameters map[string]string) interfaces.AppFunction {
	transformType, ok := parameters[TransformType]
	if !ok {
		app.lc.Errorf("Could not find '%s' parameter for Transform", TransformType)
		return nil
	}

	transform := transforms.Conversion{}

	switch strings.ToLower(transformType) {
	case TransformXml:
		return transform.TransformToXML
	case TransformJson:
		return transform.TransformToJSON
	default:
		app.lc.Errorf(
			"Invalid transform type '%s'. Must be '%s' or '%s'",
			transformType,
			TransformXml,
			TransformJson)
		return nil
	}
}

// PushToCore pushes the provided value as an event to CoreData using the device name and reading name that have been set. If validation is turned on in
// CoreServices then your deviceName and readingName must exist in the CoreMetadata and be properly registered in EdgeX.
// This function is a configuration function and returns a function pointer.
func (app *Configurable) PushToCore(parameters map[string]string) interfaces.AppFunction {
	deviceName, ok := parameters[DeviceName]
	if !ok {
		app.lc.Error("Could not find " + DeviceName)
		return nil
	}
	readingName, ok := parameters[ReadingName]
	if !ok {
		app.lc.Error("Could not find " + ReadingName)
		return nil
	}
	deviceName = strings.TrimSpace(deviceName)
	readingName = strings.TrimSpace(readingName)
	transform := transforms.CoreData{
		DeviceName:  deviceName,
		ReadingName: readingName,
	}
	return transform.PushToCoreData
}

// Compress compresses data received as either a string,[]byte, or json.Marshaller using the specified algorithm (GZIP or ZLIB)
// and returns a base64 encoded string as a []byte.
// This function is a configuration function and returns a function pointer.
func (app *Configurable) Compress(parameters map[string]string) interfaces.AppFunction {
	algorithm, ok := parameters[Algorithm]
	if !ok {
		app.lc.Errorf("Could not find '%s' parameter for Compress", Algorithm)
		return nil
	}

	transform := transforms.Compression{}

	switch strings.ToLower(algorithm) {
	case CompressGZIP:
		return transform.CompressWithGZIP
	case CompressZLIB:
		return transform.CompressWithZLIB
	default:
		app.lc.Errorf(
			"Invalid compression algorithm '%s'. Must be '%s' or '%s'",
			algorithm,
			CompressGZIP,
			CompressZLIB)
		return nil
	}
}

// Encrypt encrypts either a string, []byte, or json.Marshaller type using specified encryption
// algorithm (AES only at this time). It will return a byte[] of the encrypted data.
// This function is a configuration function and returns a function pointer.
func (app *Configurable) Encrypt(parameters map[string]string) interfaces.AppFunction {
	algorithm, ok := parameters[Algorithm]
	if !ok {
		app.lc.Errorf("Could not find '%s' parameter for Encrypt", Algorithm)
		return nil
	}

	secretPath := parameters[SecretPath]
	secretName := parameters[SecretName]
	encryptionKey := parameters[EncryptionKey]

	// SecretPath & SecretName are optional if EncryptionKey specified
	// EncryptionKey is optional if SecretPath & SecretName are specified

	// If EncryptionKey not specified, then SecretPath & SecretName must be specified
	if len(encryptionKey) == 0 && (len(secretPath) == 0 || len(secretName) == 0) {
		app.lc.Errorf("Could not find '%s' or '%s' and '%s' in configuration", EncryptionKey, SecretPath, SecretName)
		return nil
	}

	// SecretPath & SecretName both must be specified it one of them is.
	if (len(secretPath) != 0 && len(secretName) == 0) || (len(secretPath) == 0 && len(secretName) != 0) {
		app.lc.Errorf("'%s' and '%s' both must be set in configuration", SecretPath, SecretName)
		return nil
	}

	initVector, ok := parameters[InitVector]
	if !ok {
		app.lc.Error("Could not find " + InitVector)
		return nil
	}

	transform := transforms.Encryption{
		EncryptionKey:        encryptionKey,
		InitializationVector: initVector,
		SecretPath:           secretPath,
		SecretName:           secretName,
	}

	switch strings.ToLower(algorithm) {
	case EncryptAES:
		return transform.EncryptWithAES
	default:
		app.lc.Errorf(
			"Invalid encryption algorithm '%s'. Must be '%s'",
			algorithm,
			EncryptAES)
		return nil
	}
}

// HTTPExport will send data from the previous function to the specified Endpoint via http POST or PUT. If no previous function exists,
// then the event that triggered the pipeline will be used. Passing an empty string to the mimetype
// method will default to application/json.
// This function is a configuration function and returns a function pointer.
func (app *Configurable) HTTPExport(parameters map[string]string) interfaces.AppFunction {
	params, err := app.processHttpExportParameters(parameters)
	if err != nil {
		app.lc.Error(err.Error())
		return nil
	}

	var transform transforms.HTTPSender
	if len(params.secretPath) != 0 {
		transform = transforms.NewHTTPSenderWithSecretHeader(
			params.url,
			params.mimeType,
			params.persistOnError,
			params.headerName,
			params.secretPath,
			params.secretName)
	} else {
		transform = transforms.NewHTTPSender(params.url, params.mimeType, params.persistOnError)
	}

	switch strings.ToLower(params.method) {
	case ExportMethodPost:
		return transform.HTTPPost
	case ExportMethodPut:
		return transform.HTTPPut
	default:
		app.lc.Errorf(
			"Invalid HTTPExport method of '%s'. Must be '%s' or '%s'",
			params.method,
			ExportMethodPost,
			ExportMethodPut)
		return nil
	}
}

//
// MQTTExport will send data from the previous function to the specified Endpoint via MQTT publish. If no previous function exists,
// then the event that triggered the pipeline will be used.
// This function is a configuration function and returns a function pointer.
func (app *Configurable) MQTTExport(parameters map[string]string) interfaces.AppFunction {
	var err error
	qos := 0
	retain := false
	autoReconnect := false
	skipCertVerify := false

	brokerAddress, ok := parameters[BrokerAddress]
	if !ok {
		app.lc.Error("Could not find " + BrokerAddress)
		return nil
	}
	topic, ok := parameters[Topic]
	if !ok {
		app.lc.Error("Could not find " + Topic)
		return nil
	}

	secretPath, ok := parameters[SecretPath]
	if !ok {
		app.lc.Error("Could not find " + SecretPath)
		return nil
	}
	authMode, ok := parameters[AuthMode]
	if !ok {
		app.lc.Error("Could not find " + AuthMode)
		return nil
	}
	clientID, ok := parameters[ClientID]
	if !ok {
		app.lc.Error("Could not find " + ClientID)
		return nil
	}
	qosVal, ok := parameters[Qos]
	if ok {
		qos, err = strconv.Atoi(qosVal)
		if err != nil {
			app.lc.Error("Unable to parse " + Qos + " value")
			return nil
		}
	}
	retainVal, ok := parameters[Retain]
	if ok {
		retain, err = strconv.ParseBool(retainVal)
		if err != nil {
			app.lc.Error("Unable to parse " + Retain + " value")
			return nil
		}
	}
	autoreconnectVal, ok := parameters[AutoReconnect]
	if ok {
		autoReconnect, err = strconv.ParseBool(autoreconnectVal)
		if err != nil {
			app.lc.Error("Unable to parse " + AutoReconnect + " value")
			return nil
		}
	}
	skipVerifyVal, ok := parameters[SkipVerify]
	if ok {
		skipCertVerify, err = strconv.ParseBool(skipVerifyVal)
		if err != nil {
			app.lc.Errorf("Could not parse '%s' to a bool for '%s' parameter: %s", skipVerifyVal, SkipVerify, err.Error())
			return nil
		}
	}

	// These are optional and blank values result in MQTT defaults being used.
	keepAlive := parameters[KeepAlive]
	connectTimeout := parameters[ConnectTimeout]

	mqttConfig := transforms.MQTTSecretConfig{
		Retain:         retain,
		SkipCertVerify: skipCertVerify,
		AutoReconnect:  autoReconnect,
		ConnectTimeout: connectTimeout,
		KeepAlive:      keepAlive,
		QoS:            byte(qos),
		BrokerAddress:  brokerAddress,
		ClientId:       clientID,
		SecretPath:     secretPath,
		Topic:          topic,
		AuthMode:       authMode,
	}
	// PersistOnError is optional and is false by default.
	persistOnError := false
	value, ok := parameters[PersistOnError]
	if ok {
		persistOnError, err = strconv.ParseBool(value)
		if err != nil {
			app.lc.Errorf("Could not parse '%s' to a bool for '%s' parameter: %s", value, PersistOnError, err.Error())
			return nil
		}
	}
	transform := transforms.NewMQTTSecretSender(mqttConfig, persistOnError)
	return transform.MQTTSend
}

// SetResponseData sets the response data to that passed in from the previous function and the response content type
// to that set in the ResponseContentType configuration parameter. It will return an error and stop the pipeline if
// data passed in is not of type []byte, string or json.Marshaller
// This function is a configuration function and returns a function pointer.
func (app *Configurable) SetResponseData(parameters map[string]string) interfaces.AppFunction {
	transform := transforms.ResponseData{}

	value, ok := parameters[ResponseContentType]
	if ok && len(value) > 0 {
		transform.ResponseContentType = value
	}

	return transform.SetResponseData
}

// Batch sets up Batching of events based on the specified mode parameter (BatchByCount, BatchByTime or BatchByTimeAndCount)
// and mode specific parameters.
// This function is a configuration function and returns a function pointer.
func (app *Configurable) Batch(parameters map[string]string) interfaces.AppFunction {
	mode, ok := parameters[Mode]
	if !ok {
		app.lc.Errorf("Could not find '%s' parameter for Batch", Mode)
		return nil
	}

	switch strings.ToLower(mode) {
	case BatchByCount:
		batchThreshold, ok := parameters[BatchThreshold]
		if !ok {
			app.lc.Errorf("Could not find '%s' parameter for BatchByCount", BatchThreshold)
			return nil
		}

		thresholdValue, err := strconv.Atoi(batchThreshold)
		if err != nil {
			app.lc.Errorf(
				"Could not parse '%s' to an int for '%s' parameter for BatchByCount: %s",
				batchThreshold, BatchThreshold, err.Error())
			return nil
		}

		transform, err := transforms.NewBatchByCount(thresholdValue)
		if err != nil {
			app.lc.Error(err.Error())
		}
		return transform.Batch

	case BatchByTime:
		timeInterval, ok := parameters[TimeInterval]
		if !ok {
			app.lc.Errorf("Could not find '%s' parameter for BatchByTime", TimeInterval)
			return nil
		}

		transform, err := transforms.NewBatchByTime(timeInterval)
		if err != nil {
			app.lc.Error(err.Error())
		}
		return transform.Batch

	case BatchByTimeAndCount:
		timeInterval, ok := parameters[TimeInterval]
		if !ok {
			app.lc.Error("Could not find " + TimeInterval)
			return nil
		}
		batchThreshold, ok := parameters[BatchThreshold]
		if !ok {
			app.lc.Error("Could not find " + BatchThreshold)
			return nil
		}
		thresholdValue, err := strconv.Atoi(batchThreshold)
		if err != nil {
			app.lc.Errorf("Could not parse '%s' to an int for '%s' parameter: %s", batchThreshold, BatchThreshold, err.Error())
		}
		transform, err := transforms.NewBatchByTimeAndCount(timeInterval, thresholdValue)
		if err != nil {
			app.lc.Error(err.Error())
		}
		return transform.Batch

	default:
		app.lc.Errorf(
			"Invalid batch mode '%s'. Must be '%s', '%s' or '%s'",
			mode,
			BatchByCount,
			BatchByTime,
			BatchByTimeAndCount)
		return nil
	}
}

// JSONLogic ...
func (app *Configurable) JSONLogic(parameters map[string]string) interfaces.AppFunction {
	rule, ok := parameters[Rule]
	if !ok {
		app.lc.Error("Could not find " + Rule)
		return nil
	}

	transform := transforms.NewJSONLogic(rule)
	return transform.Evaluate
}

// AddTags adds the configured list of tags to Events passed to the transform.
// This function is a configuration function and returns a function pointer.
func (app *Configurable) AddTags(parameters map[string]string) interfaces.AppFunction {
	tagsSpec, ok := parameters[Tags]
	if !ok {
		app.lc.Error(fmt.Sprintf("Could not find '%s' parameter", Tags))
		return nil
	}

	tagKeyValues := util.DeleteEmptyAndTrim(strings.FieldsFunc(tagsSpec, util.SplitComma))

	tags := make(map[string]string)
	for _, tag := range tagKeyValues {
		keyValue := util.DeleteEmptyAndTrim(strings.FieldsFunc(tag, util.SplitColon))
		if len(keyValue) != 2 {
			app.lc.Errorf("Bad Tags specification format. Expect comma separated list of 'key:value'. Got `%s`", tagsSpec)
			return nil
		}

		if len(keyValue[0]) == 0 {
			app.lc.Errorf("Tag key missing. Got '%s'", tag)
			return nil
		}
		if len(keyValue[1]) == 0 {
			app.lc.Errorf("Tag value missing. Got '%s'", tag)
			return nil
		}

		tags[keyValue[0]] = keyValue[1]
	}

	transform := transforms.NewTags(tags)
	return transform.AddTags
}

func (app *Configurable) processFilterParameters(
	funcName string,
	parameters map[string]string,
	paramName string) (*transforms.Filter, bool) {
	names, ok := parameters[paramName]
	if !ok {
		app.lc.Errorf("Could not find '%s' parameter for %s", paramName, funcName)
		return nil, false
	}

	filterOutBool := false
	filterOut, ok := parameters[FilterOut]
	if ok {
		var err error
		filterOutBool, err = strconv.ParseBool(filterOut)
		if err != nil {
			app.lc.Errorf("Could not convert filterOut value `%s` to bool for %s", filterOut, funcName)
			return nil, false
		}
	}

	namesCleaned := util.DeleteEmptyAndTrim(strings.FieldsFunc(names, util.SplitComma))
	transform := transforms.Filter{
		FilterValues: namesCleaned,
		FilterOut:    filterOutBool,
	}

	return &transform, true
}

func (app *Configurable) processHttpExportParameters(
	parameters map[string]string) (*postPutParameters, error) {

	result := postPutParameters{}
	var ok bool

	result.method, ok = parameters[ExportMethod]
	if !ok {
		return nil, fmt.Errorf("HTTPExport Could not find %s", ExportMethod)
	}

	result.url, ok = parameters[Url]
	if !ok {
		return nil, fmt.Errorf("HTTPExport Could not find %s", Url)
	}
	result.mimeType, ok = parameters[MimeType]
	if !ok {
		return nil, fmt.Errorf("HTTPExport Could not find %s", MimeType)
	}

	// PersistOnError is optional and is false by default.
	var value string
	result.persistOnError = false
	value, ok = parameters[PersistOnError]
	if ok {
		var err error
		result.persistOnError, err = strconv.ParseBool(value)
		if err != nil {
			return nil,
				fmt.Errorf("HTTPExport Could not parse '%s' to a bool for '%s' parameter: %s",
					value,
					PersistOnError,
					err.Error())
		}
	}

	// ContinueOnSendError is optional and is false by default.
	result.continueOnSendError = false
	value, ok = parameters[ContinueOnSendError]
	if ok {
		var err error
		result.continueOnSendError, err = strconv.ParseBool(value)
		if err != nil {
			return nil,
				fmt.Errorf("HTTPExport Could not parse '%s' to a bool for '%s' parameter: %s",
					value,
					ContinueOnSendError,
					err.Error())
		}
	}

	// ReturnInputData is optional and is false by default.
	result.returnInputData = false
	value, ok = parameters[ReturnInputData]
	if ok {
		var err error
		result.returnInputData, err = strconv.ParseBool(value)
		if err != nil {
			return nil,
				fmt.Errorf("HTTPExport Could not parse '%s' to a bool for '%s' parameter: %s",
					value,
					ReturnInputData,
					err.Error())
		}
	}

	result.url = strings.TrimSpace(result.url)
	result.mimeType = strings.TrimSpace(result.mimeType)
	result.headerName = strings.TrimSpace(parameters[HeaderName])
	result.secretPath = strings.TrimSpace(parameters[SecretPath])
	result.secretName = strings.TrimSpace(parameters[SecretName])

	if len(result.headerName) == 0 && len(result.secretPath) != 0 && len(result.secretName) != 0 {
		return nil,
			fmt.Errorf("HTTPExport missing %s since %s & %s are specified", HeaderName, SecretPath, SecretName)
	}
	if len(result.secretPath) == 0 && len(result.headerName) != 0 && len(result.secretName) != 0 {
		return nil,
			fmt.Errorf("HTTPExport missing %s since %s & %s are specified", SecretPath, HeaderName, SecretName)
	}
	if len(result.secretName) == 0 && len(result.secretPath) != 0 && len(result.headerName) != 0 {
		return nil,
			fmt.Errorf("HTTPExport missing %s since %s & %s are specified", SecretName, SecretPath, HeaderName)
	}

	return &result, nil
}
