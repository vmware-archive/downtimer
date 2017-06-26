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
	"crypto/tls"
	"encoding/csv"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/afero"
)

type Result struct {
	Timestamp    time.Time
	ResponseTime time.Duration
	StatusCode   int
	Size         int
	Error        error
	Success      int
}

type Prober struct {
	url    string
	client http.Client
	opts   *Opts
	bosh   Bosh
}

var FS = afero.NewOsFs()

type DeploymentTimes map[int64][]string

func NewProber(opts *Opts, bosh Bosh) *Prober {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: opts.InsecureSkipVerify},
	}
	client := http.Client{Transport: transport}
	prober := Prober{opts.URL, client, opts, bosh}

	return &prober
}

func (p *Prober) RecordDowntime() error {

	/* Ticket starts ticking at instantiation. A minimal
	   sleep offset is required to ensure that boshCheckTicker
		 ticks before the prober proberTicker */
	boshCheckTicker := time.NewTicker(p.opts.Interval)
	time.Sleep(10 * time.Millisecond)

	proberTicker := time.NewTicker(p.opts.Interval)
	timeout := make(<-chan time.Time)
	boshTask := make(chan time.Time)

	boshTaskStr := p.opts.BoshTask
	if boshTaskStr != "" {
		go func() {
			for {
				select {
				case <-boshCheckTicker.C:
					taskId, err := p.bosh.GetCurrentTaskId()
					if err != nil {
						log.Println(err)
					}
					optsTaskId, err := strconv.Atoi(boshTaskStr)
					if err != nil {
						log.Println(err)
					}
					if optsTaskId != taskId {
						boshTask <- time.Now()
					}
				}
			}
		}()
	}

	if p.opts.Duration != 0 {
		timeout = time.NewTimer(p.opts.Duration).C
	}

	outfile, err := FS.Create(p.opts.OutputFile)
	if err != nil {
		return err
	}

	csvWriter := csv.NewWriter(outfile)
	defer outfile.Close()
	csvWriter.Write([]string{"timestamp", "success", "latency", "code", "size", "", "annotation"})
	for {
		select {
		case <-boshTask:
			return nil
		case <-proberTicker.C:
			row := getCvsRow(p.Probe())
			_ = csvWriter.Write(row)
			csvWriter.Flush()
		case <-timeout:
			return nil
		}
	}
	return nil
}

func (p *Prober) AnnotateWithTimestamps(timestamps DeploymentTimes) error {

	annotatedFile, err := FS.Create(p.opts.OutputFile + "-annotated")
	if err != nil {
		return err
	}

	inputFile, err := FS.Open(p.opts.OutputFile)

	if err != nil {
		return err
	}

	csvWriter := csv.NewWriter(annotatedFile)
	defer annotatedFile.Close()
	csvReader := csv.NewReader(inputFile)
	defer inputFile.Close()

	header, err := csvReader.Read()
	if err != nil {
		return err
	}
	csvReader.FieldsPerRecord = 0
	csvWriter.Write(header)

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		timestamp, err := strconv.ParseInt(record[0], 10, 64)
		if err != nil {
			return err
		}

		annotations, exists := timestamps[timestamp]

		if exists {
			annotationString := strings.Join(annotations, "\n")
			record = append(record, annotationString)
		}
		csvWriter.Write(record)
		csvWriter.Flush()
	}

	FS.Rename(p.opts.OutputFile+"-annotated", p.opts.OutputFile)
	return nil
}

func getCvsRow(result Result) []string {
	cvsRow := []string{}
	resultError := ""
	if result.Error != nil {
		resultError = result.Error.Error()
	}
	cvsRow = append(cvsRow, strconv.FormatInt(result.Timestamp.Unix(), 10), strconv.Itoa(result.Success), result.ResponseTime.String(), strconv.Itoa(result.StatusCode), strconv.Itoa(result.Size), resultError)
	return cvsRow
}

func (c *Prober) Probe() Result {
	start := time.Now()
	resp, err := c.client.Get(c.url)
	if err != nil {
		return Result{Timestamp: start, Error: err}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	end := time.Now()
	if err != nil {
		return Result{Timestamp: start, Error: err}
	}
	success := 0
	if resp.StatusCode == 200 {
		success = 1
	}
	return Result{
		Timestamp:    start,
		ResponseTime: end.Sub(start),
		StatusCode:   resp.StatusCode,
		Size:         len(body),
		Success:      success,
	}
}
