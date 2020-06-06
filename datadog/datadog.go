package datadog

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptrace"
	"time"
)

type Website struct {
	Url      string
	Interval float64
}
type Websites []Website
type Statistics struct {
	MaxResponseTime     float64
	ResponseTimes       []int64
	StatusCodesCount    map[string]int
	SuccessfulResponses int
	TotalResponses      int
}

type Monitor struct {
	wbs             Websites
	StatsPerWebsite map[Website]*Statistics
	done            chan bool
}

func newMonitor(ws Websites) *Monitor {
	return &Monitor{
		wbs:             ws,
		done:            make(chan bool),
		StatsPerWebsite: make(map[Website]*Statistics, 0),
	}
}

func (m *Monitor) Exec() {
	m.exec()
}

func (m *Monitor) exec() {

	for _, wb := range m.wbs {
		go func(wb Website) {
			var start time.Time
			req, _ := http.NewRequest("GET", wb.Url, nil)
			trace := &httptrace.ClientTrace{
				// ConnectStart: func(network, addr string) { connect = time.Now() },
				// ConnectDone: func(network, addr string, err error) {
				// 	//fmt.Printf("Connect time: %v\n", time.Since(connect))
				// 	m.StatsPerWebsite[wb]
				// },
				GotFirstResponseByte: func() {
					//fmt.Printf("Time from start to first byte: %v\n", time.Since(start))
					m.StatsPerWebsite[wb].ResponseTimes = append(m.StatsPerWebsite[wb].ResponseTimes, int64(time.Since(start)/time.Millisecond))
				},
			}
			start = time.Now()
			req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
			res, err := http.DefaultTransport.RoundTrip(req)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(res)
		}(wb)
	}
}
