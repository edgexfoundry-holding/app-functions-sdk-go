//
// Copyright (c) 2020 Intel Corporation
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

package messagebus

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/runtime"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/util"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-messaging/v2/messaging"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap"
)

// Trigger implements Trigger to support MessageBusData
type Trigger struct {
	Configuration *common.ConfigurationStruct
	Runtime       *runtime.GolangRuntime
	client        messaging.MessageClient
	topics        []types.TopicChannel
	EdgeXClients  common.EdgeXClients
}

// Initialize ...
func (trigger *Trigger) Initialize(appWg *sync.WaitGroup, appCtx context.Context, background <-chan types.MessageEnvelope) (bootstrap.Deferred, error) {
	var err error
	lc := trigger.EdgeXClients.LoggingClient

	lc.Infof("Initializing Message Bus Trigger for '%s'", trigger.Configuration.Trigger.EdgexMessageBus.Type)

	trigger.client, err = messaging.NewMessageClient(trigger.Configuration.Trigger.EdgexMessageBus)
	if err != nil {
		return nil, err
	}

	if len(strings.TrimSpace(trigger.Configuration.Trigger.SubscribeTopics)) == 0 {
		// Still allows subscribing to blank topic to receive all messages
		trigger.topics = append(trigger.topics, types.TopicChannel{Topic: trigger.Configuration.Trigger.SubscribeTopics, Messages: make(chan types.MessageEnvelope)})
	} else {
		topics := util.DeleteEmptyAndTrim(strings.FieldsFunc(trigger.Configuration.Trigger.SubscribeTopics, util.SplitComma))
		for _, topic := range topics {
			trigger.topics = append(trigger.topics, types.TopicChannel{Topic: topic, Messages: make(chan types.MessageEnvelope)})
		}
	}

	messageErrors := make(chan error)

	err = trigger.client.Connect()
	if err != nil {
		return nil, err
	}

	lc.Infof("Subscribing to topic(s): '%s' @ %s://%s:%d",
		trigger.Configuration.Trigger.SubscribeTopics,
		trigger.Configuration.Trigger.EdgexMessageBus.SubscribeHost.Protocol,
		trigger.Configuration.Trigger.EdgexMessageBus.SubscribeHost.Host,
		trigger.Configuration.Trigger.EdgexMessageBus.SubscribeHost.Port)

	if len(trigger.Configuration.Trigger.EdgexMessageBus.PublishHost.Host) > 0 {
		lc.Infof("Publishing to topic: '%s' @ %s://%s:%d",
			trigger.Configuration.Trigger.PublishTopic,
			trigger.Configuration.Trigger.EdgexMessageBus.PublishHost.Protocol,
			trigger.Configuration.Trigger.EdgexMessageBus.PublishHost.Host,
			trigger.Configuration.Trigger.EdgexMessageBus.PublishHost.Port)
	}

	// Need to have a go func for each subscription so we know with topic the data was received for.
	for _, topic := range trigger.topics {
		appWg.Add(1)
		go func(triggerTopic types.TopicChannel) {
			defer appWg.Done()
			lc.Infof("Waiting for messages from the MessageBus on the '%s' topic", triggerTopic.Topic)

			for true {
				select {
				case <-appCtx.Done():
					lc.Infof("Exiting waiting for MessageBus '%s' topic messages", triggerTopic.Topic)
					return
				case msgs := <-triggerTopic.Messages:
					go trigger.processMessage(lc, triggerTopic, msgs)
				}
			}
		}(topic)
	}

	// Need an addition go func to handle errors and background publishing to the message bus.
	appWg.Add(1)
	go func() {
		defer appWg.Done()
		for true {
			select {
			case <-appCtx.Done():
				lc.Info("Exiting waiting for MessageBus errors and background publishing")
				return

			case msgErr := <-messageErrors:
				lc.Errorf("Failed to receive message from bus, %v", msgErr)

			case bg := <-background:
				go func() {
					err := trigger.client.Publish(bg, trigger.Configuration.Trigger.PublishTopic)
					if err != nil {
						lc.Errorf("Failed to publish background Message to bus, %v", err)
						return
					}

					lc.Debugf("Published background message to bus on %s topic", trigger.Configuration.Trigger.PublishTopic)
					lc.Tracef("%s=%s", clients.CorrelationHeader, bg.CorrelationID)
				}()
			}
		}
	}()

	if err := trigger.client.Subscribe(trigger.topics, messageErrors); err != nil {
		return nil, fmt.Errorf("failed to subscribe to topic(s) '%s': %s", trigger.Configuration.Trigger.SubscribeTopics, err.Error())
	}

	deferred := func() {
		lc.Info("Disconnecting from the message bus")
		err := trigger.client.Disconnect()
		if err != nil {
			lc.Errorf("Unable to disconnect from the message bus: %s", err.Error())
		}
	}
	return deferred, nil
}

func (trigger *Trigger) processMessage(logger logger.LoggingClient, triggerTopic types.TopicChannel, message types.MessageEnvelope) {
	logger.Debugf("Received message from MessageBus on topic '%s'. Content-Type=%s", triggerTopic.Topic, message.ContentType)
	logger.Tracef("%s=%s", clients.CorrelationHeader, message.CorrelationID)

	edgexContext := &appcontext.Context{
		CorrelationID:         message.CorrelationID,
		Configuration:         trigger.Configuration,
		LoggingClient:         trigger.EdgeXClients.LoggingClient,
		EventClient:           trigger.EdgeXClients.EventClient,
		ValueDescriptorClient: trigger.EdgeXClients.ValueDescriptorClient,
		CommandClient:         trigger.EdgeXClients.CommandClient,
		NotificationsClient:   trigger.EdgeXClients.NotificationsClient,
	}

	messageError := trigger.Runtime.ProcessMessage(edgexContext, message)
	if messageError != nil {
		// ProcessMessage logs the error, so no need to log it here.
		return
	}

	if edgexContext.OutputData != nil {
		var contentType string

		if edgexContext.ResponseContentType != "" {
			contentType = edgexContext.ResponseContentType
		} else {
			contentType = clients.ContentTypeJSON
			if edgexContext.OutputData[0] != byte('{') {
				// If not JSON then assume it is CBOR
				contentType = clients.ContentTypeCBOR
			}
		}
		outputEnvelope := types.MessageEnvelope{
			CorrelationID: edgexContext.CorrelationID,
			Payload:       edgexContext.OutputData,
			ContentType:   contentType,
		}
		err := trigger.client.Publish(outputEnvelope, trigger.Configuration.Trigger.PublishTopic)
		if err != nil {
			logger.Errorf("Failed to publish Message to bus, %v", err)
			return
		}

		logger.Debugf("Published message to bus on '%s' topic", trigger.Configuration.Trigger.PublishTopic)
		logger.Tracef("%s=%s", clients.CorrelationHeader, message.CorrelationID)
	}
}
