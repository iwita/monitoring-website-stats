package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gookit/color"
	"github.com/iwita/monitoring/datadog"
	"github.com/iwita/monitoring/info"
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
	//fmt.Println(cfg)
	return nil
}

func main() {
	var cfg Configs
	err := readFile(&cfg, "input.yaml")
	if err != nil {
		fmt.Println(err)
	}
	dd := datadog.NewMonitor()
	for _, w := range cfg.Websites {
		dd.Wbs = append(dd.Wbs, datadog.Website{
			Url:      w.Url,
			Interval: w.Interval,
			Timer:    time.NewTicker(time.Millisecond * time.Duration(w.Interval)),
		})
	}
	// Initialize the two output counters
	timer1 := time.NewTicker(time.Second * time.Duration(10))
	timer2 := time.NewTicker(time.Minute * time.Duration(1))

	// Start the monitoring
	go dd.Exec()

	// When the ticker ticks, the  appropriate output is printed
	// for {
	// 	select {
	// 	case <-timer1.C:
	// 		fmt.Println(yellow("Last 10 minutes statistics"))

	// 		for _, wb := range dd.Wbs {
	// 			fmt.Printf("Website: %v\n", wb.Url)
	// 			// print alert output
	// 			dd.StatsPerWebsite[wb.Url].TwoMinutesInfo.Alert.Print()
	// 			// print all the rest
	// 			dd.StatsPerWebsite[wb.Url].TenMinutesInfo.PrintInfo()
	// 			final := time.Duration(int(dd.StatsPerWebsite[wb.Url].TenMinutesInfo.SumResponses) / dd.StatsPerWebsite[wb.Url].TenMinutesInfo.TotalResponses)
	// 			start := time.Duration(int(dd.StatsPerWebsite[wb.Url].OneHourInfo.SumResponses) / dd.StatsPerWebsite[wb.Url].OneHourInfo.TotalResponses)
	// 			percentage := float64((final - start) / start)
	// 			if percentage < 0 {
	// 				fmt.Printf("%v%% Faster than last hour average responses\n", abs(percentage)*100)
	// 			} else if percentage == 0 {
	// 				fmt.Println("Stable trend")
	// 			} else {
	// 				fmt.Printf("%v%% Slower than last hour average responses\n", percentage*100)
	// 			}
	// 			fmt.Println()
	// 		}

	// 	case <-timer2.C:
	// 		fmt.Println(yellow(blue("Last 1 hour statistics")))
	// 		for _, wb := range dd.Wbs {
	// 			fmt.Printf("Website: %v\n", wb.Url)
	// 			dd.StatsPerWebsite[wb.Url].TwoMinutesInfo.Alert.Print()
	// 			dd.StatsPerWebsite[wb.Url].OneHourInfo.PrintInfo()
	// 		}
	// 	}
	// }

	var trend string
	var res1h *info.Result = &info.Result{
		Max:          -1,
		Average:      -1,
		Percentile:   -1,
		StatusCodes:  "",
		Availability: -1,
	}
	var res10m *info.Result

	for {
		select {
		case <-timer1.C:
			for _, wb := range dd.Wbs {
				websiteName := color.FgBlue.Render(wb.Url)
				alertOut := dd.StatsPerWebsite[wb.Url].TwoMinutesInfo.Alert.PrintTest()
				res10m = dd.StatsPerWebsite[wb.Url].TenMinutesInfo.GetResult()
				final := time.Duration(int(dd.StatsPerWebsite[wb.Url].TenMinutesInfo.SumResponses) / dd.StatsPerWebsite[wb.Url].TenMinutesInfo.TotalResponses)
				start := time.Duration(int(dd.StatsPerWebsite[wb.Url].OneHourInfo.SumResponses) / dd.StatsPerWebsite[wb.Url].OneHourInfo.TotalResponses)
				percentage := float64(final-start) / float64(start)
				if percentage < 0 {
					trend = fmt.Sprintf("%.2v%% faster than past hour average responses", abs(percentage)*100)
				} else if percentage == 0 {
					trend = fmt.Sprint("Stable trend")
				} else {
					trend = fmt.Sprintf("%.2v%% slower than past hour average responses", percentage*100)
				}
				fmt.Printf(datadog.OutputTemplate, websiteName, alertOut,
					res10m.Max, res10m.Average, res10m.Percentile, res10m.Availability, trend, res10m.StatusCodes,
					res1h.Max, res1h.Average, res1h.Percentile, res1h.Availability, res1h.StatusCodes)
			}

		case <-timer2.C:
			for _, wb := range dd.Wbs {
				res1h = dd.StatsPerWebsite[wb.Url].OneHourInfo.GetResult()
			}
		}
	}
}

func abs(j float64) float64 {
	if j < 0 {
		return -j
	}
	return j
}
