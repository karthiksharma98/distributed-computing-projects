package main

import (
        "strings"
	"encoding/csv"
)

func Map(key string, value string) {
	r := csv.NewReader(strings.NewReader(input))
	s, err := r.Read()
	if err != nil {
		return 
	}
	
	if "Chardonnay" == s[9] {
		for _, word := range strings.Fields(s[2]) {
			if word == "succulent" {
				m.Emit(s[10], "1")
			}
		}
	}
}
