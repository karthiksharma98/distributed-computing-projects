package main

import (
	"strings"
)

func Map(key string, value string) {
	v := strings.TrimSpace(value)
	Emit("1", v)
}
