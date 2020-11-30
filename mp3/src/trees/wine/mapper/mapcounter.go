package main

import (
        "strings"
	"fmt"
)

func Map(key string, value string) {
	s := strings.Split(value, ",")
	fmt.Println(s[9])
	if "Chardonnay" == strings.TrimSpace(s[9]) {
		fmt.Println(s)
		for _, word := range strings.Fields(s[2]) {
			if word == "succulent" {
				Emit(s[8], "1")
			}
		}
	}
	
}
