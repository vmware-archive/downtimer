/* Copyright (C) 2017-Present Pivotal Software, Inc. All rights reserved.

This program and the accompanying materials are made available under
the terms of the under the Apache License, Version 2.0 (the "License”);
you may not use this file except in compliance with the License.

You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License. */

package clients

import (
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/cloudfoundry/bosh-cli/director"
	"github.com/cloudfoundry/bosh-cli/uaa"
	"github.com/cloudfoundry/bosh-utils/errors"
	"github.com/cloudfoundry/bosh-utils/logger"
)

type Bosh interface {
	GetDeploymentTimes(taskID string) DeploymentTimes
	GetCurrentTaskId() (int, error)
	WaitForTaskId(timeout time.Duration) int
}

type BoshImpl struct {
	director director.Director
}

func (b *BoshImpl) GetDeploymentTimes(taskID string) DeploymentTimes {
	eventsFilter := director.EventsFilter{Task: taskID}
	events, err := b.director.Events(eventsFilter)
	if err != nil {
		panic(err)
	}

	timestamps := DeploymentTimes{}
	for _, event := range events {
		if event.Action() == "update" && event.ObjectType() == "instance" {
			eventTime := event.Timestamp().Unix()
			_, ok := timestamps[eventTime]
			if !ok {
				timestamps[eventTime] = []string{}
			}
			// Event with empty context is the end time.
			instanceParts := strings.Split(event.Instance(), "/")
			instanceName := instanceParts[0]
			if len(event.Context()) == 0 {
				timestamps[eventTime] = append(timestamps[eventTime], instanceName+" done")
			} else {
				timestamps[eventTime] = append(timestamps[eventTime], instanceName+" start")
			}
		}
	}
	return timestamps
}

func (b *BoshImpl) GetCurrentTaskId() (int, error) {
	currentTasks, err := b.director.CurrentTasks(director.TasksFilter{})
	if err != nil {
		return 0, err
	}
	var currentTaskId int
	for _, task := range currentTasks {
		if task.Description() == "create deployment" {
			currentTaskId = task.ID()
			break
		}
	}
	return currentTaskId, nil
}

func (b *BoshImpl) IsAuthenticated() (bool, error) {
	return b.director.IsAuthenticated()
}

func (b *BoshImpl) WaitForTaskId(timeout time.Duration) int {
	timeoutChannel := time.After(timeout)
	tick := time.Tick(5 * time.Second)

	for {
		select {
		case <-timeoutChannel:
			log.Println("Bailed on getting the Task ID")
			return 0
		case <-tick:
			log.Println("Pulling Bosh for Deployment Task")
			id, err := b.GetCurrentTaskId()
			if err != nil {
				log.Println(err)
			}

			if id != 0 {
				return id
			}
		}
	}
}

func anonymousUserConfig(host string, port int, CACert string) director.Config {
	return director.Config{
		Host:   host,
		Port:   port,
		CACert: CACert,
	}
}

func userConfig(host string, port int, CACert, username, password string) director.Config {
	config := anonymousUserConfig(host, port, CACert)
	config.Client = username
	config.ClientSecret = password
	return config
}

func GetDirector(host string, port int, username, password, caCertFile string) (*BoshImpl, error) {

	logger := logger.NewLogger(0)
	caCertBytes, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		return nil, err
	}
	config := uaa.Config{
		Host:         host,
		Port:         port,
		CACert:       string(caCertBytes),
		Client:       username,
		ClientSecret: password,
	}

	info, err := getInfo(config.Host, config.Port, config.CACert, logger)
	if err != nil {
		return nil, err
	}

	dirConfig := userConfig(config.Host, config.Port, config.CACert, config.Client, config.ClientSecret)

	if info.Auth.Type == "uaa" {
		uaaClient, err := getUaa(info, config.Client, config.ClientSecret, config.CACert, logger)
		if err != nil {
			return nil, err
		}

		dirConfig.Client = ""
		dirConfig.ClientSecret = ""

		dirConfig.TokenFunc = uaa.NewClientTokenSession(uaaClient).TokenFunc
	}

	taskReporter := director.NewNoopTaskReporter()
	fileReporter := director.NewNoopFileReporter()

	director, err := director.NewFactory(logger).New(dirConfig, taskReporter, fileReporter)
	return &BoshImpl{director}, err
}

func getUaa(info director.Info, client, clientSecret, CACert string, logger logger.Logger) (uaa.UAA, error) {
	uaaURL := info.Auth.Options["url"]

	uaaURLStr, ok := uaaURL.(string)
	if !ok {
		return nil, errors.Errorf("Expected URL '%s' to be a string", uaaURL)
	}

	uaaConfig, err := uaa.NewConfigFromURL(uaaURLStr)
	if err != nil {
		return nil, err
	}

	uaaConfig.CACert = CACert
	uaaConfig.Client = client
	uaaConfig.ClientSecret = clientSecret

	if len(uaaConfig.Client) == 0 {
		uaaConfig.Client = "downtimer"
	}

	return uaa.NewFactory(logger).New(uaaConfig)
}

func getInfo(host string, port int, CACert string, logger logger.Logger) (director.Info, error) {
	dirConfig := anonymousUserConfig(host, port, CACert)
	directorClient, err := director.NewFactory(logger).New(dirConfig, nil, nil)
	if err != nil {
		return director.Info{}, err
	}
	return directorClient.Info()
}
