package finder

import (
	"io/ioutil"
	"regexp"
	"strings"
)

// MatchRes is the structure containing all pertinent information found
// while searching for a pattern in a log file.
type MatchRes struct {
	LineNumber     int
	FileName       string
	MatchedContent string
}

// Finder is the function that searches file at fileLoc, for pattern, using Go's
// Regex engine.
func Finder(pattern string, fileLoc string) ([]MatchRes, error) {
	retArr := make([]MatchRes, 0)
	//Create Regex from pattern
	regex, err := regexp.Compile(pattern)
	if err != nil {
		//Invalid Regex Pattern
		return make([]MatchRes, 0), err
	}

	file, err := ioutil.ReadFile(fileLoc)
	if err != nil {
		//Could not open file at fileLoc
		return make([]MatchRes, 0), err
	}

	fileString := string(file)

	// Go through and find all lines that match pattern
	for lineIndex, line := range strings.Split(fileString, "\n") {
		if regex.MatchString(line) {
			newMatch := MatchRes{lineIndex, fileLoc, line}
			retArr = append(retArr, newMatch)
		}
	}
	return retArr, nil
}
