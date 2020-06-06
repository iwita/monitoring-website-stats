package main

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

var start, connect time.Time

type Config struct {
	Website  string
	Interval float64
}
type Configs struct {
	Cfgs []Config `yaml:"websites"`
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
	fmt.Println(cfg)
	return nil
}

func main() {
	

	var cfg Configs
	readFile(&cfg, "input.yaml")

	dd = datadog.NewMonitor(cfg)
	dd.Exec()
}
