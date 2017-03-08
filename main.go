package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/pivotal-cf/downtimer/clients"
)

func main() {
	opts := clients.Opts{}
	err := ParseArgs(&opts, os.Args)
	if err != nil {
		os.Exit(1)
	}

	var bosh *clients.BoshImpl

	if useBosh(&opts) {
		bosh, err = clients.GetDirector(opts.BoshHost, 25555, opts.BoshUser, opts.BoshPassword, opts.BoshCACert)
		if err != nil {
			panic(err)
		}

		if opts.BoshTask == "" {
			opts.BoshTask = strconv.Itoa(bosh.WaitForTaskId(180 * time.Second))
		}
	}

	prober := clients.NewProber(&opts, bosh)

	log.Println(fmt.Sprintf("Starting to probe %s every %s seconds", opts.URL, opts.Interval))
	prober.RecordDowntime()

	if useBosh(&opts) {
		timestamps := bosh.GetDeploymentTimes(opts.BoshTask)
		log.Println(prober.AnnotateWithTimestamps(timestamps))
	}
}

func ParseArgs(opts *clients.Opts, args []string) error {
	_, err := flags.ParseArgs(opts, args)
	if err != nil {
		return err
	}

	if useBosh(opts) {
		if opts.BoshHost == "" || opts.BoshUser == "" || opts.BoshPassword == "" || opts.BoshCACert == "" {
			return errors.New("all bosh options must be specified")
		}
	}

	return nil
}

func useBosh(opts *clients.Opts) bool {
	return opts.BoshHost != "" || opts.BoshUser != "" || opts.BoshPassword != "" || opts.BoshCACert != ""
}
