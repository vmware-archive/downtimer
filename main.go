package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cloudfoundry/bosh-cli/director"
	flags "github.com/jessevdk/go-flags"
	"github.com/pivotal-cf/downtimer/clients"
)

type Opts struct {
	URL          string `short:"u" long:"url" description:"URL to probe"`
	Duration     string `short:"d" long:"duration" description:"How long to probe for, forever by default" default:"0s"`
	Interval     string `short:"i" long:"interval" description:"interval at which to probe" default:"1s"`
	CACert       string `short:"c" long:"ca-cert" description:"CA cert for bosh"`
	OutputFile   string `short:"o" long:"output CSV file" description:"destination for CSV rows" default:"/dev/stdout"`
	BoshHost     string `short:"b" long:"bosh" description:"bosh host"`
	BoshUser     string `short:"U" long:"user" description:"bosh user"`
	BoshPassword string `short:"P" long:"password" description:"bosh client password"`
	BoshTask     string `short:"T" long:"task" description:"bosh deployment task"` // TODO: automate this
}

func main() {
	opts := Opts{}
	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		os.Exit(1)
	}

	bosh, err := clients.GetDirector(opts.BoshHost, 25555, opts.BoshUser, opts.BoshPassword, opts.CACert)
	if err != nil {

		panic(err)
	}
	eventsFilter := director.EventsFilter{Task: opts.BoshTask}

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
			if len(event.Context()) == 0 {
				timestamps[eventTime] = append(timestamps[eventTime], event.Instance()+" done")
			} else {
				timestamps[eventTime] = append(timestamps[eventTime], event.Instance()+" start")
			}
		}
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
	prober, err := clients.NewProber(probeURL, true, opts.OutputFile)
	if err != nil {
		panic(err)
	}
	log.Println(fmt.Sprintf("Starting to probe %s every %s seconds", probeURL, interval))
	prober.RecordDowntime(interval, duration)
	prober.AnnotateWithTimestamps(timestamps)
}
