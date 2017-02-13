package clients

type Opts struct {
	URL                string `short:"u" long:"url" description:"URL to probe" required:"true"`
	Duration           string `short:"d" long:"duration" description:"How long to probe for, forever by default" default:"0s"`
	Interval           string `short:"i" long:"interval" description:"interval at which to probe" default:"1s"`
	CACert             string `short:"c" long:"ca-cert" description:"CA cert for bosh"`
	OutputFile         string `short:"o" long:"output CSV file" description:"destination for CSV rows" default:"/dev/stdout"`
	BoshHost           string `short:"b" long:"bosh" description:"bosh host"`
	BoshUser           string `short:"U" long:"user" description:"bosh user"`
	BoshPassword       string `short:"P" long:"password" description:"bosh client password"`
	BoshTask           string `short:"T" long:"task" description:"bosh deployment task"` // TODO: automate this
	InsecureSkipVerify bool   `short:"k" long:"skip-ssl-validation" description:"skip SSL validation"`
}
