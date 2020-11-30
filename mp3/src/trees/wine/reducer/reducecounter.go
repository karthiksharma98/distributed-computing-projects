package main

import (
        "strconv"
)

func Reduce(key string, values []string) {
        count := 0

        for _, v := range values {
                if s, err := strconv.Atoi(v); err == nil {
                        count += s
                }
        }
        Emit(key, strconv.Itoa(count))
}
