package datadog

import (
	"fmt"
	"net/http"
	"net/http/httptrace"
	"sync"
	"time"

	"github.com/iwita/monitoring/info"
)

const MaxInt = int(^uint(0) >> 1)

type Website struct {
	Url      string
	Interval float64
	Timer    *time.Ticker
}

type Websites []Website
type Statistics struct {
	//MaxResponseTime time.Duration
	//ResponseTimes   []time.Duration
	TenMinutesInfo *info.Info
	OneHourInfo    *info.Info
	OverallInfo    *info.Info
}

type Response struct {
	delay  time.Duration
	status int
}

type Monitor struct {
	Wbs             Websites
	UrlToWebsite    map[string]Website
	StatsPerWebsite map[string]*Statistics
	done            chan bool
	mutex           *sync.Mutex
}

func NewMonitor() *Monitor {

	return &Monitor{
		Wbs:             make(Websites, 0),
		UrlToWebsite:    make(map[string]Website, 0),
		done:            make(chan bool),
		StatsPerWebsite: make(map[string]*Statistics, 0),
		mutex:           &sync.Mutex{},
	}
}

func (m *Monitor) Exec() {
	m.exec()
}

func (m *Monitor) exec() {
	fmt.Println("Inside exec")
	var wg sync.WaitGroup
	for _, wb := range m.Wbs {
		wg.Add(1)
		go m.manageSingleWebsite(wb)

	}
	wg.Wait()
	//m.printStats()
	fmt.Println("After wait")

}

func (m *Monitor) manageSingleWebsite(wb Website) {
	for {
		select {
		case <-m.done:
			//wg.Wait()
			return
		case <-wb.Timer.C:
			m.monitorOnce(wb)
		}
	}
}

func (m *Monitor) monitorOnce(wb Website) {
	var start, connect time.Time
	var elapsedTime time.Duration
	req, _ := http.NewRequest("GET", wb.Url, nil)
	//fmt.Println("CheckPoint #1")
	trace := &httptrace.ClientTrace{
		ConnectStart: func(network, addr string) { connect = time.Now() },
		ConnectDone: func(network, addr string, err error) {
			//fmt.Printf("Website: %v, Connect time: %v\n", wb.Url, time.Since(connect))
		},
		GotFirstResponseByte: func() {
			elapsedTime = time.Since(start)

		},
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	// for {

	//fmt.Println("CheckPoint #2")
	start = time.Now()
	//fmt.Println("CheckPoint #3")
	res, err := http.DefaultTransport.RoundTrip(req)
	//fmt.Println("CheckPoint #4")
	if err != nil {
		fmt.Println(err)
	}
	//add statistics
	//fmt.Printf("Website: %v, Elapsed time: %v\n", wb.Url, elapsedTime)

	m.mutex.Lock()
	m.addStatistics(wb, elapsedTime, res.StatusCode)
	//m.printStats(wb)
	m.mutex.Unlock()
	//time.Sleep(time.Duration(wb.Interval) * time.Millisecond)
	//}

	//wg.Done()
}

func (m *Monitor) addStatistics(wb Website, elapsedTime time.Duration, status int) {

	if _, ok := m.StatsPerWebsite[wb.Url]; !ok {
		m.StatsPerWebsite[wb.Url] = &Statistics{
			TenMinutesInfo: info.NewInfo(time.Duration(10)*time.Minute, time.Duration(wb.Interval)*time.Millisecond),
			OneHourInfo:    info.NewInfo(time.Duration(1)*time.Hour, time.Duration(wb.Interval)*time.Millisecond),
			OverallInfo:    info.NewInfo(time.Duration(0)*time.Hour, time.Duration(wb.Interval)*time.Millisecond),
		}
	}

	m.StatsPerWebsite[wb.Url].TenMinutesInfo.Update(status, elapsedTime)
	m.StatsPerWebsite[wb.Url].OneHourInfo.Update(status, elapsedTime)
	m.StatsPerWebsite[wb.Url].OverallInfo.Update(status, elapsedTime)

	m.StatsPerWebsite[wb.Url].StatusCodesCount[status]++
	if status == 200 {
		m.StatsPerWebsite[wb.Url].SuccessfulResponses++
	}
	m.StatsPerWebsite[wb.Url].TotalResponses++

	m.StatsPerWebsite[wb.Url].ResponseTimes = append(m.StatsPerWebsite[wb.Url].ResponseTimes, elapsedTime)
	if elapsedTime > m.StatsPerWebsite[wb.Url].MaxResponseTime {
		m.StatsPerWebsite[wb.Url].MaxResponseTime = elapsedTime
	}

}

func (m *Monitor) printStats() {
	for _, wb := range m.Wbs {
		fmt.Println(wb.Url, m.StatsPerWebsite[wb.Url])
	}
}

/*
  Calculates the Average of the responses on the fly
*/

func getRunningAverage(n int, currentAverage, x time.Duration) time.Duration {
	return (currentAverage*time.Duration(n) + x) / (time.Duration(n + 1))
}
