package ping

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"strings"
	"time"

	"k8s.io/klog"
)

// Trace is a ping trace and all the info associated
type Trace struct {
	request       *http.Request
	firstByte     time.Time
	dnsStart      time.Time
	dnsEnd        time.Time
	connectStart  time.Time
	connectEnd    time.Time
	Host          string        `json:"host"`
	DNSLookupTime time.Duration `json:"dnsLookupTime"`
	Response      string        `json:"response"`
	ResponseTime  time.Duration `json:"responseTime"`
	RoundTripTime time.Duration `json:"roundTripTime,omitempty"`
}

// RoundTrip wraps http.DefaultTransport.RoundTrip to keep track
// of the current request.
func (t *Trace) RoundTrip(req *http.Request) (*http.Response, error) {
	t.request = req
	return http.DefaultTransport.RoundTrip(req)
}

// GotConn prints whether the connection has been used previously
// for the current request.
func (t *Trace) GotConn(info httptrace.GotConnInfo) {
	if info.Reused {
		klog.V(4).Infof("connection reused for %v", t.request.URL)
	}
}

// DNSDone is the end of DNS lookup
func (t *Trace) DNSDone(info httptrace.DNSDoneInfo) {
	t.dnsEnd = time.Now()
	klog.V(4).Infof("dns done")
}

// DNSStart is the start of DNS
func (t *Trace) DNSStart(info httptrace.DNSStartInfo) {
	t.dnsStart = time.Now()
	klog.V(4).Info("dns start")
}

// GotFirstResponseByte is the first response byte
func (t *Trace) GotFirstResponseByte() {
	t.firstByte = time.Now()
	klog.V(4).Info("got first reponse byte")
}

// ConnectStart is the beginning
func (t *Trace) ConnectStart(network, addr string) {
	t.connectStart = time.Now()
	klog.V(4).Info("connect start")
}

// ConnectDone is the end
func (t *Trace) ConnectDone(network, addr string, err error) {
	t.connectEnd = time.Now()
	klog.V(4).Info("connect end")
}

// Run does the ping trace
func (t *Trace) Run() error {
	if !strings.Contains(t.Host, "http") {
		klog.V(3).Infof("host %s does not contain valid scheme - assuming http://", t.Host)
		t.Host = fmt.Sprintf("http://%s", t.Host)
	}

	klog.V(2).Infof("starting trace on host: %s", t.Host)
	req, _ := http.NewRequest("GET", t.Host, nil)
	trace := &httptrace.ClientTrace{
		GotConn:              t.GotConn,
		DNSStart:             t.DNSStart,
		DNSDone:              t.DNSDone,
		GotFirstResponseByte: t.GotFirstResponseByte,
		ConnectStart:         t.ConnectStart,
		ConnectDone:          t.ConnectDone,
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	client := &http.Client{Transport: t}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	t.Response = string(body)

	t.calculateDNS()
	t.calculateResponseTime()

	return nil
}

func (t *Trace) calculateDNS() {
	t.DNSLookupTime = t.dnsEnd.Sub(t.dnsStart)
}

func (t *Trace) calculateResponseTime() {
	t.ResponseTime = t.firstByte.Sub(t.connectStart)
}

// FastestTrace returns the fastest of a list of traces
// Error returned on empty list
func FastestTrace(traces []Trace) (Trace, error) {
	if len(traces) == 0 {
		return Trace{}, fmt.Errorf("cannot handle empty slice of traces")
	}

	var fastest Trace = traces[0]
	for _, trace := range traces {
		klog.V(7).Infof("%v", trace)
		if trace.ResponseTime < fastest.ResponseTime {
			fastest = trace
		}
	}
	return fastest, nil
}
