package main

import (
	"fmt"
	"log"
	"os"
	"time"

	flags "github.com/jessevdk/go-flags"
	"github.com/pivotal-cf/downtimer/clients"
)

type Opts struct {
	URL      string `short:"u" long:"url" description:"URL to probe"`
	Duration string `short:"d" long:"duration" description:"How long to probe for, forever by default" default:"0s"`
	Interval string `short:"i" long:"interval" description:"interval at which to probe" default:"1s"`
}

func main() {
	opts := Opts{}
	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		os.Exit(1)
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
	log.Println(fmt.Sprintf("Starting to probe %s every %s seconds", probeURL, interval))
	prober := clients.NewProber(probeURL, true)
	prober.RecordDowntime(interval, duration)
}
