package main

import (
	"fmt"
	"log"
	"os"
	"time"

	flags "github.com/jessevdk/go-flags"
	"github.com/pivotal-cf/downdog/clients"
)

type Opts struct {
	URL      string `short:"u" long:"url" description:"URL to probe"`
	Duration string `short:"d" long:"duration" description:"How long to probe for, forever by default"`
	Interval string `short:"i" long:"interval" description:"interval at which to probe"`
}

func main() {
	opts := Opts{}
	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		// TODO: print nice usage
		panic(err)
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
