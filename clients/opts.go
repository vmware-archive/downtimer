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

import "time"

type Opts struct {
	URL                string        `short:"u" long:"url" description:"URL to probe" required:"true"`
	Duration           time.Duration `short:"d" long:"duration" description:"How long to probe for, forever by default" default:"0s"`
	Interval           time.Duration `short:"i" long:"interval" description:"interval at which to probe" default:"1s"`
	BoshCACert         string        `short:"c" long:"ca-cert" description:"CA cert for bosh" group:"bosh"`
	OutputFile         string        `short:"o" long:"output" description:"destination for CSV rows" default:"/dev/stdout"`
	LogFile            string        `short:"l" long:"logfile" description:"logfile" default:"/dev/stderr"`
	BoshHost           string        `short:"b" long:"bosh" description:"bosh host" group:"bosh"`
	BoshUser           string        `short:"U" long:"user" description:"bosh user" group:"bosh"`
	BoshPassword       string        `short:"P" long:"password" description:"bosh client password" group:"bosh"`
	BoshTask           string        `short:"T" long:"task" description:"bosh deployment task override" group:"bosh"`
	InsecureSkipVerify bool          `short:"k" long:"skip-ssl-validation" description:"skip SSL validation"`
}
