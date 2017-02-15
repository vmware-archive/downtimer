package clients

import "time"

type Opts struct {
	URL                string        `short:"u" long:"url" description:"URL to probe" required:"true"`
	Duration           time.Duration `short:"d" long:"duration" description:"How long to probe for, forever by default" default:"0s"`
	Interval           time.Duration `short:"i" long:"interval" description:"interval at which to probe" default:"1s"`
	BoshCACert         string        `short:"c" long:"ca-cert" description:"CA cert for bosh" group:"bosh"`
	OutputFile         string        `short:"o" long:"output CSV file" description:"destination for CSV rows" default:"/dev/stdout"`
	BoshHost           string        `short:"b" long:"bosh" description:"bosh host" group:"bosh"`
	BoshUser           string        `short:"U" long:"user" description:"bosh user" group:"bosh"`
	BoshPassword       string        `short:"P" long:"password" description:"bosh client password" group:"bosh"`
	BoshTask           string        `short:"T" long:"task" description:"bosh deployment task override" group:"bosh"`
	InsecureSkipVerify bool          `short:"k" long:"skip-ssl-validation" description:"skip SSL validation"`
}
