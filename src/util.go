package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
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

// MatchRes is the structure containing all pertinent information found
// while searching for a pattern in a log file.
// From MP0 best solution.
type MatchRes struct {
	MemberID       uint8
	LineNumber     int
	FileName       string
	MatchedContent string
}

var (
	Info *log.Logger
	Warn *log.Logger
	Err  *log.Logger
)

func InitLog() {
	file, err := os.OpenFile("machine.log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(file)

	Info = log.New(
		file,
		"[INFO] ",
		log.Ldate|log.Ltime,
	)

	Warn = log.New(
		file,
		"[WARN] ",
		log.Ldate|log.Ltime|log.Lshortfile,
	)

	Err = log.New(
		file,
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

// Finder is the function that searches file at fileLoc, for pattern, using Go's
// Regex engine.
// From MP0 best solution.
func Finder(pattern string, fileLoc string, memberID uint8) []MatchRes {
	retArr := make([]MatchRes, 0)
	//Create Regex from pattern
	regex, err := regexp.Compile(pattern)
	if err != nil {
		//Invalid Regex Pattern
		return make([]MatchRes, 0)
	}

	file, err := ioutil.ReadFile(fileLoc)
	if err != nil {
		//Could not open file at fileLoc
		return make([]MatchRes, 0)
	}

	fileString := string(file)

	// Go through and find all lines that match pattern
	for lineIndex, line := range strings.Split(fileString, "\n") {
		if regex.MatchString(line) {
			newMatch := MatchRes{memberID, lineIndex, fileLoc, line}
			retArr = append(retArr, newMatch)
		}
	}
	return retArr
}

func (c *Config) Print() {
	Info.Println("Detector: " + c.Service.detectorType)
	Info.Println("Introducer: " + c.Service.introducerIP + " on port " + fmt.Sprint(c.Service.port))
	Info.Println("Gossip interval: " + fmt.Sprint(c.Settings.gossipInterval))
	Info.Println("All-to-All interval: " + fmt.Sprint(c.Settings.allInterval))
	Info.Println("Failure timeout: " + fmt.Sprint(c.Settings.failTimeout))
	Info.Println("Cleanup timeout: " + fmt.Sprint(c.Settings.cleanupTimeout))
}
