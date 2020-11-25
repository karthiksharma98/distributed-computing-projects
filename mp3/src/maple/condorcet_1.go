package main

import (
	"fmt"
	"strings"
)

func (m *Mapler) Maple(input string) error {
	preferenceList := strings.Split(input, ",")
	for i := 0; i < len(preferenceList)-1; i++ {
		for j := i + 1; j < len(preferenceList); j++ {
			if preferenceList[i] < preferenceList[j] {
				m.Emit("("+fmt.Sprint(preferenceList[i])+" "+fmt.Sprint(preferenceList[j])+")", "1")
			} else {
				m.Emit("("+fmt.Sprint(preferenceList[j])+" "+fmt.Sprint(preferenceList[1])+")", "0")
			}
		}
	}
	return nil
}
