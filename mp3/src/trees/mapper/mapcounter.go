package main

import (
        "strings"
)

func Map(key string, value string) {
        for _, word := range strings.Fields(value) {
                Emit(word, "1")
        }
}
