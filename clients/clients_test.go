package clients_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"time"

	"github.com/pivotal-cf/downtimer/clients"
	"github.com/pivotal-cf/downtimer/clients/clientsfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const sampleRecordFile = `timestamp,success,latency,code,size,fill,annotation
123,1,125.75904ms,200,79,
456,1,2.860896ms,200,79,
789,1,2.564204ms,200,79,
101112,1,3.562018ms,200,79,
131415,1,3.568ms,200,79,
`

var _ = Describe("Clients", func() {
	var prober *clients.Prober
	var opts clients.Opts
	var recordFile *os.File
	var err error
	var bosh clients.Bosh
	BeforeEach(func() {
		recordFile, err = ioutil.TempFile("", "downtime-report.csv")
		Expect(err).NotTo(HaveOccurred())
		recordFile.Write([]byte(sampleRecordFile))
		opts = clients.Opts{
			OutputFile: recordFile.Name(),
			URL:        "http://localhost:54321/fake-url",
		}
	})
	JustBeforeEach(func() {
		bosh = new(clientsfakes.FakeBosh)
		prober, err = clients.NewProber(&opts, bosh)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.Remove(recordFile.Name())).ToNot(HaveOccurred())
	})

	Describe("Prober", func() {
		Context("recording downtime for given duration", func() {
			BeforeEach(func() {
				opts.Duration = 100*time.Millisecond + 2*time.Millisecond
				opts.Interval = 5 * time.Millisecond
			})
			It("records n = (duration/interval) times", func() {
				buf := make([]byte, 32*1024)
				prober.RecordDowntime()
				outputFile, err := os.Open(opts.OutputFile)
				Expect(err).NotTo(HaveOccurred())
				readBytesCount, err := outputFile.Read(buf)
				Expect(err).NotTo(HaveOccurred())
				lineCount := bytes.Count(buf[:readBytesCount], []byte{'\n'})
				Expect(lineCount).To(Equal(20 + 1)) // +1 for header
			})
		})
		Context("annotating the file", func() {
			var deploymentTimes clients.DeploymentTimes
			BeforeEach(func() {
				deploymentTimes = clients.DeploymentTimes{}
				deploymentTimes[123] = []string{"doppler done", "diego start"}
			})
			It("doesn't choke on a CSV header", func() {
				err := prober.AnnotateWithTimestamps(deploymentTimes)
				Expect(err).NotTo(HaveOccurred())
				f, err := os.Open(opts.OutputFile)
				Expect(err).NotTo(HaveOccurred())
				rewrittenFile, err := ioutil.ReadAll(f)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(rewrittenFile)).To(ContainSubstring("doppler done"))
			})
		})
	})
})
