package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hkdsun/simiload/load"
	log "github.com/sirupsen/logrus"
)

var scopeConfig = flag.String("config", "", "load config json file")

func usage() {
	fmt.Printf("Load generator tool")
	fmt.Println()
	fmt.Println("Usage: generate -config flash_sale.json <url>")
	fmt.Println()
	flag.PrintDefaults()
}

func getLoadConfig() ([]load.Load, error) {
	var loads []load.Load

	jsonFile, err := os.Open("users.json")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &loads)

	return loads, nil
}

func main() {
	if *scopeConfig == "" {
		usage()
		os.Exit(1)
	}

	loads, err := getLoadConfig()

	gen := &load.Generator{
		ServerURL: flag.Args()[0]
		Loads: loads
	}
	gen.Run()
}
