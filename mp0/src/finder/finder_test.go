package finder

import (
	"testing"
)

func TestRandLog(t *testing.T) {
	var filename = "machine.test.log"
	var pattern = ".*GET.*"

	matches, err := Finder(pattern, filename)

	if err != nil {
		panic(err)
	}

	for _, match := range matches {
		t.Log(match)
	}
}
