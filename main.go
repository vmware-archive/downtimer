package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

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
		opts.BoshTask = strconv.Itoa(bosh.WaitForTaskId(100 * time.Second))
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

	timestamps := bosh.GetDeploymentTimes(opts.BoshTask)
	log.Println(prober.AnnotateWithTimestamps(timestamps))
}
