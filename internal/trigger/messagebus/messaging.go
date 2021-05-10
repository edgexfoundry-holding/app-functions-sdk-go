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

package messagebus

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/appfunction"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/bootstrap/container"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/runtime"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/util"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	bootstrapMessaging "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-messaging/v2/messaging"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
)

// Trigger implements Trigger to support MessageBusData
type Trigger struct {
	dic     *di.Container
	runtime *runtime.GolangRuntime
	topics  []types.TopicChannel
	client  messaging.MessageClient
}

func NewTrigger(dic *di.Container, runtime *runtime.GolangRuntime) *Trigger {
	return &Trigger{
		dic:     dic,
		runtime: runtime,
	}
}

// Initialize ...
func (trigger *Trigger) Initialize(appWg *sync.WaitGroup, appCtx context.Context, background <-chan types.MessageEnvelope) (bootstrap.Deferred, error) {
	var err error
	lc := bootstrapContainer.LoggingClientFrom(trigger.dic.Get)
	config := container.ConfigurationFrom(trigger.dic.Get)

	lc.Infof("Initializing Message Bus Trigger for '%s'", config.Trigger.EdgexMessageBus.Type)

	clientConfig := trigger.createMessagingClientConfig(config.Trigger.EdgexMessageBus)

	if err := trigger.setOptionalAuthData(&clientConfig, lc); err != nil {
		return nil, err
	}

	trigger.client, err = messaging.NewMessageClient(clientConfig)
	if err != nil {
		return nil, err
	}

	subscribeTopics := strings.TrimSpace(config.Trigger.EdgexMessageBus.SubscribeHost.SubscribeTopics)

	if len(subscribeTopics) == 0 {
		// Still allows subscribing to blank topic to receive all messages
		trigger.topics = append(trigger.topics, types.TopicChannel{Topic: subscribeTopics, Messages: make(chan types.MessageEnvelope)})
	} else {
		topics := util.DeleteEmptyAndTrim(strings.FieldsFunc(subscribeTopics, util.SplitComma))
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
		subscribeTopics,
		config.Trigger.EdgexMessageBus.SubscribeHost.Protocol,
		config.Trigger.EdgexMessageBus.SubscribeHost.Host,
		config.Trigger.EdgexMessageBus.SubscribeHost.Port)

	publishTopic := config.Trigger.EdgexMessageBus.PublishHost.PublishTopic

	if len(config.Trigger.EdgexMessageBus.PublishHost.Host) > 0 {
		lc.Infof("Publishing to topic: '%s' @ %s://%s:%d",
			publishTopic,
			config.Trigger.EdgexMessageBus.PublishHost.Protocol,
			config.Trigger.EdgexMessageBus.PublishHost.Host,
			config.Trigger.EdgexMessageBus.PublishHost.Port)
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
					err := trigger.client.Publish(bg, publishTopic)
					if err != nil {
						lc.Errorf("Failed to publish background Message to bus, %v", err)
						return
					}

					lc.Debugf("Published background message to bus on %s topic", publishTopic)
					lc.Tracef("%s=%s", clients.CorrelationHeader, bg.CorrelationID)
				}()
			}
		}
	}()

	if err := trigger.client.Subscribe(trigger.topics, messageErrors); err != nil {
		return nil, fmt.Errorf("failed to subscribe to topic(s) '%s': %s", subscribeTopics, err.Error())
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

	appContext := appfunction.NewContext(message.CorrelationID, trigger.dic, message.ContentType)

	messageError := trigger.runtime.ProcessMessage(appContext, message)
	if messageError != nil {
		// ProcessMessage logs the error, so no need to log it here.
		return
	}

	if appContext.ResponseData() != nil {
		var contentType string

		if appContext.ResponseContentType() != "" {
			contentType = appContext.ResponseContentType()
		} else {
			contentType = clients.ContentTypeJSON
			if appContext.ResponseData()[0] != byte('{') {
				// If not JSON then assume it is CBOR
				contentType = clients.ContentTypeCBOR
			}
		}
		outputEnvelope := types.MessageEnvelope{
			CorrelationID: appContext.CorrelationID(),
			Payload:       appContext.ResponseData(),
			ContentType:   contentType,
		}

		config := container.ConfigurationFrom(trigger.dic.Get)
		publishTopic := config.Trigger.EdgexMessageBus.PublishHost.PublishTopic

		err := trigger.client.Publish(outputEnvelope, publishTopic)
		if err != nil {
			logger.Errorf("Failed to publish Message to bus, %v", err)
			return
		}

		logger.Debugf("Published message to bus on '%s' topic", publishTopic)
		logger.Tracef("%s=%s", clients.CorrelationHeader, message.CorrelationID)
	}
}

func (_ *Trigger) createMessagingClientConfig(localConfig common.MessageBusConfig) types.MessageBusConfig {
	clientConfig := types.MessageBusConfig{
		PublishHost: types.HostInfo{
			Host:     localConfig.PublishHost.Host,
			Port:     localConfig.PublishHost.Port,
			Protocol: localConfig.PublishHost.Protocol,
		},
		SubscribeHost: types.HostInfo{
			Host:     localConfig.SubscribeHost.Host,
			Port:     localConfig.SubscribeHost.Port,
			Protocol: localConfig.SubscribeHost.Protocol,
		},
		Type:     localConfig.Type,
		Optional: localConfig.Optional,
	}

	return clientConfig
}

func (trigger *Trigger) setOptionalAuthData(messageBusConfig *types.MessageBusConfig, lc logger.LoggingClient) error {
	authMode := strings.ToLower(strings.TrimSpace(messageBusConfig.Optional[bootstrapMessaging.AuthModeKey]))
	if len(authMode) == 0 || authMode == bootstrapMessaging.AuthModeNone {
		return nil
	}

	secretName := messageBusConfig.Optional[bootstrapMessaging.SecretNameKey]

	lc.Infof("Setting options for secure MessageBus with AuthMode='%s' and SecretName='%s", authMode, secretName)

	secretProvider := bootstrapContainer.SecretProviderFrom(trigger.dic.Get)
	if secretProvider == nil {
		return errors.New("secret provider is missing. Make sure it is specified to be used in bootstrap.Run()")
	}

	secretData, err := bootstrapMessaging.GetSecretData(authMode, secretName, secretProvider)
	if err != nil {
		return fmt.Errorf("Unable to get Secret Data for secure message bus: %w", err)
	}

	if err := bootstrapMessaging.ValidateSecretData(authMode, secretName, secretData); err != nil {
		return fmt.Errorf("Secret Data for secure message bus invalid: %w", err)
	}

	if messageBusConfig.Optional == nil {
		messageBusConfig.Optional = map[string]string{}
	}

	// Since already validated, these are the only modes that can be set at this point.
	switch authMode {
	case bootstrapMessaging.AuthModeUsernamePassword:
		messageBusConfig.Optional[bootstrapMessaging.OptionsUsernameKey] = secretData.Username
		messageBusConfig.Optional[bootstrapMessaging.OptionsPasswordKey] = secretData.Password
	case bootstrapMessaging.AuthModeCert:
		messageBusConfig.Optional[bootstrapMessaging.OptionsCertPEMBlockKey] = string(secretData.CertPemBlock)
		messageBusConfig.Optional[bootstrapMessaging.OptionsKeyPEMBlockKey] = string(secretData.KeyPemBlock)
	case bootstrapMessaging.AuthModeCA:
		messageBusConfig.Optional[bootstrapMessaging.OptionsCaPEMBlockKey] = string(secretData.CaPemBlock)
	}

	return nil
}
