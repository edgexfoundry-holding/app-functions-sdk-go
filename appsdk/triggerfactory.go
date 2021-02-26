//
// Copyright (c) 2020 Technocrats
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
	"errors"
	"fmt"
	"strings"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/runtime"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/trigger/http"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/trigger/messagebus"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/trigger/mqtt"

	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
)

func (sdk *AppFunctionsSDK) defaultTriggerMessageProcessor(edgexcontext *appcontext.Context, envelope types.MessageEnvelope) error {
	messageError := sdk.runtime.ProcessMessage(edgexcontext, envelope)

	if messageError != nil {
		// ProcessMessage logs the error, so no need to log it here.
		return messageError.Err
	}

	return nil
}

func (sdk *AppFunctionsSDK) defaultTriggerContextBuilder(env types.MessageEnvelope) *appcontext.Context {
	return &appcontext.Context{
		CorrelationID:         env.CorrelationID,
		Configuration:         sdk.config,
		LoggingClient:         sdk.LoggingClient,
		EventClient:           sdk.EdgexClients.EventClient,
		ValueDescriptorClient: sdk.EdgexClients.ValueDescriptorClient,
		CommandClient:         sdk.EdgexClients.CommandClient,
		NotificationsClient:   sdk.EdgexClients.NotificationsClient,
		SecretProvider:        sdk.secretProvider,
	}
}

// RegisterCustomTriggerFactory allows users to register builders for custom trigger types
func (sdk *AppFunctionsSDK) RegisterCustomTriggerFactory(name string,
	factory func(TriggerConfig) (Trigger, error)) error {
	nu := strings.ToUpper(name)

	if nu == TriggerTypeMessageBus ||
		nu == TriggerTypeHTTP ||
		nu == TriggerTypeMQTT {
		return errors.New(fmt.Sprintf("cannot register custom trigger for builtin type (%s)", name))
	}

	if sdk.customTriggerFactories == nil {
		sdk.customTriggerFactories = make(map[string]func(sdk *AppFunctionsSDK) (Trigger, error), 1)
	}

	sdk.customTriggerFactories[nu] = func(sdk *AppFunctionsSDK) (Trigger, error) {
		return factory(TriggerConfig{
			Config:           sdk.config,
			Logger:           sdk.LoggingClient,
			ContextBuilder:   sdk.defaultTriggerContextBuilder,
			MessageProcessor: sdk.defaultTriggerMessageProcessor,
		})
	}

	return nil
}

// setupTrigger configures the appropriate trigger as specified by configuration.
func (sdk *AppFunctionsSDK) setupTrigger(configuration *common.ConfigurationStruct, runtime *runtime.GolangRuntime) Trigger {
	var t Trigger
	// Need to make dynamic, search for the trigger that is input

	switch triggerType := strings.ToUpper(configuration.Trigger.Type); triggerType {
	case TriggerTypeHTTP:
		sdk.LoggingClient.Info("HTTP trigger selected")
		t = &http.Trigger{Configuration: configuration, Runtime: runtime, Webserver: sdk.webserver, EdgeXClients: sdk.EdgexClients}

	case TriggerTypeMessageBus:
		sdk.LoggingClient.Info("EdgeX MessageBus trigger selected")
		t = &messagebus.Trigger{Configuration: configuration, Runtime: runtime, EdgeXClients: sdk.EdgexClients}

	case TriggerTypeMQTT:
		sdk.LoggingClient.Info("External MQTT trigger selected")
		t = mqtt.NewTrigger(configuration, runtime, sdk.EdgexClients, sdk.secretProvider)

	default:
		if factory, found := sdk.customTriggerFactories[triggerType]; found {
			var err error
			t, err = factory(sdk)
			if err != nil {
				sdk.LoggingClient.Error(fmt.Sprintf("failed to initialize custom trigger [%s]: %s", triggerType, err.Error()))
				return nil
			}
		} else {
			sdk.LoggingClient.Error(fmt.Sprintf("Invalid Trigger type of '%s' specified", configuration.Trigger.Type))
		}
	}

	return t
}
