package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"math"

	"github.com/gookit/color"
	"github.com/iwita/monitoring/info"
	"github.com/iwita/monitoring/monitor"
	"gopkg.in/yaml.v2"
)

var start, connect time.Time

type Website struct {
	Url      string  `yaml:"url"`
	Interval float64 `yaml:"interval"`
}

type Configs struct {
	Websites []Website `yaml:"websites"`
}

func readFile(cfg *Configs, file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	var cfg Configs
	err := readFile(&cfg, "input.yaml")
	if err != nil {
		fmt.Println(err)
	}
	dd := monitor.NewMonitor()
	for _, w := range cfg.Websites {

		// [1] TODO:
		// include website unable to reach at the moment
		_, err := http.Get(w.Url)
		if err != nil {
			_, netErr := http.Get("https://www.google.com")
			if netErr != nil {
				fmt.Println("No internet connection")
				os.Exit(1)
			} else {
				fmt.Println("Unable to reach", w.Url)
				// [1] continue
			}
		}

		dd.Wbs = append(dd.Wbs, monitor.Website{
			Url:      w.Url,
			Interval: w.Interval,
			Timer:    time.NewTicker(time.Millisecond * time.Duration(w.Interval)),
			Res1h: &info.Result{
				Max:          -1,
				Average:      -1,
				Percentile:   -1,
				StatusCodes:  "",
				Availability: -1,
			},
		})
	}
	// Initialize the two output counters
	timer1 := time.NewTicker(time.Second * time.Duration(10))
	timer2 := time.NewTicker(time.Minute * time.Duration(1))

	// Start the monitoring
	go dd.Exec()
	var trend string
	for {
		select {
		case <-timer1.C:
			for _, wb := range dd.Wbs {
				websiteName := color.FgBlue.Render(wb.Url)
				alertOut := dd.StatsPerWebsite[wb.Url].TwoMinutesInfo.Alert.PrintTest()
				wb.Res10m = dd.StatsPerWebsite[wb.Url].TenMinutesInfo.GetResult()
				final := time.Duration(int(dd.StatsPerWebsite[wb.Url].TenMinutesInfo.SumResponses) / dd.StatsPerWebsite[wb.Url].TenMinutesInfo.TotalResponses)
				start := time.Duration(int(dd.StatsPerWebsite[wb.Url].OneHourInfo.SumResponses) / dd.StatsPerWebsite[wb.Url].OneHourInfo.TotalResponses)
				percentage := float64(final-start) / float64(start)
				if percentage < 0 {
					trend = fmt.Sprintf("%.2v%% faster than past hour", math.Abs(percentage)*100)
				} else if percentage == 0 {
					trend = fmt.Sprint("Stable trend")
				} else {
					trend = fmt.Sprintf("%.2v%% slower than past hour", percentage*100)
				}
				fmt.Printf(monitor.OutputTemplate, websiteName, alertOut,
					wb.Res10m.Max, wb.Res10m.Average, wb.Res10m.Percentile, trend, wb.Res10m.Availability, wb.Res10m.StatusCodes,
					wb.Res1h.Max, wb.Res1h.Average, wb.Res1h.Percentile, wb.Res1h.Availability, wb.Res1h.StatusCodes)
			}

		case <-timer2.C:
			for i, wb := range dd.Wbs {
				dd.Wbs[i].Res1h = dd.StatsPerWebsite[wb.Url].OneHourInfo.GetResult()
			}
		}
	}
}
