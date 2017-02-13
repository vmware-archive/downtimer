package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry/bosh-cli/director"
	flags "github.com/jessevdk/go-flags"
	"github.com/pivotal-cf/downtimer/clients"
)

func main() {
	opts := clients.Opts{}
	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		os.Exit(1)
	}

	bosh, err := clients.GetDirector(opts.BoshHost, 25555, opts.BoshUser, opts.BoshPassword, opts.CACert)
	if err != nil {

		panic(err)
	}

	if opts.BoshTask == "" {
		opts.BoshTask = strconv.Itoa(clients.WaitForTaskId(bosh, 100*time.Second))
	}

	probeURL := opts.URL
	interval, err := time.ParseDuration(opts.Interval)
	if err != nil {
		panic(err)
	}
	duration, err := time.ParseDuration(opts.Duration)
	if err != nil {
		panic(err)
	}
	prober, err := clients.NewProber(&opts, bosh)
	if err != nil {
		panic(err)
	}

	log.Println(fmt.Sprintf("Starting to probe %s every %s seconds", probeURL, interval))
	prober.RecordDowntime(interval, duration)

	timestamps := getDeploymentTimes(bosh, opts.BoshTask)
	log.Println(prober.AnnotateWithTimestamps(timestamps))
}

func getDeploymentTimes(bosh director.Director, taskID string) clients.DeploymentTimes {
	eventsFilter := director.EventsFilter{Task: taskID}
	events, err := bosh.Events(eventsFilter)
	if err != nil {
		panic(err)
	}

	timestamps := clients.DeploymentTimes{}
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
