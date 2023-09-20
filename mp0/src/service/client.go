package service

import (
	"ece428/src/config"
	"fmt"
	"log"
	"net/rpc"
	"sync"
)

// queryServer launches an RPC request and returns any received data into a shared channel
func queryServer(regex, address, port string, output chan<- string) {
	log.Println("Querying", address)
	client, err := rpc.DialHTTP("tcp", address+":"+port)
	if err != nil {
		log.Println("RPC.Dialing:", err)
		return
	}
	request := Request{regex}
	var response Response

	//Send the message to the servers to call logly
	err = client.Call("Logly.Retrieve", request, &response)
	if err != nil {
		log.Println("Error - Retrieve:", err)
	} else {
		for _, match := range response.Matches {
			//Print our output
			output <- fmt.Sprintf("FileName: %v | LineNo: %v | Text: %v", match.FileName, match.LineNumber, match.MatchedContent)
		}
	}
}

// Client performs a distributed regex search on all connected machines and prints the results
// to stdout
func Client(regex string) {
	log.Println("Performing a regex search on:", regex)

	port, err := config.Port()
	if err != nil {
		log.Fatalln("Failed to read port:", err)
	}

	addresses, err := config.IPAddress()
	if err != nil {
		log.Fatalln("Failed to read IP Addresses:", err)
	}

	// The following code establishes a shared channel to send back results on as well as
	// a wait group to coordinate when to close the shared channel. Once all goroutines have
	// finished, the channel is closed, telling Client() that it can stop waiting on strings
	// to output.

	aggregate := make(chan string)
	var wg sync.WaitGroup

	for _, address := range addresses {
		wg.Add(1)
		go func(regex, address, port string, aggregate chan<- string) {
			queryServer(regex, address, port, aggregate)
			wg.Done()
		}(regex, address, port, aggregate)
	}

	go func() {
		wg.Wait()
		close(aggregate)
	}()

	counter := 0
	for msg := range aggregate {
		fmt.Println(msg)
		counter++
	}

	fmt.Printf("Received %d Messages.\n", counter)
}
