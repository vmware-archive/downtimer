package clients

import (
	"crypto/tls"
	"encoding/csv"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
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
}

type DeploymentTimes map[int64][]string

func NewProber(opts *Opts) (*Prober, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: opts.InsecureSkipVerify},
	}

	prober := Prober{
		opts.URL,
		http.Client{Transport: transport},
		opts,
	}
	return &prober, nil
}

func (p *Prober) RecordDowntime(interval, duration time.Duration) error {
	var keepGoing func() bool
	if duration == 0 {
		keepGoing = func() bool {
			return true
		}
	} else {
		keepGoing = func() bool {
			duration = duration - interval
			return duration >= 0
		}
	}

	outfile, err := os.Create(p.opts.OutputFile)
	if err != nil {
		return err
	}

	csvWriter := csv.NewWriter(outfile)
	defer outfile.Close()
	for keepGoing() {
		go func() {
			row := getCvsRow(p.Probe())
			_ = csvWriter.Write(row)
			csvWriter.Flush()
		}()
		time.Sleep(interval)
	}
	return nil
}

func (p *Prober) AnnotateWithTimestamps(timestamps DeploymentTimes) error {

	annotatedFile, err := os.Create(p.opts.OutputFile + "-annotated")
	if err != nil {
		return err
	}

	inputFile, err := os.Open(p.opts.OutputFile)
	if err != nil {
		return err
	}

	csvWriter := csv.NewWriter(annotatedFile)
	defer annotatedFile.Close()
	csvReader := csv.NewReader(inputFile)
	defer inputFile.Close()

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
			annotationString := strings.Join(annotations, " | ")
			record = append(record, annotationString)
		}
		csvWriter.Write(record)
		csvWriter.Flush()
	}

	os.Rename(p.opts.OutputFile+"-annotated", p.opts.OutputFile)
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
