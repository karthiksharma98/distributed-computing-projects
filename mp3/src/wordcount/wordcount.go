package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

type wordCount struct {
}

func (wc wordCount) Maple(inputFilePath string) error {
	file, err := os.Open(inputFilePath)
	if err != nil {
		log.Fatal(err)
		log.Println("Error opening file")
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		fmt.Println("1," + scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
		log.Println("Error scanning tokens")
	}

	return err
}

func main() {
	if len(os.Args) < 2 {
		log.Println("Insufficient arguments to maple")
		return
	}
	var wc wordCount
	wc.Maple(os.Args[1])
}
