package ping

import (
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
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
	Host          string `json:"host"`
	DNSLookupTime string `json:"dnsLookupTime"`
	Response      string `json:"response"`
	ResponseTime  string `json:"responseTime"`
	RoundTripTime string `json:"roundTripTime,omitempty"`
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
		klog.V(2).Infof("connection reused for %v", t.request.URL)
	}
}

// DNSDone is the end of DNS lookup
func (t *Trace) DNSDone(info httptrace.DNSDoneInfo) {
	t.dnsEnd = time.Now()
	klog.V(2).Infof("dns done")
}

// DNSStart is the start of DNS
func (t *Trace) DNSStart(info httptrace.DNSStartInfo) {
	t.dnsStart = time.Now()
	klog.V(2).Info("dns start")
}

// GotFirstResponseByte is the first response byte
func (t *Trace) GotFirstResponseByte() {
	t.firstByte = time.Now()
	klog.V(2).Info("got first reponse byte")
}

// ConnectStart is the beginning
func (t *Trace) ConnectStart(network, addr string) {
	t.connectStart = time.Now()
	klog.V(2).Info("connect start")
}

// ConnectDone is the end
func (t *Trace) ConnectDone(network, addr string, err error) {
	t.connectEnd = time.Now()
	klog.V(2).Info("connect end")
}

// Run does the ping trace
func (t *Trace) Run() error {
	klog.V(2).Infof("starting host: %s", t.Host)

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
	diff := t.dnsEnd.Sub(t.dnsStart)
	t.DNSLookupTime = diff.String()
}

func (t *Trace) calculateResponseTime() {
	diff := t.firstByte.Sub(t.connectStart)
	t.ResponseTime = diff.String()
}
