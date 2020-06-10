package main

import (
	"fmt"
	"os"
	"time"

	"github.com/iwita/monitoring/datadog"

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
		//klog.Infof("Config file not found. Error: %v", err)
		return err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		//fmt.Printf("Unable to decode the config file. Error: %v", err)
		return err
	}
	//fmt.Println(cfg)
	return nil
}

func main() {

	inputArray := []int{1, 2, 3, 4, 5, 6}
	minHeap := heap.newMinHeap(len(inputArray))
	for i := 0; i < len(inputArray); i++ {
		minHeap.insert(inputArray[i])
	}
	minHeap.buildMinHeap()
	for i := 0; i < len(inputArray); i++ {
		fmt.Println(minHeap.remove())
	}
	//fmt.Scanln()

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
	for {
		select {
		case <-timer1.C:
			fmt.Println("Last 10 minutes statistics")

			for _, wb := range dd.Wbs {
				fmt.Printf("Website: %v\n", wb.Url)
				// print alert output
				dd.StatsPerWebsite[wb.Url].TwoMinutesInfo.Alert.Print()
				// print all the rest
				dd.StatsPerWebsite[wb.Url].TenMinutesInfo.PrintInfo()
			}

		case <-timer2.C:
			for _, wb := range dd.Wbs {
				fmt.Printf("Website: %v\n", wb.Url)
				dd.StatsPerWebsite[wb.Url].TwoMinutesInfo.Alert.Print()
				dd.StatsPerWebsite[wb.Url].OneHourInfo.PrintInfo()
			}
		}
	}
}
