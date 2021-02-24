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

package appsdk

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/transforms"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/util"
)

const (
	ProfileNames        = "profilenames"
	DeviceNames         = "devicenames"
	ResourceNames       = "resourcenames"
	FilterOut           = "filterout"
	EncryptionKey       = "key"
	InitVector          = "initvector"
	Url                 = "url"
	MimeType            = "mimetype"
	PersistOnError      = "persistonerror"
	SkipVerify          = "skipverify"
	Qos                 = "qos"
	Retain              = "retain"
	AutoReconnect       = "autoreconnect"
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
	Topic               = "topic"
	AuthMode            = "authmode"
	Tags                = "tags"
	ResponseContentType = "responsecontenttype"
)

// AppFunctionsSDKConfigurable contains the helper functions that return the function pointers for building the configurable function pipeline.
// They transform the parameters map from the Pipeline configuration in to the actual actual parameters required by the function.
type AppFunctionsSDKConfigurable struct {
	Sdk *AppFunctionsSDK
}

// FilterByProfileName - Specify the profile names of interest to filter for data coming from certain sensors.
// The Filter by Profile Name transform looks at the Event in the message and looks at the profile names of interest list,
// provided by this function, and filters out those messages whose Event is for profile names not in the
// profile names of interest.
// This function will return an error and stop the pipeline if a non-edgex
// event is received or if no data is received.
// For example, data generated by a motor does not get passed to functions only interested in data from a thermostat.
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) FilterByProfileName(parameters map[string]string) appcontext.AppFunction {
	profileNames, ok := parameters[ProfileNames]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("Could not find " + ProfileNames)
		return nil
	}

	filterOutBool := false
	filterOut, ok := parameters[FilterOut]
	if ok {
		var err error
		filterOutBool, err = strconv.ParseBool(filterOut)
		if err != nil {
			dynamic.Sdk.LoggingClient.Error("Could not convert filterOut value to bool " + filterOut)
			return nil
		}
	}

	profileNamesCleaned := util.DeleteEmptyAndTrim(strings.FieldsFunc(profileNames, util.SplitComma))
	transform := transforms.Filter{
		FilterValues: profileNamesCleaned,
		FilterOut:    filterOutBool,
	}
	dynamic.Sdk.LoggingClient.Debugf("Profile Name Filters (filterOut=%v) are: '%s'", filterOutBool, strings.Join(profileNamesCleaned, ","))

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
func (dynamic AppFunctionsSDKConfigurable) FilterByDeviceName(parameters map[string]string) appcontext.AppFunction {
	deviceNames, ok := parameters[DeviceNames]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("Could not find " + DeviceNames)
		return nil
	}

	filterOutBool := false
	filterOut, ok := parameters[FilterOut]
	if ok {
		var err error
		filterOutBool, err = strconv.ParseBool(filterOut)
		if err != nil {
			dynamic.Sdk.LoggingClient.Error("Could not convert filterOut value to bool " + filterOut)
			return nil
		}
	}

	deviceNamesCleaned := util.DeleteEmptyAndTrim(strings.FieldsFunc(deviceNames, util.SplitComma))
	transform := transforms.Filter{
		FilterValues: deviceNamesCleaned,
		FilterOut:    filterOutBool,
	}
	dynamic.Sdk.LoggingClient.Debugf("Device Name Filters (filterOut=%v) are: '%s'", filterOutBool, strings.Join(deviceNamesCleaned, ","))

	return transform.FilterByDeviceName
}

// FilterByResourceName - Specify the resource name of interest to filter for data from certain types of IoT objects,
// such as temperatures, motion, and so forth, that may come from an array of sensors or devices. The Filter by resource name assesses
// the data in each Event and Reading, and removes readings that have a resource name that is not in the list of
// resource names of interest for the application.
// This function will return an error and stop the pipeline if a non-edgex
// event is received or if no data is received.
// For example, pressure reading data does not go to functions only interested in motion data.
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) FilterByResourceName(parameters map[string]string) appcontext.AppFunction {
	resourceNames, ok := parameters[ResourceNames]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("Could not find " + ResourceNames)
		return nil
	}

	filterOutBool := false
	filterOut, ok := parameters[FilterOut]
	if ok {
		var err error
		filterOutBool, err = strconv.ParseBool(filterOut)
		if err != nil {
			dynamic.Sdk.LoggingClient.Error("Could not convert filterOut value to bool " + filterOut)
			return nil
		}
	}

	resourceNamesCleaned := util.DeleteEmptyAndTrim(strings.FieldsFunc(resourceNames, util.SplitComma))
	transform := transforms.Filter{
		FilterValues: resourceNamesCleaned,
		FilterOut:    filterOutBool,
	}
	dynamic.Sdk.LoggingClient.Debugf("Resource Name Filters (filterOut=%v) are `%s`", filterOutBool, strings.Join(resourceNamesCleaned, ","))

	return transform.FilterByResourceName
}

// TransformToXML transforms an EdgeX event to XML.
// It will return an error and stop the pipeline if a non-edgex
// event is received or if no data is received.
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) TransformToXML() appcontext.AppFunction {
	transform := transforms.Conversion{}
	return transform.TransformToXML
}

// TransformToJSON transforms an EdgeX event to JSON.
// It will return an error and stop the pipeline if a non-edgex
// event is received or if no data is received.
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) TransformToJSON() appcontext.AppFunction {
	transform := transforms.Conversion{}
	return transform.TransformToJSON
}

// PushToCore pushes the provided value as an event to CoreData using the device name and reading name that have been set. If validation is turned on in
// CoreServices then your deviceName and readingName must exist in the CoreMetadata and be properly registered in EdgeX.
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) PushToCore(parameters map[string]string) appcontext.AppFunction {
	deviceName, ok := parameters[DeviceName]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("Could not find " + DeviceName)
		return nil
	}
	readingName, ok := parameters[ReadingName]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("Could not find " + readingName)
		return nil
	}
	deviceName = strings.TrimSpace(deviceName)
	readingName = strings.TrimSpace(readingName)
	dynamic.Sdk.LoggingClient.Debug("PushToCore Parameters", DeviceName, deviceName, ReadingName, readingName)
	transform := transforms.CoreData{
		DeviceName:  deviceName,
		ReadingName: readingName,
	}
	return transform.PushToCoreData
}

// CompressWithGZIP compresses data received as either a string,[]byte, or json.Marshaller using gzip algorithm and returns a base64 encoded string as a []byte.
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) CompressWithGZIP() appcontext.AppFunction {
	transform := transforms.Compression{}
	return transform.CompressWithGZIP
}

// CompressWithZLIB compresses data received as either a string,[]byte, or json.Marshaller using zlib algorithm and returns a base64 encoded string as a []byte.
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) CompressWithZLIB() appcontext.AppFunction {
	transform := transforms.Compression{}
	return transform.CompressWithZLIB
}

// EncryptWithAES encrypts either a string, []byte, or json.Marshaller type using AES encryption.
// It will return a byte[] of the encrypted data.
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) EncryptWithAES(parameters map[string]string) appcontext.AppFunction {
	secretPath := parameters[SecretPath]
	secretName := parameters[SecretName]
	encryptionKey := parameters[EncryptionKey]

	// SecretPath & SecretName are optional if EncryptionKey specified
	// EncryptionKey is optional if SecretPath & SecretName are specified

	// If EncryptionKey not specified, then SecretPath & SecretName must be specified
	if len(encryptionKey) == 0 && (len(secretPath) == 0 || len(secretName) == 0) {
		dynamic.Sdk.LoggingClient.Errorf("Could not find '%s' or '%s' and '%s' in configuration", EncryptionKey, SecretPath, SecretName)
		return nil
	}

	// SecretPath & SecretName both must be specified it one of them is.
	if (len(secretPath) != 0 && len(secretName) == 0) || (len(secretPath) == 0 && len(secretName) != 0) {
		dynamic.Sdk.LoggingClient.Errorf("'%s' and '%s' both must be set in configuration", SecretPath, SecretName)
		return nil
	}

	initVector, ok := parameters[InitVector]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("Could not find " + InitVector)
		return nil
	}

	transform := transforms.Encryption{
		EncryptionKey:        encryptionKey,
		InitializationVector: initVector,
		SecretPath:           secretPath,
		SecretName:           secretName,
	}

	return transform.EncryptWithAES
}

// HTTPPost will send data from the previous function to the specified Endpoint via http POST. If no previous function exists,
// then the event that triggered the pipeline will be used. Passing an empty string to the mimetype
// method will default to application/json.
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) HTTPPost(parameters map[string]string) appcontext.AppFunction {
	url, mimeType, persistOnError, headerName, secretPath, secretName, err := dynamic.processPostPutParameters(parameters)
	if err != nil {
		dynamic.Sdk.LoggingClient.Error(err.Error())
		return nil
	}

	var transform transforms.HTTPSender
	if len(secretPath) != 0 {
		transform = transforms.NewHTTPSenderWithSecretHeader(url, mimeType, persistOnError, headerName, secretPath, secretName)
	} else {
		transform = transforms.NewHTTPSender(url, mimeType, persistOnError)
	}

	dynamic.Sdk.LoggingClient.Debugf("HTTPPost Parameters: %v", parameters)
	return transform.HTTPPost
}

// HTTPPostJSON sends data from the previous function to the specified Endpoint via http POST with a mime type of application/json.
// If no previous function exists, then the event that triggered the pipeline will be used.
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) HTTPPostJSON(parameters map[string]string) appcontext.AppFunction {
	parameters[MimeType] = "application/json"
	return dynamic.HTTPPost(parameters)
}

// HTTPPostXML sends data from the previous function to the specified Endpoint via http POST with a mime type of application/xml.
// If no previous function exists, then the event that triggered the pipeline will be used.
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) HTTPPostXML(parameters map[string]string) appcontext.AppFunction {
	parameters[MimeType] = "application/xml"
	return dynamic.HTTPPost(parameters)
}

// HTTPPut will send data from the previous function to the specified Endpoint via http PUT. If no previous function exists,
// then the event that triggered the pipeline will be used. Passing an empty string to the mimetype
// method will default to application/json.
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) HTTPPut(parameters map[string]string) appcontext.AppFunction {
	url, mimeType, persistOnError, headerName, secretPath, secretName, err := dynamic.processPostPutParameters(parameters)
	if err != nil {
		dynamic.Sdk.LoggingClient.Error(err.Error())
		return nil
	}

	var transform transforms.HTTPSender
	if len(secretPath) != 0 {
		transform = transforms.NewHTTPSenderWithSecretHeader(url, mimeType, persistOnError, headerName, secretPath, secretName)
	} else {
		transform = transforms.NewHTTPSender(url, mimeType, persistOnError)
	}

	dynamic.Sdk.LoggingClient.Debug("HTTPPut Parameters", Url, transform.URL, MimeType, transform.MimeType)
	return transform.HTTPPut
}

// HTTPPutJSON sends data from the previous function to the specified Endpoint via http PUT with a mime type of application/json.
// If no previous function exists, then the event that triggered the pipeline will be used.
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) HTTPPutJSON(parameters map[string]string) appcontext.AppFunction {
	parameters[MimeType] = "application/json"
	return dynamic.HTTPPut(parameters)
}

// HTTPPutXML sends data from the previous function to the specified Endpoint via http PUT with a mime type of application/xml.
// If no previous function exists, then the event that triggered the pipeline will be used.
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) HTTPPutXML(parameters map[string]string) appcontext.AppFunction {
	parameters[MimeType] = "application/xml"
	return dynamic.HTTPPut(parameters)
}

// SetOutputData sets the output data to that passed in from the previous function.
// It will return an error and stop the pipeline if data passed in is not of type []byte, string or json.Marshaller
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) SetOutputData(parameters map[string]string) appcontext.AppFunction {
	transform := transforms.OutputData{}

	value, ok := parameters[ResponseContentType]
	if ok && len(value) > 0 {
		transform.ResponseContentType = value
	}

	return transform.SetOutputData
}

// BatchByCount ...
func (dynamic AppFunctionsSDKConfigurable) BatchByCount(parameters map[string]string) appcontext.AppFunction {
	batchThreshold, ok := parameters[BatchThreshold]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("Could not find " + BatchThreshold)
		return nil
	}

	thresholdValue, err := strconv.Atoi(batchThreshold)
	if err != nil {
		dynamic.Sdk.LoggingClient.Error(fmt.Sprintf("Could not parse '%s' to an int for '%s' parameter", batchThreshold, BatchThreshold), "error", err)
		return nil
	}
	transform, err := transforms.NewBatchByCount(thresholdValue)
	if err != nil {
		dynamic.Sdk.LoggingClient.Error(err.Error())
	}
	dynamic.Sdk.LoggingClient.Debug("Batch by count Parameters", BatchThreshold, batchThreshold)
	return transform.Batch
}

// BatchByTime ...
func (dynamic AppFunctionsSDKConfigurable) BatchByTime(parameters map[string]string) appcontext.AppFunction {
	timeInterval, ok := parameters[TimeInterval]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("Could not find " + TimeInterval)
		return nil
	}
	transform, err := transforms.NewBatchByTime(timeInterval)
	if err != nil {
		dynamic.Sdk.LoggingClient.Error(err.Error())
	}
	dynamic.Sdk.LoggingClient.Debug("Batch by time Parameters", TimeInterval, timeInterval)
	return transform.Batch
}

// BatchByTimeAndCount ...
func (dynamic AppFunctionsSDKConfigurable) BatchByTimeAndCount(parameters map[string]string) appcontext.AppFunction {
	timeInterval, ok := parameters[TimeInterval]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("Could not find " + TimeInterval)
		return nil
	}
	batchThreshold, ok := parameters[BatchThreshold]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("Could not find " + BatchThreshold)
		return nil
	}
	thresholdValue, err := strconv.Atoi(batchThreshold)
	if err != nil {
		dynamic.Sdk.LoggingClient.Error(fmt.Sprintf("Could not parse '%s' to an int for '%s' parameter", batchThreshold, BatchThreshold), "error", err)
	}
	transform, err := transforms.NewBatchByTimeAndCount(timeInterval, thresholdValue)
	if err != nil {
		dynamic.Sdk.LoggingClient.Error(err.Error())
	}
	dynamic.Sdk.LoggingClient.Debug("Batch by time and count Parameters", BatchThreshold, batchThreshold, TimeInterval, timeInterval)
	return transform.Batch
}

// JSONLogic ...
func (dynamic AppFunctionsSDKConfigurable) JSONLogic(parameters map[string]string) appcontext.AppFunction {
	rule, ok := parameters[Rule]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("Could not find " + Rule)
		return nil
	}
	transform := transforms.NewJSONLogic(rule)
	return transform.Evaluate
}

// MQTTSecretSend
func (dynamic AppFunctionsSDKConfigurable) MQTTSecretSend(parameters map[string]string) appcontext.AppFunction {
	var err error
	qos := 0
	retain := false
	autoReconnect := false
	skipCertVerify := false

	brokerAddress, ok := parameters[BrokerAddress]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("Could not find " + BrokerAddress)
		return nil
	}
	topic, ok := parameters[Topic]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("Could not find " + Topic)
		return nil
	}

	secretPath, ok := parameters[SecretPath]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("Could not find " + SecretPath)
		return nil
	}
	authMode, ok := parameters[AuthMode]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("Could not find " + AuthMode)
		return nil
	}
	clientID, ok := parameters[ClientID]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("Could not find " + ClientID)
		return nil
	}
	qosVal, ok := parameters[Qos]
	if ok {
		qos, err = strconv.Atoi(qosVal)
		if err != nil {
			dynamic.Sdk.LoggingClient.Error("Unable to parse " + Qos + " value")
			return nil
		}
	}
	retainVal, ok := parameters[Retain]
	if ok {
		retain, err = strconv.ParseBool(retainVal)
		if err != nil {
			dynamic.Sdk.LoggingClient.Error("Unable to parse " + Retain + " value")
			return nil
		}
	}
	autoreconnectVal, ok := parameters[AutoReconnect]
	if ok {
		autoReconnect, err = strconv.ParseBool(autoreconnectVal)
		if err != nil {
			dynamic.Sdk.LoggingClient.Error("Unable to parse " + AutoReconnect + " value")
			return nil
		}
	}
	skipVerifyVal, ok := parameters[SkipVerify]
	if ok {
		skipCertVerify, err = strconv.ParseBool(skipVerifyVal)
		if err != nil {
			dynamic.Sdk.LoggingClient.Error(fmt.Sprintf("Could not parse '%s' to a bool for '%s' parameter", skipVerifyVal, SkipVerify), "error", err)
			return nil
		}
	}
	mqttConfig := transforms.MQTTSecretConfig{
		Retain:         retain,
		SkipCertVerify: skipCertVerify,
		AutoReconnect:  autoReconnect,
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
			dynamic.Sdk.LoggingClient.Error(fmt.Sprintf("Could not parse '%s' to a bool for '%s' parameter", value, PersistOnError), "error", err)
			return nil
		}
	}
	transform := transforms.NewMQTTSecretSender(mqttConfig, persistOnError)
	return transform.MQTTSend
}

// AddTags adds the configured list of tags to Events passed to the transform.
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) AddTags(parameters map[string]string) appcontext.AppFunction {
	tagsSpec, ok := parameters[Tags]
	if !ok {
		dynamic.Sdk.LoggingClient.Error(fmt.Sprintf("Could not find '%s' parameter", Tags))
		return nil
	}

	tagKeyValues := util.DeleteEmptyAndTrim(strings.FieldsFunc(tagsSpec, util.SplitComma))

	tags := make(map[string]string)
	for _, tag := range tagKeyValues {
		keyValue := util.DeleteEmptyAndTrim(strings.FieldsFunc(tag, util.SplitColon))
		if len(keyValue) != 2 {
			dynamic.Sdk.LoggingClient.Error(fmt.Sprintf("Bad Tags specification format. Expect comma separated list of 'key:value'. Got `%s`", tagsSpec))
			return nil
		}

		if len(keyValue[0]) == 0 {
			dynamic.Sdk.LoggingClient.Error(fmt.Sprintf("Tag key missing. Got '%s'", tag))
			return nil
		}
		if len(keyValue[1]) == 0 {
			dynamic.Sdk.LoggingClient.Error(fmt.Sprintf("Tag value missing. Got '%s'", tag))
			return nil
		}

		tags[keyValue[0]] = keyValue[1]
	}

	transform := transforms.NewTags(tags)
	dynamic.Sdk.LoggingClient.Debug("Add Tags", Tags, fmt.Sprintf("%v", tags))

	return transform.AddTags
}

func (dynamic AppFunctionsSDKConfigurable) processPostPutParameters(
	parameters map[string]string) (string, string, bool, string, string, string, error) {
	url, ok := parameters[Url]
	if !ok {
		return "", "", false, "", "", "", fmt.Errorf("HTTPPut Could not find %s", Url)
	}
	mimeType, ok := parameters[MimeType]
	if !ok {
		return "", "", false, "", "", "", fmt.Errorf("HTTPPut Could not find %s", MimeType)
	}

	// PersistOnError is optional and is false by default.
	persistOnError := false
	value, ok := parameters[PersistOnError]
	if ok {
		var err error
		persistOnError, err = strconv.ParseBool(value)
		if err != nil {
			return "", "", false, "", "", "",
				fmt.Errorf("HTTPPut Could not parse '%s' to a bool for '%s' parameter: %s",
					value,
					PersistOnError,
					err.Error())
		}
	}

	url = strings.TrimSpace(url)
	mimeType = strings.TrimSpace(mimeType)
	headerName := strings.TrimSpace(parameters[HeaderName])
	secretPath := strings.TrimSpace(parameters[SecretPath])
	secretName := strings.TrimSpace(parameters[SecretName])

	if len(headerName) == 0 && len(secretPath) != 0 && len(secretName) != 0 {
		return "", "", false, "", "", "",
			fmt.Errorf("HTTPPost missing %s since %s & %s are specified", HeaderName, SecretPath, SecretName)
	}
	if len(secretPath) == 0 && len(headerName) != 0 && len(secretName) != 0 {
		return "", "", false, "", "", "",
			fmt.Errorf("HTTPPost missing %s since %s & %s are specified", SecretPath, HeaderName, SecretName)
	}
	if len(secretName) == 0 && len(secretPath) != 0 && len(headerName) != 0 {
		return "", "", false, "", "", "",
			fmt.Errorf("HTTPPost missing %s since %s & %s are specified", SecretName, SecretPath, HeaderName)
	}

	return url, mimeType, persistOnError, headerName, secretPath, secretName, nil
}
