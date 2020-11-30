package main

import (
	"strings"
	"encoding/csv"
)

func (m *Mapler) Maple(input string) error {
	r := csv.NewReader(strings.NewReader(input))
	s, err := r.Read()
	if err != nil {
		return nil
	}
	
	if "Chardonnay" == s[9] {
		for _, word := range strings.Fields(s[2]) {
			if word == "succulent" {
				m.Emit(s[10], "1")
			}
		}
	}
	return nil
}
