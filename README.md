# downtimer

Record app downtime during a bosh deployment.

## Prerequisites

1. A URL that you can probe against, e.g. `http://my-sample-app.engenv.cf-app.com/`
2. Credentials for a bosh user and a CA cert with which bosh was deployed.

## Usage

* From the terminal invoke. It will wait for bosh to create a deployment.
```
downtimer -u http://my-sample-app.engenv.cf-app.com \
  -U $BOSH_USER -P $BOSH_PASS -b $BOSH_HOST \
  -c ca-cert.engenv.pem \
  -o viewer/public/my-deployment.csv
```
* Start your deployment.
* When the deployment is finished, take a look at downtime data in the CSV file. You can use our awesome downtime viewer:
```
cd $GOPATH/src/github.com/pivotal-cf/downtimer/viewer
go run main.go  # go to http://localhost:3000/index.html and select a downtime report
```

