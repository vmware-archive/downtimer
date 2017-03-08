package clients_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/pivotal-cf/downtimer/clients"
	"github.com/pivotal-cf/downtimer/clients/clientsfakes"
	"github.com/spf13/afero"

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

var mockServer *httptest.Server
var _ = BeforeSuite(func() {
	handler := http.NewServeMux()
	handler.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "I'm alive!")
	})
	handler.HandleFunc("/unavailable", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
	})
	handler.Handle("/notfound", http.NotFoundHandler())
	mockServer = httptest.NewServer(handler)
})

var _ = Describe("Clients", func() {
	var prober *clients.Prober
	var recordFile afero.File
	var err error
	var bosh *clientsfakes.FakeBosh
	var opts clients.Opts
	BeforeEach(func() {
		clients.FS = afero.NewMemMapFs()

		recordFile, err = afero.TempFile(clients.FS, "", "downtime-report.csv")
		Expect(err).NotTo(HaveOccurred())
		recordFile.Write([]byte(sampleRecordFile))
		bosh = new(clientsfakes.FakeBosh)
		opts = clients.Opts{
			OutputFile: recordFile.Name(),
			URL:        mockServer.URL + "/health",
			Duration:   1*time.Second + 2*time.Millisecond,
			Interval:   5 * time.Millisecond,
			BoshTask:   "",
		}
	})
	JustBeforeEach(func() {
		prober = clients.NewProber(&opts, bosh)
	})
	AfterEach(func() {
	})

	Describe("Prober", func() {
		Describe("Prober.Probe", func() {
			Context("when the URL responds with HTTP 200", func() {
				It("returns status 1 on success", func() {
					result := prober.Probe()
					Expect(result.StatusCode).To(Equal(200))
					Expect(result.Success).To(Equal(1))
				})
			})
			Context("when the URL is bad", func() {
				BeforeEach(func() {
					opts.URL = "unknown://scheme"
				})
				It("returns status 0 on bad url", func() {
					result := prober.Probe()
					Expect(result.StatusCode).To(Equal(0))
					Expect(result.Success).To(Equal(0))
				})
			})
			Context("when the URL responds with HTTP 404", func() {
				BeforeEach(func() {
					opts.URL = mockServer.URL + "/notfound"
				})
				It("returns status 0 on not found", func() {
					result := prober.Probe()
					Expect(result.StatusCode).To(Equal(404))
					Expect(result.Success).To(Equal(0))
				})
			})
			Context("when the URL responds with HTTP 503", func() {
				BeforeEach(func() {
					opts.URL = mockServer.URL + "/unavailable"
				})
				It("returns status 0 on internal server error", func() {
					result := prober.Probe()
					Expect(result.StatusCode).To(Equal(503))
					Expect(result.Success).To(Equal(0))
				})
			})
		})
		Context("recording downtime for given duration", func() {
			BeforeEach(func() {
				opts.Duration = 10*time.Millisecond + 2*time.Millisecond
			})
			It("records n = (duration/interval) times", func() {
				buf := make([]byte, 32*1024)
				prober.RecordDowntime()
				outputFile, err := clients.FS.Open(opts.OutputFile)
				Expect(err).NotTo(HaveOccurred())
				readBytesCount, err := outputFile.Read(buf)
				Expect(err).NotTo(HaveOccurred())
				lineCount := bytes.Count(buf[:readBytesCount], []byte{'\n'})
				Expect(lineCount).To(Equal(2 + 1)) // +1 for header
			})
		})
		Context("recording downtime for running deployment", func() {
			Context("when deployment isn't running anymore", func() {
				JustBeforeEach(func() {
					opts.Duration = 0 * time.Second
					opts.Interval = 5 * time.Millisecond
					opts.BoshTask = "111"

					bosh.GetCurrentTaskIdStub = func() (int, error) {
						return 0, nil
					}

				})
				It("should not record anything ", func() {
					buf := make([]byte, 32*1024)
					prober.RecordDowntime()
					outputFile, err := clients.FS.Open(opts.OutputFile)
					Expect(err).NotTo(HaveOccurred())
					_, err = outputFile.Read(buf)
					Expect(err).To(MatchError("EOF"))
				})
			})
			Context("when deployment is ongoing", func() {
				BeforeEach(func() {
					opts.Duration = 0 * time.Second
					opts.Interval = 100 * time.Millisecond
					opts.BoshTask = "111"

					validTaskCount := 4

					bosh.GetCurrentTaskIdStub = func() (int, error) {
						if validTaskCount > 0 {
							validTaskCount -= 1
							return 111, nil
						}
						return 0, nil
					}

				})
				It("should record for the duration of deployment", func() {
					buf := make([]byte, 32*1024)
					prober.RecordDowntime()
					outputFile, err := clients.FS.Open(opts.OutputFile)
					Expect(err).NotTo(HaveOccurred())
					readBytesCount, err := outputFile.Read(buf)
					Expect(err).NotTo(HaveOccurred())
					lineCount := bytes.Count(buf[:readBytesCount], []byte{'\n'})
					Expect(lineCount).To(Equal(4 + 1)) // +1 for header
				})
			})
		})
		Describe("Prober.AnnotateWithTimestamp", func() {
			var deploymentTimes clients.DeploymentTimes
			BeforeEach(func() {
				deploymentTimes = clients.DeploymentTimes{}
				deploymentTimes[123] = []string{"doppler done", "diego start"}
			})
			It("can handle a CSV header", func() {
				err := prober.AnnotateWithTimestamps(deploymentTimes)
				Expect(err).NotTo(HaveOccurred())
				f, err := clients.FS.Open(opts.OutputFile)
				Expect(err).NotTo(HaveOccurred())
				rewrittenFile, err := ioutil.ReadAll(f)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(rewrittenFile)).To(ContainSubstring("doppler done"))
			})
			Context("when the output file cannot be read", func() {
				BeforeEach(func() {
					corruptCsvFile := "/output.csv"
					opts.OutputFile = corruptCsvFile
					file, err := clients.FS.OpenFile(corruptCsvFile, os.O_CREATE, 0600)
					Expect(err).NotTo(HaveOccurred())
					file.WriteString(sampleRecordFile)
					file.WriteString("131415,1,3")
				})
				It("returns an error", func() {
					err := prober.AnnotateWithTimestamps(deploymentTimes)
					Expect(err.Error()).To(ContainSubstring("wrong number of fields in line"))
				})
			})
			Context("when the output file  ", func() {
				BeforeEach(func() {
					badFileName := "/output.csv"
					opts.OutputFile = badFileName
				})
				It("returns an error", func() {
					err := prober.AnnotateWithTimestamps(deploymentTimes)
					Expect(err.Error()).To(ContainSubstring("file does not exist"))
				})
			})
		})
	})
})
