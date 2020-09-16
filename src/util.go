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
	port         int    `json:"port"`
}

// Settings struct
type Settings struct {
	interval       int `json:"interval"`
	failTimeout    int `json:"fail_timeout"`
	cleanupTimeout int `json:"cleanup_timeout"`
}

// ReadConfig function to read the configuration JSON
func ReadConfig() map[string]interface{} {
	jsonFile, err := os.Open("../config.json")

	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var result map[string]interface{}
	json.Unmarshal([]byte(byteValue), &result)

	fmt.Println(result["service"])

	return result
}
