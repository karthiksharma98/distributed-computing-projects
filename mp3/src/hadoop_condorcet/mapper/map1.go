package main

import (
	"fmt"
	"strings"
)

func Map(key string, value string) {
	f := func(c rune) bool {
		return c == ','
	}

	preferenceList := strings.FieldsFunc(value, f)
	for i := 0; i < len(preferenceList)-1; i++ {
		for j := i + 1; j < len(preferenceList); j++ {
			if preferenceList[i] < preferenceList[j] {
				Emit("("+fmt.Sprint(preferenceList[i])+" "+fmt.Sprint(preferenceList[j])+")", "1")
			} else {
				Emit("("+fmt.Sprint(preferenceList[j])+" "+fmt.Sprint(preferenceList[i])+")", "0")
			}
		}
	}
}
