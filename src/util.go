package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

// Config struct
type Config struct {
	Service  Service  `json:"service"`
	Settings Settings `json:"settings"`
}

// Service struct
type Service struct {
	detectorType string  `json:"failure_detector"`
	introducerIP string  `json:"introducer_ip"`
	port         float64 `json:"port"`
}

// Settings struct
type Settings struct {
	gossipInterval float64 `json:"gossip_interval"`
	allInterval    float64 `json:"all_interval"`
	failTimeout    float64 `json:"fail_timeout"`
	cleanupTimeout float64 `json:"cleanup_timeout"`
}

var (
	Info *log.Logger
	Warn *log.Logger
	Err  *log.Logger
)

func Log(infoOut io.Writer, warnOut io.Writer, errOut io.Writer) {
	Info = log.New(
		infoOut,
		"[INFO] ",
		log.Ldate|log.Ltime,
	)

	Warn = log.New(
		warnOut,
		"[WARN] ",
		log.Ldate|log.Ltime|log.Lshortfile,
	)

	Err = log.New(
		errOut,
		"[ERROR] ",
		log.Ldate|log.Ltime|log.Lshortfile,
	)
}

// ReadConfig function to read the configuration JSON
func ReadConfig() Config {
	jsonFile, err := os.Open("../config.json")

	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var result map[string]interface{}
	json.Unmarshal([]byte(byteValue), &result)

	// Create service struct
	serviceJSON := result["service"].(map[string]interface{})
	detectType := serviceJSON["failure_detector"].(string)
	addr := serviceJSON["introducer_ip"].(string)
	port := serviceJSON["port"].(float64)

	service := Service{
		detectorType: detectType,
		introducerIP: addr,
		port:         port,
	}

	// Create settings struct
	settingsJSON := result["settings"].(map[string]interface{})
	gInterval := settingsJSON["gossip_interval"].(float64)
	aInterval := settingsJSON["all_interval"].(float64)
	fTime := settingsJSON["fail_timeout"].(float64)
	cTime := settingsJSON["cleanup_timeout"].(float64)

	settings := Settings{
		gossipInterval: gInterval,
		allInterval:    aInterval,
		failTimeout:    fTime,
		cleanupTimeout: cTime,
	}

	config := Config{
		Service:  service,
		Settings: settings,
	}
	return config
}

func (c *Config) Print() {
	Info.Println("Detector: " + c.Service.detectorType)
	Info.Println("Introducer: " + c.Service.introducerIP + " on port " + fmt.Sprint(c.Service.port))
	Info.Println("Gossip interval: " + fmt.Sprint(c.Settings.gossipInterval))
	Info.Println("All-to-All interval: " + fmt.Sprint(c.Settings.allInterval))
	Info.Println("Failure timeout: " + fmt.Sprint(c.Settings.failTimeout))
	Info.Println("Cleanup timeout: " + fmt.Sprint(c.Settings.cleanupTimeout))
}