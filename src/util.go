package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// Config struct
type Config struct {
	Service  Service  `json:"service"`
	Settings Settings `json:"settings"`
}

// Service struct
type Service struct {
	detectorType string `json:"failure_detector"`
	introducerIP string `json:"introducer_ip"`
	port         float64    `json:"port"`
}

// Settings struct
type Settings struct {
	interval       float64 `json:"interval"`
	failTimeout    float64 `json:"fail_timeout"`
	cleanupTimeout float64 `json:"cleanup_timeout"`
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
                port: port,
        }

        // Create settings struct
        settingsJSON := result["settings"].(map[string]interface{})
        interval := settingsJSON["interval"].(float64)
        fTime := settingsJSON["fail_timeout"].(float64)
        cTime := settingsJSON["cleanup_timeout"].(float64)

        settings := Settings{
                interval: interval,
                failTimeout: fTime,
                cleanupTimeout: cTime,
        }

        config := Config{
                Service: service,
                Settings: settings,
        }

	fmt.Println(result["service"])

	return config
}
