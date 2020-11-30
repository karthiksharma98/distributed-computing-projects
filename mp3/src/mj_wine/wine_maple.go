package main

import (
	"strings"
	"encoding/csv"
)

func (m *Mapler) Maple(input string) error {
	r := csv.NewReader(strings.NewReader(input))
	s, _ := r.Read()
	
	if "Chardonnay" == s[9] {
		for _, word := range strings.Fields(s[2]) {
			if word == "succulent" {
				m.Emit(s[8], "1")
			}
		}
	}
	return nil
}
