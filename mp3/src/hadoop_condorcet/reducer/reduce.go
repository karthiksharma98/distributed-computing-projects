package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var (
	keyValues map[string][]string = make(map[string][]string)
)

func CollectReduce() {
	// MapReduce feeds data line by line
	s := bufio.NewScanner(os.Stdin)
	f := func(c rune) bool {
		return c == '\t'
	}

	for s.Scan() {
		// Split strings by tab
		line := strings.FieldsFunc(s.Text(), f)
		if len(line) == 2 {
			if val, ok := keyValues[line[0]]; ok {
				keyValues[line[0]] = append(val, line[1])
				continue
			}
			keyValues[line[0]] = []string{line[1]}
		}
	}

	if err := s.Err(); err != nil {
		fmt.Println(err)
	}
}

func RunReduce() {
	CollectReduce()
	for key, values := range keyValues {
		Reduce(key, values)
	}
}

func Emit(key string, value string) {
	fmt.Printf("%s\t%s\n", key, value)
}

func main() {
	RunReduce()
}
