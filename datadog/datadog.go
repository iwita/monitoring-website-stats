package datadog

import (
	"fmt"
	"net/http"
	"net/http/httptrace"
	"sync"
	"time"

	"github.com/iwita/monitoring/alert"
	"github.com/iwita/monitoring/info"
)

const MaxInt = int(^uint(0) >> 1)

type Website struct {
	Url      string
	Interval float64
	Timer    *time.Ticker
}

type Websites []Website

// The Statistics type is a struct that includes different durations of statistics information
type Statistics struct {
	TwoMinutesInfo *info.Info
	TenMinutesInfo *info.Info
	OneHourInfo    *info.Info
	OverallInfo    *info.Info
}

// Monitor type is the main type of this package
type Monitor struct {
	Wbs             Websites
	UrlToWebsite    map[string]Website
	StatsPerWebsite map[string]*Statistics
	done            chan bool
	mutex           *sync.Mutex
	Alert           *alert.Alert
}

// Initialize the Monitor, by setting the default values and allocating space
func NewMonitor() *Monitor {
	return &Monitor{
		Wbs:             make(Websites, 0),
		UrlToWebsite:    make(map[string]Website, 0),
		done:            make(chan bool),
		StatsPerWebsite: make(map[string]*Statistics, 0),
		mutex:           &sync.Mutex{},
		Alert:           alert.NewAlert(0.8),
	}
}

func (m *Monitor) Exec() {
	m.exec()
}

func (m *Monitor) exec() {
	var wg sync.WaitGroup
	// For each website, create a new goroutine
	for _, wb := range m.Wbs {
		wg.Add(1)
		go m.manageSingleWebsite(wb)

	}
	wg.Wait()
	//fmt.Println("After wait")

}

// Go routine executed for each website
// Waits until the ticker reaches the interval's predefined value
func (m *Monitor) manageSingleWebsite(wb Website) {
	for {
		select {
		case <-m.done:
			return
		case <-wb.Timer.C:
			m.monitorOnce(wb)
		}
	}
}

// It is called when a new request needs to be sent to a website
func (m *Monitor) monitorOnce(wb Website) {
	var start time.Time
	var elapsedTime time.Duration
	req, err := http.NewRequest("GET", wb.Url, nil)
	if err != nil {
		fmt.Printf("Error while sending the request to %v : %v", wb.Url, err)
		return
	}
	trace := &httptrace.ClientTrace{
		//ConnectStart: func(network, addr string) { connectStart = time.Now() },
		//DNSStart:     func(dnsinfo httptrace.DNSStartInfo) { dnsStart = time.Now() },
		//DNSDone:      func(dnsinfo httptrace.DNSDoneInfo) { dnsLookup = time.Since(dnsStart) },
		//ConnectDone:  func(network, addr string, err error) { connectionEstablishment = time.Since(connectStart) },
		GotFirstResponseByte: func() {
			elapsedTime = time.Since(start)
		},
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	start = time.Now()
	res, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		fmt.Println(err)
	}
	m.mutex.Lock()
	m.addStatistics(wb, elapsedTime, res.StatusCode)
	m.mutex.Unlock()
}

// Adds the newly extracted metrics into the statistics of the website
func (m *Monitor) addStatistics(wb Website, elapsedTime time.Duration, status int) {

	// Handle the case, where there are no previous metrics stored
	if _, ok := m.StatsPerWebsite[wb.Url]; !ok {
		m.StatsPerWebsite[wb.Url] = &Statistics{
			TwoMinutesInfo: info.NewInfo(time.Duration(2)*time.Minute, time.Duration(wb.Interval)*time.Millisecond, true),
			TenMinutesInfo: info.NewInfo(time.Duration(10)*time.Minute, time.Duration(wb.Interval)*time.Millisecond, false),
			OneHourInfo:    info.NewInfo(time.Duration(1)*time.Hour, time.Duration(wb.Interval)*time.Millisecond, false),
			//OverallInfo:    info.NewInfo(time.Duration(0)*time.Hour, time.Duration(wb.Interval)*time.Millisecond, false),
		}
	}

	m.StatsPerWebsite[wb.Url].TwoMinutesInfo.Update(status, elapsedTime)
	m.StatsPerWebsite[wb.Url].TenMinutesInfo.Update(status, elapsedTime)
	m.StatsPerWebsite[wb.Url].OneHourInfo.Update(status, elapsedTime)
	//m.StatsPerWebsite[wb.Url].OverallInfo.Update(status, elapsedTime)
}

func (m *Monitor) printStats() {
	//Lock
	for _, wb := range m.Wbs {
		fmt.Println(wb.Url, m.StatsPerWebsite[wb.Url])
	}
	//Unlock
}
