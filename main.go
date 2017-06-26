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
