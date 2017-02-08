package clients

import (
	"crypto/tls"
	"encoding/csv"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
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
}

func NewProber(url string, insecureSkipVerify bool) *Prober {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSkipVerify},
	}
	prober := Prober{
		url,
		http.Client{Transport: transport},
	}
	return &prober
}

func (c *Prober) RecordDowntime(interval, duration time.Duration) {
	// duration == 0
	//
	// duration != 0

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

	csvWriter := csv.NewWriter(os.Stdout)
	for keepGoing() {
		go func() {
			row := getCvsRow(c.Probe())
			_ = csvWriter.Write(row)
			csvWriter.Flush()
		}()
		time.Sleep(interval)
	}
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
	// curl
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
