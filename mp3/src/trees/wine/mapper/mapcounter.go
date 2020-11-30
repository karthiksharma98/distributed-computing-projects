package main

import (
        "strings"
	"encoding/csv"
)

func Map(key string, value string) {
	//s := strings.Split(value, ",")
	//fmt.Println(s[9])
	r := csv.NewReader(strings.NewReader(value))
	s, err := r.Read()
	if err != nil {
		return
	}

	if "Chardonnay" == s[9] {
		for _, word := range strings.Fields(s[2]) {
			if word == "succulent" {
				Emit(s[8], "1")
			}
		}
	}
}
