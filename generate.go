package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/hkdsun/simiload/load"
	"github.com/hkdsun/simiload/platform"
	log "github.com/sirupsen/logrus"
)

var loadsConfigFile = flag.String("config", "", "load config json file")

func usage() {
	fmt.Printf("Load generator tool")
	fmt.Println()
	fmt.Println("Usage: generate -config flash_sale.json <url>")
	fmt.Println()
	flag.PrintDefaults()
}

func getLoadConfig() ([]*load.Load, error) {

	jsonFile, err := os.Open(*loadsConfigFile)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var loadsConfig []struct {
		ShopId     int    `json:"shop_id"`
		StartAfter string `json:"start_after"`
		Duration   string `json:"duration"`
	}

	err = json.Unmarshal(byteValue, &loadsConfig)
	if err != nil {
		return nil, err
	}

	var loads []*load.Load = make([]*load.Load, len(loadsConfig))

	for i, l := range loadsConfig {
		startAfter, err := time.ParseDuration(l.StartAfter)
		if err != nil {
			return nil, err
		}

		duration, err := time.ParseDuration(l.Duration)
		if err != nil {
			return nil, err
		}

		loads[i] = &load.Load{
			Scope:      platform.Scope{l.ShopId},
			StartAfter: startAfter,
			Duration:   duration,
		}
	}

	return loads, nil
}

func main() {
	flag.Parse()

	if *loadsConfigFile == "" {
		usage()
		os.Exit(1)
	}

	loads, err := getLoadConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	gen := &load.Generator{
		ServerURL: flag.Args()[0],
		Loads:     loads,
	}
	gen.Run()
}
