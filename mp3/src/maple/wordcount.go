package main

import (
	"strings"
)

func (m *Mapler) Maple(input string) error {
	// input is a line of text

	f := func(c rune) bool {
		return c == ','
	}

	words := strings.FieldsFunc(input, f)
	for _, w := range words {
		m.Emit(w, "1")
	}
	return nil
}
