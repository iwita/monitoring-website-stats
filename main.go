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
		//klog.Infof("Unable to decode the config file. Error: %v", err)
		return err
	}
	//fmt.Println(cfg)
	return nil
}

func main() {

	var cfg Configs
	readFile(&cfg, "input.yaml")

	dd := datadog.NewMonitor()
	for _, w := range cfg.Websites {
		dd.Wbs = append(dd.Wbs, datadog.Website{
			Url:      w.Url,
			Interval: w.Interval,
			Timer:    time.NewTicker(time.Millisecond * time.Duration(w.Interval)),
		})
	}

	//Initialize the two output counters
	timer1 := time.NewTicker(time.Second * time.Duration(10))
	timer2 := time.NewTicker(time.Minute * time.Duration(1))

	//dd.Wbs = cfg.Websites
	go dd.Exec()

	for {
		select {
		case <-timer1.C:
			fmt.Println("10 Seconds passed")
		case <-timer2.C:
			fmt.Println("1 minute passed")
		}
	}
	// for {
	// 	select:
	// }

}
